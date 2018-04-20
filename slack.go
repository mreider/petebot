package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"net/http"
	"net/url"
)

// Slack service.
type Slack struct {
	clientId         string
	clientSecret     string
	callbackUrl      string
	incomingMessages chan *IncomingSlackMessage
	outgoingMessages chan *OutgoingSlackMessage
}

// Create new slack service.
func NewSlack(clientId string, clientSecret string, callbackUrl string) *Slack {
	return &Slack{
		clientId:         clientId,
		clientSecret:     clientSecret,
		callbackUrl:      callbackUrl,
		incomingMessages: make(chan *IncomingSlackMessage, 100),
		outgoingMessages: make(chan *OutgoingSlackMessage, 100),
	}
}

// Get Slack authorization URL.
func (s *Slack) AuthorizationUrl() string {
	return fmt.Sprintf("https://slack.com/oauth/authorize?scope=bot&client_id=%s&redirect_uri=%s",
		s.clientId,
		url.QueryEscape(s.callbackUrl))
}

// Process slack authorization request.
func (s *Slack) Authorize(w http.ResponseWriter, r *http.Request) (resp *slack.OAuthResponse, err error) {
	code := r.FormValue("code")
	resp, err = slack.GetOAuthResponse(s.clientId, s.clientSecret, code, s.callbackUrl, false)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Process Slack event, send an incoming slack message to incoming messages channel.
func (s *Slack) BotCommand(w http.ResponseWriter, r *http.Request) {
	var response map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		log.Printf("Can't parse slack message: %v\n", err)
		return
	}
	if challenge, ok := response["challenge"]; ok {
		w.Write([]byte(fmt.Sprintf("{\"challenge\": \"%s\"}", challenge)))
		return
	}
	if event, ok := response["event"]; ok {
		event, _ := event.(map[string]interface{})
		event_type, _ := event["type"].(string)
		if event_type == "message" {
			teamId, _ := response["team_id"].(string)
			channelId, _ := event["channel"].(string)
			userId, _ := event["user"].(string)
			text, _ := event["text"].(string)
			s.incomingMessages <- &IncomingSlackMessage{teamId, channelId, userId, text}
		}
	}
}

// Post a Slack message.
func (s *Slack) PostMessage(message *OutgoingSlackMessage) {
	s.outgoingMessages <- message
}

// Post outgoing Slack messages.
func (s *Slack) PostOutgoingMessages() {
	for message := range s.outgoingMessages {
		api := slack.New(message.AccessToken)

		if message.Text != "" {
			api.PostMessage(message.ChannelId, message.Text, slack.PostMessageParameters{})
		}

		if message.AttachmentName != "" {
			reader := bytes.NewReader(message.AttachmentContent)
			_, err := api.UploadFile(slack.FileUploadParameters{Reader: reader, Filename: message.AttachmentName, Channels: []string{message.ChannelId}})
			if err != nil {
				log.Printf("Can't upload a file to Slack: %v\n", err)
			}
		}
	}
}

// Get Slack incoming messages channel.
func (s *Slack) IncomingMessages() <-chan *IncomingSlackMessage {
	return s.incomingMessages
}
