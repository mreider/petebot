package main

import "github.com/jinzhu/gorm"

type UserDetailsRepository struct {
	db *gorm.DB
}

func NewUserDetailsRepository(db *gorm.DB) *UserDetailsRepository {
	return &UserDetailsRepository{db: db}
}

func (repo *UserDetailsRepository) Get(stravaUserId int) (*UserDetails, error) {
	user := new(UserDetails)
	res := repo.db.First(user, &UserDetails{StravaUserId: stravaUserId})
	if res.Error == nil {
		return user, nil
	} else if res.RecordNotFound() {
		return nil, nil
	}
	return nil, res.Error
}

func (repo *UserDetailsRepository) List() ([]*UserDetails, error) {
	var userDetails []*UserDetails
	res := repo.db.Find(&userDetails)
	if res.Error != nil {
		return nil, res.Error
	}
	return userDetails, nil
}

func (repo *UserDetailsRepository) Create(user *UserDetails) error {
	res := repo.db.Create(user)
	return res.Error
}

func (repo *UserDetailsRepository) Update(user *UserDetails, values map[string]interface{}) error {
	res := repo.db.Model(user).Updates(values)
	return res.Error
}

func (repo *UserDetailsRepository) CountUnlocked() (int, error) {
	var count int
	res := repo.db.Model(&UserDetails{}).Where("unlock_code is not null and unlock_code <> ''").Count(&count)
	if res.Error != nil {
		return -1, res.Error
	}
	return count, nil
}

func (repo *UserDetailsRepository) Delete(user *UserDetails) error {
	res := repo.db.Delete(user)
	return res.Error
}
