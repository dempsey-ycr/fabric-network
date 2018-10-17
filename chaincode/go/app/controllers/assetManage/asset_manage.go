package assetManage

import (
	"encoding/json"
	. "iht-fabric-chaincode-private/chaincode/go/app/controllers"
	"iht-fabric-chaincode-private/chaincode/go/app/utils/filter"

	"iht-fabric-chaincode-private/chaincode/go/app/libraries/logging"
	"iht-fabric-chaincode-private/chaincode/go/app/models/db"
	resp "iht-fabric-chaincode-private/chaincode/go/app/utils/response"

	"iht-fabric-chaincode-private/chaincode/go/app/utils/mtime"

	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// 1. 资产类型(不动产)
const (
	IHT_ASSETMANAGE_ESTATETYPE_ALL    = iota // 全部
	IHT_ASSETMANAGE_ESTATETYPE_VILLA         // 别墅
	IHT_ASSETMANAGE_ESTATETYPE_NOW           // 现房
	IHT_ASSETMANAGE_ESTATETYPE_GROUND        // 土地
	IHT_ASSETMANAGE_ESTATETYPE_FUTURE        // 期房
)

// 2. 资产状态
const (
	IHT_ASSETMANAGE_ASSETSTATUS_ALL     = iota + 10 // 全部
	IHT_ASSETMANAGE_ASSETSTATUS_BEFORE              // 认购未开始
	IHT_ASSETMANAGE_ASSETSTATUS_ING                 // 认购中
	IHT_ASSETMANAGE_ASSETSTATUS_FAILED              // 认购失败
	IHT_ASSETMANAGE_ASSETSTATUS_SUCCEED             // 认购成功
)

// 资产简介
type AssetModelProfile struct {
	BaseInfo                  db.DataBlockBase `json:"baseInfo"`
	OwnerID                   string           `json:"ownerID"`              // 资产方ID
	OwnerName                 string           `json:"ownerName"`            // 资产方（名称）
	AssetType                 int              `json:"assetType"`            // 资产类型
	AssetID                   string           `json:"assetID"`              // 资产ID
	AssetName                 string           `json:"assetName"`            // 资产名称
	AssetPhoto                []DocumentStore  `json:"assetPhoto"`           // 资产照片（HASH）
	AssetOwnerFinancingAmount int64            `json:"ownerFinancingAmount"` // 实际融资额
	AssetSplitNum             int              `json:"assetSplitNum"`        // 拆分份额
	MemberDiscount            float32          `json:"vipDiscount"`          // 特殊会籍折扣
	IhtAirdropBounty          float32          `json:"ihtAirdropBounty"`     // IHT空投奖励
	SubscribeNum              int              `json:"subscribeNum"`         // 已被认购份额
	StartDate                 int              `json:"startDate"`            // 开始时间  20180630
	EndDate                   int              `json:"endDate"`              // 结束时间  20210630
	AssetStatus               int              `json:"assetStatus"`          // 资产状态
	//TimeLimitTotal            float32         `json:"timeLimitTotal"`     // 认购周期   产品删了
}

// 资产详情(包含)
type AssetParticulars struct {
	BaseInfo                  db.DataBlockBase `json:"baseInfo"`
	OwnerID                   string           `json:"ownerID"`              // 资产方ID
	OwnerName                 string           `json:"ownerName"`            // 资产方（名称）
	AssetType                 int              `json:"assetType"`            // 资产类型
	AssetID                   string           `json:"assetID"`              // 资产ID
	AssetName                 string           `json:"assetName"`            // 资产名称
	AssetPhoto                []DocumentStore  `json:"assetPhoto"`           // 资产照片（HASH）
	AssetOwnerFinancingAmount int64            `json:"ownerFinancingAmount"` // 实际融资额
	AssetSplitNum             int              `json:"assetSplitNum"`        // 拆分份额
	MemberDiscount            float32          `json:"vipDiscount"`          // 特殊会籍折扣
	IhtAirdropBounty          float32          `json:"ihtAirdropBounty"`     // IHT空投奖励
	SubscribeNum              int              `json:"subscribeNum"`         // 已被认购份额
	StartDate                 int              `json:"startDate"`            // 开始时间  20180630
	EndDate                   int              `json:"endDate"`              // 结束时间  20210630
	AssetStatus               int              `json:"assetStatus"`          // 资产状态
	LeastSubscribe            float32          `json:"leastSubscribe"`       // 最低认购比例
	OtherDiscount             string           `json:"otherDiscount"`        // 其他优惠
	AssetDesc                 string           `json:"assetDesc"`            // 资产描述
	AssetQualifiedFile        DocumentStore    `json:"assetQualifiedFile"`   // 资质文件（HASH）
	WhitePaper                DocumentStore    `json:"whitePaper"`           // 白皮书（HASH）
	TeamMembers               []Members        `json:"teamMembers"`          // 团队成员
}

// 团队成员
type Members struct {
	ID      string        `json:"id"`      // id(预留)
	Name    string        `json:"name"`    // 团队成员姓名
	Duty    string        `json:"duty"`    // ...成员职务
	Photo   DocumentStore `json:"photo"`   // ...成员照片（1张）
	Profile string        `json:"profile"` // ...成员简介
}

// 文件存储（picture、file...）
type DocumentStore struct {
	//DocType   int    `json:"docType"`  // TODO 预留 可能会有file{“资质文件、白皮书...”} 与 photo等分类
	//DocStatus int    `json:"doctatus"` // 默认 0 可用  10 过期 99 作废
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// ------------------------------------------------Read----------------------------------------------------------//
// 资产列表查询
type RequestReadAssetList struct {
	Limit       int    `json:"limit"`       //取多少条
	AssetType   int    `json:"assetType"`   //资产类型
	AssetStatus int    `json:"assetStatus"` //资产状态
	OwnerID     string `json:"ownerID"`     //资产方ID
}
type ResponseReadAssetList struct {
	AssetModelProfile
}

// -------------------------------------------------------------------------------------------------------------------//

type AssetManagement struct {
}

// 资产方数据入链
func (m *AssetManagement) Insert(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	// ==== Check if assetOwnerID already exists ====
	var overall AssetParticulars
	if err := json.Unmarshal([]byte(args[0]), &overall); err != nil {
		return resp.ErrorArguments("Invalid parameter structure...")
	}
	key, err := createAndCheckKey(stub, &overall)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	overall.BaseInfo.CreateTime = mtime.Now()

	// ==== Insert chaincode =====
	if err = db.PutState(stub, key, &overall); err != nil {
		return resp.ErrorDB(err.Error())
	}

	//create IDX_uers_status_2_user_email
	err = db.CreateCKeyWithNamespace(stub, KEY_ASSETDATA_TAG, KEY_ORGANIZATION, []string{overall.AssetID, overall.OwnerID})
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(nil)
}

// 删除资产详情
func (m *AssetManagement) Delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	assetOwnerID := args[0]
	if err := stub.DelState(createKeyByID(assetOwnerID)); err != nil {
		return resp.ErrorDB("Failed to delete Student from DB, key is: " + createKeyByID(assetOwnerID) + " and error: " + err.Error())
	}
	return resp.Success(nil)
}

// 修改资产详情
func (m *AssetManagement) Change(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	var update AssetParticulars
	if err := json.Unmarshal([]byte(args[0]), &update); err != nil {
		return resp.ErrorArguments("Invalid parameter structure...")
	}

	data, err := db.GetState(stub, createKeyByID(update.AssetID))
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	if len(data) == 0 {
		logging.Error("The data to be modified does not exist")
		return resp.ErrorNormal("The data to be modified does not exist")
	}
	update.BaseInfo.DBType = ST_CONSENSUS_INIT //
	update.BaseInfo.DBProfile = CONSENSUS_DESC[ST_CONSENSUS_INIT]
	update.BaseInfo.UpdateTime = mtime.Now()

	if err = db.PutState(stub, createKeyByID(update.AssetID), &update); err != nil {
		return resp.ErrorDB(err.Error())
	}

	res, err := json.Marshal(&update)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(res)
}

// 查询资产详情
func (m *AssetManagement) ReadDesc(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	assetOwnerID := args[0]
	data, err := db.GetState(stub, createKeyByID(assetOwnerID))
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(data)
}

// 溯源某一资产详情历史记录
func (m *AssetManagement) TraceHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	ownerID := args[0]

	data, err := db.GetHistoryForDocWithNamespace(stub, createKeyByID(ownerID))
	if err != nil {
		resp.ErrorNormal(err.Error())
	}
	return resp.Success(data)
}

