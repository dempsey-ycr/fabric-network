package demo

import (
	"encoding/json"
	. "iht-fabric-chaincode-private/chaincode/go/app/controllers"
	"iht-fabric-chaincode-private/chaincode/go/app/libraries/logging"
	"iht-fabric-chaincode-private/chaincode/go/app/models/db"
	resp "iht-fabric-chaincode-private/chaincode/go/app/utils/response"

	"iht-fabric-chaincode-private/chaincode/go/app/utils/filter"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	DT_USER_INFO                 string = "userInfo"
	NS_USER_INFO                 string = DT_USER_INFO + "_"
	PK_FD_USER_INFO              string = "userEmail"
	IDX_FD_USER_STATUS           string = "userStatus"
	IDX_UERS_STATUS_2_USER_EMAIL string = IDX_FD_USER_STATUS + "_2_" + PK_FD_USER_INFO
)

type DemoUserMng struct{}

type UserInfo struct {
	//docType is used to distinguish the various types of objects in state database
	DocType      string `json:"docType"`
	UserEmail    string `json:"userEmail"`    //邮箱
	UserNickname string `json:"userNickname"` //昵称
	UserPwdHash  string `json:"userPwdHash"`  //密码hash值
	UserStatus   string `json:"userStatus"`   //当前状态：00-init 99-作废
}

func (t *DemoUserMng) Insert(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// "user_email",   "user_nickname", "user_pwd_hash"
	if err := filter.CheckParamsLength(args, 3); err != nil {
		return err.(peer.Response)
	}
	// ==== Input sanitation ====
	logging.Debug("- start init user")
	if err := filter.CheckRequired(args, []int{0, 1, 2}); err != nil {
		return err.(peer.Response)
	}

	email := args[0]
	nickname := args[1]
	pwdHash := args[2]

	logging.Debug("%s %s %s", email, nickname, pwdHash)

	// ==== Check if user_info already exists ====
	userInfoAsBytes, err := db.GetDocWithNamespace(stub, NS_USER_INFO+email)
	if err != nil {
		return resp.ErrorDB("Failed to get UserInfo: " + err.Error())
	} else if userInfoAsBytes != nil {
		return resp.ErrorDataExist("This UserInfo already exists: " + email)
	}

	// ==== Create user_info object and marshal to JSON ====
	userInfo := UserInfo{DT_USER_INFO, email, nickname, pwdHash, ST_COMM_INIT}
	logging.Info(userInfo.DocType)
	userInfoJSONasBytes, err := json.Marshal(userInfo)
	if err != nil {
		return resp.ErrorNormal(err.Error())
	}

	// === Save user_info to state ===
	err = db.PutDocWithNamespace(stub, NS_USER_INFO, email, userInfoJSONasBytes)
	if err != nil {
		return resp.ErrorNormal(err.Error())
	}

	//create IDX_uers_status_2_user_email
	err = db.CreateCKeyWithNamespace(stub, NS_USER_INFO, IDX_UERS_STATUS_2_USER_EMAIL, []string{userInfo.UserStatus, userInfo.UserEmail})
	if err != nil {
		return resp.ErrorDB(err.Error())
	}

	// ==== user_info saved and indexed. Return success ====
	logging.Debug("- end init user_info")

	return resp.Success(nil)
}

// ===============================================
// readUserInfo - read a user_info from chaincode state
// ===============================================
func (t *DemoUserMng) Read(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var email string

	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	email = args[0]
	valAsbytes, err := db.GetDocWithNamespace(stub, NS_USER_INFO+email) //get the user_info from chaincode state
	if err != nil {
		return resp.ErrorDB(err.Error())
	} else if valAsbytes == nil {
		return resp.ErrorDataExist("UserInfo does not exist: " + email)
	}
	logging.Debug("Success: Get user info")
	return resp.Success(valAsbytes)
}

