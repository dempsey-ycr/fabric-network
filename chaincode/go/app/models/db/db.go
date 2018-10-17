package db

import (
	"encoding/json"
	"iht-fabric-chaincode-private/chaincode/go/app/libraries/logging"

	"iht-fabric-chaincode-private/chaincode/go/app/utils/filter"

	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// 查询世界状态
func QueryStatusByIdx(stub shim.ChaincodeStubInterface, docType string, idxKey string, idxKeyvalue string) ([]byte, error) {
	//queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", docType, idxKey, idxKeyvalue)
	selector := map[string]interface{}{
		"docType": docType,
	}
	selector[idxKey] = idxKeyvalue

	qs := map[string]interface{}{
		"selector": selector,
	}
	qsStr, err := json.Marshal(qs)
	if err != nil {
		logging.Error(err.Error())
	}
	logging.Info(string(qsStr))
	return getWorldState(stub, string(qsStr))
}

func QueryRichList(selector map[string]interface{}) string {
	cond := map[string]interface{}{
		"selector": selector,
		//"limit":    limit,
		//"skip":     skip,
		//"execution_stats": true,
	}

	data, err := json.Marshal(cond)
	if err != nil {
		logging.Error(err.Error())
	}
	logging.Debug("-----selector---  %v", string(data))

	return string(data)
}

func getWorldState(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	iterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var status []*WorldState
	for iterator.HasNext() {
		query, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		state := &WorldState{
			StateRecord: ByteArrayToStructure(query.GetValue()),
			StateKey:    query.GetKey(),
		}
		status = append(status, state)
	}
	list := &List{
		Object: status,
	}

	worldStatus, err := json.Marshal(list)
	if err != nil {
		logging.Error("Unmarshal error| return Statues [%s]", err.Error())
		return nil, err
	}
	return worldStatus, nil
}

// 查询历史记录
func GetHistoryForDocWithNamespace(stub shim.ChaincodeStubInterface, hisKey string) ([]byte, error) {
	if err := filter.CheckParamsNull(hisKey); err != nil {
		return nil, err
	}

	iterator, err := stub.GetHistoryForKey(hisKey)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var hisList []*History
	for iterator.HasNext() {
		record, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		his := &History{
			IsDelete:  record.IsDelete,
			TxId:      record.GetTxId(),
			Value:     ByteArrayToStructure(record.Value),
			Timestamp: time.Unix(record.Timestamp.Seconds, int64(record.Timestamp.Nanos)).String(),
		}
		hisList = append(hisList, his)
	}
	list := &List{
		Object: hisList,
	}

	resultList, err := json.Marshal(list)
	if err != nil {
		logging.Error("Unmarshal error| return Statues [%s]", err.Error())
		return nil, err
	}
	return resultList, nil
}

func PutState(stub shim.ChaincodeStubInterface, key string, _json interface{}) error {
	data, err := json.Marshal(_json)
	if err != nil {
		return err
	}
	return stub.PutState(key, data)
}

func GetState(stub shim.ChaincodeStubInterface, docKey string) ([]byte, error) {
	return stub.GetState(docKey)
}

// --------------------------分界线---------------------------//
func GetDocWithNamespace(stub shim.ChaincodeStubInterface, docKey string) ([]byte, error) {
	get, err := stub.GetState(docKey)
	return get, err
}

func PutDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string, bytes []byte) error {
	err := stub.PutState(ns+docKey, bytes)
	return err
}

func PutDoc(stub shim.ChaincodeStubInterface, docKey string, bytes []byte) error {
	return PutDocWithNamespace(stub, "", docKey, bytes)
}

func CreateCKeyWithNamespace(stub shim.ChaincodeStubInterface, ns string, idxName string, idxPair []string) error {
	compositeKey, err := stub.CreateCompositeKey(ns+idxName, idxPair)
	if err != nil {
		return err
	}
	value := []byte{0x00}
	err = PutDoc(stub, compositeKey, value)
	if err != nil {
		return err
	}

	return nil
}

// []byte -> interface{}
// 要求[]byte 必须是json结构序列后的结果
// 返回 (true)struct or nil(error)
func ByteArrayToStructure(data []byte) interface{} {
	var res interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		logging.Error("Json Unmarshal []byte error| %s", err.Error())
		return nil
	}
	return &res
}
