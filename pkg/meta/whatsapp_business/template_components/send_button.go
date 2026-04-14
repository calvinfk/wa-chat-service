package template_components

import (
	"wa_chat_service/pkg/utils"
)

type SendQuickReplyButton struct {
	Type       string           `json:"type" validate:"required,eq=button"`
	SubType    string           `json:"sub_type" validate:"required,eq=QUICK_REPLY"`
	Index      string           `json:"index" validate:"required,numeric"`
	Parameters []map[string]any `json:"parameters" validate:"required,dive"`
}

func (b *SendQuickReplyButton) GetType() string {
	return b.Type
}

func (b *SendQuickReplyButton) GetMap() (map[string]any, error) {
	return utils.StructToMap(b, true)
}
