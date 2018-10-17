package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

const (
	ST_share_init   string = "00" // 00-init
	ST_share_valid  string = "01" // 01-生效
	ST_share_frozen string = "02" // 02-冻结
	ST_share_niled  string = "99" // 99-作废
)

const (
	ST_DESC_share_init   string = "init"   // 00-init
	ST_DESC_share_valid  string = "valid"  // 01-生效
	ST_DESC_share_frozen string = "frozen" // 02-冻结
	ST_DESC_share_niled  string = "niled"  // 99-作废
)

const (
	DT_asset_security_share    string = "asset_security_share"
	NS_asset_security_share    string = DT_asset_security_share + "_"
	PK_FD_asset_security_share string = "share_id"
	IDX_FD_input_share_hash           = "input_share_hash"
	IDX_FD_output_share_hash          = "output_share_hash"
	//IDX_FD_asset_name = "asset_name"
	IDX_FD_owner_email                      = "owner_email"
	IDX_FD_asset_hash                       = "asset_hash"
	IDX_input_share_hash_2_share_id  string = IDX_FD_input_share_hash + "_2_" + PK_FD_asset_security_share
	IDX_output_share_hash_2_share_id string = IDX_FD_output_share_hash + "_2_" + PK_FD_asset_security_share
	IDX_asset_hash_2_share_id        string = IDX_FD_asset_hash + "_2_" + PK_FD_asset_security_share
	IDX_asset_name_2_share_id        string = IDX_FD_asset_name + "_2_" + PK_FD_asset_security_share
	IDX_owner_email_2_share_id       string = IDX_FD_owner_email + "_2_" + PK_FD_asset_security_share
)

type AssetSecurityShareMngt struct{}

type asset_security_share struct {
	ObjectType                       string   `json:"docType"`                          //文档类别"asset_security_share"
	Input_share_hash                 string   `json:"input_share_hash"`                 //份额所属输入份额哈希，当前为资产哈希
	Share_id                         string   `json:"share_id"`                         //份额ID，基于份额关键信息生成的哈希 HASH256({"unit_nos":["1-1000","2001-3000","4000"],"underwrote_id":"xxxx","owner_email":"xx@xx(init时为资产方email)","asset_name":"xxxx","share_value_ccy":"XXX","share_value_amt":xxxx.xx,"unit_total":tn,"share_annual_dividend":n.nnnn,"share_repurchase_time":"xxxxxx","share_allow_delegate_rent_income":true|false,"share_status":"00"})
	Output_share_hash                string   `json:"output_share_hash"`                //输出份额哈希，当前固定""，扩展用，便于未来交易市场份额转让
	Unit_nos                         []string `json:"unit_nos"`                         //份额编号数组： Unit_nos[i]≠Unit_nos[j] 当i≠j且i,j<unit_total时；Unit_nos[i] = "m-n" |"o" 其中m<n, o ∉[m,n]并且 m,n,o∈{1,2,3,…,unit_total}。
	Underwrote_id                    string   `json:"underwrote_id"`                    //合法承销id，可追溯承销信息
	Owner_email                      string   `json:"owner_email"`                      //份额所有方邮箱，资产方或金融机构email，初始状态为原资产方email，被承销部分会转给金融机构，未承销部分仍然为资产方
	Asset_name                       string   `json:"asset_name"`                       //资产名称，可追溯资产分拆和承销状态信息
	Share_value_ccy                  string   `json:"share_value_ccy"`                  //份额货币类型，同资产价值货币类型
	Share_value_amt                  float64  `json:"share_value_amt"`                  //份额货币类型计价的份额价值
	Unit_total                       int      `json:"unit_total"`                       //资产总拆分份额
	Share_annual_dividend            float64  `json:"share_annual_dividend"`            //份额年化收益，同资产年化收益
	Share_repurchase_time            string   `json:"share_repurchase_time"`            //份额回购时间，同资产回购截止时间
	Share_allow_delegate_rent_income bool     `json:"share_allow_delegate_rent_income"` //是否可以委托租赁产生额外收益，同资产的是否可以委托租赁产生额外收益
	Share_status                     string   `json:"share_status"`                     //份额当前状态：00-init 01-生效 02-冻结 99-作废
	Share_status_desc                string   `json:"share_status_desc"`                //份额当前状态描述 00-init 01-valid 02-fronze 99-niled
	Asset_hash                       string   `json:"asset_hash"`                       //资产哈希
	Previous_owner_signature         string   `json:"previous_owner_signature"`         //本资产份额上一个拥有者同意转给当前拥有者owner_email的签名
}

