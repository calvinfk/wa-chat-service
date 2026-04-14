package template_components

import (
	"wa_chat_service/pkg/utils"
)

type SendQuickReplyButton struct {
	Type       string                          `json:"type" validate:"required,eq=button"`
	SubType    string                          `json:"sub_type" validate:"required,eq=QUICK_REPLY"`
	Index      string                          `json:"index" validate:"required,numeric"`
	Parameters []SendQuickReplyButtonParameter `json:"parameters" validate:"required,dive"`
}

type SendQuickReplyButtonParameter struct {
	Type    string `json:"type" validate:"required,eq=payload"`
	Payload string `json:"payload" validate:"required"`
}

func (b *SendQuickReplyButton) GetType() string {
	return b.Type
}

func (b *SendQuickReplyButton) GetMap() (map[string]any, error) {
	return utils.StructToMap(b, true)
}

func (b *SendQuickReplyButton) GetSubType() string {
	return b.SubType
}
