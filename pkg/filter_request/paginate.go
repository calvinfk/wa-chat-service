package filter_request

import "gorm.io/gorm"

const (
	DEFAULT_LIMIT = 100
	MAX_LIMIT     = 100
)

// Paginate represents the pagination parameters for an API request, including the page number and page size. It includes validation logic to ensure that the page and page size values are within acceptable ranges, and a GORM scope function to apply the pagination parameters to a database query.
type Paginate struct {
	Page     int `json:"page" form:"page" query:"page"`                // -1 for all data (not implemented)
	PageSize int `json:"page_size" form:"page_size" query:"page_size"` // -1 for all data (not implemented)
}

// Sort represents the sorting parameters for an API request, including the field to sort by and the sort order (ascending or descending). It includes validation logic to ensure that the sort order is either "asc" or "desc", and a GORM scope function to apply the sorting parameters to a database query.
func (r Paginate) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Page < -1 {
		errors["page"] = "page must be greater than or equal to -1"
	}
	if r.PageSize < -1 {
		errors["page_size"] = "page_size must be greater than or equal to -1"
	}
	return errors
}

// PaginateScope takes a Paginate struct and returns a GORM scope function that applies the corresponding OFFSET and LIMIT clauses to a database query based on the specified pagination parameters. This allows for dynamic pagination of database queries based on client-provided pagination parameters in API requests.
func PaginateScope(p Paginate) func(*gorm.DB) *gorm.DB {
	offset := (p.Page - 1) * p.PageSize
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(p.PageSize)
	}
}
