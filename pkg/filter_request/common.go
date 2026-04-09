package filter_request

import (
	"strconv"
	"strings"
	"time"
	"wa_chat_service/pkg/utils"

	"github.com/google/uuid"
)

var (
	DEFAULT_OPERATOR = OpEq
)

// Operator defines the type for supported filter operators in query parameters. These operators allow clients to specify how they want to filter results when making API requests.
type QueryFilterString string

func (n QueryFilterString) GetOperator() Operator {
	split := strings.SplitN(string(n), ":", 2)
	op, err := ParseOperator(split[0])
	if err != nil {
		return DEFAULT_OPERATOR
	}
	return op
}
func (n QueryFilterString) GetValue() string {
	split := strings.SplitN(string(n), ":", 2)
	if len(split) < 2 {
		_, err := ParseOperator(split[0])
		if err != nil {
			return string(n)
		}
		return ""
	}
	return split[1]
}
func (n QueryFilterString) IsEmpty() bool {
	return n.GetValue() == ""
}

// Operator defines the type for supported filter operators in query parameters. These operators allow clients to specify how they want to filter results when making API requests.
type QueryFilterNumber string

func (n QueryFilterNumber) IsEmpty() bool {
	return n == ""
}
func (n QueryFilterNumber) IsValid() bool {
	_, err := strconv.Atoi(QueryFilterString(n).GetValue())
	return err == nil
}
func (n QueryFilterNumber) GetNumber() int {
	num, _ := strconv.Atoi(QueryFilterString(n).GetValue())
	return num
}

// Operator defines the type for supported filter operators in query parameters. These operators allow clients to specify how they want to filter results when making API requests.
type QueryFilterEmail string

func (e QueryFilterEmail) IsEmpty() bool {
	return e == ""
}
func (e QueryFilterEmail) IsValid() bool {
	return utils.ValidateEmail(QueryFilterString(e).GetValue())
}
func (e QueryFilterEmail) GetEmail() string {
	return QueryFilterString(e).GetValue()
}

// Operator defines the type for supported filter operators in query parameters. These operators allow clients to specify how they want to filter results when making API requests.
type QueryFilterUUID string

func (u QueryFilterUUID) IsEmpty() bool {
	return u == "" || u == QueryFilterUUID(uuid.Nil.String())
}
func (u QueryFilterUUID) IsValid() bool {
	_, err := uuid.Parse(QueryFilterString(u).GetValue())
	return err == nil
}
func (u QueryFilterUUID) GetUUID() uuid.UUID {
	id, err := uuid.Parse(QueryFilterString(u).GetValue())
	if err != nil {
		panic("invalid UUID format: " + string(u))
	}
	return id
}

type QueryFilterDateTime string

func (d QueryFilterDateTime) IsEmpty() bool {
	return d == ""
}
func (d QueryFilterDateTime) IsValid() bool {
	_, err := time.Parse(time.RFC3339, QueryFilterString(d).GetValue())
	return err == nil
}
