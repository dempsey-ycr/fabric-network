package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

const (
	ST_sec_bein_underwriting          string = "10" // 10-承销中
	ST_sec_halt_underwriting          string = "11" // 11-中止承销
	ST_sec_succ_underwriting          string = "12" // 12-承销成功
	ST_sec_fail_underwriting          string = "13" // 13-承销失败
	ST_sec_security_units_initialized string = "20" // 20-份额拆分完成
	ST_sec_effect_ownership_change    string = "21" // 21-权益转移生效
	ST_sec_finish_repurchasing        string = "22" // 22-回购完成
)

const (
	ST_DESC_sec_bein_underwriting          string = "underwriting"               // 10-承销中
	ST_DESC_sec_halt_underwriting          string = "halt underwriting"          // 11-中止承销
	ST_DESC_sec_succ_underwriting          string = "underwriting success"       // 12-承销成功
	ST_DESC_sec_fail_underwriting          string = "underwriting failed"        // 13-承销失败
	ST_DESC_sec_security_units_initialized string = "security units initialized" // 20-份额拆分完成
	ST_DESC_sec_effect_ownership_change    string = "ownership changes effected" // 21-权益转移生效
	ST_DESC_sec_finish_repurchasing        string = "repurchasing finished"      // 22-回购完成
)

const (
	DT_asset_securitized                  string = "asset_securitized"
	NS_asset_securitized                  string = DT_asset_securitized + "_"
	PK_FD_asset_securitized               string = "asset_name"
	IDX_FD_asset_status                   string = "asset_status"
	IDX_asset_status_2_asset_name         string = IDX_FD_asset_status + "_2_" + PK_FD_asset_securitized
	DT_asset_security_underwrote          string = "asset_security_underwrote"
	NS_asset_security_underwrote          string = DT_asset_security_underwrote + "_"
	PK_FD_asset_security_underwrote       string = "underwrote_id"
	IDX_FD_fininst_email                  string = "fininst_email"
	IDX_FD_asset_owner_email              string = "asset_owner_email"
	IDX_FD_asset_name                     string = "asset_name"
	IDX_fininst_email_2_underwrote_id     string = IDX_FD_fininst_email + "_2_" + PK_FD_asset_security_underwrote
	IDX_asset_owner_email_2_underwrote_id string = IDX_FD_asset_owner_email + "_2_" + PK_FD_asset_security_underwrote
	IDX_asset_name_2_underwrote_id        string = IDX_FD_asset_name + "_2_" + PK_FD_asset_security_underwrote
)

type AssetSecuritizedMngt struct {
}

type asset_securitized struct {
	ObjectType                       string  `json:"docType"`                          //docType is used to distinguish the various types of objects in state database
	Owner_email                      string  `json:"owner_email"`                      //资产方邮箱
	Asset_name                       string  `json:"asset_name"`                       //资产名称
	Asset_pic_hashs                  string  `json:"asset_pic_hashs"`                  //资产图片哈希值，用逗号分隔xxxxxx,yyyyyyy,zzzzzz
	Asset_desc                       string  `json:"asset_desc"`                       //资产描述
	Asset_type                       string  `json:"asset_type"`                       //资产类型
	Asset_value_ccy                  string  `json:"asset_value_ccy"`                  //资产价值货币类型
	Asset_value_amt                  float64 `json:"asset_value_amt"`                  //资产价值货币类型
	Asset_securitized_parts          int     `json:"asset_securitized_parts"`          //资产拆分份额
	Asset_annual_dividend            float64 `json:"asset_annual_dividend"`            //资产年化收益
	Asset_repurchase_period          int     `json:"asset_repurchase_period"`          //资产回购时间
	Asset_underwrite_period          int     `json:"asset_underwrite_period"`          //资产承销时限
	Asset_underwrite_bottom_percent  float64 `json:"asset_underwrite_bottom_percent"`  //资产承销比率底限
	Asset_allow_delegate_rent_income bool    `json:"asset_allow_delegate_rent_income"` //是否可以委托租赁产生额外收益
	Asset_proof_file_hashs           string  `json:"asset_proof_file_hashs"`           //资产证明文件哈希值，用逗号分隔xxxxxx,yyyyyyy,zzzzzz
	Asset_status                     string  `json:"asset_status"`                     //资产当前状态：00-init 01-正在审核 02-审核通过 03-审核不通过 10-承销中 11-中止承销 12-承销成功 13-承销失败 20-权益转移生效 21-回购完成 99-作废
	Iht_status_comments              string  `json:"iht_status_comments"`              //IHT平台修改Asset_status的备注信息
	Underwrite_start_time            string  `json:"underwrite_start_time"`            //承销开始时间，chaincode记录的启动承销接口调用时的时间戳
	Underwrite_due_time              string  `json:"underwrite_due_time"`              //根据承销时限和承销开始时间计算的承销截止时间戳
	Repurchase_due_time              string  `json:"repurchase_due_time"`              //根据回购时限和份额权益转移时间计算的回购截止时间戳
	Current_underwrite_percent       float64 `json:"current_underwrote_percent"`       //当前承销完成比率，冗余信息提高效率
	Current_underwrite_id            string  `json:"current_underwrote_id"`            //当前承销ID，冗余信息提高效率
	Asset_hash                       string  `json:"asset_hash"`                       //基于资产份额unit_id生成的资产哈希，是资产分拆份额管理的入口, HASH256({"asset_name":"xxxx","unit_ids":["unit_id1","unit_id2",…,"unit_idtn"],"owner_email":"xx@xx","asset_value_ccy":"XXX","asset_value_amt":xxxx.xx,"asset_securitized_parts":n,"asset_annual_dividend":n.nnnn,"asset_repurchase_time":"xxxxxx","unit_allow_delegate_rent_income":true|false})
	Owner_signature                  string  `json:"owner_signature"`                  //资产方对本asset_securitized关键信息的签名（保留字段，目前送“”）
	Iht_signature                    string  `json:"iht_signature"`                    //IHT平台对本asset_securitized关键信息的签名（保留字段，目前送“”）
}

