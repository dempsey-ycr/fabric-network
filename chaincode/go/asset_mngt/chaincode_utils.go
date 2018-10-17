package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	"time"
)

func queryDocsByIdxkey(stub shim.ChaincodeStubInterface, docType string, idxKey string, idxKeyvalue string) ([]byte, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", docType, idxKey, idxKeyvalue)

	return getQueryResultForQueryString(stub, queryString)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
func getHistoryForDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string) ([]byte, error) {
	if len(docKey) < 1 {
		return nil, errors.New("docKey should not be empty")
	}

	fmt.Printf("- start getHistoryForDocWithNamespace: %s\n", docKey)

	resultsIterator, err := stub.GetHistoryForKey(ns + docKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the asset_securitized_mngt
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON asset_securitized_mngt)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForDocWithNamespace returning:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func getHistoryForDoc(stub shim.ChaincodeStubInterface, docKey string) ([]byte, error) {
	return getHistoryForDocWithNamespace(stub, "", docKey)
}

func getDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string) ([]byte, error) {
	valAsbytes, err := stub.GetState(ns + docKey)
	return valAsbytes, err
}

func putDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string, bytes []byte) error {
	err := stub.PutState(ns+docKey, bytes)
	return err
}

func getDoc(stub shim.ChaincodeStubInterface, docKey string) ([]byte, error) {
	return getDocWithNamespace(stub, "", docKey)
}

func putDoc(stub shim.ChaincodeStubInterface, docKey string, bytes []byte) error {
	return putDocWithNamespace(stub, "", docKey, bytes)
}

func createCKeyWithNamespace(stub shim.ChaincodeStubInterface, ns string, idxName string, idxPair []string) error {
	compositeKey, err := stub.CreateCompositeKey(ns+idxName, idxPair)
	if err != nil {
		return err
	}
	value := []byte{0x00}
	err = putDoc(stub, compositeKey, value)
	if err != nil {
		return err
	}

	return nil
}

func createCKey(stub shim.ChaincodeStubInterface, idxName string, idxPair []string) error {
	return createCKeyWithNamespace(stub, "", idxName, idxPair)
}
