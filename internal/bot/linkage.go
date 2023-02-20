package bot

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/teris-io/shortid"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const linkageCodeExpireTime = 10 * time.Minute

type linkageCode struct {
	device      string
	code        string
	generatedAt time.Time
}

func (bot *Bot) generateLinkageCode(device string) {
	code, ok := bot.deviceToCode[device]
	if !ok {
		code, _ = shortid.Generate()
		bot.codeToDevice[code] = linkageCode{
			device:      device,
			code:        code,
			generatedAt: time.Now(),
		}
		bot.deviceToCode[device] = code
	}

	_ = bot.c.Send(comm.IncomingMessage{
		DeviceID: device,
		Message: &communication.UserMessage{
			Type:      communication.MessageType_LinkageCode,
			Text:      code,
			Timestamp: timestamppb.Now(),
		},
	})
}

func (bot *Bot) clearExpiredLinkageCodes() {
	expired := make([]string, 0, len(bot.codeToDevice))
	now := time.Now()
	for k, v := range bot.codeToDevice {
		if now.Sub(v.generatedAt) >= linkageCodeExpireTime {
			bot.l.Logf(logger.WarnLevel, "Code %s for device %s expired", k, v.device)
			expired = append(expired, k)
			delete(bot.deviceToCode, v.device)
		}
	}

	for _, e := range expired {
		delete(bot.codeToDevice, e)
	}
}

func (bot *Bot) linkUserToDevice(user *tgbotapi.User, chatID int64, code linkageCode) error {
	u := model.DeviceUser{UserID: user.ID, ChatID: chatID, NickName: user.UserName}
	if err := bot.db.LinkUserToDevice(code.device, u); err != nil {
		return err
	}
	l := bot.linkages[code.device]
	l.Users = append(l.Users, u)
	bot.linkages[code.device] = l
	bot.userToDevice[user.ID] = code.device

	delete(bot.codeToDevice, code.code)
	delete(bot.deviceToCode, code.device)

	return nil
}
