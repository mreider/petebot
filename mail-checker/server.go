package main

import (
	"github.com/jinzhu/gorm"
	"google.golang.org/api/gmail/v1"
	"log"
	"net/http"
)

type PushMessage struct {
	Data string `json:"data"`
}

type PushNotification struct {
	Message PushMessage `json:"message"`
}

type GmailNotification struct {
	HistoryId uint64 `json:"historyId"`
}

// HTTP Server.
type Server struct {
	settings     *Settings
	db           *gorm.DB
	gmailService *gmail.Service
	processor    *MailProcessor
}

// Create new HTTP server.
func NewServer(settings *Settings, db *gorm.DB, gmailService *gmail.Service, appDir string) *Server {
	return &Server{
		settings:     settings,
		db:           db,
		gmailService: gmailService,
		processor:    NewMailProcessor(settings, db, gmailService, appDir),
	}
}

// Run the server.
func (s *Server) Run() {
	go s.processor.Run()

	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/push", s.push)
	http.HandleFunc("/test", s.test)
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(s.settings.ServerInternalAddress, nil))
}

func (s *Server) push(w http.ResponseWriter, r *http.Request) {
	s.processor.NotifyPush()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) test(w http.ResponseWriter, r *http.Request) {
	s.processor.NotifyPush()
}
