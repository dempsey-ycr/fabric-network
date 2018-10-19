#!/bin/bash

# getGlobals
. scripts/globals.sh


# cc_name cc_version
setChaincodeEnv $1 $2
LANGUAGE="golang"
CHANNEL_NAME="mychannel"
TIMEOUT=10
DELAY=3
readonly PACK_NAME="chaincode-pack.out"
readonly SIGN_PACK_NAME="sign-$CC_NAME-$CC_VERSION.out"
readonly CC_PATH="fabric-network/chaincode/go/dempsey_ycr/go"

# packaging chaincode
packageChaincode(){
  set -x
  peer chaincode package $PACK_NAME -n $CC_NAME -v $CC_VERSION -s -S -p ${CC_PATH} >&log.txt
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode packaging has failed"
  echo "===================== Chaincode is packaged successful ===================== "
  echo
}

# signing chaincode
signChaincode(){
  set -x
  peer chaincode signpackage $PACK_NAME $SIGN_PACK_NAME >&log.txt
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode signing has failed"
  echo "===================== Chaincode is signed successful ===================== "
  rm $PACK_NAME
  echo

}

# install chaincode
installChaincode() {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG
  set -x
  # peer chaincode install -n app -v ${VERSION} -l ${LANGUAGE} -p ${CC_IHT_PATH} >&log.txt
  peer chaincode install $SIGN_PACK_NAME >&log.txt
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode installation on peer${PEER}.org${ORG} has failed"
  echo "===================== Chaincode is installed on peer${PEER}.org${ORG} ===================== "
  echo
}

# instantiate Chaincode 
instantiateChaincode() {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG

  # while 'peer chaincode' command can get the orderer endpoint from the peer
  # (if join was successful), let's supply it directly as we know it using
  # the "-o" option
  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
    set -x
    peer chaincode instantiate -o orderer.example.com:7050 -C $CHANNEL_NAME -n $CC_NAME -l ${LANGUAGE} -v $CC_VERSION -c '{"Args":["init","a","100","b","200"]}' -P "AND ('Org1MSP.peer','Org2MSP.peer')" >&log.txt
    res=$?
    set +x
  else
    set -x
    peer chaincode instantiate -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME -l ${LANGUAGE} -v $CC_VERSION -c '{"Args":["init","a","100","b","200"]}' -P "AND ('Org1MSP.peer','Org2MSP.peer')" >&log.txt
    res=$?
    set +x
  fi
  cat log.txt
  verifyResult $res "Chaincode instantiation on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' failed"
  echo "===================== Chaincode is instantiated on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
  echo
}

# parsePeerConnectionParameters $@
# Helper function that takes the parameters from a chaincode operation
# (e.g. invoke, query, instantiate) and checks for an even number of
# peers and associated org, then sets $PEER_CONN_PARMS and $PEERS
parsePeerConnectionParameters() {
  # check for uneven number of peer and org parameters
  if [ $(($# % 2)) -ne 0 ]; then
    exit 1
  fi

  PEER_CONN_PARMS=""
  PEERS=""
  while [ "$#" -gt 0 ]; do
    PEER="peer$1.org$2"
    PEERS="$PEERS $PEER"
    PEER_CONN_PARMS="$PEER_CONN_PARMS --peerAddresses $PEER.example.com:7051"
    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "true" ]; then
      TLSINFO=$(eval echo "--tlsRootCertFiles \$PEER$1_ORG$2_CA")
      PEER_CONN_PARMS="$PEER_CONN_PARMS $TLSINFO"
    fi
    # shift by two to get the next pair of peer/org parameters
    shift
    shift
  done
  # remove leading space for output
  PEERS="$(echo -e "$PEERS" | sed -e 's/^[[:space:]]*//')"
}

# chaincodeInvoke <peer> <org> ...
# Accepts as many peer/org pairs as desired and requests endorsement from each
chaincodeInvoke() {
  parsePeerConnectionParameters $@
  res=$?
  verifyResult $res "Invoke transaction failed on channel '$CHANNEL_NAME' due to uneven number of peer and org parameters "

  # while 'peer chaincode' command can get the orderer endpoint from the
  # peer (if join was successful), let's supply it directly as we know
  # it using the "-o" option
  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
    set -x
    peer chaincode invoke -o orderer.example.com:7050 -C $CHANNEL_NAME -n $CC_NAME $PEER_CONN_PARMS -c '{"Args":["invoke","a","b","10"]}' >&log.txt
    res=$?
    set +x
  else
    set -x
    peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME $PEER_CONN_PARMS -c '{"Args":["invoke","a","b","10"]}' >&log.txt res=$?
    set +x
  fi
  cat log.txt
  verifyResult $res "Invoke execution on $PEERS failed "
  echo "===================== Invoke transaction successful on $PEERS on channel '$CHANNEL_NAME' ===================== "
  echo
}

chaincodeQuery() {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG
  EXPECTED_RESULT=$3
  echo "===================== Querying on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME'... ===================== "
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while
    test "$(($(date +%s) - starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
    sleep $DELAY
    echo "Attempting to Query peer${PEER}.org${ORG} ...$(($(date +%s) - starttime)) secs"
    set -x
    peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"Args":["query","a"]}' >&log.txt
    res=$?
    set +x
    test $res -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
    test "$VALUE" = "$EXPECTED_RESULT" && let rc=0
    # removed the string "Query Result" from peer chaincode query command
    # result. as a result, have to support both options until the change
    # is merged.
    test $rc -ne 0 && VALUE=$(cat log.txt | egrep '^[0-9]+$')
    test "$VALUE" = "$EXPECTED_RESULT" && let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0; then
    echo "===================== Query successful on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
  else
    echo "!!!!!!!!!!!!!!! Query result on peer${PEER}.org${ORG} is INVALID !!!!!!!!!!!!!!!!"
    echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
    echo
    exit 1
  fi
}

if [ $# -gt 2 ]; then
  echo "==================current is : $# ====================="
else
  # package
  echo "Package chaincode..."
  packageChaincode

  # sign
  echo "Signature chaincode..."
  signChaincode

  # install
  echo "Install chaincode on top three peer..."
  installChaincode 0 1
  installChaincode 1 1
  installChaincode 0 2

  # instantiate
  echo "Instantiate chaincode on peer0 org1"
  instantiateChaincode 0 1

  # querying
  echo "Querying chaincode on peer0.org1..."
  chaincodeQuery 0 1 100

  # Invoke chaincode on peer0.org1 and peer0.org2
  echo "Sending invoke transaction on peer0.org1 peer0.org2..."
  chaincodeInvoke 0 1 0 2


  # Install chaincode on peer1 org2
  installChaincode 1 2

  # Query on chaincode on peer1.org2, check if the result is 90
  echo "Querying chaincode on peer1.org2..."
  chaincodeQuery 1 2 90

  echo
  echo "========= All GOOD, Install chaincode and test completed =========== "
  echo

  echo
  echo " _____   _   _   ____   "
  echo "| ____| | \ | | |  _ \  "
  echo "|  _|   |  \| | | | | | "
  echo "| |___  | |\  | | |_| | "
  echo "|_____| |_| \_| |____/  "
  echo

  exit 0

fi