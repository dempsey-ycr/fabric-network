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
	DT_fininst            string = "fininst"
	DT_fininst_detail     string = "fininst_detail"
	NS_fininst            string = DT_fininst + "_"
	NS_fininst_detail     string = DT_fininst_detail + "_"
	PK_FD_fininst         string = "fininst_email"
	IDX_FD_fininst_status        = "fininst_status"
	//IDX_FD_detail_status = "detail_status"
	IDX_fininst_status_2_fininst_email string = IDX_FD_fininst_status + "_2_" + PK_FD_fininst
	IDX_detail_status_2_fininst_email  string = IDX_FD_detail_status + "_2_" + PK_FD_fininst
)

type FininstMngt struct{}

type fininst struct {
	ObjectType           string `json:"docType"`              //docType is used to distinguish the various types of objects in state database
	Fininst_email        string `json:"fininst_email"`        //邮箱
	Fininst_nickname     string `json:"fininst_nickname"`     //昵称
	Fininst_pwd_hash     string `json:"fininst_pwd_hash"`     //密码hash值
	Fininst_status       string `json:"fininst_status"`       //资产方当前状态：00-init 01-正在审核  02-审核通过  03-审核不通过  99-作废
	Iht_approve_comments string `json:"iht_approve_comments"` //IHT平台审核资产方结果
}

type fininst_detail struct {
	ObjectType                    string `json:"docType"`                    //docType is used to distinguish the various types of objects in state database
	Fininst_email                 string `json:"fininst_email"`              //邮箱
	Fininst_realname              string `json:"fininst_realname"`           //企业名称
	Fininst_contact               string `json:"fininst_contact"`            //联系方式
	Fininst_identity_pic1_hash    string `json:"fininst_identity_pic1_hash"` //营业执照照片hash
	Fininst_identity_pic2_hash    string `json:"fininst_identity_pic2_hash"`
	Fininst_identity_pic3_hash    string `json:"fininst_identity_pic3_hash"`
	Fininst_asset_proof_pic1_hash string `json:"fininst_asset_proof_pic1_hash"` //资产证明及经营状况照片hash
	Fininst_asset_proof_pic2_hash string `json:"fininst_asset_proof_pic2_hash"`
	Fininst_asset_proof_pic3_hash string `json:"fininst_asset_proof_pic3_hash"`
	Fininst_asset_proof_pic4_hash string `json:"fininst_asset_proof_pic4_hash"`
	Detail_status                 string `json:"detail_status"`        //资产方明细信息当前状态：00-init 01-正在审核  02-审核通过  03-审核不通过  99-作废
	Iht_approve_comments          string `json:"iht_approve_comments"` //IHT平台审核资产方结果
	Iht_signature                 string `json:"iht_signature"`        //IHT平台对本fininst_detail的签名（保留字段，目前送“”）
}

