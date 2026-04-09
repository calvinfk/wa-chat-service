package filter_request

import (
	"context"
	"fmt"
	"maps"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DEFAULT_PAGE_SIZE = 100
	DEFAULT_PAGE      = 1
)

// Validatable is an interface that defines a method for validating the fields of a struct. Any struct that implements this interface can be used as the SpecificFilter in a FilterRequest, allowing it to perform custom validation logic and return any validation errors in a standardized format.
type Validatable interface {
	Validate() map[string]string
}

type (
	// FilterRequest is a generic struct that represents a request for filtering, sorting, and paginating results in an API endpoint. It includes a SpecificFilter of a generic type T that must implement the Validatable interface, allowing for custom validation logic specific to the type of filter being used. The struct also includes embedded Paginate and Sort structs to handle pagination and sorting parameters in a standardized way.
	FilterRequest[T Validatable] struct {
		SpecificFilter T
		Paginate
		Sort
	}

	// FilterResponse is a generic struct that represents the response for a filtered API request. It includes a Result slice of a generic type T, as well as pagination information such as the current page, page size, total pages, and total items. This struct is used to standardize the format of responses for filtered requests across different endpoints in the API.
	FilterResponse[T any] struct {
		Results    []T   `json:"results"`
		Page       int   `json:"page"`
		PageSize   int   `json:"page_size"`
		TotalPages int   `json:"total_pages"`
		TotalItems int64 `json:"total_items"`
	}
)

func (r *FilterRequest[T]) Validate() map[string]string {
	errors := make(map[string]string)
	maps.Copy(errors, r.SpecificFilter.Validate())
	maps.Copy(errors, r.Paginate.Validate())
	maps.Copy(errors, r.Sort.Validate())
	if len(errors) > 0 {
		return errors
	}
	return nil
}

// isEmpty checks if a given string is empty, "null", or equal to the string representation of a nil UUID. This function is used to determine whether a filter value should be considered as not provided or invalid when processing filter requests.
func isEmpty(str string) bool {
	return str == "" || str == "null" || str == uuid.Nil.String()
}

// NewFilterResponse creates a new instance of FilterResponse based on the provided result slice, pagination information, and total item count. It calculates the total number of pages based on the total items and page size, and returns a structured response that can be used to standardize the format of API responses for filtered requests.
func NewFilterResponse[T any](results []T, paginate Paginate, totalItems int64) FilterResponse[T] {
	r := FilterResponse[T]{
		Page:       paginate.Page,
		TotalItems: totalItems,
		PageSize:   paginate.PageSize,
	}
	if len(results) == 0 {
		r.Results = nil
	} else {
		r.Results = results
	}
	if r.PageSize > 0 {
		r.TotalPages = int((totalItems + int64(r.PageSize) - 1) / int64(r.PageSize))
	}
	return r
}

// InitializeFilter takes a filter request struct, converts it to a map, and processes the map to create a slice of Filter structs.
func InitializeFilter[T Validatable](filterRequest FilterRequest[T], allowedFilterFields []string, allowedSortFields []string) ([]Filter, Sort, Paginate, error) {
	var filters []Filter
	var sort Sort = filterRequest.Sort
	var paginate Paginate = filterRequest.Paginate
	if isEmpty(sort.SortBy) {
		sort.SortBy = DEFAULT_SORT_BY
	}
	if isEmpty(sort.SortOrder) {
		sort.SortOrder = DEFAULT_SORT_ORDER
	}
	sortFieldMap := map[string]bool{}
	for _, f := range allowedSortFields {
		sortFieldMap[f] = true
	}
	if !sortFieldMap[sort.SortBy] {
		return nil, filterRequest.Sort, filterRequest.Paginate, fmt.Errorf("invalid sort_by field: %s", sort.SortBy)
	}
	// TODO: implement get all
	if paginate.Page <= 0 {
		paginate.Page = DEFAULT_PAGE
	}
	if paginate.PageSize <= 0 {
		paginate.PageSize = DEFAULT_PAGE_SIZE
	}
	mapStruct, err := utils.StructToMap(filterRequest.SpecificFilter, false)
	if err != nil {
		return nil, filterRequest.Sort, filterRequest.Paginate, err
	}
	filterFieldMap := map[string]bool{}
	for _, f := range allowedFilterFields {
		filterFieldMap[f] = true
	}
	for key, value := range mapStruct {
		if !filterFieldMap[key] {
			continue
		}
		if value == nil {
			continue
		}
		strValue := fmt.Sprintf("%v", value)
		op, val, err := parseFilterValue(strValue)
		if err != nil {
			return nil, filterRequest.Sort, filterRequest.Paginate, err
		}
		filters = append(filters, Filter{
			Field:    key,
			Operator: op,
			Value:    val,
		})
	}
	return filters, sort, paginate, nil
}

func ApplyFilterGorm[T any](query *gorm.DB, data *[]T, filters []Filter, paginate Paginate, sort Sort) (int64, error) {
	var totalData int64
	query.Scopes(
		FilterScope(filters),
	)
	if err := query.Count(&totalData).Error; err != nil {
		return 0, err
	}
	query.Scopes(
		SortScope(sort),
		PaginateScope(paginate),
	)
	err := query.Find(data).Error
	if err != nil {
		return 0, err
	}
	return totalData, nil
}

func ApplyFilterFirestore(ctx context.Context, query firestore.Query, filters []Filter, sort Sort, paginate Paginate) ([]*firestore.DocumentSnapshot, int64, error) {
	for _, filter := range filters {
		query = query.Where(filter.Field, parseOperatorToFirestoreCondition(filter.Operator), filter.Value)
	}
	countResult, err := query.NewAggregationQuery().WithCount("all").Get(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, ok := countResult["all"]
	if !ok {
		return nil, 0, ErrInvalidCountAlias
	}
	totalData := count.(*firestorepb.Value).GetIntegerValue()
	startsAt := (paginate.Page - 1) * paginate.PageSize
	page := query.OrderBy(
		sort.SortBy, parseSortOrder(sort.SortOrder),
	).Offset(startsAt).Limit(paginate.PageSize).Documents(ctx)
	docs, err := page.GetAll()
	if err != nil {
		return nil, 0, err
	}
	// for _, doc := range docs {
	// 	var data T
	// 	docData := doc.Data()
	// 	log.Println("[DEBUG][ApplyFilterFirestore] docData:", docData)
	// 	docData[firestore.DocumentID] = doc.Ref.ID
	// 	err := utils.MapToStruct(docData, &data)
	// 	if err != nil {
	// 		return nil, 0, err
	// 	}
	// 	*results = append(*results, data)
	// }
	return docs, totalData, nil
}
