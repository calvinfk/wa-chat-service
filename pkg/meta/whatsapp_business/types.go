package whatsapp_business

var (
	mimeTypeExtensionMap = map[string]string{
		"audio/aac":  ".aac",
		"audio/amr":  ".amr",
		"audio/mpeg": ".mp3",
		"audio/mp4":  ".m4a",
		"audio/ogg":  ".ogg",

		"text/plain":               ".txt",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   ".docx",
		"application/vnd.ms-powerpoint":                                             ".ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"application/pdf": ".pdf",

		"image/jpeg": ".jpeg",
		"image/png":  ".png",

		"image/webp": ".webp",

		"video/3gpp": ".3gp",
		"video/mp4":  ".mp4",

		// "image/gif": ".gif",
	}
)

type Client struct {
	BaseURL         string
	WabaID          string
	PhoneNumberID   string
	Version         string
	UserAccessToken string
}

type MessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

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

type UploadMediaResponse struct {
	ID string `json:"id"`
}

type GetMediaURLResponse struct {
	MessagingProduct string `json:"messaging_product"`
	URL              string `json:"url"`
	MimeType         string `json:"mime_type"`
	Sha256           string `json:"sha256"`
	FileSize         int64  `json:"file_size"`
	ID               string `json:"id"`
}

type DeleteMediaResponse struct {
	Success bool `json:"success"`
}