// ============================================================
// initFininst - create a new fininst_mngt, store into chaincode state
// ============================================================
func (t *FininstMngt) initFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 0                       1                  2
	// "rodger0514@163.com",   "fininst_nickname1", "fininst_pwd_hash"
	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 3"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start init fininst")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}
	if len(args[1]) <= 0 {
		return errorPbResponse(errors.New("2nd argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 0 {
		return errorPbResponse(errors.New("3rd argument must be a non-empty string"), true)
	}

	fininst_email := args[0]
	fininst_nickname := args[1]
	fininst_pwd_hash := args[2]

	// ==== Check if fininst already exists ====
	FininstAsBytes, err := getDocWithNamespace(stub, NS_fininst, fininst_email)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get fininst: "+err.Error()), true)
	} else if FininstAsBytes != nil {
		fmt.Println("This fininst already exists: " + fininst_email)
		return errorPbResponse(errors.New("This fininst already exists: "+fininst_email), true)
	}

	// ==== Create fininst_mngt object and marshal to JSON ====
	Fininst := &fininst{DT_fininst, fininst_email, fininst_nickname, fininst_pwd_hash, ST_comm_init, ST_DESC_comm_init}
	FininstJSONasBytes, err := json.Marshal(Fininst)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//Alternatively, build the fininst json string manually if you don't want to use struct marshalling
	//FininstJSONasString := `{"docType":"fininst_mngt",  "Trade_no": "` + trade_no + `", "Mer_id": "` + mer_id + `", "Card_no": ` + card_no + `, "Trade_amt": "` + trade_amt + `"}`
	//FininstJSONasBytes := []byte(FininstJSONasString)

	// === Save fininst to state ===
	err = putDocWithNamespace(stub, NS_fininst, fininst_email, FininstJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	// create IDX_fininst_status_2_fininst_email
	err = createCKeyWithNamespace(stub, NS_fininst, IDX_fininst_status_2_fininst_email, []string{Fininst.Fininst_status, Fininst.Fininst_email})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== fininst_mngt saved and indexed. Return success ====
	fmt.Println("- end init fininst")
	return successPbResponse(nil)
}

// ============================================================
// initFininst - create a new fininst_mngt, store into chaincode state
// ============================================================
func (t *FininstMngt) initFininstDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 0                       1                  2                 3                           4                           5                           6                              7                              8                              9
	// "rodger0514@163.com",   "fininst_realname1", "	fininst_contact", "fininst_identity_pic1_hash", "fininst_identity_pic2_hash", "fininst_identity_pic3_hash", "fininst_asset_proof_pic1_hash", "fininst_asset_proof_pic2_hash", "fininst_asset_proof_pic3_hash", "fininst_asset_proof_pic4_hash"
	if len(args) != 10 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 10"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start init fininst_detail")
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
		return errorPbResponse(errors.New("7th 8th 9th argument must have at least 1 non-empty string"), true)
	}

	Fininst_email := args[0]
	Fininst_realname := args[1]
	Fininst_contact := args[2]
	Fininst_identity_pic1_hash := args[3]
	Fininst_identity_pic2_hash := args[4]
	Fininst_identity_pic3_hash := args[5]
	Fininst_asset_proof_pic1_hash := args[6]
	Fininst_asset_proof_pic2_hash := args[7]
	Fininst_asset_proof_pic3_hash := args[8]
	Fininst_asset_proof_pic4_hash := args[9]
	Iht_signature := ""

	// ==== Check if fininst already exists and status is 02 ====
	FininstAsBytes, err := getDocWithNamespace(stub, NS_fininst, Fininst_email)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get fininst: "+Fininst_email+", "+err.Error()), true)
	}

	if FininstAsBytes == nil {
		return errorPbResponse(errors.New("Failed to get fininst: "+Fininst_email+", fininst not found"), true)
	}
	fininst := fininst{}
	err = json.Unmarshal(FininstAsBytes, &fininst)
	if err != nil {
		msg := "fininst exists: " + Fininst_email + ", but failed to be unmarshalled as an fininst json"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	if ST_comm_approved != fininst.Fininst_status {
		msg := "fininst exists: " + Fininst_email + ", but current Fininst_status is " + fininst.Fininst_status + ", status 02-approved required"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	// ==== Check if fininst_datail already exists ====
	FininstDetailAsBytes, err2 := getDocWithNamespace(stub, NS_fininst_detail, Fininst_email)
	if err2 != nil {
		return errorPbResponse(errors.New("Failed to get fininst_detail: "+err.Error()), true)
	} else if FininstDetailAsBytes != nil {
		fmt.Println("This fininst_detail already exists: " + Fininst_email)
		return errorPbResponse(errors.New("This fininst_detail already exists: "+Fininst_email), true)
	}

	// ==== Create fininst_mngt object and marshal to JSON ====
	FininstDetail := &fininst_detail{DT_fininst_detail, Fininst_email, Fininst_realname, Fininst_contact, Fininst_identity_pic1_hash, Fininst_identity_pic2_hash, Fininst_identity_pic3_hash, Fininst_asset_proof_pic1_hash, Fininst_asset_proof_pic2_hash, Fininst_asset_proof_pic3_hash, Fininst_asset_proof_pic4_hash, ST_comm_init, ST_DESC_comm_init, Iht_signature}
	FininstDetailJSONasBytes, err := json.Marshal(FininstDetail)
	if err != nil {
		return errorPbResponse(err, false)
	}
	//Alternatively, build the fininst json string manually if you don't want to use struct marshalling
	//FininstJSONasString := `{"docType":"fininst_mngt",  "Trade_no": "` + trade_no + `", "Mer_id": "` + mer_id + `", "Card_no": ` + card_no + `, "Trade_amt": "` + trade_amt + `"}`
	//FininstJSONasBytes := []byte(FininstJSONasString)

	// === Save fininst_mngt to state ===
	err = putDocWithNamespace(stub, NS_fininst_detail, Fininst_email, FininstDetailJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	// create IDX_detail_status_2_fininst_email
	err = createCKeyWithNamespace(stub, NS_fininst_detail, IDX_detail_status_2_fininst_email, []string{FininstDetail.Detail_status, FininstDetail.Fininst_email})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== fininst_mngt saved and indexed. Return success ====
	fmt.Println("- end init fininst_detail")
	return successPbResponse(nil)
}

// ===============================================
// readFininst - read a fininst_mngt from chaincode state
// ===============================================
func (t *FininstMngt) readFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	email = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_fininst, email) //get the fininst from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// readFininst - read a fininst_mngt from chaincode state
// ===============================================
func (t *FininstMngt) readFininstDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	email = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_fininst_detail, email) //get the fininst from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst_detail + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst_detail does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// queryFininstsByStatus - read a fininst_mngt from chaincode state
// ===============================================
func (t *FininstMngt) queryFininstsByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "status_01"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting detail_status"), true)
	}

	fininstStatus := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_fininst, IDX_FD_fininst_status, fininstStatus)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryFininstsByStatus - read a fininst_mngt from chaincode state
