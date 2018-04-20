package main

import (
	"github.com/jinzhu/gorm"
)

type SentCodesRepository struct {
	db *gorm.DB
}

func NewSentCodesRepository(db *gorm.DB) *SentCodesRepository {
	return &SentCodesRepository{db: db}
}

func (repo *SentCodesRepository) Get(confirmCode string) (*SentCode, error) {
	sentCode := new(SentCode)
	res := repo.db.First(sentCode, &SentCode{Code: confirmCode})
	if res.Error == nil {
		return sentCode, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *SentCodesRepository) Update(sentCode *SentCode, values map[string]interface{}) error {
	res := repo.db.Model(sentCode).Updates(values)
	return res.Error
}
