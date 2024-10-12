package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/users"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
)

func UserList(raw json.RawMessage) any {
	var p provider.Condition
	_ = json.Unmarshal(raw, &p)
	res := utils.Pagination{
		Page:  p.Page,
		Limit: p.PageSize,
	}
	res.TotalCount, res.Result, _ = users.Traversal(p)
	return res
}