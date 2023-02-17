package bot

import (
	"bytes"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
)

type keyboardStyle int

const (
	chatKeyboardStyle keyboardStyle = iota
	messageKeyboardStyle
)

type botMessage interface {
	compose(chatID int64) tgbotapi.Chattable
}

type videoUploading struct {
	name     string
	mimeType string
	rd       io.Reader
}

type photoUploading struct {
	name     string
	mimeType string
	image    []byte
}

type photoURL string

type chatMessage struct {
	text             string
	replyToMessageID int
	keyboardStyle    keyboardStyle
	attachment       interface{}

	buttons []button
}

type button struct {
	command string
	title   string
}

func (m *chatMessage) compose(chatID int64) tgbotapi.Chattable {
	var msg tgbotapi.Chattable
	var keyboard interface{}

	if len(m.buttons) > 0 {
		if m.keyboardStyle == chatKeyboardStyle {
			buttons := make([]tgbotapi.KeyboardButton, 0)
			for _, button := range m.buttons {
				buttons = append(buttons, tgbotapi.NewKeyboardButton(button.command))
			}

			keyboard = tgbotapi.NewReplyKeyboard(buttons)
		} else {
			buttons := make([]tgbotapi.InlineKeyboardButton, 0)
			for _, action := range m.buttons {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(action.title, action.command))
			}

			keyboard = tgbotapi.NewInlineKeyboardMarkup(buttons)
		}
	}

	if m.attachment != nil {
		switch attach := m.attachment.(type) {
		case *photoURL:
			photoMessage := tgbotapi.PhotoConfig{}
			photoMessage.Caption = m.text
			photoMessage.FileID = string(*attach)
			photoMessage.ParseMode = "HTML"
			photoMessage.UseExisting = true
			photoMessage.ReplyMarkup = keyboard
			photoMessage.ChatID = chatID

			msg = photoMessage

		case *photoUploading:
			fileBytes := tgbotapi.FileBytes{
				Name:  attach.name,
				Bytes: attach.image,
			}

			photoMessage := tgbotapi.NewPhotoUpload(chatID, fileBytes)
			photoMessage.MimeType = attach.mimeType
			photoMessage.Caption = m.text
			photoMessage.ParseMode = "HTML"
			photoMessage.ReplyMarkup = keyboard

			msg = photoMessage

		case *videoUploading:
			fileReader := tgbotapi.FileReader{
				Name:   attach.name,
				Size:   -1,
				Reader: attach.rd,
			}
			videoMessage := tgbotapi.NewVideoUpload(chatID, fileReader)
			videoMessage.MimeType = attach.mimeType
			videoMessage.Caption = m.text
			videoMessage.ParseMode = "HTML"
			videoMessage.ReplyMarkup = keyboard

			msg = videoMessage

		}
	} else {
		textMessage := tgbotapi.NewMessage(chatID, m.text)
		textMessage.ParseMode = "HTML"
		textMessage.ReplyToMessageID = m.replyToMessageID
		textMessage.ReplyMarkup = keyboard

		msg = &textMessage
	}

	return msg
}

func newMessage(text string, replyToMessageID int) *chatMessage {
	return &chatMessage{
		text:             text,
		replyToMessageID: replyToMessageID,
	}
}

func (m *chatMessage) addButton(title string, command string) {
	m.buttons = append(m.buttons, button{
		command: command,
		title:   title,
	})
}

func (m *chatMessage) setPhotoURL(url string) {
	u := photoURL(url)
	m.attachment = &u
}

func (m *chatMessage) uploadPhoto(name, mimeType string, image []byte) {
	m.attachment = &photoUploading{
		name:     name,
		mimeType: mimeType,
		image:    image,
	}
}

func (m *chatMessage) setKeyboardStyle(style keyboardStyle) {
	m.keyboardStyle = style
}

func (m *chatMessage) uploadVideo(name, mimeType string, video []byte) {
	m.attachment = &videoUploading{
		name:     name,
		mimeType: mimeType,
		rd:       bytes.NewReader(video),
	}
}
