package response

import (
	"encoding/json"

	"iht-fabric-chaincode-private/chaincode/go/app/libraries/logging"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	RESP_CODE_SUCESS                   int = 200  // 成功
	RESP_CODE_ERROR_ARGUMENTS          int = 2000 // 2000 - 参数错误
	RESP_CODE_ERROR_DATA_ALREADY_EXIST int = 2010 // 2010 - 数据已经存在
	RESP_CODE_ERROR_DATA_NOT_EXISTED   int = 2020 // 2020 - 数据不存在

	RESP_CODE_ERROR_DB     int = 9970 // 9970 - 存取错误
	RESP_CODE_ERROR_NORMAL int = 9980 // 9980 - 其它错误
	RESP_CODE_ERROR_SYSTEM int = 9990 // 9990 - 系统错误
)

type IHTResponse struct {
	Code   int         `json:"code"`  // 返回的code
	ErrMsg string      `json:"error"` // 失败描述
	Data   interface{} `json:"data"`  // 成功有可能会返回这个，json对象或数组或字符串
}

func normal(code int, message string, data []byte) peer.Response {
	var ss interface{}
	if err := json.Unmarshal(data, &ss); err != nil && len(data) != 0 {
		logging.Error("Not a usable json structure.")
	}
	res := &IHTResponse{
		Code:   code,
		ErrMsg: message,
		Data:   &ss,
	}
	resp, err := json.Marshal(res)
	if err != nil {
		logging.Error(err.Error())
		return shim.Error(err.Error())
	}
	return shim.Success(resp)
}

func Success(data []byte) peer.Response {
	//logging.Debug("response success: [code:%d; data:%s]", RESP_CODE_SUCESS, string(data))
	return normal(RESP_CODE_SUCESS, "", data)
}

func Error(code int, message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", code, message)
	return normal(code, message, nil)
}
func ErrorArguments(message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", RESP_CODE_ERROR_ARGUMENTS, message)
	return normal(RESP_CODE_ERROR_ARGUMENTS, message, nil)
}
func ErrorDB(message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", RESP_CODE_ERROR_DB, message)
	return normal(RESP_CODE_ERROR_DB, message, nil)
}
func ErrorSystem(message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", RESP_CODE_ERROR_SYSTEM, message)
	return normal(RESP_CODE_ERROR_SYSTEM, message, nil)
}

func ErrorDataExist(message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", RESP_CODE_ERROR_NORMAL, message)
	return normal(RESP_CODE_ERROR_DATA_ALREADY_EXIST, message, nil)
}

func ErrorNotExist(message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", RESP_CODE_ERROR_NORMAL, message)
	return normal(RESP_CODE_ERROR_DATA_NOT_EXISTED, message, nil)
}

func ErrorNormal(message string) peer.Response {
	logging.Error("response error:[code:%d; message: %s]", RESP_CODE_ERROR_NORMAL, message)
	return normal(RESP_CODE_ERROR_NORMAL, message, nil)
}
