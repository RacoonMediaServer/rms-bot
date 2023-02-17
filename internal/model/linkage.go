package model

// Linkage represents link between device and user chat
type Linkage struct {
	DeviceID string `gorm:"primaryKey"`
	ChatID   int64  `gorm:"unique"`
}
