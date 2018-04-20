package main

import (
	"fmt"
	"log"
	"strings"
)

// Slack command executor.
type CommandExecutor struct {
	settings               *Settings
	repo                   *Repository
	strava                 *Strava
	poster                 *ActivitiesPoster
	jobsExecutor           *JobsExecutor
	slack                  *Slack
	addedBots              chan *BotDetails
	commands               []Command
	activityTemplateEngine *TemplateEngine
	activitiesPoster       *ActivitiesPoster
	teamsUnlockInfo        *TeamsUnlockInfo
}

// Create new command executor.
func NewCommandExecutor(settings *Settings, repo *Repository, strava *Strava, poster *ActivitiesPoster, jobsExecutor *JobsExecutor, slack *Slack, activityTemplateEngine *TemplateEngine, activitiesPoster *ActivitiesPoster, teamsUnlockInfo *TeamsUnlockInfo) *CommandExecutor {
	commands := []Command{
		&HelpCommand{},
		&ClubCommand{},
		&UnitCommand{},
		//&UnlockCommand{},
		&DonateCommand{},
		&MessageCommand{},
		&RecentCommand{},
		//&AdminUnlockCommand{},
		&AdminInfoCommand{},
		&AdminUsersCommand{},
	}

	return &CommandExecutor{
		settings:               settings,
		repo:                   repo,
		strava:                 strava,
		poster:                 poster,
		jobsExecutor:           jobsExecutor,
		slack:                  slack,
		addedBots:              make(chan *BotDetails, 100),
		commands:               commands,
		activityTemplateEngine: activityTemplateEngine,
		activitiesPoster:       activitiesPoster,
		teamsUnlockInfo:        teamsUnlockInfo,
	}
}

// Run command executor.
func (e *CommandExecutor) Run() {
	bots := e.loadExistingBots()

	for {
		select {
		case bot := <-e.addedBots:
			bots[bot.TeamId] = bot
		case message := <-e.slack.IncomingMessages():
			if bot, ok := bots[message.TeamId]; ok {
				response, attachmentName, attachmentContent := e.processCommand(bot.BotId, message)
				if response != "" || attachmentName != "" {
					e.slack.PostMessage(&OutgoingSlackMessage{message.ChannelId, bot.BotAccessToken, response, attachmentName, attachmentContent})
				}
			}
		}
	}
}

// Load existing bot details from the database.
func (e *CommandExecutor) loadExistingBots() map[string]*BotDetails {
	bots := make(map[string]*BotDetails)
	existingBots, err := e.repo.BotDetails.List()
	if err != nil {
		log.Printf("Can't load bot details from the database: %v\n", err)
	}
	for _, bot := range existingBots {
		bots[bot.TeamId] = bot
	}
	return bots
}

// Add new bot details.
func (e *CommandExecutor) AddBot(bot *BotDetails) {
	e.addedBots <- bot
}

// Process bot command.
func (e *CommandExecutor) processCommand(botId string, message *IncomingSlackMessage) (response, attachmentName string, attachmentContent []byte) {
	messageText := strings.TrimSpace(message.Text)
	mention := fmt.Sprintf("<@%s>", botId)
	commandText := ""
	if messageText == mention {
		commandText = "help"
	} else {
		mention += " "
		if strings.HasPrefix(messageText, mention) {
			commandText = strings.TrimSpace(strings.TrimPrefix(messageText, mention))
		}
	}

	if commandText == "" {
		return "", "", nil
	}

	if strings.HasPrefix(commandText, "uploaded a file: ") {
		return "", "", nil
	}

	command, params := e.getCommand(commandText)

	if command == nil {
		return "Unknown command", "", nil
	}

	commandWithAttachment, ok := command.(CommandWithAttachment)
	if ok {
		response, attachmentName, attachmentContent, err := commandWithAttachment.ExecuteWithAttachment(params, message, e)
		if err != nil {
			return "Something is wrong", "", nil
		}
		return response, attachmentName, attachmentContent
	}

	response, err := command.Execute(params, message, e)
	if err != nil {
		log.Printf("Error when executing %s: %v\n", commandText, err)
		return "Something is wrong", "", nil
	}
	return response, "", nil
}

func (e *CommandExecutor) getCommand(commandText string) (Command, []string) {
	for _, command := range e.commands {
		commandName := strings.ToLower(command.Name())
		if strings.ToLower(commandText) == commandName || strings.HasPrefix(strings.ToLower(commandText), fmt.Sprintf("%s ", commandName)) {
			return command, parseParams(commandText[len(commandName):])
		}
	}
	return nil, nil
}

func (e *CommandExecutor) unlockUser(stravaUserId int, unlockCode string, teamId string) error {
	userDetails, err := e.repo.UserDetails.Get(stravaUserId)
	if err != nil {
		return err
	}

	if userDetails != nil && userDetails.UnlockCode != "" {
		return nil
	}

	if userDetails == nil {
		userDetails := &UserDetails{StravaUserId: stravaUserId, UnlockCode: unlockCode}
		err = e.repo.UserDetails.Create(userDetails)
		if err != nil {
			return err
		}
	} else {
		err = e.repo.UserDetails.Update(userDetails, map[string]interface{}{"UnlockCode": unlockCode})
		if err != nil {
			return err
		}
	}

	e.teamsUnlockInfo.AddUnlockUser(teamId)
	return nil
}

func (e *CommandExecutor) checkFromAdmin(message *IncomingSlackMessage) (string, error) {
	accessDetails, err := e.repo.AccessDetails.GetForUser(message.TeamId, message.UserId)
	if err != nil {
		return "", err
	}
	if accessDetails == nil {
		return COMMAND_STRAVA_NOT_CONNECTED, nil
	}
	if !e.settings.IsAdmin(accessDetails.StravaUserId) {
		return COMMAND_SHOULD_BE_ADMIN, nil
	}
	return "", nil
}

func parseParams(paramsStr string) []string {
	parts := strings.Split(paramsStr, " ")
	params := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			params = append(params, part)
		}
	}
	return params
}
