package whatsapp_business

type Client struct {
	Version         string
	UserAccessToken string
}

type MessageComponent interface {
	GetType() string
	GetPayload() map[string]any
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
