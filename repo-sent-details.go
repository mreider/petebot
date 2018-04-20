package main

import (
	"github.com/jinzhu/gorm"
)

type SentDetailsRepository struct {
	db *gorm.DB
}

func NewSentDetailsRepository(db *gorm.DB) *SentDetailsRepository {
	return &SentDetailsRepository{db: db}
}

func (repo *SentDetailsRepository) Get(clubId, channelId string) (*SentDetails, error) {
	details := new(SentDetails)
	res := repo.db.First(details, &SentDetails{ClubId: clubId, ChannelId: channelId})
	if res.Error == nil {
		return details, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *SentDetailsRepository) List(clubId string) ([]*SentDetails, error) {
	var details []*SentDetails
	res := repo.db.Find(&details, &SentDetails{ClubId: clubId})
	if res.Error != nil {
		return nil, res.Error
	}
	return details, nil
}

func (repo *SentDetailsRepository) Create(details *SentDetails) error {
	res := repo.db.Create(details)
	return res.Error
}

func (repo *SentDetailsRepository) Update(details *SentDetails, values map[string]interface{}) error {
	res := repo.db.Model(details).Updates(values)
	return res.Error
}
