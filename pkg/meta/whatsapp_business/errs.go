package whatsapp_business

import (
	"encoding/json"
	"fmt"
)

type WhatsappErrorCode int

// Authorization errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#authorization-errors
const (
	ERR_AUTH_EXCEPTION       WhatsappErrorCode = 0
	ERR_API_METHOD           WhatsappErrorCode = 3
	ERR_PERMISSION_DENIED    WhatsappErrorCode = 10
	ERR_ACCESS_TOKEN_EXPIRED WhatsappErrorCode = 190
	ERR_API_PERMISSION       WhatsappErrorCode = 200
)

// Integrity errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#integrity-errors
const (
	ERR_BLOCKED_POLICY_VIOLATION   WhatsappErrorCode = 368
	ERR_MESSAGE_RESTRICTED_COUNTRY WhatsappErrorCode = 130497
	ERR_ACCOUNT_LOCKED             WhatsappErrorCode = 131031
)

// Template creation errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#template-creation-errors
const (
	ERR_CHARACTER_LIMIT_EXCEEDED               WhatsappErrorCode = 2388040
	ERR_MESSAGE_HEADER_FORMAT_INCORRECT        WhatsappErrorCode = 2388047
	ERR_MESSAGE_BODY_FORMAT_INCORRECT          WhatsappErrorCode = 2388072
	ERR_MESSAGE_FOOTER_FORMAT_INCORRECT        WhatsappErrorCode = 2388073
	ERR_PARAMETER_WORD_RATIO_EXCEEDED          WhatsappErrorCode = 2388293
	ERR_LEADING_TRAILING_PARAMETER_NOT_ALLOWED WhatsappErrorCode = 2388299
)

// Send Template Errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#send-template-errors
const (
	ERR_MESSAGE_TEMPLATE_LIMIT_EXCEEDED WhatsappErrorCode = 2388019
)

// Phone migration errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#phone-migration-errors
const (
	ERR_PHONE_NUMBER_ALREADY_EXISTS               WhatsappErrorCode = 2388012
	ERR_PHONE_NUMBER_INELIGIBLE_RECEIVE_MIGRATION WhatsappErrorCode = 2388091
	ERR_PHONE_NUMBER_INELIGIBLE_VERIFY_MIGRATION  WhatsappErrorCode = 2388093
	ERR_PHONE_NUMBER_CANNOT_MIGRATE               WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_NOT_ADDED                    WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_NAME_NOT_REGISTERED          WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_NOT_SETUP_PROPERLY           WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_PAYMENT_NOT_FOUND            WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_MIGRATION_ERROR              WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_BELONGS_TO_OTHER_BUSINESS    WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_NOT_APPROVED                 WhatsappErrorCode = 2388103
	ERR_PHONE_NUMBER_MESSAGING_FOR_NOT_APPROVED   WhatsappErrorCode = 2388103
	ERR_ACCOUNT_IS_IN_MAINTENANCE_MODE            WhatsappErrorCode = 2494100
)

// Template insights errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#template-insights-errors
const (
	ERR_TEMPALTE_INSIGHTS_UNAVAILABLE WhatsappErrorCode = 200005
	ERR_TEMPLATE_INSIGHTS_PERSISTENT  WhatsappErrorCode = 200006 // Cannot disable template insights
	ERR_TEMPLATE_INSIGHTS_NOT_ENABLED WhatsappErrorCode = 200007
)

// WhatsApp Business Account errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#whatsapp-business-account-errors
const (
	ERR_WABA_MIGRATION_IN_PROGRESS        WhatsappErrorCode = 2593079
	ERR_WABA_INELIGIBLE_FOR_OBO_MIGRATION WhatsappErrorCode = 2593080
)

// Synchronization errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#synchronization-errors
const (
	ERR_SYNCHRONIZATION_LIMIT_EXCEEDED         WhatsappErrorCode = 2593107
	ERR_SYNCHRONIZATION_OUTSIDE_ALLOWED_WINDOW WhatsappErrorCode = 2593108
)

// Throttling errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#throttling-errors
const (
	ERR_THROTTLE_LIMIT_EXCEEDED                  WhatsappErrorCode = 4
	ERR_THROTTLE_WABA_LIMIT_EXCEEDED             WhatsappErrorCode = 80007
	ERR_THROTTLE_CLOUD_MESSAGE_LIMIT_EXCEEDED    WhatsappErrorCode = 130429
	ERR_THROTTLE_SPAM_LIMIT_EXCEEDED             WhatsappErrorCode = 131048
	ERR_THROTTLE_SPAM_SAME_NUMBER_LIMIT_EXCEEDED WhatsappErrorCode = 131056
	ERR_THROTTLE_REGISTRATION_LIMIT_EXCEEDED     WhatsappErrorCode = 133016
)

// Other errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#other-errors
const (
	ERR_UNKNOWN_API                         WhatsappErrorCode = 1
	ERR_API_DOWN                            WhatsappErrorCode = 2
	ERR_BUSINESS_PHONE_NUMBER_DELETED       WhatsappErrorCode = 33
	ERR_UNSUPPORTED_OR_MISSPELLED_PARAMETER WhatsappErrorCode = 100
	ERR_USER_IN_EXPERIMENT                  WhatsappErrorCode = 130472
	ERR_MESSAGE_NOT_SENT                    WhatsappErrorCode = 131000 // Message failed to send due to an unknown error.
	ERR_ACCESS_DENIED                       WhatsappErrorCode = 131005
	ERR_REQUIRED_PARAMETER_MISSING          WhatsappErrorCode = 131008
	ERR_PARAMETER_NOT_VALID                 WhatsappErrorCode = 131009
	ERR_SERVICE_UNAVAILABLE                 WhatsappErrorCode = 131016
	ERR_SAME_SENDER_RECEIVER                WhatsappErrorCode = 131021
	ERR_MESSAGE_UNDELIVERABLE               WhatsappErrorCode = 131026
	ERR_DISPLAY_NAME_NEED_APPROVAL          WhatsappErrorCode = 131037
	ERR_PAYMENT_METHOD_ERROR                WhatsappErrorCode = 131042
	ERR_PHONE_NUMBER_REGISTRATION_ERROR     WhatsappErrorCode = 131045
	// ...
)

// Marketing Messages API for WhatsApp Error codes
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#marketing-messages-api-for-whatsapp-error-codes
const (
	ERR_INVALID_PARAMETER WhatsappErrorCode = 100
)

type WhatsAppBusinessError struct {
	ErrorData struct {
		Message   string            `json:"message"`
		Type      string            `json:"type"`
		Code      WhatsappErrorCode `json:"code"`
		ErrorData struct {
			MessagingProduct string `json:"messaging_product"`
			Details          string `json:"details"`
		} `json:"error_data"`
		ErrorSubcode int    `json:"error_subcode"`
		FbtraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

func (v WhatsAppBusinessError) Error() string {
	return v.ErrorData.Message
}

func parseMetaErrorResponse[T any](emptyResponse T, body []byte, httpCode int) (T, int, error) {
	var responseError WhatsAppBusinessError
	if err := json.Unmarshal(body, &responseError); err != nil {
		return emptyResponse, httpCode, fmt.Errorf("unexpected http code: %d", httpCode)
	}
	return emptyResponse, httpCode, responseError
}
