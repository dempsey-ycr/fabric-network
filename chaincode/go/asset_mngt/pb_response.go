package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type PbResponse struct {
	Success bool
	Data    interface{}
	Err     string
}

func successPbResponseString(data string) pb.Response {
	response := PbResponse{
		true,
		data,
		"",
	}
	bytes, err := StructToJSONBytes(response)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(bytes)
	//return shim.Success(data)
}

func successPbResponse(data []byte) pb.Response {
	//	response := PbResponse{
	//		true,
	//		data,
	//		"",
	//	}
	//	bytes, err := StructToJSONBytes(response)
	//	if err != nil {
	//		return shim.Error(err.Error())
	//	}
	//	return shim.Success(bytes)
	return shim.Success(data)
}

func errorPbResponse(err error, isAppError bool) pb.Response {
	if err != nil && isAppError {
		response := PbResponse{
			false,
			nil,
			err.Error(),
		}
		bytes, errParse := StructToJSONBytes(response)
		if errParse != nil {
			fmt.Println("Failed to transform to json bytes for app error: " + errParse.Error() + "   original error is " + err.Error())
			return shim.Error("Failed to transform to json bytes for app error, the error is: " + err.Error() + ", returned as pb.Response system error")
		}
		return shim.Success(bytes)
	}

	return shim.Error(err.Error())
}
