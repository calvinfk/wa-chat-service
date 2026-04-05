package whatsapp_business

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
