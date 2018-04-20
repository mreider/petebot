package main

import (
	"encoding/base64"
	"fmt"
	"github.com/jinzhu/gorm"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	CODE_LENGTH = 7
)

var (
	SpacesRe = regexp.MustCompile(`(?m)\s+`)
)

type MailProcessor struct {
	settings     *Settings
	db           *gorm.DB
	gmailService *gmail.Service
	appDirectory string

	pushNotifications chan bool
}

type MessageData struct {
	From    string
	Subject string
	Body    string
}

func NewMailProcessor(settings *Settings, db *gorm.DB, gmailService *gmail.Service, appDirectory string) *MailProcessor {
	return &MailProcessor{
		settings:          settings,
		db:                db,
		gmailService:      gmailService,
		appDirectory:      appDirectory,
		pushNotifications: make(chan bool, 10),
	}
}

func (p *MailProcessor) NotifyPush() {
	p.pushNotifications <- true
}

func (p *MailProcessor) Run() {
	profile, err := p.gmailService.Users.GetProfile("me").Do()
	if err != nil {
		log.Fatalf("Can't get gmail profile: %v\n", err)
	}
	fmt.Printf("Gmail account: %s\n", profile.EmailAddress)

	for {
		select {
		case <-p.pushNotifications:
			lastMessageIdFromDB := p.getLastMessageIdFromDB()
			unknowMessageIds := p.getUnknownInboxMessageIds(lastMessageIdFromDB)
			if len(unknowMessageIds) > 0 {
				newLastMessageId := ""
				for i := len(unknowMessageIds) - 1; i >= 0; i-- {
					messageId := unknowMessageIds[i]
					message := p.getMessageData(messageId)
					if message == nil {
						break
					}

					// FOR DEBUG
					ioutil.WriteFile(filepath.Join(p.appDirectory, "letters", messageId), []byte(message.Body), 0666)

					body := strings.ToLower(message.Body)
					body = SpacesRe.ReplaceAllString(body, " ")

					matchedPattern := ""
					for pattern, patternParts := range p.settings.Patterns {
						for _, patternPart := range patternParts {
							if strings.Contains(body, strings.ToLower(patternPart)) {
								matchedPattern = pattern
								break
							}
						}
					}

					if matchedPattern != "" {
						unlockCode := GenerateRandomCode(CODE_LENGTH)
						for p.existConfirmationCodeInDB(unlockCode) {
							unlockCode = GenerateRandomCode(CODE_LENGTH)
						}
						unlockSubject := p.getEmailText(UNLOCK_EMAIL_SUBJECT, matchedPattern, unlockCode)
						unlockBody := p.getEmailText(UNLOCK_EMAIL_BODY, matchedPattern, unlockCode)
						if p.sendMessage(profile.EmailAddress, message.From, unlockSubject, unlockBody) {
							p.saveConfirmationCodeToDB(message.From, unlockCode)
						}
					}
					newLastMessageId = messageId
				}
				if newLastMessageId != "" {
					p.saveLastMessageIdToDB(newLastMessageId)
				}
			}
		}
	}
}

func (p *MailProcessor) getLastMessageIdFromDB() string {
	var lastMessage LastMessage
	res := p.db.First(&lastMessage)
	if res.RecordNotFound() {
		return ""
	} else if res.Error == nil {
		return lastMessage.MessageId
	} else {
		log.Printf("Can't get last message id from DB: %v\n", res.Error)
		return ""
	}
}

func (p *MailProcessor) saveLastMessageIdToDB(newLastMessageId string) {
	var count int
	res := p.db.Model(&LastMessage{}).Count(&count)
	if res.Error != nil {
		log.Printf("Can't get last message count from DB: %v\n", res.Error)
		return
	}

	if count == 0 {
		res := p.db.Create(&LastMessage{MessageId: newLastMessageId})
		if res.Error != nil {
			log.Printf("Can't create last message in DB: %v\n", res.Error)
			return
		}
	} else {
		var lastMessage LastMessage
		res := p.db.First(&lastMessage)
		if res.Error != nil {
			log.Printf("Can't get last message from DB: %v\n", res.Error)
			return
		}

		if isMessageIdGreaterThen(newLastMessageId, lastMessage.MessageId) {
			lastMessage.MessageId = newLastMessageId
			res := p.db.Save(&lastMessage)
			if res.Error != nil {
				log.Printf("Can't update last message in DB: %v\n", res.Error)
				return
			}
		}
	}
}

func (p *MailProcessor) getUnknownInboxMessageIds(lastMessageIdFromDB string) []string {
	response, err := p.gmailService.Users.Messages.List("me").LabelIds(INBOX_LABEL).Do()
	if err != nil {
		log.Printf("Can't get gmail inbox messages: %v\n", err)
		return nil
	}
	result := make([]string, 0, 4)
	for _, message := range response.Messages {
		if isMessageIdGreaterThen(message.Id, lastMessageIdFromDB) {
			result = append(result, message.Id)
		} else {
			break
		}
	}
	return result
}

func (p *MailProcessor) getMessageData(messageId string) *MessageData {
	response, err := p.gmailService.Users.Messages.Get("me", messageId).Format("full").Do()
	if err != nil {
		log.Printf("Can't get gmail message: %v\n", err)
		return nil
	}

	from := getHeaderValue(response.Payload.Headers, "From")
	subject := getHeaderValue(response.Payload.Headers, "Subject")

	bodyData := response.Payload.Body.Data
	if bodyData == "" && len(response.Payload.Parts) > 0 {
		bodyData = response.Payload.Parts[0].Body.Data
	}

	body, _ := base64.URLEncoding.DecodeString(bodyData)
	return &MessageData{From: from, Subject: subject, Body: string(body)}
}

func (p *MailProcessor) sendMessage(from, to, subject, body string) bool {
	messageStr := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, to, subject, body))

	var gmailMessage gmail.Message
	gmailMessage.Raw = base64.URLEncoding.EncodeToString(messageStr)

	_, err := p.gmailService.Users.Messages.Send("me", &gmailMessage).Do()
	if err != nil {
		log.Printf("Can't send gmail message: %v\n", err)
		return false
	}
	return true
}

func isMessageIdGreaterThen(left, right string) bool {
	if len(left) > len(right) {
		return true
	}
	if len(left) < len(right) {
		return false
	}
	return left > right
}

func getHeaderValue(headers []*gmail.MessagePartHeader, headerName string) string {
	for _, header := range headers {
		if header.Name == headerName {
			return header.Value
		}
	}
	return ""
}

func (p *MailProcessor) existConfirmationCodeInDB(code string) bool {
	var sentCode SentCode
	res := p.db.First(&sentCode, &SentCode{Code: code})
	return res.Error == nil
}

func (p *MailProcessor) saveConfirmationCodeToDB(recipient, code string) {
	res := p.db.Create(&SentCode{Recipient: recipient, Code: code})
	if res.Error != nil {
		log.Printf("Can't create sent code in DB: %v\n", res.Error)
		return
	}
}

func (p *MailProcessor) getEmailText(text, charity, code string) string {
	result := strings.Replace(text, "{charity}", charity, -1)
	result = strings.Replace(result, "{code}", code, -1)
	return result
}
