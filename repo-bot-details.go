package main

import (
	"github.com/jinzhu/gorm"
)

type BotDetailsRepository struct {
	db *gorm.DB
}

func NewBotDetailsRepository(db *gorm.DB) *BotDetailsRepository {
	return &BotDetailsRepository{db: db}
}

func (repo *BotDetailsRepository) Get(teamId string) (*BotDetails, error) {
	bot := new(BotDetails)
	res := repo.db.First(bot, &BotDetails{TeamId: teamId})
	if res.Error == nil {
		return bot, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *BotDetailsRepository) List() ([]*BotDetails, error) {
	var bots []*BotDetails
	res := repo.db.Find(&bots)
	if res.Error != nil {
		return nil, res.Error
	}
	return bots, nil
}

func (repo *BotDetailsRepository) Create(bot *BotDetails) error {
	res := repo.db.Create(bot)
	return res.Error
}

func (repo *BotDetailsRepository) Update(bot *BotDetails, values map[string]interface{}) error {
	res := repo.db.Model(bot).Updates(values)
	return res.Error
}
