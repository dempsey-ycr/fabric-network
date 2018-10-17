package main

import (
	"fmt"
	"iht-fabric-chaincode-private/chaincode/go/app/controllers/assetManage"
	"iht-fabric-chaincode-private/chaincode/go/app/controllers/assets"
	"iht-fabric-chaincode-private/chaincode/go/app/controllers/demo"
	"iht-fabric-chaincode-private/chaincode/go/app/libraries/logging"
	"iht-fabric-chaincode-private/chaincode/go/app/utils/filter"
	resp "iht-fabric-chaincode-private/chaincode/go/app/utils/response"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type peerFunc func(shim.ChaincodeStubInterface, []string) peer.Response

// var map_functions map[string]peerFunc // goroutine时考虑并发安全性

// AppManage app management
type AppManage struct {
	DemoUserMng   *demo.DemoUserMng
	AssetMng      *assetManage.AssetManagement
	AssetProe     *assetManage.AssetProcess
	Assets        *assets.AssetManage
	AssetSide     *assets.AssetsideManage
	map_functions map[string]peerFunc
}

func main() {
	err := shim.Start(new(AppManage))
	if err != nil {
		logging.Error("Error starting AppManage chaincode: %s", err.Error())
	}
}

// Init ...
func (p *AppManage) Init(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return resp.ErrorArguments("Incorrect number of arguments. Expecting 1")
	}

	p.initFunctions()
	stub.PutState("test", []byte(args[0]))
	return resp.Success(nil)
}

// Invoke ...
func (p *AppManage) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "init" {
		return p.Init(stub)
	} else if function == "test" {
		return p.test(stub, args)
	}
	return p.exec(stub, function, args)
}

// sdk function <-> chaincode function
func (p *AppManage) initFunctions() {

	// user demo
	p.map_functions = map[string]peerFunc{
		"test":                     p.test,
		"insertTest":               p.insertTest,
		"readTest":                 p.readTest,
		"upgrade":                  p.upgrade,
		"demoUser.Insert":          p.DemoUserMng.Insert,
		"demoUser.Read":            p.DemoUserMng.Read,
		"demoUser.Change":          p.DemoUserMng.Change,
		"demoUser.Delete":          p.DemoUserMng.Delete,
		"demoUser.QueryWithStatus": p.DemoUserMng.QueryWithStatus,
		"demoUser.QueryHistory":    p.DemoUserMng.QueryHistory,

		// 资产详情的增、删、改、查、溯源、rich select
		"assetManagement.Insert":       p.AssetMng.Insert,
		"assetManagement.Delete":       p.AssetMng.Delete,
		"assetManagement.Change":       p.AssetMng.Change,
		"assetManagement.ReadDesc":     p.AssetMng.ReadDesc,
		"assetManagement.TraceHistory": p.AssetMng.TraceHistory,
		"assetManagement.ReadList":     p.AssetMng.ReadList,
		// 资产进度
		"assetProcess.Insert":   p.AssetProe.Insert,
		"assetProcess.Read":     p.AssetProe.Read,
		"assetProcess.Change":   p.AssetProe.Change,
		"assetProcess.History":  p.AssetProe.History,
		"assetProcess.Delete":   p.AssetProe.Delete,
		"assetProcess.ReadList": p.AssetProe.ReadList,

		// 资产信息
		"assets.Insert":       p.Assets.Insert,
		"assets.Delete":       p.Assets.Delete,
		"assets.Change":       p.Assets.Change,
		"assets.ReadDesc":     p.Assets.ReadDesc,
		"assets.ReadList":     p.Assets.ReadList,
		"assets.TraceHistory": p.Assets.TraceHistory,

		// 资产方信息
		"assetSide.Insert":       p.AssetSide.Insert,
		"assetSide.Delete":       p.AssetSide.Delete,
		"assetSide.Change":       p.AssetSide.Change,
		"assetSide.ReadDesc":     p.AssetSide.ReadDesc,
		"assetSide.ReadList":     p.AssetSide.ReadList,
		"assetSide.TraceHistory": p.AssetSide.TraceHistory,
	}
}

/************************************************分界线********************************************************/

func (p *AppManage) exec(stub shim.ChaincodeStubInterface, function string, args []string) peer.Response {
	f, ok := p.map_functions[function]
	if ok {
		logging.Debug("Invoke Success: functiong name [%s]", function)
		return f(stub, args) // 具体执行
	}
	return resp.ErrorNormal("Received unknown function invocation function: " + function)
}

// fabric 网络测试函数
func (p *AppManage) test(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	p.initFunctions()
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}
	res, err := stub.GetState("test")
	if err != nil {
		return resp.ErrorNormal("Failed to get state for " + args[0])
	}

	logging.Debug("TEST: SUCCESS. The chaincode network...")
	return shim.Success(res)
}

func init() {
	logging.SetLogModel(true, true)
}

/***************************************** test ****************************************/
// args: args[0]- function name; args[1]- key; args[2]- value
func (p *AppManage) insertTest(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 2); err != nil {
		return err.(peer.Response)
	}

	if err := stub.PutState(args[0], []byte(args[1])); err != nil {
		return resp.ErrorNormal("insertTest PutState err: " + err.Error())
	}
	return shim.Success(nil)
}

// args: args[0]- function name; args[1]- key
func (p *AppManage) readTest(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if err := filter.CheckParamsLength(args, 1); err != nil {
		return err.(peer.Response)
	}

	res, err := stub.GetState(args[0])
	if err != nil {
		return resp.ErrorNormal("insertTest PutState err: " + err.Error())
	}
	return shim.Success(res)
}

func (p *AppManage) upgrade(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("chaincode upgrade succeed....")
	return shim.Success(nil)
}
