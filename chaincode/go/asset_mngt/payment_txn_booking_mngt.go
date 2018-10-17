package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	//"time"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	DT_payment_txn_booking string = "payment_txn_booking"
	NS_payment_txn_booking string = DT_payment_txn_booking + "_"
	PK_FD_payment_txn_id   string = "payment_txn_id"
	//IDX_FD_asset_hash = "asset_hash"
	IDX_FD_share_id                        = "share_id"
	IDX_asset_hash_2_payment_txn_id string = IDX_FD_asset_hash + "_2_" + PK_FD_payment_txn_id
	IDX_share_id_2_payment_txn_id   string = IDX_FD_share_id + "_2_" + PK_FD_payment_txn_id
)

type PaymentTxnBookingMngt struct{}

type payment_txn_booking struct {
	ObjectType       string  `json:"docType"`          //文档类别"payment_book"
	Asset_hash       string  `json:"asset_hash"`       //支付交易相关的资产哈希，如果有
	Share_id         string  `json:"share_id"`         //支付交易相关的资产份额ID，如果有
	Payment_txn_id   string  `json:"payment_txn_id"`   //支付交易ID，对于数字资产的交易就是txn_id
	Debit1_ccy       string  `json:"debit1_ccy"`       //借1币种，“IHT”，“CNY”，“USD”
	Debit1_acct      string  `json:"debit1_acct"`      //借1账户号，对于数字权益就是私钥拥有者的公钥或者支付地址
	Debit1_acct_sub  string  `json:"debit1_acct_sub"`  //借1子账户
	Debit1_amt       float64 `json:"debit1_amt"`       //借1金额
	Debit1_owner     string  `json:"debit1_owner"`     //借1 owner email
	Credit1_ccy      string  `json:"credit1_ccy"`      //贷1币种，“IHT”，“CNY”，“USD”
	Credit1_acct     string  `json:"credit1_acct"`     //贷1账户号，对于数字权益就是私钥拥有者的公钥或者支付地址
	Credit1_acct_sub string  `json:"credit1_acct_sub"` //贷1子账户
	Credit1_amt      float64 `json:"credit1_amt"`      //贷1金额
	Credit1_owner    string  `json:"credit1_owner"`    //贷1 owner email
	Debit2_ccy       string  `json:"debit2_ccy"`       //借2币种，“IHT”，“CNY”，“USD”
	Debit2_acct      string  `json:"debit2_acct"`      //借2账户号，对于数字权益就是私钥拥有者的公钥或者支付地址
	Debit2_acct_sub  string  `json:"debit2_acct_sub"`  //借2子账户
	Debit2_amt       float64 `json:"debit2_amt"`       //借2金额
	Debit2_owner     string  `json:"debit2_owner"`     //借2 owner email
	Credit2_ccy      string  `json:"credit2_ccy"`      //贷2币种，“IHT”，“CNY”，“USD”
	Credit2_acct     string  `json:"credit2_acct"`     //贷2账户号，对于数字权益就是私钥拥有者的公钥或者支付地址
	Credit2_acct_sub string  `json:"credit2_acct_sub"` //贷2子账户
	Credit2_amt      float64 `json:"credit2_amt"`      //贷2金额
	Credit2_owner    string  `json:"credit2_owner"`    //贷2 owner email
	Payment_desc     string  `json:"payment_desc"`     //支付事由描述
}

