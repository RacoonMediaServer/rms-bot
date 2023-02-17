package bot

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func deserializeMessage(chat int64, msg *communication.BotMessage) tgbotapi.Chattable {
	m := newMessage(msg.Text, 0)
	return m.compose(chat)
}

func serializeMessage(msg *tgbotapi.Message) *communication.UserMessage {
	return &communication.UserMessage{
		Text:      msg.Text,
		Timestamp: timestamppb.New(msg.Time()),
	}
}