// ===============================================
func (t *FininstMngt) queryFininstDetailsByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "status_01"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting detail_status"), true)
	}

	fininstDetailStatus := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_fininst_detail, IDX_FD_detail_status, fininstDetailStatus)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// approveFininst - read a fininst_mngt from chaincode state
// ===============================================
func (t *FininstMngt) approveFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var email string

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	email = args[0]

	return t.changeFininstStatus(stub, []string{email, ST_comm_approved, ST_DESC_comm_approved})

}

// ===============================================
// approveFininstDetail - read a fininst_mngt from chaincode state
// ===============================================
func (t *FininstMngt) approveFininstDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, ihtsig string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email, iht_signature"), true)
	}

	email = args[0]
	ihtsig = args[1]

	return t.changeFininstDetailStatus(stub, []string{email, ihtsig, ST_comm_approved, ST_DESC_comm_approved})
}

// ===============================================
// rejectFininst - reject a fininst as "03" status
// ===============================================
func (t *FininstMngt) rejectFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var email string

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	email = args[0]

	return t.changeFininstStatus(stub, []string{email, ST_comm_rejected, ST_DESC_comm_rejected})

}

// ===============================================
// rejectFininstDetail - reject a fininst_detail as "03" status
// ===============================================
func (t *FininstMngt) rejectFininstDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, ihtsig string

	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email, iht_signature"), true)
	}

	email = args[0]
	ihtsig = args[1]

	return t.changeFininstDetailStatus(stub, []string{email, ihtsig, ST_comm_rejected, ST_DESC_comm_rejected})
}

