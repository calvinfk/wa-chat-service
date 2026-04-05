package whatsapp_business

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
	PHONE_NUMBER_ALREADY_EXISTS               WhatsappErrorCode = 2388012
	PHONE_NUMBER_INELIGIBLE_RECEIVE_MIGRATION WhatsappErrorCode = 2388091
	PHONE_NUMBER_INELIGIBLE_VERIFY_MIGRATION  WhatsappErrorCode = 2388093
	PHONE_NUMBER_CANNOT_MIGRATE               WhatsappErrorCode = 2388103
	PHONE_NUMBER_NOT_ADDED                    WhatsappErrorCode = 2388103
	PHONE_NUMBER_NAME_NOT_REGISTERED          WhatsappErrorCode = 2388103
	PHONE_NUMBER_NOT_SETUP_PROPERLY           WhatsappErrorCode = 2388103
	PHONE_NUMBER_PAYMENT_NOT_FOUND            WhatsappErrorCode = 2388103
	PHONE_NUMBER_MIGRATION_ERROR              WhatsappErrorCode = 2388103
	PHONE_NUMBER_BELONGS_TO_OTHER_BUSINESS    WhatsappErrorCode = 2388103
	PHONE_NUMBER_NOT_APPROVED                 WhatsappErrorCode = 2388103
	PHONE_NUMBER_MESSAGING_FOR_NOT_APPROVED   WhatsappErrorCode = 2388103
	ACCOUNT_IS_IN_MAINTENANCE_MODE            WhatsappErrorCode = 2494100
)

// Template insights errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#template-insights-errors
const (
	TEMPALTE_INSIGHTS_UNAVAILABLE WhatsappErrorCode = 200005
	TEMPLATE_INSIGHTS_PERSISTENT  WhatsappErrorCode = 200006 // Cannot disable template insights
	TEMPLATE_INSIGHTS_NOT_ENABLED WhatsappErrorCode = 200007
)

// WhatsApp Business Account errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#whatsapp-business-account-errors
const (
	WABA_MIGRATION_IN_PROGRESS        WhatsappErrorCode = 2593079
	WABA_INELIGIBLE_FOR_OBO_MIGRATION WhatsappErrorCode = 2593080
)

// Synchronization errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#synchronization-errors
const (
	SYNCHRONIZATION_LIMIT_EXCEEDED         WhatsappErrorCode = 2593107
	SYNCHRONIZATION_OUTSIDE_ALLOWED_WINDOW WhatsappErrorCode = 2593108
)

// Throttling errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#throttling-errors
const (
	THROTTLE_LIMIT_EXCEEDED                  WhatsappErrorCode = 4
	THROTTLE_WABA_LIMIT_EXCEEDED             WhatsappErrorCode = 80007
	THROTTLE_CLOUD_MESSAGE_LIMIT_EXCEEDED    WhatsappErrorCode = 130429
	THROTTLE_SPAM_LIMIT_EXCEEDED             WhatsappErrorCode = 131048
	THROTTLE_SPAM_SAME_NUMBER_LIMIT_EXCEEDED WhatsappErrorCode = 131056
	THROTTLE_REGISTRATION_LIMIT_EXCEEDED     WhatsappErrorCode = 133016
)

// Other errors
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#other-errors
const (
	UNKNOWN_API                         WhatsappErrorCode = 1
	API_DOWN                            WhatsappErrorCode = 2
	BUSINESS_PHONE_NUMBER_DELETED       WhatsappErrorCode = 33
	UNSUPPORTED_OR_MISSPELLED_PARAMETER WhatsappErrorCode = 100
	USER_IN_EXPERIMENT                  WhatsappErrorCode = 130472
	MESSAGE_NOT_SENT                    WhatsappErrorCode = 131000 // Message failed to send due to an unknown error.
	ACCESS_DENIED                       WhatsappErrorCode = 131005
	REQUIRED_PARAMETER_MISSING          WhatsappErrorCode = 131008
	PARAMETER_NOT_VALID                 WhatsappErrorCode = 131009
	SERVICE_UNAVAILABLE                 WhatsappErrorCode = 131016
	SAME_SENDER_RECEIVER                WhatsappErrorCode = 131021
	MESSAGE_UNDELIVERABLE               WhatsappErrorCode = 131026
	DISPLAY_NAME_NEED_APPROVAL          WhatsappErrorCode = 131037
	PAYMENT_METHOD_ERROR                WhatsappErrorCode = 131042
	PHONE_NUMBER_REGISTRATION_ERROR     WhatsappErrorCode = 131045
	// ...
)

// Marketing Messages API for WhatsApp Error codes
//
// https://developers.facebook.com/documentation/business-messaging/whatsapp/support/error-codes#marketing-messages-api-for-whatsapp-error-codes
const (
	INVALID_PARAMETER WhatsappErrorCode = 100
)
