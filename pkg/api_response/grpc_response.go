package api_response

import (
	"wa_chat_service/pkg/errs"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// https://grpc.io/docs/guides/status-codes/
func NewGRPCErrorResponse(serverError bool, err error) error {
	var code codes.Code
	var msg string
	if serverError {
		code = codes.Internal
		msg = "Internal Server Error"
	} else {
		if errs.IsUnauthenticatedError(err) {
			code = codes.Unauthenticated
		} else {
			switch err {
			case errs.ErrGenericAlreadyExists:
				code = codes.AlreadyExists
			case errs.ErrGenericForbidden:
				code = codes.PermissionDenied
			default:
				code = codes.InvalidArgument
			}
			msg = err.Error()
		}
	}
	return status.Error(code, msg)
}
