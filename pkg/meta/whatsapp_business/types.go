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
