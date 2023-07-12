package bot

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
)

func deserializeMessage(chat int64, msg *communication.BotMessage) tgbotapi.Chattable {
	m := newMessage(msg.Text, 0)
	for _, btn := range msg.Buttons {
		m.addButton(btn.Title, btn.Command)
	}
	if msg.KeyboardStyle == communication.KeyboardStyle_Message {
		m.setKeyboardStyle(messageKeyboardStyle)
	}
	if msg.Attachment != nil {
		switch msg.Attachment.Type {
		case communication.Attachment_PhotoURL:
			// Зачем скачивать фото? обход блокировки российского IP от TMDb...
			// будет работать, если сервер на зарубежном хостинге
			u := string(msg.Attachment.Content)
			photo, err := downloadPhoto(u)
			if err == nil {
				m.uploadPhoto("photo", http.DetectContentType(photo), photo)
			} else {
				m.setPhotoURL(u)
			}
		case communication.Attachment_Photo:
			m.uploadPhoto("photo", msg.Attachment.MimeType, msg.Attachment.Content)
		case communication.Attachment_Video:
			m.uploadVideo("video", msg.Attachment.MimeType, msg.Attachment.Content)
		}
	}
	return m.compose(chat)
}

func serializeMessage(msg *tgbotapi.Message) *communication.UserMessage {
	return &communication.UserMessage{
		Text: msg.Text,
		User: int32(msg.From.ID),
	}
}