// 根据各种条件查询资产列表(包括状态查询)
func (m *AssetManagement) ReadList(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	request := &RequestReadAssetList{}
	if err := json.Unmarshal([]byte(args[0]), request); err != nil {
		return resp.ErrorArguments(err.Error())
	}

	// 以资产方ID、 资产状态作为查询条件
	cond := map[string]interface{}{
		"baseInfo.docTag": KEY_ASSETDATA_TAG + KEY_ORGANIZATION,
	}
	if request.AssetType != IHT_ASSETMANAGE_ESTATETYPE_ALL { // 不限
		cond["assetType"] = request.AssetType
	}
	if request.OwnerID != "" {
		cond["ownerID"] = request.OwnerID
	}

	selector := db.QueryRichList(cond)
	iterator, err := stub.GetQueryResult(selector)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	defer iterator.Close()

	var status []*db.WorldState
	for iterator.HasNext() {
		query, err := iterator.Next()
		if err != nil {
			return resp.ErrorDB(err.Error())
		}

		assetPro, err := getAssetSingle(query.GetValue())
		if err != nil {
			return resp.ErrorDB(err.Error())
		}

		// 以资产状态筛选返回结果
		var state *db.WorldState
		if assetPro.AssetStatus != IHT_ASSETMANAGE_ASSETSTATUS_ALL { //不限
			state = &db.WorldState{
				StateRecord: assetPro,
				StateKey:    query.GetKey(),
			}
		} else {
			if assetPro.AssetStatus == request.AssetStatus {
				state = &db.WorldState{
					StateRecord: assetPro,
					StateKey:    query.GetKey(),
				}
			} else {
				continue
			}
		}
		status = append(status, state)
	}
	list := &db.List{
		Object: status,
	}

	worldStatus, err := json.Marshal(list)
	if err != nil {
		logging.Error("Unmarshal error| return Statues [%s]", err.Error())
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(worldStatus)
}

// 获取资产链表
// 链里只存了一份资产详情数据， 故资产链表需要资产详情数据做处理
func getAssetSingle(data []byte) (*ResponseReadAssetList, error) {
	if len(data) == 0 {
		logging.Error("Asset lists have no data to use...")
		return nil, fmt.Errorf("%s", "Asset lists have no data to use...")
	}
	var detail AssetParticulars           // 资产详情
	profile := new(ResponseReadAssetList) //资产简介

	if err := json.Unmarshal(data, &detail); err != nil {
		logging.Error("Asset list processing, the structure does not match |%v", err.Error())
		return nil, fmt.Errorf("Asset list processing, the structure does not match |%v", err.Error())
	}

	profile.OwnerID = detail.OwnerID
	profile.OwnerName = detail.OwnerName
	profile.AssetType = detail.AssetType
	profile.AssetID = detail.AssetID
	profile.AssetName = detail.AssetName
	profile.AssetPhoto = detail.AssetPhoto
	profile.AssetOwnerFinancingAmount = detail.AssetOwnerFinancingAmount
	profile.AssetSplitNum = detail.AssetSplitNum
	profile.MemberDiscount = detail.MemberDiscount
	profile.IhtAirdropBounty = detail.IhtAirdropBounty
	profile.SubscribeNum = detail.SubscribeNum
	profile.StartDate = detail.StartDate
	profile.EndDate = detail.EndDate
	profile.AssetStatus = subscribeStatus(&detail)

	base := db.DataBlockBase{
		DocTag:     detail.BaseInfo.DocTag,
		DBType:     detail.BaseInfo.DBType,
		DBProfile:  detail.BaseInfo.DBProfile,
		CreateTime: detail.BaseInfo.CreateTime,
		UpdateTime: detail.BaseInfo.UpdateTime,
	}
	profile.BaseInfo = base
	logging.Debug("Asset lists value: %v", profile)
	return profile, nil
}
