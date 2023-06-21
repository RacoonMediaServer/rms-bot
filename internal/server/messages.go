package server

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
)

func getDeviceConnectedMessage(deviceID string) comm.OutgoingMessage {
	return comm.OutgoingMessage{
		DeviceID: deviceID,
		Message: &communication.BotMessage{
			Text: "Устройство сново в сети",
		},
	}
}

func getDeviceDisconnectedMessage(deviceID string) comm.OutgoingMessage {
	return comm.OutgoingMessage{
		DeviceID: deviceID,
		Message: &communication.BotMessage{
			Text: "Связь с устройством потеряна. Возможные причины:\n\t-устройство отключено;\n\t-потеря соединения с Интернет\n\t-отсутствие электричества\n\t-программный сбой;\n\t-обновление системы.",
		},
	}
}
