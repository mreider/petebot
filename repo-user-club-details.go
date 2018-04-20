package main

import "github.com/jinzhu/gorm"

type UserClubDetailsRepository struct {
	db *gorm.DB
}

func NewUserClubDetailsRepository(db *gorm.DB) *UserClubDetailsRepository {
	return &UserClubDetailsRepository{db: db}
}

func (repo *UserClubDetailsRepository) Get(stravaUserId int, clubId string) (*UserClubDetails, error) {
	userClubDetails := new(UserClubDetails)
	res := repo.db.First(userClubDetails, &UserClubDetails{StravaUserId: stravaUserId, ClubId: clubId})
	if res.Error == nil {
		return userClubDetails, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *UserClubDetailsRepository) Create(userClubDetails *UserClubDetails) error {
	res := repo.db.Create(userClubDetails)
	return res.Error
}

func (repo *UserClubDetailsRepository) Update(userClubDetails *UserClubDetails, values map[string]interface{}) error {
	res := repo.db.Model(userClubDetails).Updates(values)
	return res.Error
}

func (repo *UserClubDetailsRepository) Remove(stravaUserId int) error {
	res := repo.db.Exec("delete from user_club_details where strava_user_id = ?;", stravaUserId)
	return res.Error
}
