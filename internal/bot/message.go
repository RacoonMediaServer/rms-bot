package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"os"
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
	name string
	rd   io.ReadCloser
}

type photoUploading struct {
	name  string
	image []byte
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
		case photoURL:
			photoMessage := tgbotapi.NewPhotoUpload(chatID, attach)
			photoMessage.Caption = m.text
			photoMessage.ParseMode = "HTML"
			photoMessage.UseExisting = true
			photoMessage.ReplyMarkup = keyboard

			msg = photoMessage

		case *photoUploading:
			fileBytes := tgbotapi.FileBytes{
				Name:  attach.name,
				Bytes: attach.image,
			}

			photoMessage := tgbotapi.NewPhotoUpload(chatID, fileBytes)
			photoMessage.MimeType = "image/jpeg"
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
			videoMessage.MimeType = "video/mp4"
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

func (m *chatMessage) uploadPhoto(name string, image []byte) {
	m.attachment = &photoUploading{
		name:  name,
		image: image,
	}
}

func (m *chatMessage) setKeyboardStyle(style keyboardStyle) {
	m.keyboardStyle = style
}

func (m *chatMessage) uploadVideo(name, path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	m.attachment = &videoUploading{
		name: name,
		rd:   f,
	}

	return nil
}
