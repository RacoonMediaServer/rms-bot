package bot

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go-micro.dev/v4/logger"
	"net/http"
)

func deserializeMessage(chat int64, msg *communication.BotMessage) tgbotapi.Chattable {
	m := newMessage(msg.Text, int(msg.ReplyID))
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
			photo, err := download(u)
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

func (bot *Bot) serializeMessage(msg *tgbotapi.Message) *communication.UserMessage {
	msgToDevice := &communication.UserMessage{
		Text:      msg.Text,
		User:      int32(msg.From.ID),
		MessageID: int32(msg.MessageID),
	}
	if msg.ReplyToMessage != nil {
		msgToDevice.ReplyID = int32(msg.ReplyToMessage.MessageID)
	}

	var fileID *string

	if msg.Document != nil {
		msgToDevice.Attachment = &communication.Attachment{
			Type:     communication.Attachment_Document,
			MimeType: msg.Document.MimeType,
		}
		fileID = &msg.Document.FileID
	} else if msg.Audio != nil {
		msgToDevice.Attachment = &communication.Attachment{
			Type:     communication.Attachment_Audio,
			MimeType: msg.Audio.MimeType,
		}
		fileID = &msg.Audio.FileID
	} else if msg.Voice != nil {
		msgToDevice.Attachment = &communication.Attachment{
			Type:     communication.Attachment_Voice,
			MimeType: msg.Voice.MimeType,
		}
		fileID = &msg.Voice.FileID
	} else if msg.Video != nil {
		msgToDevice.Attachment = &communication.Attachment{
			Type:     communication.Attachment_Video,
			MimeType: msg.Video.MimeType,
		}
		fileID = &msg.Video.FileID
	}

	if fileID != nil {
		var err error
		msgToDevice.Attachment.Content, err = bot.downloadTelegramFile(*fileID)
		if err != nil {
			bot.l.Logf(logger.ErrorLevel, "Download user file failed: %s", err)
		}
		msgToDevice.Text = msg.Caption
	}

	return msgToDevice
}

func (bot *Bot) downloadTelegramFile(fileID string) ([]byte, error) {
	fcfg := tgbotapi.FileConfig{FileID: fileID}
	f, err := bot.api.GetFile(fcfg)
	if err != nil {
		return nil, err
	}
	return download(f.Link(bot.api.Token))
}
