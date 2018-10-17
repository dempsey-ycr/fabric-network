package assetManage

import (
	"encoding/json"
	"iht-fabric-chaincode-private/chaincode/go/app/utils/filter"

	"fmt"
	"iht-fabric-chaincode-private/chaincode/go/app/libraries/logging"
	"iht-fabric-chaincode-private/chaincode/go/app/models/db"
	"iht-fabric-chaincode-private/chaincode/go/app/utils/mtime"
	resp "iht-fabric-chaincode-private/chaincode/go/app/utils/response"
	. "iht-fabric-chaincode-private/chaincode/go/app/controllers"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	KEY_ASSETPROCESS_TAG = "AssetProcess_" // 资产进度基准key
)

// 资产进度
type AssetProcess struct {
	db.DataBlockBase
	AssetID            string          `json:"assetID"`            //资产ID
	ProcessTime        string          `json:"processTime"`        //资产进度时间
	ProcessStauts      string          `json:"processStauts"`      // 资产描述
	AssetQualifiedFile []DocumentStore `json:"assetQualifiedFile"` // 资质文件（HASH）
}

// 资产进度入链
func (m *AssetProcess) Insert(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	// ==== Check if assetOwnerID already exists ====
	var overall AssetProcess
	if err := json.Unmarshal([]byte(args[0]), &overall); err != nil {
		return resp.ErrorArguments("Invalid parameter structure...")
	}
	key, err := createAndCheckByKey(stub, &overall)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	overall.CreateTime = mtime.Now()

	// ==== Insert chaincode =====
	if err = db.PutState(stub, key, &overall); err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(nil)
}

// 查询资产进度
func (m *AssetProcess) Read(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	assetID := args[0]
	data, err := db.GetState(stub, createAssertProcessKeyByID(assetID))
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(data)
}

// 修改资产进度
func (m *AssetProcess) Change(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	var update AssetProcess
	if err := json.Unmarshal([]byte(args[0]), &update); err != nil {
		return resp.ErrorArguments("Invalid parameter structure...")
	}

	data, err := db.GetState(stub, createAssertProcessKeyByID(update.AssetID))
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	if len(data) == 0 {
		logging.Error("The asset process to be modified does not exist")
		return resp.ErrorNormal("The asset process to be modified does not exist")
	}
	update.UpdateTime = mtime.Now()

	if err = db.PutState(stub, createAssertProcessKeyByID(update.AssetID), &update); err != nil {
		return resp.ErrorDB(err.Error())
	}

	res, err := json.Marshal(&update)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(res)
}

// 删除资产详情
func (m *AssetProcess) Delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	assetID := args[0]
	if err := stub.DelState(createAssertProcessKeyByID(assetID)); err != nil {
		return resp.ErrorDB("Failed to delete Student from DB, key is: " + createKeyByID(assetID) + " and error: " + err.Error())
	}
	return resp.Success(nil)
}

// 溯源某一资产详情历史记录
func (m *AssetProcess) History(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	ownerID := args[0]
	fmt.Println("--------------")
	data, err := db.GetHistoryForDocWithNamespace(stub, createAssertProcessKeyByID(ownerID))
	if err != nil {
		resp.ErrorNormal(err.Error())
	}
	return resp.Success(data)
}

// 根据条件查询资产进度
func (m *AssetProcess) ReadList(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	request := &AssetProcess{}
	if err := json.Unmarshal([]byte(args[0]), request); err != nil {
		return resp.ErrorArguments(err.Error())
	}
	// 以资产方ID、 资产状态作为查询条件
	cond := map[string]interface{}{}

	if request.AssetID != "" { // 不限
		cond["assetID"] = request.AssetID
	}
	if request.ProcessStauts != "" {
		cond["processStauts"] = request.ProcessStauts
	}
	selector := db.QueryRichList(cond)
	iterator, err := stub.GetQueryResult(selector)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	defer iterator.Close()
	var status []*AssetProcess
	for iterator.HasNext() {
		query, err := iterator.Next()
		if err != nil {
			return resp.ErrorDB(err.Error())
		}
		detail := &AssetProcess{}
		if err := json.Unmarshal(query.GetValue(), detail); err != nil {
			logging.Error("Asset list processing, the structure does not match |%v", err.Error())
			return resp.ErrorDB(err.Error())
		}
		status = append(status, detail)
	}
	list := &db.List{
		Object: status,
	}
	jlist, err := json.Marshal(list)
	return resp.Success(jlist)
}

// 创建并检查key
func createAndCheckByKey(stub shim.ChaincodeStubInterface, desc *AssetProcess) (string, error) {
	desc.DocTag = KEY_ASSETPROCESS_TAG + KEY_ORGANIZATION
	key := desc.DocTag + desc.AssetID
	if data, err := db.GetDocWithNamespace(stub, key); err != nil {
		return "", err
	} else {
		if len(data) == 0 {
			desc.DBProfile = "可用"
			return key, nil
		}
	}
	return "", fmt.Errorf("The key[%s] is Exist !!!", key)
}

// 创建key
func createAssertProcessKeyByID(assetID string) string {
	return KEY_ASSETPROCESS_TAG + KEY_ORGANIZATION + assetID
}
