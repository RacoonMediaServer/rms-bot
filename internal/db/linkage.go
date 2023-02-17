package db

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
	"gorm.io/gorm"
)

func (d *Database) LoadLinkages() (map[string]int64, error) {
	linkages := make([]model.Linkage, 0)
	if err := d.conn.Find(&linkages).Error; err != nil {
		return nil, err
	}
	result := map[string]int64{}
	for _, l := range linkages {
		result[l.DeviceID] = l.ChatID
	}
	return result, nil
}

func (d *Database) StoreLinkage(linkage model.Linkage) error {
	return d.conn.Transaction(func(tx *gorm.DB) error {
		// удаляем все старые привязки, если они есть
		if err := d.conn.Model(&model.Linkage{}).Unscoped().Delete(&linkage).Error; err != nil {
			return err
		}

		if err := d.conn.Model(&model.Linkage{}).Unscoped().Where("chat_id = ?", linkage.ChatID).Delete(&model.Linkage{}).Error; err != nil {
			return err
		}

		return d.conn.Create(&linkage).Error
	})
}

func (d *Database) RemoveLinkage(deviceID string) error {
	l := &model.Linkage{DeviceID: deviceID}
	return d.conn.Model(l).Unscoped().Delete(l).Error
}
