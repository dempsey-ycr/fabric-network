#!/bin/bash

#. scripts/globals.sh
. scripts/installCC.sh $1 $2 "out"

# Instantiate on peer0Org1 by default
upgradeChaincode() {
  PEER=0
  ORG=1
  setGlobals $PEER $ORG

  set -x
  peer chaincode upgrade -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME -v $CC_VERSION -c '{"Args":["init","a","100","b","200"]}' -P "AND ('Org1MSP.peer','Org2MSP.peer')" res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode upgrade on peer${PEER}.org${ORG} has failed"
  echo "===================== Chaincode is upgraded on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
  echo 
 }

 # package
echo "Package chaincode..."
packageChaincode

# sign
echo "Signature chaincode..."
signChaincode


## Install chaincode on peer0.org1 and peer0.org2
echo "Installing chaincode on peer0.org1..."
installChaincode 0 1

echo "Installling chaincode on peer0.org2..."
installChaincode 0 2

# upgrade chaincode on peer0.org2
echo "upgrade chaincode on peer0.org2..."
upgradeChaincode 0 2

# Query chaincode on peer0.org1
echo "Querying chaincode on peer0.org1..."
chaincodeQuery 0 1 100

# Invoke chaincode on peer0.org1 and peer0.org2
echo "Sending invoke transaction on peer0.org1 peer0.org2..."
chaincodeInvoke 0 1 0 2

## Install chaincode on peer1.org2
echo "Installing chaincode on peer1.org1 and peer1.org2..."
installChaincode 1 1
installChaincode 1 2


# Query on chaincode on peer1.org2, check if the result is 90
echo "Querying chaincode on peer1.org2..."
chaincodeQuery 1 2 90

# init chaincode
echo "initialize chaincode on peer0.org2..."
initchaincode 0 1 0 2

echo
echo "========= All GOOD, Upgrade chaincode and test completed =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0