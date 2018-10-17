package filter

import (
	"fmt"
	resp "iht-fabric-chaincode-private/chaincode/go/app/utils/response"

	"errors"
)

// 判断必填参数是否为空
func CheckParamsNull(args ...string) error {
	for _, v := range args {
		if v == "" || v == " " {
			return errors.New("param error null...")
		}
	}
	return nil
}

func CheckParamsLength(args []string, lens int) interface{} {
	if len(args) != lens {
		for i := 0; i < len(args); i++ {
			fmt.Print(args[i], "---")
		}
		return resp.ErrorArguments(fmt.Sprintf("Incorrect number of arguments. Expecting %d != %d", len(args), lens))
	}
	return nil
}

func CheckRequired(argsArray []string, indexArray []int) interface{} {
	for _, index := range indexArray {
		if len(argsArray[index]) <= 0 {
			return resp.ErrorArguments(fmt.Sprintf("%vst argument must be a non-empty string ]", index))
		}
	}
	return nil
}
