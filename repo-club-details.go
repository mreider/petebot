package main

import "github.com/jinzhu/gorm"

type ClubDetailsRepository struct {
	db *gorm.DB
}

func NewClubDetailsRepository(db *gorm.DB) *ClubDetailsRepository {
	return &ClubDetailsRepository{db: db}
}

func (repo *ClubDetailsRepository) Get(clubId string) (*ClubDetails, error) {
	club := new(ClubDetails)
	res := repo.db.First(club, &ClubDetails{ClubId: clubId})
	if res.Error == nil {
		return club, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *ClubDetailsRepository) List() ([]*ClubDetails, error) {
	var clubs []*ClubDetails
	res := repo.db.Find(&clubs)
	if res.Error != nil {
		return nil, res.Error
	}
	return clubs, nil
}

func (repo *ClubDetailsRepository) Create(club *ClubDetails) error {
	res := repo.db.Create(club)
	return res.Error
}

func (repo *ClubDetailsRepository) Update(club *ClubDetails, values map[string]interface{}) error {
	res := repo.db.Model(club).Updates(values)
	return res.Error
}
