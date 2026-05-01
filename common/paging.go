package common

const (
	DefaultPageNum  = 1
	DefaultPageSize = 20
	DefaultMaxPage  = 10000
)

// OrderDefinition describes a single sort column.
type OrderDefinition struct {
	Column string
	Asc    bool
}

// GetOffset computes the SQL OFFSET for the given page number and page size.
func GetOffset(pageNum, pageSize int) int {
	if pageNum > 0 {
		return pageSize * (pageNum - 1)
	}
	return 0
}

// GetPages computes the total number of pages.
func GetPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	if total%pageSize == 0 {
		return total / pageSize
	}
	return total/pageSize + 1
}
