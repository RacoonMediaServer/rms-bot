package db

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
)

func (d *Database) LoadLinks(domain string) (result []*model.Link, err error) {
	err = d.conn.Model(&model.Link{}).Where("domain = ?", domain).Find(&result).Error
	return
}

func (d *Database) AddLink(link *model.Link) error {
	return d.conn.Create(link).Error
}

func (d *Database) DelLink(link *model.Link) error {
	return d.conn.Delete(link).Error
}

func (d *Database) RemoveAll(domain, device string) error {
	return d.conn.Model(&model.Link{}).Where("domain = ? AND device = ?", domain, device).Error
}
