package main

import (
	"encoding/json"
	"fmt"
	. "iht-fabric-chaincode-private/chaincode/go/app/controllers/assetManage"
	"testing"

	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//----------------------------------init----------------------------------//

func _init(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

//func TestApp_init(t *testing.T) {
//	scc := new(IPropertyManage)
//	stub := shim.NewMockStub("ex02", scc)
//
//	bs := [][]byte{[]byte("init"), []byte("200-A1235")}
//	_init(t, stub, bs)
//}

//----------------------------------invoke----------------------------------//

func invoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Error Invoke: failed", string(res.Message))
		t.FailNow()
	}
	fmt.Println(string(args[0]), "::::invoke functions success  ", res.Message, res.Status, string(res.Payload), res.GetMessage())
}

func TestApp_invoke(t *testing.T) {
	scc := new(AppManage)
	stub := shim.NewMockStub("ex02", scc)

	// Init
	_init(t, stub, [][]byte{[]byte("init"), []byte("A567")})
	invoke(t, stub, [][]byte{[]byte("test"), []byte("test-key")})

	// "userTestQuery"  -> "QueryTest"
	// "userTestInsert" -> "InsertTest"
	//invoke(t, stub, [][]byte{[]byte("userTestInsert"), []byte("A1235"), []byte("suchangqin@gmail.com"), []byte("user stub test")})
	//invoke(t, stub, [][]byte{[]byte("demoUser.Insert"), []byte("")})
	//invoke(t, stub, [][]byte{[]byte("demoUser.Read"), []byte("")})

	// "userInit" -> "InitUserInfo"
	// "userRead" -> "ReadUserInfo"
	invoke(t, stub, [][]byte{[]byte("demoUser.Insert"), []byte("xuxiaodong@gmail.com"), []byte("kity-xu"), []byte("HASH_PASSWORD")})
	invoke(t, stub, [][]byte{[]byte("demoUser.Read"), []byte("xuxiaodong@gmail.com")})

	// "userChange" -> "ChangeUserInfo"
	// "userDelete" -> "DeleteUserinfo"
	invoke(t, stub, [][]byte{[]byte("demoUser.Insert"), []byte("xuxiaodong2@gmail.com"), []byte("xu"), []byte("HASH_PASSWORD222")})   // put
	invoke(t, stub, [][]byte{[]byte("demoUser.Read"), []byte("xuxiaodong2@gmail.com"), []byte("df")})                                 // get
	invoke(t, stub, [][]byte{[]byte("demoUser.Change"), []byte("xuxiaodong2@gmail.com"), []byte("lang"), []byte("HASH_PASSWORD333")}) // change
	invoke(t, stub, [][]byte{[]byte("demoUser.Read"), []byte("xuxiaodong2@gmail.com")})                                               // get
	invoke(t, stub, [][]byte{[]byte("demoUser.Delete"), []byte("xuxiaodong2@gmail.com")})                                             // delete
	invoke(t, stub, [][]byte{[]byte("demoUser.Read"), []byte("xuxiaodong2@gmail.com")})

	// AssetManage test

	assetOwnerID := "05"
	str := createJson(assetOwnerID)
	invoke(t, stub, [][]byte{[]byte("assetManagement.Insert"), str})                                                                                                               // get
	invoke(t, stub, [][]byte{[]byte("assetManagement.ReadDesc"), []byte(assetOwnerID)})                                                                                            // get
	invoke(t, stub, [][]byte{[]byte("assetManagement.ReadList"), []byte("{\"docTag\":\"AssetData_IHT_\", \"profile.assetOwnerInfo.id\": \"Owner_ID\",\"profile.assetType\": 2}")}) // get

	// "userQueryWithStatus" -> "QueryUserInfoByStatus"
	// "userQueryHistory"    -> "GetHistoryForUserInfo"

	// 用户状态和历史记录，由于chaincode没部署上链，查不到区块信息
	// invoke(t, stub, [][]byte{[]byte("demoUser.QueryWithStatus"), []byte("99")}) // 99
	// invoke(t, stub, [][]byte{[]byte("demoUser.QueryHistory"), []byte("xuxiaodong2@gmail.com")})
}

// 资产数据json串生成
func createJson(ID string) []byte {
	var mems []Members
	for i, v := range []string{"xiaodong", "yanliang", "xiaosu"} {

		m := Members{
			ID:      "00" + strconv.Itoa(i),
			Name:    v,
			Duty:    "董事长",
			Photo:   DocumentStore{Path: "/use/iht/fabric/file/", Hash: "HASH_AASDBCSKKN"},
			Profile: "都是牛逼人",
		}
		mems = append(mems, m)
	}

	//// 资产简介
	//profile := AssetModelProfile{
	//	OwnerID:                   "Owner_ID",
	//	OwnerName:                 "孔雀城钻石湾",
	//	AssetType:                 2,
	//	AssetID:                   ID,
	//	AssetPhoto:                getPhoto(),
	//	AssetName:                 "5楼508",
	//	AssetOwnerFinancingAmount: 8977989,
	//	AssetSplitNum:             1000,
	//	MemberDiscount:            0.1,
	//	IhtAirdropBounty:          0.2,
	//	SubscribeNum:              860,
	//	StartDate:                 20180523,
	//	EndDate:                   20211124,
	//}
	// 基类
	//binfo := db.DataBlockBase{
	//	DocTag:     "IHT_TAG",
	//	DBProfile:  "可用",
	//	CreateTime: mtime.Now(),
	//}

	//资产详情
	res := &AssetParticulars{
		//BaseInfo:                  binfo,
		OwnerID:                   "Owner_ID",
		OwnerName:                 "孔雀城钻石湾",
		AssetType:                 2,
		AssetID:                   ID,
		AssetPhoto:                getPhoto(),
		AssetName:                 "5楼508",
		AssetOwnerFinancingAmount: 8977989,
		AssetSplitNum:             1000,
		MemberDiscount:            0.1,
		IhtAirdropBounty:          0.2,
		SubscribeNum:              860,
		StartDate:                 20180523,
		EndDate:                   20211124,
		LeastSubscribe:            0.5,
		OtherDiscount:             "无",
		AssetDesc:                 "买了你就发了",
		AssetQualifiedFile:        DocumentStore{"文件加密", "HANS_文件加密"},
		WhitePaper:                DocumentStore{"白皮书加密", "HANS_白皮书加密"},
		TeamMembers:               mems,
	}

	data, _ := json.Marshal(res)
	return data
}

func getPhoto() []DocumentStore {
	var docs []DocumentStore
	for range []int{0, 1, 2} {
		doc := DocumentStore{
			Path: "/use/local/images",
			Hash: "HASH_ImageASBDM",
		}
		docs = append(docs, doc)
	}
	return docs
}
