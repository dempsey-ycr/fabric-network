package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	//"strconv"
	//"time"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	DT_asset_owner                  string = "asset_owner"
	DT_asset_owner_detail           string = "asset_owner_detail"
	NS_asset_owner                  string = DT_asset_owner + "_"
	NS_asset_owner_detail           string = DT_asset_owner_detail + "_"
	PK_FD_asset_owner               string = "owner_email"
	IDX_FD_owner_status                    = "owner_status"
	IDX_FD_detail_status                   = "detail_status"
	IDX_owner_status_2_owner_email  string = IDX_FD_owner_status + "_2_" + PK_FD_asset_owner
	IDX_detail_status_2_owner_email string = IDX_FD_detail_status + "_2_" + PK_FD_asset_owner
)

type AssetOwnerMngt struct{}

type asset_owner struct {
	ObjectType           string `json:"docType"`              //docType is used to distinguish the various types of objects in state database
	Owner_email          string `json:"owner_email"`          //邮箱
	Owner_nickname       string `json:"owner_nickname"`       //昵称
	Owner_pwd_hash       string `json:"owner_pwd_hash"`       //密码hash值
	Owner_status         string `json:"owner_status"`         //资产方当前状态：00-init 01-正在审核  02-审核通过  03-审核不通过  99-作废
	Iht_approve_comments string `json:"iht_approve_comments"` //IHT平台审核资产方结果
}

type asset_owner_detail struct {
	ObjectType                  string `json:"docType"`                  //docType is used to distinguish the various types of objects in state database
	Owner_email                 string `json:"owner_email"`              //邮箱
	Owner_realname              string `json:"owner_realname"`           //企业或个人真实名称
	Owner_contact               string `json:"owner_contact"`            //联系方式
	Owner_identity_pic1_hash    string `json:"owner_identity_pic1_hash"` //营业执照或身份证照片hash
	Owner_identity_pic2_hash    string `json:"owner_identity_pic2_hash"`
	Owner_identity_pic3_hash    string `json:"owner_identity_pic3_hash"`
	Owner_asset_proof_pic1_hash string `json:"owner_asset_proof_pic1_hash"` //资产证明及经营状况照片hash
	Owner_asset_proof_pic2_hash string `json:"owner_asset_proof_pic2_hash"`
	Owner_asset_proof_pic3_hash string `json:"owner_asset_proof_pic3_hash"`
	Owner_asset_proof_pic4_hash string `json:"owner_asset_proof_pic4_hash"`
	Detail_status               string `json:"detail_status"`        //资产方明细信息当前状态：00-init 01-正在审核  02-审核通过  03-审核不通过  99-作废
	Iht_approve_comments        string `json:"iht_approve_comments"` //IHT平台审核资产方结果
	Iht_signature               string `json:"iht_signature"`        //IHT平台对本asset_owner_detail的签名（保留字段，目前送“”）
}

// ============================================================
// initAssetOwner - create a new asset_owner_mngt, store into chaincode state
// ============================================================
func (t *AssetOwnerMngt) initAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 0                       1                  2
	// "rodger0514@163.com",   "owner_nickname1", "owner_pwd_hash"
	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 3"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start init asset_owner")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 0 {
		return errorPbResponse(errors.New("3rd argument must be a non-empty string"), true)
	}

	owner_email := args[0]
	owner_nickname := args[1]
	owner_pwd_hash := args[2]

	// ==== Check if asset_owner already exists ====
	AssetOwnerAsBytes, err := getDocWithNamespace(stub, NS_asset_owner, owner_email)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_owner: "+err.Error()), true)
	} else if AssetOwnerAsBytes != nil {
		fmt.Println("This asset_owner already exists: " + owner_email)
		return errorPbResponse(errors.New("This asset_owner already exists: "+owner_email), true)
	}

	// ==== Create asset_owner_mngt object and marshal to JSON ====
	AssetOwner := &asset_owner{DT_asset_owner, owner_email, owner_nickname, owner_pwd_hash, ST_comm_init, ST_DESC_comm_init}
	AssetOwnerJSONasBytes, err := json.Marshal(AssetOwner)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//Alternatively, build the asset_owner json string manually if you don't want to use struct marshalling
	//AssetOwnerJSONasString := `{"docType":"asset_owner_mngt",  "Trade_no": "` + trade_no + `", "Mer_id": "` + mer_id + `", "Card_no": ` + card_no + `, "Trade_amt": "` + trade_amt + `"}`
	//AssetOwnerJSONasBytes := []byte(AssetOwnerJSONasString)

	// === Save asset_owner to state ===
	err = putDocWithNamespace(stub, NS_asset_owner, owner_email, AssetOwnerJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	//create IDX_owner_status_2_owner_email
	err = createCKeyWithNamespace(stub, NS_asset_owner, IDX_owner_status_2_owner_email, []string{AssetOwner.Owner_status, AssetOwner.Owner_email})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== asset_owner_mngt saved and indexed. Return success ====
	fmt.Println("- end init asset_owner")
	return successPbResponse(nil)
}

