package pagination

import "strconv"

type Pagination struct {
	Page     int   `form:"page" json:"page"`
	PageSize int   `form:"page_size" json:"page_size"`
	Total    int64 `json:"total"`
}

// SetDefault 设置默认分页参数
func (p *Pagination) SetDefault() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// GetOffset 获取偏移量
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取每页数量
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

// GetPage 获取当前页码
func (p *Pagination) GetPage() int {
	return p.Page
}

// GetTotalPages 获取总页数
func (p *Pagination) GetTotalPages() int {
	if p.Total == 0 || p.PageSize <= 0 {
		return 0
	}
	totalPages := int(p.Total) / p.PageSize
	if int(p.Total)%p.PageSize > 0 {
		totalPages++
	}
	return totalPages
}

// ParseFromQuery 从查询参数解析分页信息
func ParseFromQuery(pageStr, pageSizeStr string) *Pagination {
	p := &Pagination{}

	if page, err := strconv.Atoi(pageStr); err == nil {
		p.Page = page
	}
	if size, err := strconv.Atoi(pageSizeStr); err == nil {
		p.PageSize = size
	}

	p.SetDefault()
	return p
}
