package main

type LastMessage struct {
	Id        int    `gorm:"primary_key;AUTO_INCREMENT"`
	MessageId string `gorm:"not null;size:50;"`
}

// The same model as for the bot service.
type SentCode struct {
	Id        int    `gorm:"primary_key;AUTO_INCREMENT"`
	Recipient string `gorm:"not null;" sql:"type:VARCHAR(200) CHARACTER SET utf8 COLLATE utf8_general_ci"`
	Code      string `gorm:"not null;unique_index;size:7;"`
	Used      bool   `gorm:"not null;"`
}
