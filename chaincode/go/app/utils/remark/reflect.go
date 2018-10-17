package reflect

import (
	resp "iht-fabric-chaincode-private/chaincode/go/app/utils/response"
	"reflect"

	"github.com/hyperledger/fabric/protos/peer"
)

func call(function interface{}, params ...interface{}) peer.Response {
	f := reflect.ValueOf(function)
	if len(params) != f.Type().NumIn() {
		return resp.ErrorNormal("The number of params is not adapted.")
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	var resultValue = f.Call(in)[0]
	result := resultValue.Interface()
	s, ok := result.(peer.Response)
	if ok {
		return s
	}
	return resp.ErrorNormal("Function return value type does not match")
}