// ==================================================
// delete - delete a fininst key/value pair from state
// ==================================================
func (t *FininstMngt) deleteFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var jsonResp string
	//var FininstJSON fininst
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}
	email := args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_fininst, email) //get the fininst from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	fininstToUpdate := fininst{}
	err = json.Unmarshal(valAsbytes, &fininstToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if ST_comm_niled == fininstToUpdate.Fininst_status {
		jsonResp = "{\"Info\":\"" + email + "'s fininst was already deleted!\"}"
		fmt.Println("- end delete fininst (success) " + jsonResp)
	} else {

		fininstToUpdate.Fininst_status = ST_comm_niled //change the fininst status to "02" as approve status
		fininstToUpdate.Iht_approve_comments = ST_DESC_comm_niled

		fininstJSONasBytes, _ := json.Marshal(fininstToUpdate)
		err = putDocWithNamespace(stub, NS_fininst, email, fininstJSONasBytes)
		if err != nil {
			return errorPbResponse(err, false)
		}

		fmt.Println("- end delete fininst (success)")
	}

	valAsbytes2, err := getDocWithNamespace(stub, NS_fininst_detail, email) //get the fininst_detail from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst_detail + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes2 == nil {
		jsonResp = "{\"Error\":\"fininst_detail does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	fininstDetailToUpdate := fininst_detail{}
	err = json.Unmarshal(valAsbytes2, &fininstDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	if ST_comm_niled == fininstDetailToUpdate.Detail_status {
		jsonResp = "{\"Info\":\"" + email + "'s fininst_detail was already deleted!\"}"
		fmt.Println("- end delete fininst_detail (success) " + jsonResp)
	} else {

		fininstDetailToUpdate.Detail_status = ST_comm_niled //change the fininst status to "02" as approve status
		fininstDetailToUpdate.Iht_approve_comments = ST_DESC_comm_niled

		fininstDetailJSONasBytes, _ := json.Marshal(fininstDetailToUpdate)
		err = putDocWithNamespace(stub, NS_fininst_detail, email, fininstDetailJSONasBytes) //rewrite the asset fininst detail
		if err != nil {
			return errorPbResponse(err, false)
		}

		fmt.Println("- end delete fininst_detail (success)")
	}

	return successPbResponse(nil)
}

func (t *FininstMngt) changeFininstStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, status, desc, jsonResp string
	var err error

	if len(args) != 3 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting email, status, desc of the fininst to change status"), true)
	}

	email = args[0]
	status = args[1]
	desc = args[2]

	fmt.Println("- start changeFininstStatus: email " + email + " , status " + status + " , desc " + desc)

	if status != ST_comm_approving && status != ST_comm_approved && status != ST_comm_rejected && status != ST_comm_niled {
		return errorPbResponse(errors.New("Incorrect changed status. Expecting 01-正在审核  02-审核通过  03-审核不通过  99-作废 for fininst to change status"), true)
	}

	valAsbytes, err := getDocWithNamespace(stub, NS_fininst, email) //get the fininst from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	fininstToUpdate := fininst{}
	err = json.Unmarshal(valAsbytes, &fininstToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if fininstToUpdate.Fininst_status == status && fininstToUpdate.Iht_approve_comments == desc {
		fmt.Println("status " + status + " , desc " + desc + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}
	fininstToUpdate.Fininst_status = status
	fininstToUpdate.Iht_approve_comments = desc

	fininstJSONasBytes, err := json.Marshal(fininstToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_fininst, email, fininstJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeFininstStatus (success)")
	return successPbResponse(nil)
}

func (t *FininstMngt) changeFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	// 0                       1                  2
	// "rodger0514@163.com",   "fininst_nickname1", "fininst_pwd_hash"
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

	fininst_email := args[0]
	fininst_nickname := args[1]
	fininst_pwd_hash := args[2]

	fmt.Println("- start changeFininst: fininst_email " + fininst_email + " , fininst_nickname " + fininst_nickname + " , fininst_pwd_hash " + fininst_pwd_hash)

	valAsbytes, err := getDocWithNamespace(stub, NS_fininst, fininst_email) //get the fininst from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst + fininst_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst does not exist: " + fininst_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	fininstToUpdate := fininst{}
	err = json.Unmarshal(valAsbytes, &fininstToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	var isChanged bool = false

	if fininstToUpdate.Fininst_nickname != fininst_nickname {
		fininstToUpdate.Fininst_nickname = fininst_nickname
		isChanged = true
	}
	if fininstToUpdate.Fininst_pwd_hash != fininst_pwd_hash {
		fininstToUpdate.Fininst_pwd_hash = fininst_pwd_hash
		isChanged = true
	}

	if !isChanged {
		fmt.Println("- end changeFininst (no change no commit)")
		return successPbResponse(nil)
	}

	//reset status to 00-init if isChanged
	if fininstToUpdate.Fininst_status == ST_comm_rejected {
		fininstToUpdate.Fininst_status = ST_comm_init
		fininstToUpdate.Iht_approve_comments = ST_DESC_comm_init_change
	}

	fininstJSONasBytes, err := json.Marshal(fininstToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_fininst, fininst_email, fininstJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeFininst (success)")
	return successPbResponse(nil)
}

func (t *FininstMngt) changeFininstDetailStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email, ihtsig, status, desc, jsonResp string
	var err error

	if len(args) != 4 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting email, iht_signature, status, desc of the fininst_detail to change status"), true)
	}

	email = args[0]
	ihtsig = args[1]
	status = args[2]
	desc = args[3]

	fmt.Println("- start changeFininstStatus: email " + email + " , status " + status + " , desc " + desc)

	if status != ST_comm_approving && status != ST_comm_approved && status != ST_comm_rejected && status != ST_comm_niled {
		return errorPbResponse(errors.New("Incorrect changed status. Expecting 01-正在审核  02-审核通过  03-审核不通过  99-作废 for fininst to change status"), true)
	}

	valAsbytes, err := getDocWithNamespace(stub, NS_fininst_detail, email) //get the fininst from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst_detail + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst_detail does not exist: " + email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	fininstDetailToUpdate := fininst_detail{}
	err = json.Unmarshal(valAsbytes, &fininstDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	if fininstDetailToUpdate.Detail_status == status && fininstDetailToUpdate.Iht_approve_comments == desc {
		fmt.Println("status " + status + " , desc " + desc + " are the same as current status, omit and return success")
		return successPbResponse(nil)
	}

	fininstDetailToUpdate.Iht_signature = ihtsig
	fininstDetailToUpdate.Detail_status = status
	fininstDetailToUpdate.Iht_approve_comments = desc

	fininstDetailJSONasBytes, err := json.Marshal(fininstDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_fininst_detail, email, fininstDetailJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeFininstDetailStatus (success)")
	return successPbResponse(nil)

}

func (t *FininstMngt) changeFininstDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	if len(args) != 10 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 10"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start changeFininstDetail")
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

	Fininst_email := args[0]
	Fininst_realname := args[1]
	Fininst_contact := args[2]
	Fininst_identity_pic1_hash := args[3]
	Fininst_identity_pic2_hash := args[4]
	Fininst_identity_pic3_hash := args[5]
	Fininst_asset_proof_pic1_hash := args[6]
	Fininst_asset_proof_pic2_hash := args[7]
	Fininst_asset_proof_pic3_hash := args[8]
	Fininst_asset_proof_pic4_hash := args[9]

	valAsbytes, err := getDocWithNamespace(stub, NS_fininst_detail, Fininst_email) //get the asset_owner from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_fininst_detail + Fininst_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"fininst_detail does not exist: " + Fininst_email + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	fininstDetailToUpdate := fininst_detail{}
	err = json.Unmarshal(valAsbytes, &fininstDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	if fininstDetailToUpdate.Detail_status != ST_comm_init && fininstDetailToUpdate.Detail_status != ST_comm_rejected {
		msg := "current fininst_detail status " + fininstDetailToUpdate.Detail_status + "-" + fininstDetailToUpdate.Iht_approve_comments + " not allowed to change fininst_detail, only 00-init or 03-IHT rejected allowed"
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	var isChanged bool = false

	if fininstDetailToUpdate.Fininst_realname != Fininst_realname {
		fininstDetailToUpdate.Fininst_realname = Fininst_realname
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_contact != Fininst_contact {
		fininstDetailToUpdate.Fininst_contact = Fininst_contact
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_identity_pic1_hash != Fininst_identity_pic1_hash {
		fininstDetailToUpdate.Fininst_identity_pic1_hash = Fininst_identity_pic1_hash
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_identity_pic2_hash != Fininst_identity_pic2_hash {
		fininstDetailToUpdate.Fininst_identity_pic2_hash = Fininst_identity_pic2_hash
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_identity_pic3_hash != Fininst_identity_pic3_hash {
		fininstDetailToUpdate.Fininst_identity_pic3_hash = Fininst_identity_pic3_hash
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_asset_proof_pic1_hash != Fininst_asset_proof_pic1_hash {
		fininstDetailToUpdate.Fininst_asset_proof_pic1_hash = Fininst_asset_proof_pic1_hash
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_asset_proof_pic2_hash != Fininst_asset_proof_pic2_hash {
		fininstDetailToUpdate.Fininst_asset_proof_pic2_hash = Fininst_asset_proof_pic2_hash
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_asset_proof_pic3_hash != Fininst_asset_proof_pic3_hash {
		fininstDetailToUpdate.Fininst_asset_proof_pic3_hash = Fininst_asset_proof_pic3_hash
		isChanged = true
	}
	if fininstDetailToUpdate.Fininst_asset_proof_pic4_hash != Fininst_asset_proof_pic4_hash {
		fininstDetailToUpdate.Fininst_asset_proof_pic4_hash = Fininst_asset_proof_pic4_hash
		isChanged = true
	}

	if !isChanged {
		fmt.Println("- end changeFininstDetail (no change no commit)")
		return successPbResponse(nil)
	}

	//reset status to 00-init if isChanged
	if fininstDetailToUpdate.Detail_status == ST_comm_rejected {
		fininstDetailToUpdate.Detail_status = ST_comm_init
		fininstDetailToUpdate.Iht_approve_comments = ST_DESC_comm_init_change
		fininstDetailToUpdate.Iht_signature = ""
	}

	fininstDetailJSONasBytes, err := json.Marshal(fininstDetailToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}

	err = putDocWithNamespace(stub, NS_fininst_detail, Fininst_email, fininstDetailJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	fmt.Println("- end changeFininstDetail (success)")
	return successPbResponse(nil)

}

func (t *FininstMngt) getHistoryForFininst(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	fininst_email := args[0]

	fmt.Printf("- start getHistoryForFininst: %s\n", fininst_email)

	bytes, err := getHistoryForDocWithNamespace(stub, NS_fininst, fininst_email)

	if err != nil {
		return errorPbResponse(err, true)
	}

	return successPbResponse(bytes)

}

func (t *FininstMngt) getHistoryForFininstDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting fininst_email"), true)
	}

	fininst_email := args[0]

	fmt.Printf("- start getHistoryForFininstDetail: %s\n", fininst_email)

	bytes, err := getHistoryForDocWithNamespace(stub, NS_fininst_detail, fininst_email)

	if err != nil {
		return errorPbResponse(err, true)
	}

	return successPbResponse(bytes)

}
