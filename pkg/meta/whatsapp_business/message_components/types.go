package message_components

type MessageType string

const (
	AudioMessageType       MessageType = "audio"
	ButtonMessageType      MessageType = "button"
	ContactsMessageType    MessageType = "contacts"
	DocumentMessageType    MessageType = "document"
	EditMessageType        MessageType = "edit"
	ImageMessageType       MessageType = "image"
	InteractiveMessageType MessageType = "interactive"
	LocationMessageType    MessageType = "location"
	OrderMessageType       MessageType = "order"
	ReactionMessageType    MessageType = "reaction"
	RevokeMessageType      MessageType = "revoke"
	StickerMessageType     MessageType = "sticker"
	SystemMessageType      MessageType = "system"
	TextMessageType        MessageType = "text"
	UnsupportedMessageType MessageType = "unsupported"
	VideoMessageType       MessageType = "video"
	TemplateMessageType    MessageType = "template"
)

type Interactive struct {
	Type string `json:"type" validate:"required,oneof=cta_url list carousel button"`
}

type InteractiveBody struct {
	Text string `json:"text" validate:"required,max=1024"`
}

type InteractiveFooter struct {
	Text string `json:"text" validate:"required,max=60"`
}

// type InteractiveHeader struct {
// 	Type     string  `json:"type" validate:"required,oneof=text image document video"`
// 	Text     *string `json:"text" validate:"omitempty,max=60"`
// 	Image    *Media  `json:"image,omitempty"`
// 	Document *Media  `json:"document,omitempty"`
// 	Video    *Media  `json:"video,omitempty"`
// }

type Media struct {
	ID   *string `json:"id,omitempty" validate:"required_without=Link,excluded_with=Link,omitempty,min=1"` // Only if using uploaded media, Required if using uploaded media, otherwise omit.
	Link *string `json:"link,omitempty" validate:"required_without=ID,excluded_with=ID,omitempty,uri"`     // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
}

type MediaAssetURL struct {
	Link string `json:"link,omitempty" validate:"uri"` // Only if using hosted media (not recommended), Required if using hosted media, otherwise omit.
}

type QuickReplyButton struct {
	Type       string                     `json:"type" validate:"required,eq=quick_reply"`
	QuickReply QuickReplyButtonQuickReply `json:"quick_reply" validate:"required"`
}

type QuickReplyButtonQuickReply struct {
	ID          string `json:"id" validate:"required,max=256"`
	DisplayText string `json:"display_text" validate:"required,max=20"`
}
