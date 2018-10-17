/*
Licensed
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke AssetOwners ====
// peer chaincode invoke -C myc1 -n AssetOwners -c '{"Args":["initAssetOwner","AssetOwner1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n AssetOwners -c '{"Args":["initAssetOwner","AssetOwner2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n AssetOwners -c '{"Args":["initAssetOwner","AssetOwner3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n AssetOwners -c '{"Args":["transferAssetOwner","AssetOwner2","jerry"]}'
// peer chaincode invoke -C myc1 -n AssetOwners -c '{"Args":["transferAssetOwnersBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n AssetOwners -c '{"Args":["delete","AssetOwner1"]}'

// ==== Query AssetOwners ====
// peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["readAssetOwner","AssetOwner1"]}'
// peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["getAssetOwnersByRange","AssetOwner1","AssetOwner3"]}'
// peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["getHistoryForAssetOwner","AssetOwner1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["queryAssetOwnersByOwner","tom"]}'
//   peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["queryAssetOwners","{\"selector\":{\"owner\":\"tom\"}}"]}'

//The following examples demonstrate creating indexes on CouchDB
//Example hostname:port configurations
//
//Docker or vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for chaincodeid, docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":["chaincodeid","data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"chaincodeid\",\"data.docType\",\"data.owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1/_index
//

// Index for chaincodeid, docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"chaincodeid":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"data.size\":\"desc\"},{\"chaincodeid\":\"desc\"},{\"data.docType\":\"desc\"},{\"data.owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["queryAssetOwners","{\"selector\":{\"docType\":\"asset_owner_mngt\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n AssetOwners -c '{"Args":["queryAssetOwners","{\"selector\":{\"docType\":{\"$eq\":\"asset_owner_mngt\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

const (
	ST_comm_init      string = "00" // init
	ST_comm_approving string = "01" // 01-正在审核
	ST_comm_approved  string = "02" // 02-审核通过
	ST_comm_rejected  string = "03" // 03-审核不通过
	ST_comm_niled     string = "99" // 99-作废
)

const (
	ST_DESC_comm_init        string = "init"             // init
	ST_DESC_comm_init_change string = "init with change" // 00-init with change
	ST_DESC_comm_approving   string = "IHT approving"    // 01-正在审核
	ST_DESC_comm_approved    string = "IHT approved"     // 02-审核通过
	ST_DESC_comm_rejected    string = "IHT rejected"     // 03-审核不通过
	ST_DESC_comm_niled       string = "IHT deleted"      // 99-作废
)

// AssetMngtChaincode example simple Chaincode implementation
type AssetMngtChaincode struct {
	assetOwnerMngt         *AssetOwnerMngt
	fininstMngt            *FininstMngt
	assetSecuritizedMngt   *AssetSecuritizedMngt
	assetSecurityShareMngt *AssetSecurityShareMngt
	paymentTxnBookingMngt  *PaymentTxnBookingMngt
	myTest                 *MyTest
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(AssetMngtChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *AssetMngtChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("asset_owner_mngt chaincode Is Starting Up")
	_, args := stub.GetFunctionAndParameters()
	var Aval int
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// convert numeric string to integer
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting a numeric string argument to Init()")
	}

	// store compaitible asset_owner_mngt application version
	err = stub.PutState("asset_mngt_ui", []byte("1.0"))
	if err != nil {
		return shim.Error(err.Error())
	}

	// this is a very simple dumb test.  let's write to the ledger and error on any errors
	err = stub.PutState("selftest", []byte(strconv.Itoa(Aval))) //making a test var "selftest", its handy to read this right away to test the network
	if err != nil {
		return shim.Error(err.Error()) //self-test fail
	}

	// init modules
	t.assetOwnerMngt = new(AssetOwnerMngt)
	t.fininstMngt = new(FininstMngt)
	t.assetSecuritizedMngt = new(AssetSecuritizedMngt)
	t.assetSecurityShareMngt = new(AssetSecurityShareMngt)
	t.paymentTxnBookingMngt = new(PaymentTxnBookingMngt)

	fmt.Println(" - ready for action") //self-test pass
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *AssetMngtChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "read" { //selftest
		return t.read(stub, args)
	} else if function == "initAssetOwner" { //create a new asset_owner_mngt
		return t.assetOwnerMngt.initAssetOwner(stub, args)
	} else if function == "initAssetOwnerDetail" { //create a new asset_owner_mngt
		return t.assetOwnerMngt.initAssetOwnerDetail(stub, args)
	} else if function == "readAssetOwner" { //read a asset_owner
		return t.assetOwnerMngt.readAssetOwner(stub, args)
	} else if function == "readAssetOwnerDetail" { //read a asset_owner_detail
		return t.assetOwnerMngt.readAssetOwnerDetail(stub, args)
	} else if function == "approveAssetOwner" { //approve a asset_owner
		return t.assetOwnerMngt.approveAssetOwner(stub, args)
	} else if function == "approveAssetOwnerDetail" { //approve a asset_owner_detail
		return t.assetOwnerMngt.approveAssetOwnerDetail(stub, args)
	} else if function == "rejectAssetOwner" { //reject a asset_owner
		return t.assetOwnerMngt.rejectAssetOwner(stub, args)
	} else if function == "rejectAssetOwnerDetail" { //reject a asset_owner_detail
		return t.assetOwnerMngt.rejectAssetOwnerDetail(stub, args)
	} else if function == "deleteAssetOwner" { //delete asset_owner
		return t.assetOwnerMngt.deleteAssetOwner(stub, args)
	} else if function == "queryAssetOwnersByStatus" { //query asset_owners by status
		return t.assetOwnerMngt.queryAssetOwnersByStatus(stub, args)
	} else if function == "queryAssetOwnerDetailsByStatus" { //query asset_owner_details by status
		return t.assetOwnerMngt.queryAssetOwnerDetailsByStatus(stub, args)
	} else if function == "getHistoryForAssetOwner" { //get history of values for a asset_owner
		return t.assetOwnerMngt.getHistoryForAssetOwner(stub, args)
	} else if function == "getHistoryForAssetOwnerDetail" { //get history of values for a asset_owner_detail
		return t.assetOwnerMngt.getHistoryForAssetOwnerDetail(stub, args)
	} else if function == "changeAssetOwner" { //changeAssetOwner
		return t.assetOwnerMngt.changeAssetOwner(stub, args)
	} else if function == "changeAssetOwnerDetail" { //changeAssetOwnerDetail
		return t.assetOwnerMngt.changeAssetOwnerDetail(stub, args)
	} else if function == "initAssetSecuritized" { //init asset securitized
		return t.assetSecuritizedMngt.initAssetSecuritized(stub, args)
	} else if function == "readAssetSecuritized" { //init asset securitized
		return t.assetSecuritizedMngt.readAssetSecuritized(stub, args)
	} else if function == "queryAssetSecuritizedsByStatus" { //init asset securitized
		return t.assetSecuritizedMngt.queryAssetSecuritizedsByStatus(stub, args)
	} else if function == "approveAssetSecuritized" { //init asset securitized
		return t.assetSecuritizedMngt.approveAssetSecuritized(stub, args)
	} else if function == "rejectAssetSecuritized" { //init asset securitized
		return t.assetSecuritizedMngt.rejectAssetSecuritized(stub, args)
	} else if function == "deleteAssetSecuritized" { //init asset securitized
		return t.assetSecuritizedMngt.deleteAssetSecuritized(stub, args)
	} else if function == "getHistoryForAssetSecuritized" { //init asset securitized
		return t.assetSecuritizedMngt.getHistoryForAssetSecuritized(stub, args)
	} else if function == "changeAssetSecuritized" { //changeAssetSecuritized
		return t.assetSecuritizedMngt.changeAssetSecuritized(stub, args)
	} else if function == "initFininst" { //create a new asset_owner_mngt
		return t.fininstMngt.initFininst(stub, args)
	} else if function == "initFininstDetail" { //create a new asset_owner_mngt
		return t.fininstMngt.initFininstDetail(stub, args)
	} else if function == "readFininst" { //read a asset_owner
		return t.fininstMngt.readFininst(stub, args)
	} else if function == "readFininstDetail" { //read a asset_owner_detail
		return t.fininstMngt.readFininstDetail(stub, args)
	} else if function == "approveFininst" { //approve a asset_owner
		return t.fininstMngt.approveFininst(stub, args)
	} else if function == "approveFininstDetail" { //approve a asset_owner_detail
		return t.fininstMngt.approveFininstDetail(stub, args)
	} else if function == "rejectFininst" { //reject a asset_owner
		return t.fininstMngt.rejectFininst(stub, args)
	} else if function == "rejectFininstDetail" { //reject a asset_owner_detail
		return t.fininstMngt.rejectFininstDetail(stub, args)
	} else if function == "deleteFininst" { //delete asset_owner
		return t.fininstMngt.deleteFininst(stub, args)
	} else if function == "queryFininstsByStatus" { //query asset_owners by status
		return t.fininstMngt.queryFininstsByStatus(stub, args)
	} else if function == "queryFininstDetailsByStatus" { //query asset_owner_details by status
		return t.fininstMngt.queryFininstDetailsByStatus(stub, args)
	} else if function == "getHistoryForFininst" { //get history of values for a asset_owner
		return t.fininstMngt.getHistoryForFininst(stub, args)
	} else if function == "getHistoryForFininstDetail" { //get history of values for a asset_owner_detail
		return t.fininstMngt.getHistoryForFininstDetail(stub, args)
	} else if function == "changeFininst" { //changeFininst
		return t.fininstMngt.changeFininst(stub, args)
	} else if function == "changeFininstDetail" { //changeFininstDetail
		return t.fininstMngt.changeFininstDetail(stub, args)
	} else if function == "startUnderwrite" { //start underwriting
		return t.assetSecuritizedMngt.startUnderwrite(stub, args)
	} else if function == "haltUnderwrite" { //haltUnderwrite
		return t.assetSecuritizedMngt.haltUnderwrite(stub, args)
	} else if function == "resumeUnderwrite" { //resumeUnderwrite
		return t.assetSecuritizedMngt.resumeUnderwrite(stub, args)
	} else if function == "stopUnderwriteAndCalculateResults" { //stopUnderwriteAndCalculateResults
		return t.assetSecuritizedMngt.stopUnderwriteAndCalculateResults(stub, args)
	} else if function == "underwrite" { //underwrite
		return t.assetSecuritizedMngt.underwrite(stub, args)
	} else if function == "readUnderwrote" { //readUnderwrote
		return t.assetSecuritizedMngt.readUnderwrote(stub, args)
	} else if function == "repurchase" { //repurchase
		return t.assetSecuritizedMngt.repurchase(stub, args)
	} else if function == "queryUnderwrotesByFininstEmail" { //queryUnderwrotesByFininstEmail
		return t.assetSecuritizedMngt.queryUnderwrotesByFininstEmail(stub, args)
	} else if function == "queryUnderwrotesByAssetOwnerEmail" { //queryUnderwrotesByAssetOwnerEmail
		return t.assetSecuritizedMngt.queryUnderwrotesByAssetOwnerEmail(stub, args)
	} else if function == "queryUnderwrotesByAssetName" { //queryUnderwrotesByAssetName
		return t.assetSecuritizedMngt.queryUnderwrotesByAssetName(stub, args)
	} else if function == "initAssetSecurityShares" { //initAssetSecurityShares
		return t.assetSecurityShareMngt.initAssetSecurityShares(stub, args)
	} else if function == "readAssetSecurityShare" { //readAssetSecurityShare
		return t.assetSecurityShareMngt.readAssetSecurityShare(stub, args)
	} else if function == "enableUnderwroteAssetSharesOwnerships" { //enableUnderwroteAssetSharesOwnerships
		return t.assetSecurityShareMngt.enableUnderwroteAssetSharesOwnerships(stub, args)
	} else if function == "queryAssetSecuritySharesByAssetName" { //queryAssetSecuritySharesByAssetName
		return t.assetSecurityShareMngt.queryAssetSecuritySharesByAssetName(stub, args)
	} else if function == "queryAssetSecuritySharesByOwnerEmail" { //queryAssetSecuritySharesByOwnerEmail
		return t.assetSecurityShareMngt.queryAssetSecuritySharesByOwnerEmail(stub, args)
	} else if function == "queryAssetSecuritySharesByAssetHash" { //queryAssetSecuritySharesByAssetHash
		return t.assetSecurityShareMngt.queryAssetSecuritySharesByAssetHash(stub, args)
	} else if function == "initPaymentTxnBooking" { //initPaymentTxnBooking
		return t.paymentTxnBookingMngt.initPaymentTxnBooking(stub, args)
	} else if function == "readPaymentTxnBooking" { //readPaymentTxnBooking
		return t.paymentTxnBookingMngt.readPaymentTxnBooking(stub, args)
	} else if function == "queryBookedPaymentTxnsByAssetHash" { //queryBookedPaymentTxnsByAssetHash
		return t.paymentTxnBookingMngt.queryBookedPaymentTxnsByAssetHash(stub, args)
	} else if function == "queryBookedPaymentTxnsByShareId" { //queryBookedPaymentTxnsByShareId
		return t.paymentTxnBookingMngt.queryBookedPaymentTxnsByShareId(stub, args)
	} else if function == "testInsert" { //queryBookedPaymentTxnsByShareId
		return t.myTest.Insert(stub, args)
	} else if function == "testQuery" { //queryBookedPaymentTxnsByShareId
		return t.myTest.Query(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================================================================================
// Read - read a generic variable from ledger
//
// Shows Off GetState() - reading a key/value from the ledger
//
// Inputs - Array of strings
//  0
//  key
//  "abc"
//
// Returns - string
// ============================================================================================================================
func (t *AssetMngtChaincode) read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	var err error
	fmt.Println("starting read")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key) //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("- end read")
	return shim.Success(valAsbytes) //send it onward
}