type asset_security_underwrote struct {
	ObjectType        string `json:"docType"`           //文档类别"asset_security_underwrote"
	Underwrote_id     string `json:"underwrote_id"`     //合法承销id，必须全局唯一，由前端生成
	Fininst_email     string `json:"fininst_email"`     //金融机构邮箱，合法状态的金融机构才可以登记承销份额
	Asset_owner_email string `json:"asset_owner_email"` //资产方邮箱
	Asset_name        string `json:"asset_name"`        //资产名称
	Underwrote_units  int    `json:"underwrote_units"`  //承销资产分拆份额
	Underwrote_time   string `json:"underwrote_time"`   //承销时间，由chaincode记录的underwrite接口调用时间，∈[underwrite_start_time, underwrite_due_time]
	Fininst_signature string `json:"fininst_signature"` //金融机构对本asset_security_underwrote关键信息的签名（保留字段，目前送“”）
}

// ============================================================
// initAssetSecuritized - create a new asset_securitized_mngt, store into chaincode state
// ============================================================
func (t *AssetSecuritizedMngt) initAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 0            1           2                3           4           5                6                7                        8                      9                        10                       11                               12                                13                      14
	// owner_email, asset_name, asset_pic_hashs, asset_desc, asset_type, asset_value_ccy, asset_value_amt, asset_securitized_parts, asset_annual_dividend, asset_repurchase_period, asset_underwrite_period, asset_underwrite_bottom_percent, asset_allow_delegate_rent_income, asset_proof_file_hashs, owner_signature
	if len(args) != 15 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 15"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start initAssetSecuritized")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("owner_email 1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("asset_name 2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 5 {
		return errorPbResponse(errors.New("asset_pic_hashs 3rd argument must be a valid comma seperated hashs string"), true)
	}
	if len(args[4]) <= 0 {
		return errorPbResponse(errors.New("asset_type 5th argument must be a non-empty string"), true)
	}
	if len(args[5]) <= 0 {
		return errorPbResponse(errors.New("asset_value_ccy 6th argument must be a non-empty string"), true)
	}
	if len(args[6]) <= 0 {
		return errorPbResponse(errors.New("asset_value_amt 7th argument must be a non-empty string"), true)
	}
	if len(args[7]) <= 0 {
		return errorPbResponse(errors.New("asset_securitized_parts 8th argument must be a non-empty string"), true)
	}
	if len(args[8]) <= 0 {
		return errorPbResponse(errors.New("asset_annual_dividend 9th argument must be a non-empty string"), true)
	}
	if len(args[9]) <= 0 {
		return errorPbResponse(errors.New("asset_repurchase_period 10th argument must be a non-empty string"), true)
	}
	if len(args[10]) <= 0 {
		return errorPbResponse(errors.New("asset_underwrite_period 11th argument must be a non-empty string"), true)
	}
	if len(args[11]) <= 0 {
		return errorPbResponse(errors.New("asset_underwrite_bottom_percent 12th argument must be a non-empty string"), true)
	}
	if len(args[12]) <= 0 {
		return errorPbResponse(errors.New("asset_allow_delegate_rent_income 13th argument must be a non-empty string"), true)
	}
	if len(args[13]) <= 5 {
		return errorPbResponse(errors.New("asset_proof_file_hashs 14th argument must be a valid comma seperated hashs string"), true)
	}

	owner_email := args[0]
	asset_name := args[1]
	asset_pic_hashs := args[2]
	asset_desc := args[3]
	asset_type := args[4]
	asset_value_ccy := args[5]
	asset_value_amt, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		return errorPbResponse(errors.New("asset_value_amt 7th argument must be a valid float64 string"), true)
	}
	asset_securitized_parts, err := strconv.Atoi(args[7])
	if err != nil {
		return errorPbResponse(errors.New("asset_securitized_parts 8th argument must be a valid int string"), true)
	}
	asset_annual_dividend, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return errorPbResponse(errors.New("asset_annual_dividend 9th argument must be a valid float64 string"), true)
	}
	asset_repurchase_period, err := strconv.Atoi(args[9])
	if err != nil {
		return errorPbResponse(errors.New("asset_repurchase_period 10th argument must be a valid int string"), true)
	}
	asset_underwrite_period, err := strconv.Atoi(args[10])
	if err != nil {
		return errorPbResponse(errors.New("asset_underwrite_period 11th argument must be a valid int string"), true)
	}
	asset_underwrite_bottom_percent, err := strconv.ParseFloat(args[11], 64)
	if err != nil {
		return errorPbResponse(errors.New("asset_underwrite_bottom_percent 12th argument must be a valid float64 string"), true)
	}
	asset_allow_delegate_rent_income, err := strconv.ParseBool(args[12])
	if err != nil {
		return errorPbResponse(errors.New("asset_allow_delegate_rent_income 13th argument must be a valid bool string"), true)
	}
	asset_proof_file_hashs := args[13]
	owner_signature := args[14]
	//TODO verify signature of owner signature

	asset_status := ST_comm_init
	iht_status_comments := ST_DESC_comm_init
	underwrite_start_time := ""
	underwrite_due_time := ""
	repurchase_due_time := ""
	current_underwrite_percent := 0.0000
	current_underwrite_id := ""
	asset_hash := ""

	iht_signature := ""

	// ==== Check if asset_owner already exists and status is 02 ====
	AssetOwnerAsBytes, err := getDocWithNamespace(stub, NS_asset_owner, owner_email)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_owner: "+owner_email+", "+err.Error()), true)
	}
	if AssetOwnerAsBytes == nil {
		return errorPbResponse(errors.New("Failed to get asset_owner: "+owner_email+", asset_owner not found"), true)
	}

	assetOwner := asset_owner{}
	err = json.Unmarshal(AssetOwnerAsBytes, &assetOwner)
	if err != nil {
		msg := "asset_owner exists: " + owner_email + ", but failed to be unmarshalled as an asset_owner json"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if ST_comm_approved != assetOwner.Owner_status {
		msg := "asset_owner exists: " + owner_email + ", but current Owner_status is " + assetOwner.Owner_status + ", status 02-approved required"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	// ==== Check if asset_owner_detail already exists and status is 02 ====
	AssetOwnerDetailAsBytes, err := getDocWithNamespace(stub, NS_asset_owner_detail, owner_email)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_owner_detail: "+owner_email+", "+err.Error()), true)
	}
	if AssetOwnerDetailAsBytes == nil {
		return errorPbResponse(errors.New("Failed to get asset_owner_detail: "+owner_email+", asset_owner_detail not found"), true)
	}

	assetOwnerDetail := asset_owner_detail{}
	err = json.Unmarshal(AssetOwnerDetailAsBytes, &assetOwnerDetail)
	if err != nil {
		msg := "asset_owner_detail exists: " + owner_email + ", but failed to be unmarshalled as an asset_owner_detail json"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if ST_comm_approved != assetOwnerDetail.Detail_status {
		msg := "asset_owner_detail exists: " + owner_email + ", but current Detail_status is " + assetOwnerDetail.Detail_status + ", status 02-approved required"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	// ==== Check if asset_securitized already exists ====
	AssetSecuritizedAsBytes, err2 := getDocWithNamespace(stub, NS_asset_securitized, asset_name)
	if err2 != nil {
		return errorPbResponse(errors.New("Failed to get asset_securitized: "+err.Error()), true)
	} else if AssetSecuritizedAsBytes != nil {
		fmt.Println("This asset_securitized already exists: " + asset_name)
		return errorPbResponse(errors.New("This asset_securitized already exists: "+asset_name), true)
	}

	// ==== Create asset_securitized_mngt object and marshal to JSON ====
	objectType := DT_asset_securitized
	AssetSecuritized := &asset_securitized{objectType, owner_email, asset_name, asset_pic_hashs, asset_desc, asset_type, asset_value_ccy, asset_value_amt, asset_securitized_parts, asset_annual_dividend, asset_repurchase_period, asset_underwrite_period, asset_underwrite_bottom_percent, asset_allow_delegate_rent_income, asset_proof_file_hashs, asset_status, iht_status_comments, underwrite_start_time, underwrite_due_time, repurchase_due_time, current_underwrite_percent, current_underwrite_id, asset_hash, owner_signature, iht_signature}
	AssetSecuritizedJSONasBytes, err := json.Marshal(AssetSecuritized)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//Alternatively, build the asset_securitized json string manually if you don't want to use struct marshalling
	//AssetSecuritizedJSONasString := `{"docType":"asset_securitized_mngt",  "Trade_no": "` + trade_no + `", "Mer_id": "` + mer_id + `", "Card_no": ` + card_no + `, "Trade_amt": "` + trade_amt + `"}`
	//AssetSecuritizedJSONasBytes := []byte(AssetSecuritizedJSONasString)

	// === Save asset_securitized_mngt to state ===
	err = putDocWithNamespace(stub, NS_asset_securitized, asset_name, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	//add IDX_asset_status_2_asset_name
	err = createCKeyWithNamespace(stub, NS_asset_securitized, IDX_asset_status_2_asset_name, []string{AssetSecuritized.Asset_status, AssetSecuritized.Asset_name})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== asset_securitized_mngt saved and indexed. Return success ====
	fmt.Println("- end initAssetSecuritized (success)")
	return successPbResponse(nil)
}

// ===============================================
// readAssetSecuritized - read a asset_securitized_mngt from chaincode state
// ===============================================
func (t *AssetSecuritizedMngt) readAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name of the asset_securitized to query"), true)
	}

	assetName = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// queryAssetSecuritizedsByStatus - read a asset_securitized_mngt from chaincode state
// ===============================================
func (t *AssetSecuritizedMngt) queryAssetSecuritizedsByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "status_01"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 1"), true)
	}

	assetStatus := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_securitized, IDX_FD_asset_status, assetStatus)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// approveAssetSecuritized - read a asset_securitized_mngt from chaincode state
