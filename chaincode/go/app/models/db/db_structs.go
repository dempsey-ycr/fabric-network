package db

// 每条入链的数据都会有该属性，标识该数据处于的状态
type DataBlockBase struct {
	DocTag     string `json:"docTag"`
	DBType     int    `json:"dbType"`     // 状态类型
	DBProfile  string `json:"dbProfile"`  // 状态描述
	CreateTime string `json:"createTime"` // 创建时间
	UpdateTime string `json:"updateTime"` // 更新时间
}

func (b *DataBlockBase) GetDBType() int {
	return b.DBType
}

func (b *DataBlockBase) GetDBProfile() string {
	return b.DBProfile
}

// demo_user WorldState
type WorldState struct {
	StateRecord interface{} `json:"Record"`
	StateKey    string      `json:"Key"`
	Namespace   string      `json:"namespace"`
}

// demo_user history
type History struct {
	IsDelete  bool        `json:"IsDelete"`
	TxId      string      `json:"TxId"`
	Value     interface{} `json:"Value"`
	Timestamp string      `json:"Timestamp"`
}

// list
type List struct {
	Object interface{} `json:"list"`
}