// ==================================================
// delete - delete a user_info key/value pair from state
// ==================================================
func (t *DemoUserMng) Delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// "UserEmail"
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	// ==== Input sanitation ====
	if err := filter.CheckRequired(args, []int{0}); err != nil {
		return err.(peer.Response)
	}

	email := args[0]

	logging.Debug("- start DeleteUserinfo: UserEmail " + email)

	ValAsbytes, err := db.GetDocWithNamespace(stub, NS_USER_INFO+email) //get the UserInfo from chaincode state
	if err != nil {
		return resp.ErrorDB("Failed to get doc for " + NS_USER_INFO + email + ":" + err.Error())
	} else if ValAsbytes == nil {
		return resp.ErrorNotExist("user_info does not exist: " + email + ":" + err.Error())
	}

	userInfoToUpdate := UserInfo{}
	err = json.Unmarshal(ValAsbytes, &userInfoToUpdate)
	if err != nil {
		return resp.ErrorNormal(err.Error())
	}

	if userInfoToUpdate.UserStatus == ST_COMM_NILED {
		logging.Debug("- end delete user_info (success) " + email + "s UserInfo was already deleted!")
	} else {
		userInfoToUpdate.UserStatus = ST_COMM_NILED

		userInfoJSONasBytes, err := json.Marshal(userInfoToUpdate)
		if err != nil {
			return resp.ErrorNormal(err.Error())
		}

		err = db.PutDocWithNamespace(stub, NS_USER_INFO, email, userInfoJSONasBytes)
		if err != nil {
			return resp.ErrorDB(err.Error())
		}
	}
	logging.Debug("- end DeleteUserinfo (success)")
	return resp.Success(nil)
}

// ==================================================
// Change UserInfo key/value pair from state
// ==================================================
func (t *DemoUserMng) Change(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// "UserEmail",   "user_nickname", "user_pwd_hash"
	if err := filter.CheckParamsLength(args, 3); err != nil {
		return err.(peer.Response)
	}
	// ==== Input sanitation ====
	logging.Debug("- start init user")
	if err := filter.CheckRequired(args, []int{0, 1, 2}); err != nil {
		return err.(peer.Response)
	}
	email := args[0]
	nickname := args[1]
	pwdHash := args[2]

	logging.Debug("- start ChangeUserInfo: UserEmail " + email + " , UserNickname " + nickname + " , UserPwdHash " + pwdHash)

	ValAsbytes, err := db.GetDocWithNamespace(stub, NS_USER_INFO+email) //get the UserInfo from chaincode state
	if err != nil {
		return resp.ErrorDB("Failed to get doc for " + NS_USER_INFO + email + ":" + err.Error())
	} else if ValAsbytes == nil {
		return resp.ErrorNotExist("user_info does not exist: " + email)
	}

	userInfoToUpdate := UserInfo{}
	err = json.Unmarshal(ValAsbytes, &userInfoToUpdate)
	if err != nil {
		return resp.ErrorNormal(err.Error())
	}

	var isChanged bool
	isChanged = false

	if userInfoToUpdate.UserNickname != nickname {
		userInfoToUpdate.UserNickname = nickname
		isChanged = true
	}
	if userInfoToUpdate.UserPwdHash != pwdHash {
		userInfoToUpdate.UserPwdHash = pwdHash
		isChanged = true
	}

	if !isChanged {
		logging.Debug("- end changeUserInfo (no change no commit)")
		return resp.Success(nil)
	}

	userInfoJSONasBytes, err := json.Marshal(userInfoToUpdate)
	if err != nil {
		return resp.ErrorNormal(err.Error())
	}

	err = db.PutDocWithNamespace(stub, NS_USER_INFO, email, userInfoJSONasBytes)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}

	logging.Debug("- end ChangeUserInfo (success)")
	return resp.Success(nil)
}

// ===============================================
// queryUserInfoByStatus - read a user_info from chaincode state
// ===============================================
func (t *DemoUserMng) QueryWithStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 0
	// "status_01"
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	userStatus := args[0]

	queryResults, err := db.QueryStatusByIdx(stub, DT_USER_INFO, IDX_FD_USER_STATUS, userStatus)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}
	return resp.Success(queryResults)
}

func (t *DemoUserMng) QueryHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	email := args[0]
	logging.Debug("- start getHistoryForAssetOwner: " + email)

	historyUserInfoBytes, err := db.GetHistoryForDocWithNamespace(stub, NS_USER_INFO+email)
	if err != nil {
		return resp.ErrorDB(err.Error())
	}

	return resp.Success(historyUserInfoBytes)
}
