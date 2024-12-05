package model

// Link represents Telegram user connected to a device with specific domain
type Link struct {
	Domain   string `gorm:"primaryKey"`
	Device   string `gorm:"primaryKey"`
	TgUserID int    `gorm:"primaryKey"`
	TgChatID int64  `gorm:"unique"`
	NickName string
}