// ============================================================
// initAssetSecurityShares
// ============================================================
func (t *AssetSecurityShareMngt) initAssetSecurityShares(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetName, ihtsig string
	var err error

	// 0           1
	// asset_name, iht_signature
	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name and iht_signature"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start initAssetSecurityShares")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}

	assetName = args[0]
	ihtsig = args[1]

	// ==== Check if asset_securitized exists and in 12-承销成功 status only====
	AssetSecuritizedAsBytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_securitized: "+err.Error()), true)
	}

	if AssetSecuritizedAsBytes == nil {
		msg := "This asset_securitized does not exist: " + assetName
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(AssetSecuritizedAsBytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_succ_underwriting {
		msg := "Asset should be underwrote successfully before generating security units, current asset status " + AssetSecuritizedToUpdate.Asset_status + "-" + AssetSecuritizedToUpdate.Iht_status_comments + " operation not allowed"
		return errorPbResponse(errors.New(msg), true)
	}

	// ==== Check if asset_security_share already exists ====
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", DT_asset_security_share, IDX_FD_asset_name, assetName)

	fmt.Printf("- initAssetSecurityShares queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return errorPbResponse(err, false)
	}
	defer resultsIterator.Close()
	if resultsIterator.HasNext() {
		msg := "ATTENTION!! existing asset_security_share for assetName: " + assetName + ", there must be something wrong in somewhere, pls check!!!"
		return errorPbResponse(errors.New(msg), true)
	}

	// ==== Query current underwrote records, and generating shares for each underwrote record
	queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", DT_asset_security_underwrote, IDX_FD_asset_name, assetName)

	fmt.Printf("- initAssetSecurityShares queryString:\n%s\n", queryString)

	resultsIterator, err = stub.GetQueryResult(queryString)
	if err != nil {
		return errorPbResponse(err, false)
	}
	defer resultsIterator.Close()

	var unitNo int = 1
	var processedUnits int = 0
	var shareIds []string
	var assetShares []asset_security_share
	shareIds = make([]string, 0)
	assetShares = make([]asset_security_share, 0)
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return errorPbResponse(err, false)
		}
		AssetSecurityUnderwrote := asset_security_underwrote{}
		err = json.Unmarshal(queryResponse.Value, &AssetSecurityUnderwrote)
		if err != nil {
			return errorPbResponse(err, false)
		}
		//generate asset_security_share for underwrote fininsts
		unitNoStart := unitNo
		unitNo = unitNo + AssetSecurityUnderwrote.Underwrote_units
		unitNoEnd := unitNo - 1
		var unit_nos []string = make([]string, 1)
		unit_nos[0] = strconv.Itoa(unitNoStart) + "-" + strconv.Itoa(unitNoEnd)
		unit_nos_str := "[\"" + unit_nos[0] + "\"]"
		underwrote_id := AssetSecurityUnderwrote.Underwrote_id
		owner_email := AssetSecuritizedToUpdate.Owner_email
		asset_name := AssetSecuritizedToUpdate.Asset_name
		share_value_ccy := AssetSecuritizedToUpdate.Asset_value_ccy
		share_value_amt := AssetSecuritizedToUpdate.Asset_value_amt * float64(AssetSecurityUnderwrote.Underwrote_units) / float64(AssetSecuritizedToUpdate.Asset_securitized_parts)
		unit_total := AssetSecuritizedToUpdate.Asset_securitized_parts
		share_annual_dividend := AssetSecuritizedToUpdate.Asset_annual_dividend
		share_repurchase_time := AssetSecuritizedToUpdate.Repurchase_due_time
		share_allow_delegate_rent_income := AssetSecuritizedToUpdate.Asset_allow_delegate_rent_income
		share_status := ST_share_init
		share_status_desc := ST_DESC_share_init

		securityShareJson := fmt.Sprintf("{\"unit_nos\":%s,\"underwrote_id\":\"%s\",\"owner_email\":\"%s\",\"asset_name\":\"%s\",\"share_value_ccy\":\"%s\",\"share_value_amt\":%.2f,\"unit_total\":%d,\"share_annual_dividend\":%.6f,\"share_repurchase_time\":\"%s\",\"share_allow_delegate_rent_income\":%t,\"share_status\":\"%s\",\"share_status_desc\":\"%s\"}", unit_nos_str, underwrote_id, owner_email, asset_name, share_value_ccy, share_value_amt, unit_total, share_annual_dividend, share_repurchase_time, share_allow_delegate_rent_income, share_status, share_status_desc)
		share_id := ComputeSHA256Base16UpperCase(securityShareJson)
		fmt.Printf("%s: %s, SHA256: %s\n", GetRFC3339TimeStr(time.Now().Local()), securityShareJson, share_id)

		shareIds = append(shareIds, share_id)

		input_share_hash := ""
		output_share_hash := ""
		asset_hash := ""
		previous_owner_signature := ""

		AssetSecurityShare := &asset_security_share{DT_asset_security_share, input_share_hash, share_id, output_share_hash, unit_nos, underwrote_id, owner_email, asset_name, share_value_ccy, share_value_amt, unit_total, share_annual_dividend, share_repurchase_time, share_allow_delegate_rent_income, share_status, share_status_desc, asset_hash, previous_owner_signature}
		assetShares = append(assetShares, *AssetSecurityShare)

		processedUnits = processedUnits + AssetSecurityUnderwrote.Underwrote_units
	}

	//generate un-underwrote share to asset_owner itself with frozen status and underwrote_id empty string
	if processedUnits < AssetSecuritizedToUpdate.Asset_securitized_parts {
		var unit_nos []string = make([]string, 1)
		var unit_nos_str string
		if unitNo == AssetSecuritizedToUpdate.Asset_securitized_parts {
			unit_nos_str = "\"" + strconv.Itoa(unitNo) + "\""
		} else {
			unitNoStart := unitNo
			unitNoEnd := AssetSecuritizedToUpdate.Asset_securitized_parts

			unit_nos[0] = strconv.Itoa(unitNoStart) + "-" + strconv.Itoa(unitNoEnd)
			unit_nos_str = "[\"" + unit_nos[0] + "\"]"
		}
		underwrote_id := ""
		owner_email := AssetSecuritizedToUpdate.Owner_email
		asset_name := AssetSecuritizedToUpdate.Asset_name
		share_value_ccy := AssetSecuritizedToUpdate.Asset_value_ccy
		share_value_amt := AssetSecuritizedToUpdate.Asset_value_amt * float64(AssetSecuritizedToUpdate.Asset_securitized_parts-processedUnits) / float64(AssetSecuritizedToUpdate.Asset_securitized_parts)
		unit_total := AssetSecuritizedToUpdate.Asset_securitized_parts
		share_annual_dividend := AssetSecuritizedToUpdate.Asset_annual_dividend
		share_repurchase_time := AssetSecuritizedToUpdate.Repurchase_due_time
		share_allow_delegate_rent_income := AssetSecuritizedToUpdate.Asset_allow_delegate_rent_income
		share_status := ST_share_frozen
		share_status_desc := ST_DESC_share_frozen

		securityShareJson := fmt.Sprintf("{\"unit_nos\":%s,\"underwrote_id\":\"%s\",\"owner_email\":\"%s\",\"asset_name\":\"%s\",\"share_value_ccy\":\"%s\",\"share_value_amt\":%.2f,\"unit_total\":%d,\"share_annual_dividend\":%.6f,\"share_repurchase_time\":\"%s\",\"share_allow_delegate_rent_income\":%t,\"share_status\":\"%s\",\"share_status_desc\":\"%s\"}", unit_nos_str, underwrote_id, owner_email, asset_name, share_value_ccy, share_value_amt, unit_total, share_annual_dividend, share_repurchase_time, share_allow_delegate_rent_income, share_status, share_status_desc)
		share_id := ComputeSHA256Base16UpperCase(securityShareJson)
		fmt.Printf("%s: %s, SHA256: %s\n", GetRFC3339TimeStr(time.Now().Local()), securityShareJson, share_id)

		shareIds = append(shareIds, share_id)

		input_share_hash := ""
		output_share_hash := ""
		asset_hash := ""
		previous_owner_signature := ""

		AssetSecurityShare := &asset_security_share{DT_asset_security_share, input_share_hash, share_id, output_share_hash, unit_nos, underwrote_id, owner_email, asset_name, share_value_ccy, share_value_amt, unit_total, share_annual_dividend, share_repurchase_time, share_allow_delegate_rent_income, share_status, share_status_desc, asset_hash, previous_owner_signature}
		assetShares = append(assetShares, *AssetSecurityShare)
	}

	//generate asset_hash
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for j := 0; j < len(shareIds); j++ {

		buffer.WriteString("\"" + shareIds[j] + "\"")
		if j != len(shareIds)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
	assetSecuritizedJson := fmt.Sprintf("{\"asset_name\":\"%s\",\"share_ids\":%s,\"owner_email\":\"%s\",\"asset_value_ccy\":\"%s\",\"asset_value_amt\":%.2f,\"asset_securitized_parts\":%d,\"asset_annual_dividend\":%.6f,\"asset_repurchase_time\":\"%s\",\"asset_allow_delegate_rent_income\":%t}", AssetSecuritizedToUpdate.Asset_name, buffer.String(), AssetSecuritizedToUpdate.Owner_email, AssetSecuritizedToUpdate.Asset_value_ccy, AssetSecuritizedToUpdate.Asset_value_amt, AssetSecuritizedToUpdate.Asset_securitized_parts, AssetSecuritizedToUpdate.Asset_annual_dividend, AssetSecuritizedToUpdate.Repurchase_due_time, AssetSecuritizedToUpdate.Asset_allow_delegate_rent_income)
	fmt.Println(assetSecuritizedJson)
	asset_hash := ComputeSHA256Base16UpperCase(assetSecuritizedJson)

	for i := 0; i < len(assetShares); i++ {
		// === Update asset_security_share object with newly calculated sha256 hash and marshal to JSON ====
		assetShares[i].Asset_hash = asset_hash
		assetShares[i].Input_share_hash = asset_hash
		err := t.saveAssetSecurityShare(stub, assetShares[i], true)
		if err != nil {
			return errorPbResponse(err, false)
		}
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = ST_sec_security_units_initialized
	AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_security_units_initialized
	AssetSecuritizedToUpdate.Asset_hash = asset_hash

	AssetSecuritizedJSONasBytes, err := json.Marshal(AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ====all asset_security_share saved and indexed and asset_securitized updated asset_status. Return success ====
	fmt.Println("- end initAssetSecurityShares (success)")
	return successPbResponse(nil)
}

func (t *AssetSecurityShareMngt) saveAssetSecurityShare(stub shim.ChaincodeStubInterface, share asset_security_share, isInit bool) error {
	AssetSecurityShareJSONasBytes, err := json.Marshal(share)
	if err != nil {
		return err
	}

	// === Save asset_security_share to state ===
	err = putDocWithNamespace(stub, NS_asset_security_share, share.Share_id, AssetSecurityShareJSONasBytes)
	if err != nil {
		return err
	}

	if isInit {
		// create IDX_asset_hash_2_share_id
		err = createCKeyWithNamespace(stub, NS_asset_security_share, IDX_asset_hash_2_share_id, []string{share.Asset_hash, share.Share_id})
		if err != nil {
			return err
		}

		// create IDX_asset_name_2_share_id
		err = createCKeyWithNamespace(stub, NS_asset_security_share, IDX_asset_name_2_share_id, []string{share.Asset_name, share.Share_id})
		if err != nil {
			return err
		}

		// create IDX_owner_email_2_share_id
		err = createCKeyWithNamespace(stub, NS_asset_security_share, IDX_owner_email_2_share_id, []string{share.Owner_email, share.Share_id})
		if err != nil {
			return err
		}
	}

	return nil
}

// ===============================================
// readAssetSecurityShare - read a asset_security_share_mngt from chaincode state
// ===============================================
func (t *AssetSecurityShareMngt) readAssetSecurityShare(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var shareId, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting share_id to query"), true)
	}

	shareId = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_asset_security_share, shareId) //get the asset_security_share from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get" + DT_asset_security_share + " doc for " + shareId + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"asset_security_share does not exist: " + shareId + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// enableUnderwroteAssetSharesOwnerships
// ===============================================
func (t *AssetSecurityShareMngt) enableUnderwroteAssetSharesOwnerships(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	// 0             2
	// "asset_name", "iht_signature"
	if len(args) != 2 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name and iht_signature"), true)
	}

	assetName := args[0]
	ihtsig := args[1]

	// ==== Check if asset_securitized exists and in  20-份额拆分完成 status only====
	AssetSecuritizedAsBytes, err := getDocWithNamespace(stub, NS_asset_securitized, assetName)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get asset_securitized: "+err.Error()), true)
	}

	if AssetSecuritizedAsBytes == nil {
		msg := "This asset_securitized does not exist: " + assetName
		fmt.Println(msg)
		return errorPbResponse(errors.New(msg), true)
	}

	AssetSecuritizedToUpdate := asset_securitized{}
	err = json.Unmarshal(AssetSecuritizedAsBytes, &AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	if AssetSecuritizedToUpdate.Asset_status != ST_sec_security_units_initialized {
		msg := "Asset should be in " + ST_sec_security_units_initialized + "-" + ST_DESC_sec_security_units_initialized + " status before enable underwrote fininst ownerships, current asset status " + AssetSecuritizedToUpdate.Asset_status + "-" + AssetSecuritizedToUpdate.Iht_status_comments + " operation not allowed"
		return errorPbResponse(errors.New(msg), true)
	}

	//query all security shares by assetName
	//for all Underwrote_id not "" change staus from 00 to 01
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", DT_asset_security_share, IDX_FD_asset_name, assetName)

	fmt.Printf("- enableUnderwroteAssetSharesOwnerships queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return errorPbResponse(err, false)
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return errorPbResponse(err, false)
		}
		AssetSecurityShareToUpdate := asset_security_share{}
		err = json.Unmarshal(queryResponse.Value, &AssetSecurityShareToUpdate)
		if err != nil {
			return errorPbResponse(err, false)
		}
		if AssetSecurityShareToUpdate.Underwrote_id != "" && AssetSecurityShareToUpdate.Share_status == ST_share_init {
			AssetSecurityShareToUpdate.Share_status = ST_share_valid
			AssetSecurityShareToUpdate.Share_status_desc = ST_DESC_share_valid
			//TODO considering let owner provide Previous_owner_signature with asset_owner for right transfering

			underwroteId := AssetSecurityShareToUpdate.Underwrote_id
			valAsbytes, err := getDocWithNamespace(stub, NS_asset_security_underwrote, underwroteId)
			if err != nil {
				jsonResp := "{\"Error\":\"Failed to get" + DT_asset_security_underwrote + " doc for " + underwroteId + "\"}"
				return errorPbResponse(errors.New(jsonResp), true)
			} else if valAsbytes == nil {
				jsonResp := "{\"Error\":\"asset_security_underwrote does not exist: " + underwroteId + "\"}"
				return errorPbResponse(errors.New(jsonResp), true)
			}
			AssetSecurityUnderwrote := asset_security_underwrote{}
			err = json.Unmarshal(valAsbytes, &AssetSecurityUnderwrote)
			if err != nil {
				return errorPbResponse(err, false)
			}
			AssetSecurityShareToUpdate.Owner_email = AssetSecurityUnderwrote.Fininst_email
			err = t.saveAssetSecurityShare(stub, AssetSecurityShareToUpdate, false)
			if err != nil {
				return errorPbResponse(err, false)
			}
		}
	}

	//TODO check ihtsig
	AssetSecuritizedToUpdate.Iht_signature = ihtsig
	AssetSecuritizedToUpdate.Asset_status = ST_sec_effect_ownership_change
	AssetSecuritizedToUpdate.Iht_status_comments = ST_DESC_sec_effect_ownership_change
	AssetSecuritizedJSONasBytes, err := json.Marshal(AssetSecuritizedToUpdate)
	if err != nil {
		return errorPbResponse(err, false)
	}
	err = putDocWithNamespace(stub, NS_asset_securitized, assetName, AssetSecuritizedJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(nil)
}

// ===============================================
// queryAssetSecuritySharesByOwnerEmail
// ===============================================
func (t *AssetSecurityShareMngt) queryAssetSecuritySharesByOwnerEmail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "owner_email"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting owner_email"), true)
	}

	ownerEmail := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_security_share, IDX_FD_owner_email, ownerEmail)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryAssetSecuritySharesByAssetName
// ===============================================
func (t *AssetSecurityShareMngt) queryAssetSecuritySharesByAssetName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "asset_name"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_name"), true)
	}

	assetName := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_security_share, IDX_FD_asset_name, assetName)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryAssetSecuritySharesByAssetHash
// ===============================================
func (t *AssetSecurityShareMngt) queryAssetSecuritySharesByAssetHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "asset_hash"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_hash"), true)
	}

	assetHash := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_asset_security_share, IDX_FD_asset_hash, assetHash)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}
