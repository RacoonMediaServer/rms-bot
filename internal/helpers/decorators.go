package helpers

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/db"
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
)

type LinksDomainDecorator struct {
	Domain   string
	Database *db.Database
}

func (d LinksDomainDecorator) LoadLinks() (result map[string][]*model.Link, err error) {
	result = map[string][]*model.Link{}
	var list []*model.Link
	list, err = d.Database.LoadLinks(d.Domain)
	if err != nil {
		return
	}
	for _, link := range list {
		result[link.Device] = append(result[link.Device], link)
	}
	return
}

func (d LinksDomainDecorator) AddLink(link *model.Link) error {
	link.Domain = d.Domain
	return d.Database.AddLink(link)
}

func (d LinksDomainDecorator) DelLink(link *model.Link) error {
	link.Domain = d.Domain
	return d.Database.DelLink(link)
}
