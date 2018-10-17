package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type MyTest struct{}

func (t *MyTest) Insert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return successPbResponseString("test insert success 22222")
}
func (t *MyTest) Query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	return successPbResponseString("test query success 22222")
}
