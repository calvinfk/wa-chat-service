package whatsapp_business_component

type Text struct {
	EnableLinkPreview *bool  `json:"enable_link_preview,omitempty"`
	Body              string `json:"body"`
}

func (t *Text) GetType() string {
	return "text"
}

func (t *Text) GetPayload() map[string]any {
	payload := map[string]any{
		"body": t.Body,
	}
	if t.EnableLinkPreview != nil {
		payload["enable_link_preview"] = *t.EnableLinkPreview
	}
	return map[string]any{
		"type": t.GetType(),
		"text": payload,
	}
}
