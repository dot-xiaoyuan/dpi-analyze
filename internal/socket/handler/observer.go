package handler

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture/observer"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/provider"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/types"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/utils"
)

func Observer(raw json.RawMessage) any {
	var p provider.Condition
	_ = json.Unmarshal(raw, &p)
	res := utils.Pagination{
		Page:  p.Page,
		Limit: p.PageSize,
	}
	switch p.Type {
	case types.TTL:
		res.TotalCount, res.Result, _ = observer.TTLObserver.Traversal(p)
		break
	case types.Mac:
		res.TotalCount, res.Result, _ = observer.MacObserver.Traversal(p)
		break
	case types.UserAgent:
		res.TotalCount, res.Result, _ = observer.UaObserver.Traversal(p)
		break
	}
	return res
}
