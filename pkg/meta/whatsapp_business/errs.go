package whatsapp_business

type WhatsAppBusinessError struct {
	Error struct {
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

type WhatsappErrorCode int

// Authorization errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#authorization-errors
const (
	AUTH_EXCEPTION       WhatsappErrorCode = 0
	API_METHOD           WhatsappErrorCode = 3
	PERMISSION_DENIED    WhatsappErrorCode = 10
	ACCESS_TOKEN_EXPIRED WhatsappErrorCode = 190
	API_PERMISSION       WhatsappErrorCode = 200
)

// Integrity errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#integrity-errors
const (
	BLOCKED_POLICY_VIOLATION   WhatsappErrorCode = 368
	MESSAGE_RESTRICTED_COUNTRY WhatsappErrorCode = 130497
	ACCOUNT_LOCKED             WhatsappErrorCode = 131031
)

// Template creation errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#template-creation-errors
const (
	CHARACTER_LIMIT_EXCEEDED               WhatsappErrorCode = 2388040
	MESSAGE_HEADER_FORMAT_INCORRECT        WhatsappErrorCode = 2388047
	MESSAGE_BODY_FORMAT_INCORRECT          WhatsappErrorCode = 2388072
	MESSAGE_FOOTER_FORMAT_INCORRECT        WhatsappErrorCode = 2388073
	PARAMETER_WORD_RATIO_EXCEEDED          WhatsappErrorCode = 2388293
	LEADING_TRAILING_PARAMETER_NOT_ALLOWED WhatsappErrorCode = 2388299
)

// Send Template Errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#send-template-errors
const (
	MESSAGE_TEMPLATE_LIMIT_EXCEEDED WhatsappErrorCode = 2388019
)

// Phone migration errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#phone-migration-errors
const (
	PHONE_NUMBER_ALREADY_EXISTS         WhatsappErrorCode = 2388012
	PHONE_NUMBER_NOT_ELIGIBLE_MIGRATION WhatsappErrorCode = 2388091
	PHONE_NUMBER_NOT_MIGRATABLE         WhatsappErrorCode = 2388103
)

// Marketing Messages API for WhatsApp Error codes
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#marketing-messages-api-for-whatsapp-error-codes
const (
	INVALID_PARAMETER WhatsappErrorCode = 100
)
