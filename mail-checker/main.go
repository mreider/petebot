package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"google.golang.org/api/gmail/v1"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	INBOX_LABEL = "INBOX"
)

func main() {
	appDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	os.MkdirAll(filepath.Join(appDir, "letters"), 0777)

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

	authenticator := NewGmailAuthenticator(appDir)
	client := authenticator.getClient()

	gmailService, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	watch(gmailService, settings.TopicName)
	if err != nil {
		log.Fatalf("Unable to watch gmail account: %v", err)
	}
	go watchPeriodically(gmailService, settings.TopicName)

	server := NewServer(settings, db.Debug(), gmailService, appDir)
	server.Run()
}

func watch(srv *gmail.Service, topicName string) error {
	srv.Users.Stop("me").Do()
	_, err := srv.Users.Watch("me", &gmail.WatchRequest{LabelIds: []string{INBOX_LABEL}, TopicName: topicName}).Do()
	return err
}

func watchPeriodically(srv *gmail.Service, topicName string) {
	for {
		time.Sleep(5 * 24 * time.Hour)
		err := watch(srv, topicName)
		if err != nil {
			fmt.Printf("Can't wath: %v\n", err)
		}
	}
}

// Run database migrations.
func runMigrations(db *gorm.DB) error {
	res := db.AutoMigrate(&LastMessage{}, &SentCode{})
	return res.Error
}
