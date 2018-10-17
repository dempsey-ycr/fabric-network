package assetManage

import (
	"fmt"
	"iht-fabric-chaincode-private/chaincode/go/app/models/db"

	"iht-fabric-chaincode-private/chaincode/go/app/utils/mtime"
	."iht-fabric-chaincode-private/chaincode/go/app/controllers"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const DEFAULT_LIMIT = 50

func checkParamsDefault(r *RequestReadAssetList) {
	if r.Limit == 0 {
		r.Limit = DEFAULT_LIMIT
	}
}

// 资产目前被认购比例
func subscribeRatio(info *AssetParticulars) float32 {
	return float32(info.SubscribeNum) / float32(info.AssetSplitNum)
}

//认购状态
func subscribeStatus(info *AssetParticulars) int {
	today := mtime.TodayDateInt()
	if today < info.StartDate {
		return IHT_ASSETMANAGE_ASSETSTATUS_BEFORE
	} else if info.EndDate < today {
		if subscribeRatio(info) < info.LeastSubscribe {
			return IHT_ASSETMANAGE_ASSETSTATUS_FAILED
		} else {
			return IHT_ASSETMANAGE_ASSETSTATUS_SUCCEED
		}
	} else {
		return IHT_ASSETMANAGE_ASSETSTATUS_ING
	}
}

// 创建并检查key
func createAndCheckKey(stub shim.ChaincodeStubInterface, desc *AssetParticulars) (string, error) {
	desc.BaseInfo.DocTag = KEY_ASSETDATA_TAG + KEY_ORGANIZATION
	key := desc.BaseInfo.DocTag + desc.AssetID
	if data, err := db.GetDocWithNamespace(stub, key); err != nil {
		return "", err
	} else {
		if len(data) == 0 {
			desc.BaseInfo.DBProfile = "可用"
			return key, nil
		}
	}
	return "", fmt.Errorf("The key[%s] is Exist !!!", key)
}

// 创建key
func createKeyByID(assetID string) string {
	return KEY_ASSETDATA_TAG + KEY_ORGANIZATION + assetID
}