// ===============================================
func (t *AssetSecuritizedMngt) approveAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var assetName, ihtsig string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name and iht_signature to approve the asset securitization"), true)
	}

	assetName = args[0]
	ihtsig = args[1]

	return t.changeAssetSecuritizedStatus(stub, []string{assetName, ihtsig, ST_comm_approved, ST_DESC_comm_approved})

}

// ===============================================
// rejectAssetSecuritized - reject a asset_securitized as "03" status
// ===============================================
func (t *AssetSecuritizedMngt) rejectAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name and iht_signature to reject the asset securitization"), true)
	}

	assetName = args[0]
	ihtsig = args[1]

	return t.changeAssetSecuritizedStatus(stub, []string{assetName, ihtsig, ST_comm_rejected, ST_DESC_comm_rejected})
}

// ==================================================
// delete - delete a asset_securitized key/value pair from state
// ==================================================
func (t *AssetSecuritizedMngt) deleteAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var jsonResp string
	//var AssetSecuritizedJSON asset_securitized
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 1"), true)
	}
	assetName := args[0]
	fmt.Println("- start delete asset_securitized, assetName " + assetName)
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if ST_comm_niled == AssetSecuritizedToUpdate.Asset_status {
		jsonResp = "{\"Info\":\"" + assetName + "'s asset_securitized was already deleted!\"}"
		fmt.Println("- end delete asset_securitized (success) " + jsonResp)
	} else {

		AssetSecuritizedToUpdate.Asset_status = ST_comm_niled //change the securitized status to "02" as approve status
		AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_comm_niled

		AssetSecuritizedJSONasBytes, _ := json.Marshal(AssetSecuritizedToUpdate)
		err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
		if err != nil {
			return errorPbResponse(err, false)
		}

		fmt.Println("- end delete asset_securitized (success)")
	}

	return successPbResponse(nil)
}