// ============================================================
// initAssetOwner - create a new asset_owner_mngt, store into chaincode state
// ============================================================
func (t *AssetOwnerMngt) initAssetOwnerDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 0                       1                  2                 3                           4                           5                           6                              7                              8                              9
	// "rodger0514@163.com",   "owner_realname1", "	owner_contact", "owner_identity_pic1_hash", "owner_identity_pic2_hash", "owner_identity_pic3_hash", "owner_asset_proof_pic1_hash", "owner_asset_proof_pic2_hash", "owner_asset_proof_pic3_hash", "owner_asset_proof_pic4_hash"
	if len(args) != 10 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 10"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start init asset_owner_detail")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 0 {
		return errorPbResponse(errors.New("3rd argument must be a non-empty string"), true)
	}
	if len(args[3]) <= 0 && len(args[4]) <= 0 && len(args[5]) <= 0 {
		return errorPbResponse(errors.New("4th 5th 6th argument must have at least 1 non-empty string"), true)
	}
	if len(args[6]) <= 0 && len(args[7]) <= 0 && len(args[8]) <= 0 && len(args[9]) <= 0 {
		return errorPbResponse(errors.New("7th 8th 9th 10th argument must have at least 1 non-empty string"), true)
	}

	Owner_email := args[0]
	Owner_realname := args[1]
	Owner_contact := args[2]
	Owner_identity_pic1_hash := args[3]
	Owner_identity_pic2_hash := args[4]
	Owner_identity_pic3_hash := args[5]
	Owner_asset_proof_pic1_hash := args[6]
	Owner_asset_proof_pic2_hash := args[7]
	Owner_asset_proof_pic3_hash := args[8]
	Owner_asset_proof_pic4_hash := args[9]
	Iht_signature := ""
	// ==== Check if asset_owner already exists and status is 02 ====
	AssetOwnerAsBytes, err := getDocWithNamespace(stub, NS_asset_owner, Owner_email)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_owner: "+Owner_email+", "+err.Error()), true)
	}
	if AssetOwnerAsBytes == nil {
		return errorPbResponse(errors.New("Failed to get asset_owner: "+Owner_email+", asset_owner not found"), true)
	}
	assetOwner := asset_owner{}
	err = json.Unmarshal(AssetOwnerAsBytes, &assetOwner)
	if err != nil {
		msg := "asset_owner exists: " + Owner_email + ", but failed to be unmarshalled as an asset_owner json"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if ST_comm_approved != assetOwner.Owner_status {
		msg := "asset_owner exists: " + Owner_email + ", but current Owner_status is " + assetOwner.Owner_status + ", status 02-approved required"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	// ==== Check if asset_owner_datail already exists ====
	AssetOwnerDetailAsBytes, err2 := getDocWithNamespace(stub, NS_asset_owner_detail, Owner_email)
	if err2 != nil {
		return errorPbResponse(errors.New("Failed to get asset_owner_detail: "+err.Error()), true)
	} else if AssetOwnerDetailAsBytes != nil {
		fmt.Println("This asset_owner_detail already exists: " + Owner_email)
		return errorPbResponse(errors.New("This asset_owner_detail already exists: "+Owner_email), true)
	}

	// ==== Create asset_owner_mngt object and marshal to JSON ====
	AssetOwnerDetail := &asset_owner_detail{DT_asset_owner_detail, Owner_email, Owner_realname, Owner_contact, Owner_identity_pic1_hash, Owner_identity_pic2_hash, Owner_identity_pic3_hash, Owner_asset_proof_pic1_hash, Owner_asset_proof_pic2_hash, Owner_asset_proof_pic3_hash, Owner_asset_proof_pic4_hash, ST_comm_init, ST_DESC_comm_init, Iht_signature}
	AssetOwnerDetailJSONasBytes, err := json.Marshal(AssetOwnerDetail)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//Alternatively, build the asset_owner json string manually if you don't want to use struct marshalling
	//AssetOwnerJSONasString := `{"docType":"asset_owner_mngt",  "Trade_no": "` + trade_no + `", "Mer_id": "` + mer_id + `", "Card_no": ` + card_no + `, "Trade_amt": "` + trade_amt + `"}`
	//AssetOwnerJSONasBytes := []byte(AssetOwnerJSONasString)

	// === Save asset_owner_mngt to state ===
	err = putDocWithNamespace(stub, NS_asset_owner_detail, Owner_email, AssetOwnerDetailJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	//create IDX_detail_status_2_owner_email
	err = createCKeyWithNamespace(stub, NS_asset_owner_detail, IDX_detail_status_2_owner_email, []string{AssetOwnerDetail.Detail_status, AssetOwnerDetail.Owner_email})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== asset_owner_mngt saved and indexed. Return success ====
	fmt.Println("- end init asset_owner_detail")
	return successPbResponse(nil)
}

// ===============================================
// readAssetOwner - read a asset_owner_mngt from chaincode state
// ===============================================
func (t *AssetOwnerMngt) readAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email"), true)
	}

	email = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner, email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// readAssetOwner - read a asset_owner_mngt from chaincode state
// ===============================================
func (t *AssetOwnerMngt) readAssetOwnerDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email"), true)
	}

	email = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner_detail, email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner_detail + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner_detail does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// queryAssetOwnersByStatus - read a asset_owner_mngt from chaincode state
