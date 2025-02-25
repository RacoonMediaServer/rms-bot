package bot

import (
	"time"

	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/teris-io/shortid"
	"go-micro.dev/v4/logger"
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
			Type: communication.MessageType_LinkageCode,
			Text: code,
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
	link := model.Link{
		Device:   code.device,
		TgUserID: user.ID,
		TgChatID: chatID,
		NickName: user.UserName,
	}
	if err := bot.db.AddLink(&link); err != nil {
		return err
	}

	bot.deviceToUsers[code.device] = append(bot.deviceToUsers[code.device], &link)
	bot.chatToDevice[chatID] = code.device

	delete(bot.codeToDevice, code.code)
	delete(bot.deviceToCode, code.device)

	return nil
}

func (bot *Bot) unlinkUserFromDevice(user int, device string) {
	links, ok := bot.deviceToUsers[device]
	if !ok {
		bot.l.Logf(logger.WarnLevel, "Cannot unlink user %d from device %s: no linkage", user, device)
		return
	}

	userIndex := -1
	foundLink := &model.Link{}
	for i, u := range links {
		if u.TgUserID == user {
			userIndex = i
			foundLink = u
			break
		}
	}
	if userIndex == -1 {
		bot.l.Logf(logger.WarnLevel, "Cannot unlink user %d from device %s: not assigned", user, device)
		return
	}

	if err := bot.db.DelLink(foundLink); err != nil {
		bot.l.Logf(logger.WarnLevel, "Cannot unlink user %d from device %s: %s", user, device, err)
		return
	}

	delete(bot.chatToDevice, foundLink.TgChatID)

	links = append(links[:userIndex], links[userIndex+1:]...)
	bot.deviceToUsers[device] = links
}
