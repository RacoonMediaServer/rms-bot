package bot

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			m.setPhotoURL(string(msg.Attachment.Content))
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
		Text:      msg.Text,
		Timestamp: timestamppb.New(msg.Time()),
	}
}
