package whatsapp_business

import "wa_chat_service/pkg/meta/whatsapp_business/message_components"

var mimeTypeExtensionMap = map[string]string{
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
}

var messageRegistry = map[message_components.MessageType]MessageComponent{
	message_components.AudioMessageType:    &message_components.Audio{},
	message_components.ContactsMessageType: &message_components.Contacts{},
	message_components.DocumentMessageType: &message_components.Document{},
	message_components.ImageMessageType:    &message_components.Image{},
	message_components.LocationMessageType: &message_components.Location{},
	message_components.ReactionMessageType: &message_components.Reaction{},
	message_components.StickerMessageType:  &message_components.Sticker{},
	message_components.TextMessageType:     &message_components.Text{},
	message_components.VideoMessageType:    &message_components.Video{},
	message_components.TemplateMessageType: &message_components.Template{},
}

var interactiveMessageRegistry = map[string]MessageComponent{
	"cta_url":                  &message_components.InteractiveCTAUrl{},
	"list":                     &message_components.InteractiveList{},
	"carousel":                 &message_components.InteractiveCarousel{},
	"button":                   &message_components.InteractiveButton{},
	"location_request_message": &message_components.LocationRequest{},
}
