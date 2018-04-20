package main

type IncomingSlackMessage struct {
	TeamId    string
	ChannelId string
	UserId    string
	Text      string
}

type OutgoingSlackMessage struct {
	ChannelId         string
	AccessToken       string
	Text              string
	AttachmentName    string
	AttachmentContent []byte
}

type AccessDetails struct {
	Id           int    `gorm:"primary_key;AUTO_INCREMENT"`
	TeamId       string `gorm:"size:100;index:idx_teamid_slackuserid"`
	SlackToken   string `gorm:"size:100"`
	StravaToken  string `gorm:"size:100"`
	SlackUserId  string `gorm:"size:100;index:idx_teamid_slackuserid"`
	StravaUserId int
}

type BotDetails struct {
	Id             int    `gorm:"primary_key;AUTO_INCREMENT"`
	TeamId         string `gorm:"size:100;unique_index"`
	BotId          string `gorm:"size:100"`
	BotAccessToken string `gorm:"size:100"`
}

type JobDetails struct {
	Id          int    `gorm:"primary_key;AUTO_INCREMENT"`
	JobId       string `gorm:"size:100"`
	TeamId      string `gorm:"size:100;index:idx_teamid_channelid_clubid"`
	ChannelId   string `gorm:"size:100;index:idx_teamid_channelid_clubid"`
	ClubId      string `gorm:"size:100;index:idx_teamid_channelid_clubid"`
	SlackUserId string `gorm:"size:100"`
}

type SentDetails struct {
	Id             int    `gorm:"primary_key;AUTO_INCREMENT"`
	ClubId         string `gorm:"size:100;index:idx_clubid_channelid"`
	ChannelId      string `gorm:"size:100;index:idx_clubid_channelid"`
	LastActivityId int
}

// The same model as for the mail service.
type SentCode struct {
	Id        int    `gorm:"primary_key;AUTO_INCREMENT"`
	Recipient string `gorm:"not null;" sql:"type:VARCHAR(200) CHARACTER SET utf8 COLLATE utf8_general_ci"`
	Code      string `gorm:"not null;unique_index;size:7;"`
	Used      bool   `gorm:"not null;"`
}

type ClubDetails struct {
	Id     int    `gorm:"primary_key;AUTO_INCREMENT"`
	ClubId string `gorm:"size:100;unique;"`
	Unit   string `gorm:"size:10"`
}

type UserDetails struct {
	Id           int    `gorm:"primary_key;AUTO_INCREMENT"`
	StravaUserId int    `gorm:"unique;"`
	UserName     string `gorm:"size:1000" sql:"type:VARCHAR(1000) CHARACTER SET utf8 COLLATE utf8_general_ci"`
	UnlockCode   string `gorm:"size:7;"`
}

type UserClubDetails struct {
	Id           int    `gorm:"primary_key;AUTO_INCREMENT"`
	StravaUserId int    `gorm:"unique_index:idx_userid_clubid;"`
	ClubId       string `gorm:"size:100;unique_index:idx_userid_clubid;"`
	MessageText  string `gorm:"size:1000" sql:"type:VARCHAR(1000) CHARACTER SET utf8 COLLATE utf8_general_ci"`
}

type PendingUnlock struct {
	Id          int    `gorm:"primary_key;AUTO_INCREMENT"`
	SlackUserId string `gorm:"size:100;unique"`
}
