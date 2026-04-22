package filter_request

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	DEFAULT_SORT_BY    = "created_at"
	DEFAULT_SORT_ORDER = "desc"
)

// Sort represents the sorting parameters for an API request, including the field to sort by and the sort order (ascending or descending). It includes validation logic to ensure that the sort order is either "asc" or "desc", and a GORM scope function to apply the sorting parameters to a database query.
type Sort struct {
	SortBy    string `json:"sort_by" form:"sort_by" query:"sort_by"`
	SortOrder string `json:"sort_order" form:"sort_order" query:"sort_order"`
}

// Validate checks the Sort struct for valid values. It ensures that the SortOrder is either "asc" or "desc", and returns a map of errors if any validation rules are violated. This method is used to validate sorting parameters in API requests before applying them to database queries.
func (r Sort) Validate() map[string]string {
	errors := make(map[string]string)
	if r.SortOrder != "" && strings.ToLower(r.SortOrder) != "asc" && strings.ToLower(r.SortOrder) != "desc" {
		errors["sort_order"] = "must be 'asc' or 'desc'"
	}
	return errors
}

// SortScope takes a Sort struct and returns a GORM scope function that applies the corresponding ORDER BY clause to a database query based on the specified sorting parameters. This allows for dynamic sorting of database queries based on client-provided sorting parameters in API requests.
func SortScope(sort Sort) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		dir := strings.ToLower(sort.SortOrder)
		if dir != "asc" && dir != "desc" {
			dir = "asc"
		}
		return db.Order(fmt.Sprintf("%s %s", sort.SortBy, dir))
	}
}
