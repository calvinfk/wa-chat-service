package template_components

import (
	"wa_chat_service/pkg/utils"
)

type CreateCopyCodeButton struct {
	Type    string `json:"type" validate:"required,eq=COPY_CODE"`
	Example string `json:"example" validate:"required,max=15"`
}

func (b *CreateCopyCodeButton) GetType() string {
	return b.Type
}

func (b *CreateCopyCodeButton) GetText() string {
	return b.Example
}

func (b *CreateCopyCodeButton) GetMap() (map[string]any, error) {
	return utils.StructToMap(b, true)
}

type CreatePhoneNumberButton struct {
	Type        string `json:"type" validate:"required,eq=PHONE_NUMBER"`
	Text        string `json:"text" validate:"required,max=25"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
}

func (b *CreatePhoneNumberButton) GetMap() (map[string]any, error) {
	return utils.StructToMap(b, true)
}

func (b *CreatePhoneNumberButton) GetType() string {
	return b.Type
}

func (b *CreatePhoneNumberButton) GetText() string {
	return b.Text
}

type CreateQuickReplyButton struct {
	Type string `json:"type" validate:"required,eq=QUICK_REPLY"`
	Text string `json:"text" validate:"required,max=25"`
}

func (b *CreateQuickReplyButton) GetType() string {
	return b.Type
}

func (b *CreateQuickReplyButton) GetText() string {
	return b.Text
}

func (b *CreateQuickReplyButton) GetMap() (map[string]any, error) {
	return utils.StructToMap(b, true)
}

type CreateURLButton struct {
	Type    string    `json:"type" validate:"required,eq=URL"`
	Text    string    `json:"text" validate:"required,max=25"`
	URL     string    `json:"url" validate:"required,url,max=2000"`
	Example *[]string `json:"example,omitempty" validate:"omitempty,dive,max=1"`
}

func (b *CreateURLButton) GetType() string {
	return b.Type
}

func (b *CreateURLButton) GetText() string {
	return b.Text
}

func (b *CreateURLButton) GetMap() (map[string]any, error) {
	return utils.StructToMap(b, true)
}