// ============================================================
// initPaymentTxnBooking
// ============================================================
func (t *PaymentTxnBookingMngt) initPaymentTxnBooking(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//asset_hash,share_id,payment_txn_id,debit1_ccy,debit1_acct,debit1_acct_sub,debit1_amt,debit1_owner,credit1_ccy,credit1_acct,credit1_acct_sub,credit1_amt,credit1_owner,debit2_ccy,debit2_acct,debit2_acct_sub,debit2_amt,debit2_owner,credit2_ccy,credit2_acct,credit2_acct_sub,credit2_amt,credit2_owner,payment_desc
	if len(args) != 24 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting 24"), true)
	}

	// ==== Input sanitation ====
	fmt.Println("- start initPaymentTxnBooking")
	if len(args[0]) <= 0 {
		return errorPbResponse(errors.New("1st argument must be a non-empty string"), true)
	}
	if len(args[2]) <= 0 {
		return errorPbResponse(errors.New("2nd argument must be a non-empty string"), true)
	}
	if (len(args[3]) > 0 || len(args[4]) > 0 || len(args[6]) > 0 || len(args[7]) > 0) && (len(args[3]) <= 0 || len(args[4]) <= 0 || len(args[6]) <= 0 || len(args[7]) <= 0) {
		return errorPbResponse(errors.New("debit1_ccy, debit1_acct, debit1_amt, debit1_owner must or must not be empty string at the same time"), true)
	}
	if (len(args[8]) > 0 || len(args[9]) > 0 || len(args[11]) > 0 || len(args[12]) > 0) && (len(args[8]) <= 0 || len(args[9]) <= 0 || len(args[11]) <= 0 || len(args[12]) <= 0) {
		return errorPbResponse(errors.New("credit1_ccy, credit1_acct, credit1_amt, credit1_owner must or must not be empty string at the same time"), true)
	}
	if (len(args[13]) > 0 || len(args[14]) > 0 || len(args[16]) > 0 || len(args[17]) > 0) && (len(args[13]) <= 0 || len(args[14]) <= 0 || len(args[16]) <= 0 || len(args[17]) <= 0) {
		return errorPbResponse(errors.New("debit2_ccy, debit2_acct, debit2_amt, debit2_owner must or must not be empty string at the same time"), true)
	}
	if (len(args[18]) > 0 || len(args[19]) > 0 || len(args[21]) > 0 || len(args[22]) > 0) && (len(args[18]) <= 0 || len(args[19]) <= 0 || len(args[21]) <= 0 || len(args[22]) <= 0) {
		return errorPbResponse(errors.New("credit2_ccy, credit2_acct, credit2_amt, credit2_owner must or must not be empty string at the same time"), true)
	}
	if len(args[3]) <= 0 && len(args[13]) <= 0 {
		return errorPbResponse(errors.New("For debit1 or debit2 elements there must be appeared at least one group for a valid booking transaction"), true)
	}
	if len(args[8]) <= 0 && len(args[18]) <= 0 {
		return errorPbResponse(errors.New("For credit1 or credit2 elements there must be appeared at least one group for a valid booking transaction"), true)
	}

	asset_hash := args[0]
	share_id := args[1]
	payment_txn_id := args[2]
	debit1_ccy := args[3]
	debit1_acct := args[4]
	debit1_acct_sub := args[5]
	debit1_amt := 0.00
	fmt.Printf("debit1_amt args[6]:%s\n", args[6])
	if args[6] != "" {
		debit1_amt, err = strconv.ParseFloat(args[6], 64)
		if err != nil {
			return errorPbResponse(errors.New("debit1_amt 7th argument must be a valid float64 string"), true)
		}
	}
	debit1_owner := args[7]
	credit1_ccy := args[8]
	credit1_acct := args[9]
	credit1_acct_sub := args[10]
	credit1_amt := 0.00
	fmt.Printf("credit1_amt args[11]:%s\n", args[11])
	if args[11] != "" {
		credit1_amt, err = strconv.ParseFloat(args[11], 64)
		if err != nil {
			return errorPbResponse(errors.New("credit1_amt 12th argument must be a valid float64 string"), true)
		}
	}
	credit1_owner := args[12]
	debit2_ccy := args[13]
	debit2_acct := args[14]
	debit2_acct_sub := args[15]
	debit2_amt := 0.00
	fmt.Printf("debit2_amt args[16]:%s\n", args[11])
	if args[16] != "" {
		debit2_amt, err = strconv.ParseFloat(args[16], 64)
		if err != nil {
			return errorPbResponse(errors.New("debit2_amt 17th argument must be a valid float64 string"), true)
		}
	}
	debit2_owner := args[17]
	credit2_ccy := args[18]
	credit2_acct := args[19]
	credit2_acct_sub := args[20]
	credit2_amt := 0.00
	fmt.Printf("credit2_amt args[16]:%s\n", args[21])
	if args[21] != "" {
		credit2_amt, err = strconv.ParseFloat(args[21], 64)
		if err != nil {
			return errorPbResponse(errors.New("credit2_amt 22th argument must be a valid float64 string"), true)
		}
	}
	credit2_owner := args[22]
	payment_desc := args[23]

	// ==== Check if payment_txn_booking already exists ====
	PaymentTxnBookingAsBytes, err := getDocWithNamespace(stub, NS_payment_txn_booking, payment_txn_id)
	if err != nil {
		return errorPbResponse(errors.New("Failed to get payment_txn_booking: "+err.Error()), true)
	} else if PaymentTxnBookingAsBytes != nil {
		fmt.Println("This payment_txn_booking already exists: " + payment_txn_id)
		return errorPbResponse(errors.New("This payment_txn_booking already exists: "+payment_txn_id), true)
	}

	// ==== Create payment_txn_booking_mngt object and marshal to JSON ====
	PaymentTxnBooking := &payment_txn_booking{DT_payment_txn_booking, asset_hash, share_id, payment_txn_id, debit1_ccy, debit1_acct, debit1_acct_sub, debit1_amt, debit1_owner, credit1_ccy, credit1_acct, credit1_acct_sub, credit1_amt, credit1_owner, debit2_ccy, debit2_acct, debit2_acct_sub, debit2_amt, debit2_owner, credit2_ccy, credit2_acct, credit2_acct_sub, credit2_amt, credit2_owner, payment_desc}
	PaymentTxnBookingJSONasBytes, err := json.Marshal(PaymentTxnBooking)
	if err != nil {
		return errorPbResponse(err, false)
	}
	// === Save payment_txn_booking to state ===
	err = putDocWithNamespace(stub, NS_payment_txn_booking, payment_txn_id, PaymentTxnBookingJSONasBytes)
	if err != nil {
		return errorPbResponse(err, false)
	}

	// create IDX_asset_hash_2_payment_txn_id
	err = createCKeyWithNamespace(stub, NS_payment_txn_booking, IDX_asset_hash_2_payment_txn_id, []string{PaymentTxnBooking.Asset_hash, PaymentTxnBooking.Payment_txn_id})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// create IDX_share_id_2_payment_txn_id
	err = createCKeyWithNamespace(stub, NS_payment_txn_booking, IDX_share_id_2_payment_txn_id, []string{PaymentTxnBooking.Share_id, PaymentTxnBooking.Payment_txn_id})
	if err != nil {
		return errorPbResponse(err, false)
	}

	// ==== payment_txn_booking_mngt saved and indexed. Return success ====
	fmt.Println("- end initPaymentTxnBooking (success)")
	return successPbResponse(nil)
}