func (t *AssetSecuritizedMngt) changeAssetSecuritizedStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig, status, desc, jsonResp string
	var err error

	if len(args) != 4 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting assetName, iht_signature, status, desc of the asset_securitized to change status"), true)
	}

	assetName = args[0]
	ihtsig = args[1]
	status = args[2]
	desc = args[3]

	fmt.Println("- start changeAssetSecuritizedStatus: assetName " + assetName + " , status " + status + " , desc " + desc)

	if status != ST_comm_approving && status != ST_comm_approved && status != ST_comm_rejected && status != ST_comm_niled {
		return errorPbResponse(errors.New("Incorrect changed status. Expecting 01-正在审核  02-审核通过  03-审核不通过  99-作废 for asset_securitized to change status"), true)
	}

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if AssetSecuritizedToUpdate.Asset_status == status && AssetSecuritizedToUpdate.Iht_status_comments == desc {
		fmt.Println("status " + status + " , desc " + desc + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = status
	AssetSecuritizedToUpdate.Iht_status_comments = desc

	AssetSecuritizedJSONasBytes, err := json.Marshal(AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeAssetSecuritizedStatus (success)")
	return successPbResponse(nil)
}

func (t *AssetSecuritizedMngt) changeAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	// 0            1           2                3           4           5                6                7                        8                      9                        10                       11                               12                                13                      14
	// owner_email, asset_name, asset_pic_hashs, asset_desc, asset_type, asset_value_ccy, asset_value_amt, asset_securitized_parts, asset_annual_dividend, asset_repurchase_period, asset_underwrite_period, asset_underwrite_bottom_percent, asset_allow_delegate_rent_income, asset_proof_file_hashs, owner_signature
	if len(args) != 15 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 15"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start changeAssetSecuritized")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("owner_email 1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("asset_name 2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 5 {
		return errorPbResponse(errors.New("asset_pic_hashs 3rd argument must be a valid comma seperated hashs string"), true)
	}
	if len(args[4]) <= 0 {
		return errorPbResponse(errors.New("asset_type 5th argument must be a non-empty string"), true)
	}
	if len(args[5]) <= 0 {
		return errorPbResponse(errors.New("asset_value_ccy 6th argument must be a non-empty string"), true)
	}
	if len(args[6]) <= 0 {
		return errorPbResponse(errors.New("asset_value_amt 7th argument must be a non-empty string"), true)
	}
	if len(args[7]) <= 0 {
		return errorPbResponse(errors.New("asset_securitized_parts 8th argument must be a non-empty string"), true)
	}
	if len(args[8]) <= 0 {
		return errorPbResponse(errors.New("asset_annual_dividend 9th argument must be a non-empty string"), true)
	}
	if len(args[9]) <= 0 {
		return errorPbResponse(errors.New("asset_repurchase_period 10th argument must be a non-empty string"), true)
	}
	if len(args[10]) <= 0 {
		return errorPbResponse(errors.New("asset_underwrite_period 11th argument must be a non-empty string"), true)
	}
	if len(args[11]) <= 0 {
		return errorPbResponse(errors.New("asset_underwrite_bottom_percent 12th argument must be a non-empty string"), true)
	}
	if len(args[12]) <= 0 {
		return errorPbResponse(errors.New("asset_allow_delegate_rent_income 13th argument must be a non-empty string"), true)
	}
	if len(args[13]) <= 5 {
		return errorPbResponse(errors.New("asset_proof_file_hashs 14th argument must be a valid comma seperated hashs string"), true)
	}

	owner_email := args[0]
	asset_name := args[1]
	asset_pic_hashs := args[2]
	asset_desc := args[3]
	asset_type := args[4]
	asset_value_ccy := args[5]
	asset_value_amt, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		return errorPbResponse(errors.New("asset_value_amt 7th argument must be a valid float64 string"), true)
	}
	asset_securitized_parts, err := strconv.Atoi(args[7])
	if err != nil {
		return errorPbResponse(errors.New("asset_securitized_parts 8th argument must be a valid int string"), true)
	}
	asset_annual_dividend, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return errorPbResponse(errors.New("asset_annual_dividend 9th argument must be a valid float64 string"), true)
	}
	asset_repurchase_period, err := strconv.Atoi(args[9])
	if err != nil {
		return errorPbResponse(errors.New("asset_repurchase_period 10th argument must be a valid int string"), true)
	}
	asset_underwrite_period, err := strconv.Atoi(args[10])
	if err != nil {
		return errorPbResponse(errors.New("asset_underwrite_period 11th argument must be a valid int string"), true)
	}
	asset_underwrite_bottom_percent, err := strconv.ParseFloat(args[11], 64)
	if err != nil {
		return errorPbResponse(errors.New("asset_underwrite_bottom_percent 12th argument must be a valid float64 string"), true)
	}
	asset_allow_delegate_rent_income, err := strconv.ParseBool(args[12])
	if err != nil {
		return errorPbResponse(errors.New("asset_allow_delegate_rent_income 13th argument must be a valid bool string"), true)
	}
	asset_proof_file_hashs := args[13]
	owner_signature := args[14]
	//TODO verify signature of owner signature

	// ==== Check if asset_securitized already exists ====
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, asset_name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + asset_name + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + asset_name + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if AssetSecuritizedToUpdate.Asset_status != ST_comm_init && AssetSecuritizedToUpdate.Asset_status != ST_comm_rejected {
		msg := "current asset_securitized asset_status " + AssetSecuritizedToUpdate.Asset_status + "-" + AssetSecuritizedToUpdate.Iht_status_comments + " not allowed to change asset_securitized, only 00-init or 03-IHT rejected allowed"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if AssetSecuritizedToUpdate.Owner_email != owner_email {
		msg := "asset " + asset_name + " does not belong to asset_owner " + owner_email + ", asset_securitized change request rejected"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	var isChanged bool = false

	if AssetSecuritizedToUpdate.Asset_pic_hashs != asset_pic_hashs {
		AssetSecuritizedToUpdate.Asset_pic_hashs = asset_pic_hashs
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_desc != asset_desc {
		AssetSecuritizedToUpdate.Asset_desc = asset_desc
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_type != asset_type {
		AssetSecuritizedToUpdate.Asset_type = asset_type
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_value_ccy != asset_value_ccy {
		AssetSecuritizedToUpdate.Asset_value_ccy = asset_value_ccy
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_value_amt != asset_value_amt {
		AssetSecuritizedToUpdate.Asset_value_amt = asset_value_amt
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_securitized_parts != asset_securitized_parts {
		AssetSecuritizedToUpdate.Asset_securitized_parts = asset_securitized_parts
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_annual_dividend != asset_annual_dividend {
		AssetSecuritizedToUpdate.Asset_annual_dividend = asset_annual_dividend
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_repurchase_period != asset_repurchase_period {
		AssetSecuritizedToUpdate.Asset_repurchase_period = asset_repurchase_period
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_underwrite_period != asset_underwrite_period {
		AssetSecuritizedToUpdate.Asset_underwrite_period = asset_underwrite_period
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_underwrite_bottom_percent != asset_underwrite_bottom_percent {
		AssetSecuritizedToUpdate.Asset_underwrite_bottom_percent = asset_underwrite_bottom_percent
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_allow_delegate_rent_income != asset_allow_delegate_rent_income {
		AssetSecuritizedToUpdate.Asset_allow_delegate_rent_income = asset_allow_delegate_rent_income
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Asset_proof_file_hashs != asset_proof_file_hashs {
		AssetSecuritizedToUpdate.Asset_proof_file_hashs = asset_proof_file_hashs
		isChanged = true
	}
	if AssetSecuritizedToUpdate.Owner_signature != owner_signature {
		AssetSecuritizedToUpdate.Owner_signature = owner_signature
		isChanged = true
	}

	if !isChanged {
		fmt.Println("- end changeAssetSecuritized (no change no commit)")
		return successPbResponse(nil)
	}

	//reset status to 00-init if isChanged
	if AssetSecuritizedToUpdate.Asset_status == ST_comm_rejected {
		AssetSecuritizedToUpdate.Asset_status = ST_comm_init
		AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_comm_init_change
		AssetSecuritizedToUpdate.Iht_signature = ""
	}

	AssetSecuritizedToUpdate.Asset_desc = asset_desc
	AssetSecuritizedToUpdate.Asset_type = asset_type
	AssetSecuritizedToUpdate.Asset_value_ccy = asset_value_ccy
	AssetSecuritizedToUpdate.Asset_value_amt = asset_value_amt
	AssetSecuritizedToUpdate.Asset_securitized_parts = asset_securitized_parts
	AssetSecuritizedToUpdate.Asset_annual_dividend = asset_annual_dividend
	AssetSecuritizedToUpdate.Asset_repurchase_period = asset_repurchase_period
	AssetSecuritizedToUpdate.Asset_underwrite_period = asset_underwrite_period
	AssetSecuritizedToUpdate.Asset_underwrite_bottom_percent = asset_underwrite_bottom_percent
	AssetSecuritizedToUpdate.Asset_allow_delegate_rent_income = asset_allow_delegate_rent_income
	AssetSecuritizedToUpdate.Asset_proof_file_hashs = asset_proof_file_hashs
	AssetSecuritizedToUpdate.Owner_signature = owner_signature

	// === Save asset_securitized to state ===
	AssetSecuritizedJSONasBytes, err := json.Marshal(AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_asset_securitized, asset_name, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== asset_securitized updated. Return success ====
	fmt.Println("- end changeAssetSecuritized (success)")
	return successPbResponse(nil)
}

func (t *AssetSecuritizedMngt) getHistoryForAssetSecuritized(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 1"), true)
	}

	assetName := args[0]

	fmt.Printf("- start getHistoryForAssetSecuritized: %s\n", assetName)

	bytes, err := getHistoryForDocWithNamespace(stub, NS_asset_securitized, assetName)

	if err != nil {
		return errorPbResponse(err, true)
	}

	return successPbResponse(bytes)
}

func (t *AssetSecuritizedMngt) startUnderwrite(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig, jsonResp string
	var err error

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting assetName, iht_signature, status, desc of the asset_securitized to change underwrite status"), true)
	}

	assetName = args[0]
	ihtsig = args[1]

	fmt.Println("- start startUnderwrite: assetName " + assetName)

	//check if assetName exists
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	//already in underwriting
	if AssetSecuritizedToUpdate.Asset_status == ST_sec_bein_underwriting && AssetSecuritizedToUpdate.Iht_status_comments == ST_DESC_sec_bein_underwriting {
		jsonResp = "{\"Error\":\"asset name " + assetName + " already begun underwriting, starting time " + AssetSecuritizedToUpdate.Underwrite_start_time + "\"}"
		fmt.Println(jsonResp)
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//current asset_status must be 02-审核通过 status
	if AssetSecuritizedToUpdate.Asset_status != ST_comm_approved {
		jsonResp = "{\"Error\":\"asset_securitized not in approved status, cannot startUnderwrite: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = ST_sec_bein_underwriting
	AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_bein_underwriting
	startTime := time.Now().Local()
	dueTime := startTime.AddDate(0, 0, AssetSecuritizedToUpdate.Asset_underwrite_period)
	repurchaseTime := dueTime.AddDate(0, 0, AssetSecuritizedToUpdate.Asset_repurchase_period)
	fmt.Println("start time %s, due time %s, repurchase time %s", GetTmStr(startTime, "y-m-d H:i:s"), GetTmStr(dueTime, "y-m-d H:i:s"), GetTmStr(repurchaseTime, "y-m-d H:i:s"))
	AssetSecuritizedToUpdate.Underwrite_start_time = GetRFC3339TimeStr(startTime)
	AssetSecuritizedToUpdate.Underwrite_due_time = GetRFC3339TimeStr(dueTime)
	AssetSecuritizedToUpdate.Repurchase_due_time = GetRFC3339TimeStr(repurchaseTime)
	AssetSecuritizedToUpdate.Current_underwrite_percent = 0.0000

	AssetSecuritizedJSONasBytes, _ := json.Marshal(AssetSecuritizedToUpdate)
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end startUnderwrite (success)")
	return successPbResponse(nil)
}

// ===============================================
// haltUnderwrite
// ===============================================
func (t *AssetSecuritizedMngt) haltUnderwrite(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig, jsonResp string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name and iht_signature to halt underwriting asset securitization"), true)
	}

	assetName = args[0]
	ihtsig = args[1]

	fmt.Println("- start haltUnderwrite: assetName " + assetName)

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	//there must be changes
	if AssetSecuritizedToUpdate.Asset_status == ST_sec_halt_underwriting && AssetSecuritizedToUpdate.Iht_status_comments == ST_DESC_sec_halt_underwriting {
		fmt.Println("status " + ST_sec_halt_underwriting + " , desc " + ST_DESC_sec_halt_underwriting + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}

	//current should be in underwriting status
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_bein_underwriting {
		jsonResp = "{\"Error\":\"asset_securitized not in underwriting status, cannot halt, asset_name: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = ST_sec_halt_underwriting
	AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_halt_underwriting

	AssetSecuritizedJSONasBytes, _ := json.Marshal(AssetSecuritizedToUpdate)
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end haltUnderwrite (success)")
	return successPbResponse(nil)
}

// ===============================================
// resumeUnderwrite
// ===============================================
func (t *AssetSecuritizedMngt) resumeUnderwrite(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig, jsonResp string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name and iht_signature to resume underwriting asset securitization"), true)
	}

	assetName = args[0]
	ihtsig = args[1]

	fmt.Println("- start resumeUnderwrite: assetName " + assetName)

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//there must be changes
	if AssetSecuritizedToUpdate.Asset_status == ST_sec_bein_underwriting && AssetSecuritizedToUpdate.Iht_status_comments == ST_DESC_sec_bein_underwriting {
		fmt.Println("status " + ST_sec_bein_underwriting + " , desc " + ST_DESC_sec_bein_underwriting + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}
	//current should be in halt underwriting status
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_halt_underwriting {
		jsonResp = "{\"Error\":\"asset_securitized not in halt underwriting status, cannot resume, asset_name: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = ST_sec_bein_underwriting
	AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_bein_underwriting

	AssetSecuritizedJSONasBytes, _ := json.Marshal(AssetSecuritizedToUpdate)
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end resumeUnderwrite (success)")
	return successPbResponse(nil)
}

// ===============================================
// stopUnderwriteAndCalculateResults
// ===============================================
func (t *AssetSecuritizedMngt) stopUnderwriteAndCalculateResults(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig, jsonResp string
	var ignoreDueTime bool
	var err error

	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name, isDueTimeIgnore(bool), iht_signature to halt underwriting asset securitization"), true)
	}

	assetName = args[0]
	ignoreDueTime, err = strconv.ParseBool(args[1])
	if err != nil {
		return errorPbResponse(errors.New("ignoreDueTime 2nd argument must be a valid bool string"), true)
	}
	ihtsig = args[2]

	fmt.Println("- start stopUnderwriteAndCalculateResults: assetName " + assetName)

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//current should be in underwriting or halt underwriting status
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_bein_underwriting && AssetSecuritizedToUpdate.Asset_status != ST_sec_halt_underwriting {
		jsonResp = "{\"Error\":\"asset_securitized not in underwriting or halt underwriting status, current status: " + AssetSecuritizedToUpdate.Asset_status + "-" + AssetSecuritizedToUpdate.Iht_status_comments + ", cannot stop underwriting: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig

	//ignoreDueTime?
	currentTime := time.Now().Local()
	if !ignoreDueTime {
		dueTime, er := ParseRFC3339TimeStr(AssetSecuritizedToUpdate.Underwrite_due_time)
		if er != nil {
			jsonResp = "{\"Error\":\"asset_securitized underwrite_due_time parse error, underWriteDueTime: " + AssetSecuritizedToUpdate.Underwrite_due_time + ", asset_name " + assetName + ", error " + er.Error() + "\"}"
			return errorPbResponse(errors.New(jsonResp), true)
		}
		if dueTime.After(currentTime) {
			jsonResp = "{\"Error\":\"asset_securitized underwrite_due_time not reach, underWriteDueTime: " + AssetSecuritizedToUpdate.Underwrite_due_time + ", asset_name " + assetName + ", try using isDueTimeIgnore(true) or just wait till underwrite due time\"}"
			return errorPbResponse(errors.New(jsonResp), true)
		}
	}

	//TODO read all asset_security_underwrote and added up together, not just using Current_underwrite_percent of asset_securitized
	if AssetSecuritizedToUpdate.Current_underwrite_percent < AssetSecuritizedToUpdate.Asset_underwrite_bottom_percent {
		AssetSecuritizedToUpdate.Asset_status = ST_sec_fail_underwriting
		AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_fail_underwriting
	} else {
		AssetSecuritizedToUpdate.Asset_status = ST_sec_succ_underwriting
		AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_succ_underwriting
	}

	AssetSecuritizedJSONasBytes, _ := json.Marshal(AssetSecuritizedToUpdate)
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end stopUnderwriteAndCalculateResults (success), result: " + AssetSecuritizedToUpdate.Iht_status_comments)
	return successPbResponse(nil)
}

// ===============================================
// underwrite
// ===============================================
func (t *AssetSecuritizedMngt) underwrite(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var underwroteId, fininstEmail, assetName, fininstsig, jsonResp string
	var underwroteUnits int

	if len(args) != 5 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting underwrote_id, fininst_email, asset_owner_email, asset_name, underwrote_units, fininst_signature to underwrite asset securitization"), true)
	}

	underwroteId = args[0]
	fininstEmail = args[1]
	assetName = args[2]
	underwroteUnits, er := strconv.Atoi(args[3])
	if er != nil {
		return errorPbResponse(errors.New("underwrote_units 4th argument must be a valid int string"), true)
	}
	fininstsig = args[4]
	currentTime := time.Now().Local()
	fmt.Println("- start underwrite: assetName " + assetName + " for fininst " + fininstEmail + ", claiming units " + strconv.Itoa(underwroteUnits))

	//fininst email must exist and fininst in 02-approved status
	FininstAsBytes, err := getDocWithNamespace(stub, NS_fininst, fininstEmail)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get " + DT_fininst + " doc for fininst_email " + fininstEmail + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if FininstAsBytes == nil {
		jsonResp = "{\"Error\":\"" + DT_fininst + " does not exist: " + fininstEmail + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}
	fininst := fininst{}
	err = json.Unmarshal(FininstAsBytes, &fininst)
	if err != nil {
		msg := "fininst exists: " + fininstEmail + ", but failed to be unmarshalled as an fininst json"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if ST_comm_approved != fininst.Fininst_status {
		msg := "fininst exists: " + fininstEmail + ", but current Fininst_status is " + fininst.Fininst_status + ", status 02-approved required"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	//fininst_detail email must exist and fininst_detail in 02-approved status
	FininstDetailAsBytes, err := getDocWithNamespace(stub, NS_fininst_detail, fininstEmail)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get " + DT_fininst_detail + " doc for fininst_email " + fininstEmail + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if FininstDetailAsBytes == nil {
		jsonResp = "{\"Error\":\"" + DT_fininst_detail + " does not exist: " + fininstEmail + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}
	fininstDetail := fininst_detail{}
	err = json.Unmarshal(FininstDetailAsBytes, &fininstDetail)
	if err != nil {
		msg := "fininst_detail exists: " + fininstEmail + ", but failed to be unmarshalled as an fininst_detail json"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if ST_comm_approved != fininstDetail.Detail_status {
		msg := "fininst_detail exists: " + fininstEmail + ", but current Detail_status is " + fininstDetail.Detail_status + ", status 02-approved required"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	//asset_name must exist and in underwriting status
	AssetSecuritizedAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get " + DT_asset_securitized + " doc for asset_name" + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if AssetSecuritizedAsbytes == nil {
		jsonResp = "{\"Error\":\"" + DT_asset_securitized + " does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(AssetSecuritizedAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//current should be in underwriting status
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_bein_underwriting {
		jsonResp = "{\"Error\":\"asset_securitized not in underwriting status, cannot execute underwrite: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	underwrote_id := underwroteId
	fininst_email := fininstEmail
	asset_owner_email := AssetSecuritizedToUpdate.Owner_email
	asset_name := assetName
	underwrote_units := underwroteUnits
	underwrote_time := GetTmStr(currentTime, "y-m-d H:i:s")
	fininst_signature := fininstsig
	//TODO check fininstsig

	// ==== Check if asset_security_underwrote already exists ====
	AssetSecurityUnderwroteAsBytes, err := getDocWithNamespace(stub, NS_asset_security_underwrote, underwrote_id)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_security_underwrote: "+err.Error()), true)
	} else if AssetSecurityUnderwroteAsBytes != nil {
		fmt.Println("This asset_security_underwrote already exists: " + underwrote_id)
		return errorPbResponse(errors.New("This asset_security_underwrote already exists: "+underwrote_id), true)
	}

	// ================ BUSINESS RULES start ================
	// ==== rule1: still within underwrite due time
	dueTime, er := ParseRFC3339TimeStr(AssetSecuritizedToUpdate.Underwrite_due_time)
	if er != nil {
		jsonResp = "{\"Error\":\"asset_securitized underwrite_due_time parse error, underWriteDueTime: " + AssetSecuritizedToUpdate.Underwrite_due_time + ", asset_name " + assetName + ", error " + er.Error() + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}
	if dueTime.Before(currentTime) {
		jsonResp = "{\"Error\":\"asset_securitized underwrite_due_time reached, underWriteDueTime: " + AssetSecuritizedToUpdate.Underwrite_due_time + ",currentTime: " + GetRFC3339TimeStr(currentTime) + ", asset_name " + assetName + ", underwriting not allowed anymore\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	// ==== rule2: Check if current underwrote exceeds 100% if added
	updatedTotal := underwrote_units + int(AssetSecuritizedToUpdate.Current_underwrite_percent*(float64(AssetSecuritizedToUpdate.Asset_securitized_parts)))

	if updatedTotal > AssetSecuritizedToUpdate.Asset_securitized_parts {
		msg := "current underwrote units " + strconv.Itoa(underwrote_units) + " if added will exceed total supply: " + strconv.Itoa(AssetSecuritizedToUpdate.Asset_securitized_parts) + ", current underwrote percent reached " + strconv.FormatFloat(AssetSecuritizedToUpdate.Current_underwrite_percent, 'f', -1, 64)
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)

	}

	// ================ BUSINESS RULES end ================

	// ==== Create asset_security_underwrote object and marshal to JSON ====
	AssetSecurityUnderwrote := &asset_security_underwrote{DT_asset_security_underwrote, underwrote_id, fininst_email, asset_owner_email, asset_name, underwrote_units, underwrote_time, fininst_signature}
	AssetSecurityUnderwroteJSONasBytes, err := json.Marshal(AssetSecurityUnderwrote)
	if err != nil {
		return errorPbResponse(err, false)
	}
	// === Save asset_security_underwrote to state ===
	err = putDocWithNamespace(stub, NS_asset_security_underwrote, underwrote_id, AssetSecurityUnderwroteJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	// === create IDX_fininst_email_2_underwrote_id
	err = createCKeyWithNamespace(stub, NS_asset_security_underwrote, IDX_fininst_email_2_underwrote_id, []string{AssetSecurityUnderwrote.Fininst_email, AssetSecurityUnderwrote.Underwrote_id})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// === create IDX_asset_owner_email_2_underwrote_id
	err = createCKeyWithNamespace(stub, NS_asset_security_underwrote, IDX_asset_owner_email_2_underwrote_id, []string{AssetSecurityUnderwrote.Asset_owner_email, AssetSecurityUnderwrote.Underwrote_id})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// === create IDX_asset_name_2_underwrote_id
	err = createCKeyWithNamespace(stub, NS_asset_security_underwrote, IDX_asset_name_2_underwrote_id, []string{AssetSecurityUnderwrote.Asset_name, AssetSecurityUnderwrote.Underwrote_id})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// === update asset_securitized with Current_underwrite_percent, Current_underwrite_id
	AssetSecuritizedToUpdate.Current_underwrite_id = AssetSecurityUnderwrote.Underwrote_id
	//AssetSecuritizedToUpdate.Current_underwrite_percent = AssetSecuritizedToUpdate.Current_underwrite_percent + float64(AssetSecurityUnderwrote.Underwrote_units)/float64(AssetSecuritizedToUpdate.Asset_securitized_parts)
	count, err := t.getCurrentUnderwroteUnitsCount(stub, []string{AssetSecuritizedToUpdate.Asset_name})
	if err != nil {
		return errorPbResponse(err, false)
	}
	currentTotal := count + AssetSecurityUnderwrote.Underwrote_units
	AssetSecuritizedToUpdate.Current_underwrite_percent = float64(currentTotal) / float64(AssetSecuritizedToUpdate.Asset_securitized_parts)
	AssetSecuritizedJSONasBytes, err := json.Marshal(AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}
	fmt.Println("- end underwrite(success)")
	return successPbResponse(nil)
}

func (t *AssetSecuritizedMngt) getCurrentUnderwroteUnitsCount(stub shim.ChaincodeStubInterface, args []string) (int, error) {
	// 0
	// "asset_name"
	if len(args) != 1 {
		return 0, errors.New("Incorrect number of arguments. Expecting asset_name")
	}

	assetName := args[0]

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", DT_asset_security_underwrote, IDX_FD_asset_name, assetName)

	fmt.Printf("- getCurrentUnderwroteUnitsCount queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return 0, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var count int
	count = 0

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return 0, err
		}
		AssetSecurityUnderwrote := asset_security_underwrote{}
		err = json.Unmarshal(queryResponse.Value, &AssetSecurityUnderwrote)
		if err != nil {
			return 0, err
		}

		count = count + AssetSecurityUnderwrote.Underwrote_units

	}

	fmt.Printf("- getCurrentUnderwroteUnitsCount count: %d", count)

	return count, nil
}

// ===============================================
// readUnderwrote
// ===============================================
func (t *AssetSecuritizedMngt) readUnderwrote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var underwroteId, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting underwrote_id"), true)
	}

	underwroteId = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_security_underwrote, underwroteId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get" + DT_asset_security_underwrote + " doc for " + underwroteId + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_security_underwrote does not exist: " + underwroteId + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// repurchase
// ===============================================
func (t *AssetSecuritizedMngt) repurchase(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig, jsonResp string
	var ignoreDueTime bool
	var err error

	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name, isDueTimeIgnore(bool), iht_signature to change repurchase status"), true)
	}

	assetName = args[0]
	ignoreDueTime, err = strconv.ParseBool(args[1])
	if err != nil {
		return errorPbResponse(errors.New("ignoreDueTime 2nd argument must be a valid bool string"), true)
	}
	ihtsig = args[2]

	fmt.Println("- start startUnderwrite: assetName " + assetName)

	//check if assetName exists
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName) //get the asset_securitized from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_securitized does not exist: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(valAsbytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	//already in underwriting
	if AssetSecuritizedToUpdate.Asset_status == ST_sec_finish_repurchasing && AssetSecuritizedToUpdate.Iht_status_comments == ST_DESC_sec_finish_repurchasing {
		jsonResp = "{\"Error\":\"asset name " + assetName + " already begun repurchased }"
		fmt.Println(jsonResp)
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//current asset_status must be 21-权益转移生效 status  ST_sec_effect_ownership_change
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_effect_ownership_change {
		jsonResp = "{\"Error\":\"asset_securitized not in effect_ownership status, cannot repurchase: " + assetName + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = ST_sec_finish_repurchasing
	AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_finish_repurchasing

	//ignoreDueTime?
	currentTime := time.Now().Local()
	if !ignoreDueTime {
		dueTime, er := ParseRFC3339TimeStr(AssetSecuritizedToUpdate.Repurchase_due_time)
		if er != nil {
			jsonResp = "{\"Error\":\"asset_securitized Repurchase_due_time parse error, underWriteDueTime: " + AssetSecuritizedToUpdate.Repurchase_due_time + ", asset_name " + assetName + ", error " + er.Error() + "\"}"
			return errorPbResponse(errors.New(jsonResp), true)
		}
		if dueTime.After(currentTime) {
			jsonResp = "{\"Error\":\"asset_securitized Repurchase_due_time not reach, underWriteDueTime: " + AssetSecuritizedToUpdate.Repurchase_due_time + ", asset_name " + assetName + ", try using isDueTimeIgnore(true) or just wait till underwrite due time\"}"
			return errorPbResponse(errors.New(jsonResp), true)
		}
	}

	AssetSecuritizedJSONasBytes, _ := json.Marshal(AssetSecuritizedToUpdate)
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end startUnderwrite (success)")
	return successPbResponse(nil)
}

// ===============================================
// queryUnderwroteByFininstEmail
// ===============================================
func (t *AssetSecuritizedMngt) queryUnderwrotesByFininstEmail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "fininst_email"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	fininstEmail := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_security_underwrote, IDX_FD_fininst_email, fininstEmail)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryUnderwroteByAssetOwnerEmail
// ===============================================
func (t *AssetSecuritizedMngt) queryUnderwrotesByAssetOwnerEmail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "asset_owner_email"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_owner_email"), true)
	}

	assetOwnerEmail := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_security_underwrote, IDX_FD_asset_owner_email, assetOwnerEmail)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryUnderwroteByAssetName
// ===============================================
func (t *AssetSecuritizedMngt) queryUnderwrotesByAssetName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "asset_name"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name"), true)
	}

	assetName := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_security_underwrote, IDX_FD_asset_name, assetName)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}
