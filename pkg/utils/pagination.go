package utils

import "strconv"

type Pagination struct {
	Page       int64       `json:"page"`
	Limit      int64       `json:"limit"`
	TotalCount int64       `json:"total_count"`
	Result     interface{} `json:"result"`
}

func NewPagination(p, l string) *Pagination {
	page, _ := strconv.ParseInt(p, 10, 32)
	limit, _ := strconv.ParseInt(l, 10, 32)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	return &Pagination{
		Page:  page,
		Limit: limit,
	}
}
