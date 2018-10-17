package controllers

// 共识审核
const (
	ST_COMM_INIT      string = "00" // init
	ST_COMM_APPROVING string = "01" // 01-正在审核
	ST_COMM_APPROVED  string = "02" // 02-审核通过
	ST_COMM_REJECTED  string = "03" // 03-审核不通过
	ST_COMM_NILED     string = "99" // 99-作废
)

const (
	ST_CONSENSUS_INIT      = 0  // init
	ST_CONSENSUS_APPROVING = 1  // 01-正在审核
	ST_CONSENSUS_APPROVED  = 2  // 02-审核通过
	ST_CONSENSUS_REJECTED  = 3  // 03-审核不通过
	ST_CONSENSUS_NILED     = 99 // 99-作废
)

var CONSENSUS_DESC = []string{
	"数据初始化",
	"数据正在审核",
	"数据审核通过",
	"数据审核不通过",
}

// iht 资产 doctag
const (
	KEY_ASSETDATA_TAG = "AssetData_" // 资产数据基准key
	KEY_ORGANIZATION  = "IHT_"       // 组织机构
)
