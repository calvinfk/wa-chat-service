package filter_request

import (
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"gorm.io/gorm"
)

// Operator defines the type for supported filter operators in query parameters. These operators allow clients to specify how they want to filter results when making API requests.
type Operator string

const (
	OpEq    Operator = "eq"
	OpNeq   Operator = "neq"
	OpGt    Operator = "gt"
	OpGte   Operator = "gte"
	OpLt    Operator = "lt"
	OpLte   Operator = "lte"
	OpLike  Operator = "like"
	OpIlike Operator = "ilike"
	OpIn    Operator = "in"
)

// Filter represents a single filter condition that can be applied to a database query.
type Filter struct {
	Field    string
	Operator Operator
	Value    any
}

// Validatable is an interface that defines a method for validating the fields of a struct. Any struct that implements this interface can be used as the SpecificFilter in a FilterRequest, allowing it to perform custom validation logic and return any validation errors in a standardized format.
func ParseOperator(opStr string) (Operator, error) {
	var op Operator
	switch strings.ToLower(opStr) {
	case "eq":
		op = OpEq
	case "neq":
		op = OpNeq
	case "gt":
		op = OpGt
	case "gte":
		op = OpGte
	case "lt":
		op = OpLt
	case "lte":
		op = OpLte
	case "like":
		op = OpLike
	case "ilike":
		op = OpIlike
	case "in":
		op = OpIn
	}
	if op == "" {
		return "", ErrInvalidOperator
	}
	return op, nil
}

// parseFilterValue takes a filter value string, splits it into operator and value parts, and returns the corresponding Operator and value. If the operator is not specified, it defaults to OpEq. This function is used to process filter values from query parameters in API requests.
func parseFilterValue(value string) (Operator, any, error) {
	parts := strings.Split(value, ":")
	if len(parts) == 1 {
		return OpEq, value, nil
	}
	op, err := ParseOperator(parts[0])
	if err != nil {
		return "", nil, err
	}
	var val any = parts[1]
	// If the operator is "in", we need to split the value into a slice of values
	if op == OpIn {
		cleanStr := strings.TrimSpace(parts[1])
		cleanStr = strings.TrimPrefix(cleanStr, "[")
		cleanStr = strings.TrimSuffix(cleanStr, "]")
		val = strings.Split(cleanStr, ",")
	}
	return op, val, nil
}

func parseOperatorToFirestoreCondition(op Operator) string {
	switch op {
	case OpEq:
		return "=="
	case OpNeq:
		return "!="
	case OpGt:
		return ">"
	case OpGte:
		return ">="
	case OpLt:
		return "<"
	case OpLte:
		return "<="
	case OpIn:
		return "in"
	default:
		return ""
	}
}

func parseOperatorToMeiliCondition(op Operator) string {
	switch op {
	case OpEq:
		return "="
	case OpNeq:
		return "!="
	case OpGt:
		return ">"
	case OpGte:
		return ">="
	case OpLt:
		return "<"
	case OpLte:
		return "<="
	case OpIn:
		return "IN"
	case OpLike, OpIlike:
		return "CONTAINS"
	default:
		return ""
	}
}

func parseSortOrder(order string) firestore.Direction {
	if strings.ToLower(order) == "desc" {
		return firestore.Desc
	}
	return firestore.Asc
}

// FilterScope takes a slice of Filter structs and returns a GORM scope function that applies the corresponding WHERE conditions to a database query based on the specified filters. This allows for dynamic filtering of database queries based on client-provided filter parameters in API requests.
func FilterScope(filters []Filter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, filter := range filters {
			var condition string
			switch filter.Operator {
			case OpEq:
				condition = fmt.Sprintf("%s = ?", filter.Field)
			case OpNeq:
				condition = fmt.Sprintf("%s <> ?", filter.Field)
			case OpGt:
				condition = fmt.Sprintf("%s > ?", filter.Field)
			case OpGte:
				condition = fmt.Sprintf("%s >= ?", filter.Field)
			case OpLt:
				condition = fmt.Sprintf("%s < ?", filter.Field)
			case OpLte:
				condition = fmt.Sprintf("%s <= ?", filter.Field)
			case OpLike:
				condition = fmt.Sprintf("%s LIKE ?", filter.Field)
				filter.Value = fmt.Sprintf("%%%v%%", filter.Value)
			case OpIlike:
				condition = fmt.Sprintf("%s ILIKE ?", filter.Field)
				filter.Value = fmt.Sprintf("%%%v%%", filter.Value)
			default:
				continue
			}
			db = db.Where(condition, filter.Value)
		}
		return db
	}
}
