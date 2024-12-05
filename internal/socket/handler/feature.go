package handler

import (
	"encoding/json"
	"errors"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/application"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands_keyword"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/components/features/brands_root"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/loader"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"net/http"
)

func FeatureLibrary(raw json.RawMessage) any {
	type Feature struct {
		Name    string           `json:"name"`
		Count   int              `json:"count"`
		Version string           `json:"version"`
		Module  string           `json:"module"`
		History []loader.History `json:"history"`
	}

	res := []Feature{
		{
			Name:    "应用特征",
			Count:   len(application.Feature),
			Version: application.LoaderManger.Version(),
			Module:  "application",
			History: application.LoaderManger.History(),
		},
		{
			Name:    "品牌特征",
			Count:   len(brands.Manager.Feature),
			Version: brands.Manager.Loader.Version(),
			Module:  "brands",
			History: brands.Manager.Loader.History(),
		},
		{
			Name:    "品牌关键词特征",
			Count:   len(brands_keyword.Manager.Feature),
			Version: brands_keyword.Manager.Loader.Version(),
			Module:  "brands_keyword",
			History: brands_keyword.Manager.Loader.History(),
		},
		{
			Name:    "品牌根域名特征",
			Count:   len(brands_root.Manager.Feature),
			Version: brands_root.Manager.Loader.Version(),
			Module:  "brands_root",
			History: brands_root.Manager.Loader.History(),
		},
	}

	return res
}

// FeatureUpdate 特征更新
func FeatureUpdate(raw json.RawMessage) any {
	var req struct {
		Filepath string `json:"filepath"`
		Module   string `json:"module"`
	}
	res := &models.Response{
		Code: http.StatusBadRequest,
	}
	err := json.Unmarshal(raw, &req)
	if err != nil {
		res.Message = err.Error()
		return res
	}
	if req.Filepath == "" {
		return "file path is required"
	}

	switch req.Module {
	case "application":
		err = application.Update(req.Filepath)
		break
	case "brands":
		err = brands.Manager.Update(req.Filepath)
		break
	case "brands_keyword":
		err = brands_keyword.Manager.Update(req.Filepath)
		break
	case "brands_root":
		err = brands_root.Manager.Update(req.Filepath)
		break
	default:
		err = errors.New("invalid module")
		break
	}
	if err != nil {
		res.Message = err.Error()
		return res
	}
	res.Code = http.StatusOK
	res.Message = "update successful!"
	return res
}
