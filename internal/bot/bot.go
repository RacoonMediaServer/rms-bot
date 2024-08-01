package bot

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go-micro.dev/v4/logger"
)

// Database is a set of methods for accessing table, which map devices to Telegram users
type Database interface {
	LoadLinkages() (map[string]model.Linkage, error)
	LinkUserToDevice(deviceID string, u model.DeviceUser) error
	UnlinkUser(deviceID string, u model.DeviceUser) error
}

// Bot implements a Telegram bot
type Bot struct {
	l   logger.Logger
	wg  sync.WaitGroup
	api *tgbotapi.BotAPI
	cmd chan interface{}
	c   comm.DeviceCommunicator
	db  Database

	linkages     map[string]model.Linkage
	userToDevice map[int]string

	codeToDevice map[string]linkageCode
	deviceToCode map[string]string
}

type stopCommand struct{}

func NewBot(token string, database Database, c comm.DeviceCommunicator) (*Bot, error) {
	var err error
	bot := &Bot{
		l:            logger.DefaultLogger.Fields(map[string]interface{}{"from": "bot"}),
		db:           database,
		c:            c,
		linkages:     map[string]model.Linkage{},
		userToDevice: map[int]string{},
		codeToDevice: map[string]linkageCode{},
		deviceToCode: map[string]string{},
	}

	bot.linkages, err = database.LoadLinkages()
	if err != nil {
		return nil, fmt.Errorf("load linkages failed: %w", err)
	}
	for id, l := range bot.linkages {
		for _, u := range l.Users {
			bot.userToDevice[u.UserID] = id
		}
	}

	bot.api, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.cmd = make(chan interface{})
	bot.wg.Add(1)
	go func() {
		defer bot.wg.Done()
		bot.loop()
	}()

	return bot, nil
}

func (bot *Bot) loop() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(updateConfig)

	if err != nil {
		logger.Errorf("Get bot updates failed: %+v", err)
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msgToUser := <-bot.c.OutgoingChannel():
			if msgToUser.Message.Type == communication.MessageType_Interaction {
				bot.sendMessageToUser(msgToUser)
			} else if msgToUser.Message.Type == communication.MessageType_AcquiringCode {
				bot.generateLinkageCode(msgToUser.DeviceID)
			} else if msgToUser.Message.Type == communication.MessageType_UnlinkUser {
				bot.unlinkUserFromDevice(int(msgToUser.Message.User), msgToUser.DeviceID)
			}

		case <-ticker.C:
			// чистим просроченные коды
			bot.clearExpiredLinkageCodes()

		case update := <-updates:
			var message *tgbotapi.Message
			if update.Message == nil {
				if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
					message = update.CallbackQuery.Message
					message.Text = update.CallbackQuery.Data
					message.From = update.CallbackQuery.From
				}
			} else {
				message = update.Message
				if update.Message.From != nil {
					incomingMessagesCounter.WithLabelValues(update.Message.From.UserName).Inc()
				}
			}
			if message == nil {
				continue
			}
			bot.sendMessageToDevice(message)

		case command := <-bot.cmd:
			switch command.(type) {
			case *stopCommand:
				return
			default:
			}
		}
	}
}

func (bot *Bot) Stop() {
	bot.cmd <- &stopCommand{}
	bot.wg.Wait()
}

func (bot *Bot) send(msg tgbotapi.Chattable) int {
	tgMsg, err := bot.api.Send(msg)
	if err != nil {
		bot.l.Logf(logger.ErrorLevel, "Send message failed: %s", err)
		return 0
	}
	return tgMsg.MessageID
}

func (bot *Bot) sendMessageToUser(message comm.OutgoingMessage) {
	l, ok := bot.linkages[message.DeviceID]
	if !ok {
		return
	}
	for _, u := range l.Users {
		if message.Message.User == 0 || message.Message.User == int32(u.UserID) {
			msg := deserializeMessage(u.ChatID, message.Message)
			msgId := bot.send(msg)
			if message.Message.Pin == communication.BotMessage_ThisMessage {
				cfg := tgbotapi.PinChatMessageConfig{
					ChatID:    u.ChatID,
					MessageID: msgId,
				}
				_, err := bot.api.PinChatMessage(cfg)
				if err != nil {
					bot.l.Logf(logger.WarnLevel, "Pin message %d failed: %s", msgId, err)
				}
			} else if message.Message.Pin == communication.BotMessage_Drop {
				_, err := bot.api.UnpinChatMessage(tgbotapi.UnpinChatMessageConfig{ChatID: u.ChatID})
				if err != nil {
					bot.l.Logf(logger.WarnLevel, "Unpin message failed: %s", err)
				}
			}
		}
	}
}

func (bot *Bot) sendMessageToDevice(message *tgbotapi.Message) {
	token, ok := bot.userToDevice[message.From.ID]
	if !ok {
		if code, ok := bot.codeToDevice[message.Text]; ok {
			if err := bot.linkUserToDevice(message.From, message.Chat.ID, code); err != nil {
				bot.l.Logf(logger.ErrorLevel, "Link %d to device %s failed: %s", message.Chat.ID, code.device, err)
				bot.sendTextMessage(message.Chat.ID, "Не удалось связать чат с устройством")
				return
			}

			bot.l.Logf(logger.InfoLevel, "Device '%s' linked to chat %d", code.device, message.Chat.ID)
			bot.sendTextMessage(message.Chat.ID, "Текущий чат связан с устройством. Ура, ничего не сломалось...")
			return
		}
		bot.sendTextMessage(message.Chat.ID, "Необходимо привязать устройство к текущему чату. Для этого необходимо ввести здесь код из веб-интерфейса")
		return
	}

	msg := bot.serializeMessage(message)
	if err := bot.c.Send(comm.IncomingMessage{DeviceID: token, Message: msg}); err != nil {
		bot.l.Logf(logger.ErrorLevel, "Cannot send message to the device: %s", err)
		text := ""
		if errors.Is(err, comm.ErrDeviceIsNotConnected) {
			text = "Устройство не в сети, команда не доставлена"
		} else {
			text = "Что-то пошло не так..."
		}
		bot.sendTextMessage(message.Chat.ID, text)
	}
}

func (bot *Bot) sendTextMessage(chat int64, text string) {
	msg := newMessage(text, 0)
	bot.send(msg.compose(chat))
}