// ===============================================
func (t *AssetOwnerMngt) queryAssetOwnersByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "status_01"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_status"), true)
	}

	ownerStatus := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_owner, IDX_FD_owner_status, ownerStatus)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryAssetOwnersByStatus - read a asset_owner_mngt from chaincode state
// ===============================================
func (t *AssetOwnerMngt) queryAssetOwnerDetailsByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "status_01"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting detail_status"), true)
	}

	ownerDetailStatus := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_owner_detail, IDX_FD_detail_status, ownerDetailStatus)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// approveAssetOwner - read a asset_owner_mngt from chaincode state
// ===============================================
func (t *AssetOwnerMngt) approveAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var email string

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email"), true)
	}

	email = args[0]

	return t.changeAssetOwnerStatus(stub, []string{email, ST_comm_approved, ST_DESC_comm_approved})

}

// ===============================================
// approveAssetOwnerDetail - read a asset_owner_mngt from chaincode state
// ===============================================
func (t *AssetOwnerMngt) approveAssetOwnerDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, ihtsig string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email, iht_signature"), true)
	}

	email = args[0]
	ihtsig = args[1]

	return t.changeAssetOwnerDetailStatus(stub, []string{email, ihtsig, ST_comm_approved, ST_DESC_comm_approved})
}

// ===============================================
// rejectAssetOwner - reject a asset_owner as "03" status
// ===============================================
func (t *AssetOwnerMngt) rejectAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var email string

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting email of the asset_owner"), true)
	}

	email = args[0]

	return t.changeAssetOwnerStatus(stub, []string{email, ST_comm_rejected, ST_DESC_comm_rejected})

}

// ===============================================
// rejectAssetOwnerDetail - reject a asset_owner_detail as "03" status
// ===============================================
func (t *AssetOwnerMngt) rejectAssetOwnerDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, ihtsig string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email, iht_signature"), true)
	}

	email = args[0]
	ihtsig = args[1]
	return t.changeAssetOwnerDetailStatus(stub, []string{email, ihtsig, ST_comm_rejected, ST_DESC_comm_rejected})
}

