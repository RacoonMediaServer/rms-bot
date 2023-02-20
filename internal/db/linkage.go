package db

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
)

func (d *Database) LoadLinkages() (map[string]model.Linkage, error) {
	var linkages []model.Linkage
	result := map[string]model.Linkage{}

	err := d.conn.Model(&model.Linkage{}).Preload("Users").Find(&linkages).Error
	if err == nil {
		for _, l := range linkages {
			result[l.DeviceID] = l
		}
	}
	return result, err
}

func (d *Database) LinkUserToDevice(deviceID string, u model.DeviceUser) error {
	l := model.Linkage{DeviceID: deviceID}
	if err := d.conn.FirstOrCreate(&l).Error; err != nil {
		return err
	}
	return d.conn.Model(&l).Association("Users").Append(&u)
}

func (d *Database) UnlinkUser(deviceID string, u model.DeviceUser) error {
	l := model.Linkage{DeviceID: deviceID}
	return d.conn.Model(&l).Association("Users").Delete(&u)
}

func (d *Database) RemoveDevice(deviceID string) error {
	l := model.Linkage{DeviceID: deviceID}
	if err := d.conn.Model(&l).Association("Users").Clear(); err != nil {
		return err
	}
	d.conn.Delete(&l)
	return nil
}
