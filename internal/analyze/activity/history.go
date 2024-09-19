package activity

import (
	"sync"
)

// IP 活动记录

var IPTables sync.Map

// 新增/更新 IP表
