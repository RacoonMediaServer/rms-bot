package model

// Linkage represents link between device and users
type Linkage struct {
	DeviceID string       `gorm:"primaryKey"`
	Users    []DeviceUser `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// DeviceUser represents a Telegram user
type DeviceUser struct {
	LinkageDeviceID string
	UserID          int `gorm:"primaryKey;autoIncrement:false"`
	ChatID          int64
	NickName        string
}