// ==================================================
// delete - delete a asset_owner key/value pair from state
// ==================================================
func (t *AssetOwnerMngt) deleteAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var jsonResp string
	//var AssetOwnerJSON asset_owner
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 1"), true)
	}
	email := args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner, email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	assetOwnerToUpdate := asset_owner{}
	err = json.Unmarshal(valAsbytes, &assetOwnerToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if ST_comm_niled == assetOwnerToUpdate.Owner_status {
		jsonResp = "{\"Info\":\"" + email + "'s asset_owner was already deleted!\"}"
		fmt.Println("- end delete asset_owner (success) " + jsonResp)
	} else {

		assetOwnerToUpdate.Owner_status = ST_comm_niled //change the owner status to "02" as approve status
		assetOwnerToUpdate.Iht_approve_comments = ST_DESC_comm_niled

		assetOwnerJSONasBytes, _ := json.Marshal(assetOwnerToUpdate)
		err = putDocWithNamespace(stub, NS_asset_owner, email, assetOwnerJSONasBytes)
		if err != nil {
			return errorPbResponse(err, false)
		}

		fmt.Println("- end delete asset_owner (success)")
	}

	valAsbytes2, err := getDocWithNamespace(stub, NS_asset_owner_detail, email) //get the asset_owner_detail from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner_detail + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes2 == nil {
		jsonResp = "{\"Error\":\"asset_owner_detail does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	assetOwnerDetailToUpdate := asset_owner_detail{}
	err = json.Unmarshal(valAsbytes2, &assetOwnerDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	if ST_comm_niled == assetOwnerDetailToUpdate.Detail_status {
		jsonResp = "{\"Info\":\"" + email + "'s asset_owner_detail was already deleted!\"}"
		fmt.Println("- end delete asset_owner_detail (success) " + jsonResp)
	} else {

		assetOwnerDetailToUpdate.Detail_status = ST_comm_niled //change the owner status to "02" as approve status
		assetOwnerDetailToUpdate.Iht_approve_comments = ST_DESC_comm_niled

		assetOwnerDetailJSONasBytes, _ := json.Marshal(assetOwnerDetailToUpdate)
		err = putDocWithNamespace(stub, NS_asset_owner_detail, email, assetOwnerDetailJSONasBytes) //rewrite the asset owner detail
		if err != nil {
			return errorPbResponse(err, false)
		}

		fmt.Println("- end delete asset_owner_detail (success)")
	}

	//TODO, delete 资产分拆信息

	return successPbResponse(nil)
}

func (t *AssetOwnerMngt) changeAssetOwnerStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, status, desc, jsonResp string
	var err error

	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting email, status, desc of the asset_owner to change status"), true)
	}

	email = args[0]
	status = args[1]
	desc = args[2]

	fmt.Println("- start changeAssetOwnerStatus: email " + email + " , status " + status + " , desc " + desc)

	if status != ST_comm_approving && status != ST_comm_approved && status != ST_comm_rejected && status != ST_comm_niled {
		return errorPbResponse(errors.New("Incorrect changed status. Expecting 01-正在审核  02-审核通过  03-审核不通过  99-作废 for asset_owner to change status"), true)
	}

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner, email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	assetOwnerToUpdate := asset_owner{}
	err = json.Unmarshal(valAsbytes, &assetOwnerToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if assetOwnerToUpdate.Owner_status == status && assetOwnerToUpdate.Iht_approve_comments == desc {
		fmt.Println("status " + status + " , desc " + desc + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}
	assetOwnerToUpdate.Owner_status = status
	assetOwnerToUpdate.Iht_approve_comments = desc

	assetOwnerJSONasBytes, err := json.Marshal(assetOwnerToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_asset_owner, email, assetOwnerJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeAssetOwnerStatus (success)")
	return successPbResponse(nil)
}

func (t *AssetOwnerMngt) changeAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	// 0                       1                  2
	// "rodger0514@163.com",   "owner_nickname1", "owner_pwd_hash"
	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 3"), true)
	}

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 0 {
		return errorPbResponse(errors.New("3rd argument must be a non-empty string"), true)
	}

	owner_email := args[0]
	owner_nickname := args[1]
	owner_pwd_hash := args[2]

	fmt.Println("- start changeAssetOwner: owner_email " + owner_email + " , owner_nickname " + owner_nickname + " , owner_pwd_hash " + owner_pwd_hash)

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner, owner_email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner + owner_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner does not exist: " + owner_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	assetOwnerToUpdate := asset_owner{}
	err = json.Unmarshal(valAsbytes, &assetOwnerToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	var isChanged bool = false

	if assetOwnerToUpdate.Owner_nickname != owner_nickname {
		assetOwnerToUpdate.Owner_nickname = owner_nickname
		isChanged = true
	}
	if assetOwnerToUpdate.Owner_pwd_hash != owner_pwd_hash {
		assetOwnerToUpdate.Owner_pwd_hash = owner_pwd_hash
		isChanged = true
	}

	if !isChanged {
		fmt.Println("- end changeAssetOwner (no change no commit)")
		return successPbResponse(nil)
	}

	//reset status to 00-init if isChanged
	if assetOwnerToUpdate.Owner_status == ST_comm_rejected {
		assetOwnerToUpdate.Owner_status = ST_comm_init
		assetOwnerToUpdate.Iht_approve_comments = ST_DESC_comm_init_change
	}

	assetOwnerJSONasBytes, err := json.Marshal(assetOwnerToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_asset_owner, owner_email, assetOwnerJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeAssetOwner (success)")
	return successPbResponse(nil)
}

func (t *AssetOwnerMngt) changeAssetOwnerDetailStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, ihtsig, status, desc, jsonResp string
	var err error

	if len(args) != 4 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting email, iht_signature, status, desc of the asset_owner_detail to change status"), true)
	}

	email = args[0]
	ihtsig = args[1]
	status = args[2]
	desc = args[3]

	fmt.Println("- start changeAssetOwnerDetailStatus: email " + email + " , status " + status + " , desc " + desc)

	if status != ST_comm_approving && status != ST_comm_approved && status != ST_comm_rejected && status != ST_comm_niled {
		return errorPbResponse(errors.New("Incorrect changed status. Expecting 01-正在审核  02-审核通过  03-审核不通过  99-作废 for asset_owner to change status"), true)
	}

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner_detail, email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner_detail + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner_detail does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	assetOwnerDetailToUpdate := asset_owner_detail{}
	err = json.Unmarshal(valAsbytes, &assetOwnerDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	if assetOwnerDetailToUpdate.Detail_status == status && assetOwnerDetailToUpdate.Iht_approve_comments == desc {
		fmt.Println("status " + status + " , desc " + desc + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}
	assetOwnerDetailToUpdate.Iht_signature = ihtsig
	assetOwnerDetailToUpdate.Detail_status = status
	assetOwnerDetailToUpdate.Iht_approve_comments = desc

	assetOwnerDetailJSONasBytes, err := json.Marshal(assetOwnerDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_asset_owner_detail, email, assetOwnerDetailJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeAssetOwnerDetailStatus (success)")
	return successPbResponse(nil)

}

func (t *AssetOwnerMngt) changeAssetOwnerDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	if len(args) != 10 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 10"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start changeAssetOwnerDetail")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 0 {
		return errorPbResponse(errors.New("3rd argument must be a non-empty string"), true)
	}
	if len(args[3]) <= 0 && len(args[4]) <= 0 && len(args[5]) <= 0 {
		return errorPbResponse(errors.New("4th 5th 6th argument must have at least 1 non-empty string"), true)
	}
	if len(args[6]) <= 0 && len(args[7]) <= 0 && len(args[8]) <= 0 && len(args[9]) <= 0 {
		return errorPbResponse(errors.New("7th 8th 9th 10th argument must have at least 1 non-empty string"), true)
	}

	Owner_email := args[0]
	Owner_realname := args[1]
	Owner_contact := args[2]
	Owner_identity_pic1_hash := args[3]
	Owner_identity_pic2_hash := args[4]
	Owner_identity_pic3_hash := args[5]
	Owner_asset_proof_pic1_hash := args[6]
	Owner_asset_proof_pic2_hash := args[7]
	Owner_asset_proof_pic3_hash := args[8]
	Owner_asset_proof_pic4_hash := args[9]

	valAsbytes, err := getDocWithNamespace(stub, NS_asset_owner_detail, Owner_email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_asset_owner_detail + Owner_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_owner_detail does not exist: " + Owner_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	assetOwnerDetailToUpdate := asset_owner_detail{}
	err = json.Unmarshal(valAsbytes, &assetOwnerDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	if assetOwnerDetailToUpdate.Detail_status != ST_comm_init && assetOwnerDetailToUpdate.Detail_status != ST_comm_rejected {
		msg := "current asset_owner_detail status " + assetOwnerDetailToUpdate.Detail_status + "-" + assetOwnerDetailToUpdate.Iht_approve_comments + " not allowed to change asset_owner_detail, only 00-init or 03-IHT rejected allowed"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	var isChanged bool = false

	if assetOwnerDetailToUpdate.Owner_realname != Owner_realname {
		assetOwnerDetailToUpdate.Owner_realname = Owner_realname
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_contact != Owner_contact {
		assetOwnerDetailToUpdate.Owner_contact = Owner_contact
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_identity_pic1_hash != Owner_identity_pic1_hash {
		assetOwnerDetailToUpdate.Owner_identity_pic1_hash = Owner_identity_pic1_hash
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_identity_pic2_hash != Owner_identity_pic2_hash {
		assetOwnerDetailToUpdate.Owner_identity_pic2_hash = Owner_identity_pic2_hash
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_identity_pic3_hash != Owner_identity_pic3_hash {
		assetOwnerDetailToUpdate.Owner_identity_pic3_hash = Owner_identity_pic3_hash
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_asset_proof_pic1_hash != Owner_asset_proof_pic1_hash {
		assetOwnerDetailToUpdate.Owner_asset_proof_pic1_hash = Owner_asset_proof_pic1_hash
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_asset_proof_pic2_hash != Owner_asset_proof_pic2_hash {
		assetOwnerDetailToUpdate.Owner_asset_proof_pic2_hash = Owner_asset_proof_pic2_hash
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_asset_proof_pic3_hash != Owner_asset_proof_pic3_hash {
		assetOwnerDetailToUpdate.Owner_asset_proof_pic3_hash = Owner_asset_proof_pic3_hash
		isChanged = true
	}
	if assetOwnerDetailToUpdate.Owner_asset_proof_pic4_hash != Owner_asset_proof_pic4_hash {
		assetOwnerDetailToUpdate.Owner_asset_proof_pic4_hash = Owner_asset_proof_pic4_hash
		isChanged = true
	}

	if !isChanged {
		fmt.Println("- end changeAssetOwnerDetail (no change no commit)")
		return successPbResponse(nil)
	}

	//reset status to 00-init if isChanged
	if assetOwnerDetailToUpdate.Detail_status == ST_comm_rejected {
		assetOwnerDetailToUpdate.Detail_status = ST_comm_init
		assetOwnerDetailToUpdate.Iht_approve_comments = ST_DESC_comm_init_change
		assetOwnerDetailToUpdate.Iht_signature = ""
	}

	assetOwnerDetailJSONasBytes, err := json.Marshal(assetOwnerDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_asset_owner_detail, Owner_email, assetOwnerDetailJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeAssetOwnerDetail (success)")
	return successPbResponse(nil)

}

func (t *AssetOwnerMngt) getHistoryForAssetOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email"), true)
	}

	owner_email := args[0]

	fmt.Printf("- start getHistoryForAssetOwner: %s\n", owner_email)

	bytes, err := getHistoryForDocWithNamespace(stub, NS_asset_owner, owner_email)

	if err != nil {
		return errorPbResponse(err, true)
	}

	return successPbResponse(bytes)

}

func (t *AssetOwnerMngt) getHistoryForAssetOwnerDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email"), true)
	}

	owner_email := args[0]

	fmt.Printf("- start getHistoryForAssetOwnerDetail: %s\n", owner_email)

	bytes, err := getHistoryForDocWithNamespace(stub, NS_asset_owner_detail, owner_email)

	if err != nil {
		return errorPbResponse(err, true)
	}

	return successPbResponse(bytes)

}
