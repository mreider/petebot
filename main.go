package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/strava/go.strava"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"path/filepath"
)

// Main function.
func main() {
	appDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join(appDir, "service.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
	})

	defaultSettingsPath := filepath.Join(appDir, "settings.default.json")
	localSettingsPath := filepath.Join(appDir, "settings.local.json")

	settings, err := LoadSettings(defaultSettingsPath, localSettingsPath)
	if err != nil {
		log.Fatalf("Can't load settings: %v\n", err)
	}

	db, err := gorm.Open("mysql", settings.DatabaseUri)
	if err != nil {
		log.Fatalf("Can't connect to the database: %v\n", err)
	}

	err = runMigrations(db)
	if err != nil {
		log.Fatalf("Can't execute database migrations: %v\n", err)
	}

	repo := NewRepository(db.Debug())
	fixMissingStravaUsersId(repo)
	fixMissingStravaUserNames(repo)

	server := NewServer(settings, repo, appDir)
	server.Run()
}

// Run database migrations.
func runMigrations(db *gorm.DB) error {
	res := db.AutoMigrate(&AccessDetails{}, &BotDetails{}, &JobDetails{}, &SentDetails{}, &ClubDetails{}, &UserDetails{}, &UserClubDetails{}, &PendingUnlock{})
	return res.Error
}

func fixMissingStravaUsersId(repo *Repository) {
	accessDetails, err := repo.AccessDetails.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, access := range accessDetails {
		if access.StravaUserId == 0 {
			client := strava.NewClient(access.StravaToken)
			service := strava.NewCurrentAthleteService(client)
			athlete, err := service.Get().Do()
			if err == nil {
				repo.AccessDetails.Update(access, map[string]interface{}{"StravaUserId": athlete.Id})
			}
		}
	}
}

func fixMissingStravaUserNames(repo *Repository) {
	userDetails, err := repo.UserDetails.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range userDetails {
		if user.UserName == "" {
			access, err := repo.AccessDetails.GetByStravaUserId(user.StravaUserId)
			if err == nil {
				client := strava.NewClient(access.StravaToken)
				service := strava.NewCurrentAthleteService(client)
				athlete, err := service.Get().Do()
				if err == nil {
					repo.UserDetails.Update(user, map[string]interface{}{"UserName": athlete.FirstName + " " + athlete.LastName})
				}

			}
		}
	}
}
