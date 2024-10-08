package utils

type Pagination struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total_count"`
	Data       interface{} `json:"data"`
}

func NewPagination(page, limit int) *Pagination {
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
