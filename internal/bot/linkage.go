package bot

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	"github.com/teris-io/shortid"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const linkageCodeExpireTime = 10 * time.Minute

type linkageCode struct {
	token       string
	code        string
	generatedAt time.Time
}

func (bot *Bot) generateLinkageCode(token string) {
	code, _ := shortid.Generate()
	bot.linkageCodes[code] = linkageCode{
		token:       token,
		code:        code,
		generatedAt: time.Now(),
	}
	_ = bot.c.Send(comm.IncomingMessage{
		Token: token,
		Message: &communication.UserMessage{
			Type:      communication.MessageType_LinkageCode,
			Text:      code,
			Timestamp: timestamppb.Now(),
		},
	})
}

func (bot *Bot) clearExpiredLinkageCodes() {
	expired := make([]string, 0, len(bot.linkageCodes))
	now := time.Now()
	for k, v := range bot.linkageCodes {
		if now.Sub(v.generatedAt) >= linkageCodeExpireTime {
			bot.l.Logf(logger.WarnLevel, "Code %s for device %s expired", k, v.token)
			expired = append(expired, k)
		}
	}

	for _, e := range expired {
		delete(bot.linkageCodes, e)
	}
}

func (bot *Bot) linkUserToDevice(id int64, code linkageCode) {
	bot.tokenToChat[code.token] = id
	bot.chatToToken[id] = code.token
	// TODO: запись в БД
	delete(bot.linkageCodes, code.code)

	bot.l.Logf(logger.InfoLevel, "Device '%s' linked to chat %d", code.token, id)
}
