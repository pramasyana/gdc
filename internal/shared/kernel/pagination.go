package kernel

// Pagination describes an offset/limit window.
type Pagination struct {
	Page  int
	Limit int
}

// Offset returns the SQL offset for the page.
func (p Pagination) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}
