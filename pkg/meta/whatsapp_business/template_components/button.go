package template_components

type CopyCodeButton struct {
	Type    string `json:"type" validate:"required,eq=COPY_CODE"`
	Example string `json:"example" validate:"required,max=15"`
}

type PhoneNumberButton struct {
	Type        string `json:"type" validate:"required,eq=PHONE_NUMBER"`
	Text        string `json:"text" validate:"required,max=25"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
}

type QuickReplyButton struct {
	Type string `json:"type" validate:"required,eq=QUICK_REPLY"`
	Text string `json:"text" validate:"required,max=25"`
}

type URLButton struct {
	Type    string    `json:"type" validate:"required,eq=URL"`
	Text    string    `json:"text" validate:"required,max=25"`
	URL     string    `json:"url" validate:"required,url,max=2000"`
	Example *[]string `json:"example,omitempty" validate:"omitempty,dive,max=1"`
}