// ===============================================
// readPaymentTxnBooking
// ===============================================
func (t *PaymentTxnBookingMngt) readPaymentTxnBooking(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var payment_txn_id, jsonResp string
	var err error

	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting payment_txn_id of the payment_txn_booking to query"), true)
	}

	payment_txn_id = args[0]
	valAsbytes, err := getDocWithNamespace(stub, NS_payment_txn_booking, payment_txn_id)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get doc for " + NS_payment_txn_booking + payment_txn_id + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"payment_txn_booking does not exist: " + payment_txn_id + "\"}"
		return errorPbResponse(errors.New(jsonResp), true)
	}

	return successPbResponse(valAsbytes)
}

// ===============================================
// queryBookedPaymentTxnsByAssetHash
// ===============================================
func (t *PaymentTxnBookingMngt) queryBookedPaymentTxnsByAssetHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "asset_hash"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting asset_hash"), true)
	}

	asset_hash := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_payment_txn_booking, IDX_FD_asset_hash, asset_hash)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}

// ===============================================
// queryBookedPaymentTxnsByShareId
// ===============================================
func (t *PaymentTxnBookingMngt) queryBookedPaymentTxnsByShareId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "share_id"
	if len(args) != 1 {
		return errorPbResponse(errors.New("Incorrect number of arguments. Expecting share_id"), true)
	}

	share_id := args[0]

	queryResults, err := queryDocsByIdxkey(stub, DT_payment_txn_booking, IDX_FD_share_id, share_id)
	if err != nil {
		return errorPbResponse(err, false)
	}
	return successPbResponse(queryResults)
}
