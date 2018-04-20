package main

import (
	"bytes"
	"fmt"
	"github.com/strava/go.strava"
	"log"
	"strings"
	"time"
)

// Poster that posts activities to slack.
type ActivitiesPoster struct {
	slack                  *Slack
	repo                   *Repository
	activityTemplateEngine *TemplateEngine
}

// Create new activities poster.
func NewActivitiesPoster(slack *Slack, repo *Repository, activityTemplateEngine *TemplateEngine) *ActivitiesPoster {
	return &ActivitiesPoster{
		slack: slack,
		repo:  repo,
		activityTemplateEngine: activityTemplateEngine,
	}
}

// Post activities from specified club to slack channels.
func (p *ActivitiesPoster) PostActivities(clubId string, activities []*strava.ActivitySummary, jobs []*JobDetails) {
	lastSentActivityIdentifiers := p.loadLastSentActivityIdentifiers(clubId)
	bots := p.loadBotDetails(jobs)

	fromTime := time.Now().UTC().Add(-48 * time.Hour)
	for _, job := range jobs {
		clubDetails, _ := p.repo.ClubDetails.Get(clubId)
		clubUnit := DefaultUnit
		if clubDetails != nil {
			clubUnit = GetKnownUnit(clubDetails.Unit)
		}

		if bot, ok := bots[job.TeamId]; ok {
			lastSentId := lastSentActivityIdentifiers[job.ChannelId]
			newLastSentId := lastSentId
			activitiesLines := make([]string, 0, 4)

			for _, activity := range activities {
				if activity.StartDate.Before(fromTime) {
					continue
				}

				messageText := ""
				userClubDetails, err := p.repo.UserClubDetails.Get(int(activity.Athlete.Id), clubId)
				if err == nil && userClubDetails != nil {
					messageText = userClubDetails.MessageText
				}

				activityText := p.getActivityText(activity, messageText, clubUnit)
				if activity.Id > int64(lastSentId) {
					activitiesLines = append(activitiesLines, activityText)

					if newLastSentId < int(activity.Id) {
						newLastSentId = int(activity.Id)
					}
				}
			}
			if len(activitiesLines) > 0 {
				p.slack.PostMessage(&OutgoingSlackMessage{ChannelId: job.ChannelId, AccessToken: bot.BotAccessToken, Text: strings.Join(activitiesLines, "\n")})
			}
			if newLastSentId > lastSentId {
				p.saveUpdatedSentDetails(clubId, job.ChannelId, newLastSentId)
			}
		}
	}
}

// Post activities from specified club to slack channels.
func (p *ActivitiesPoster) GetActivitiesText(clubId string, activities []*strava.ActivitySummary) (string, error) {
	clubDetails, _ := p.repo.ClubDetails.Get(clubId)
	clubUnit := DefaultUnit
	if clubDetails != nil {
		clubUnit = GetKnownUnit(clubDetails.Unit)
	}

	activitiesLines := make([]string, 0, len(activities))
	for _, activity := range activities {
		messageText := ""
		userClubDetails, err := p.repo.UserClubDetails.Get(int(activity.Athlete.Id), clubId)
		if err == nil && userClubDetails != nil {
			messageText = userClubDetails.MessageText
		}

		activityText := p.getActivityText(activity, messageText, clubUnit)
		activityText = fmt.Sprintf("%s: %s", activity.StartDateLocal.Format("Mon, 02-Jan-06 15:04:05"), activityText)
		activitiesLines = append(activitiesLines, activityText)
	}
	return strings.Join(activitiesLines, "\n"), nil
}

// Load last sent activity identifiers (by channels) for specified club from the database.
func (p *ActivitiesPoster) loadLastSentActivityIdentifiers(clubId string) map[string]int {
	sentDetails, err := p.repo.SentDetails.List(clubId)
	if err != nil {
		log.Printf("Failed to load sent details from the database: %v\n", err)
	}

	sentInfo := make(map[string]int)
	for _, details := range sentDetails {
		sentInfo[details.ChannelId] = details.LastActivityId
	}
	return sentInfo
}

// Load bot details for specified jobs from the database.
func (p *ActivitiesPoster) loadBotDetails(jobs []*JobDetails) map[string]*BotDetails {
	bots := make(map[string]*BotDetails)

	for _, job := range jobs {
		bot, err := p.repo.BotDetails.Get(job.TeamId)
		if err != nil {
			log.Printf("Failed to get bot details from the database: %v\n", err)
		} else if bot != nil {
			bots[job.TeamId] = bot
		} else {
			log.Printf("Can't find bot details in the database for team=%s\n", job.TeamId)
		}
	}
	return bots
}

// Save updated sent details to the database.
func (p *ActivitiesPoster) saveUpdatedSentDetails(clubId string, channelId string, newLastSentId int) {
	log.Printf("Saving sent details to the database: clubId=%s, channelId=%s, newLastSentId=%d\n", clubId, channelId, newLastSentId)

	sentDetails, err := p.repo.SentDetails.Get(clubId, channelId)
	if err != nil {
		log.Printf("Failed to get sent details from the database: %v\n", err)
	} else if sentDetails == nil {
		sentDetails = &SentDetails{ClubId: clubId, ChannelId: channelId, LastActivityId: newLastSentId}
		err := p.repo.SentDetails.Create(sentDetails)
		if err != nil {
			log.Printf("Failed to save new sent details to the database: %v\n", err)
		}
	} else {
		err := p.repo.SentDetails.Update(sentDetails, map[string]interface{}{"LastActivityId": newLastSentId})
		if err != nil {
			log.Printf("Failed to update exisitng sent details in the database: %v\n", err)
		}
	}
}

// Get activity text.
func (p *ActivitiesPoster) getActivityText(activity *strava.ActivitySummary, messageText string, unit *Unit) string {
	activityUrl := fmt.Sprintf("https://strava.com/activities/%d", activity.Id)
	distanceInUnit := activity.Distance / unit.MeterFactor
	averageSpeed := float64(activity.MovingTime) / 60.0 / distanceInUnit
	averageSpeedMinutes := int(averageSpeed)
	averageSpeedSeconds := int((averageSpeed-float64(averageSpeedMinutes))*60 + 0.5)
	averageSpeedStr := fmt.Sprintf("%d.%02d", averageSpeedMinutes, averageSpeedSeconds)

	if messageText == "" {
	} else {
		tmpl, err := p.activityTemplateEngine.Compile(messageText)
		if err == nil {
			var buf bytes.Buffer
			tmpl.Execute(&buf, activity)
			return fmt.Sprintf("%s %s", buf.String(), activityUrl)
		}
	}

	return fmt.Sprintf("%s %s %.2f %s at %s /%s %s", activity.Athlete.FirstName, getActivityTypeText(activity.Type), distanceInUnit, unit.Name, averageSpeedStr, unit.Name, activityUrl)
}

func getActivityTypeText(activityType strava.ActivityType) string {
	result := string(activityType)
	switch activityType {
	case strava.ActivityTypes.Run:
		result = "ran"
	case strava.ActivityTypes.Swim:
		result = "swam"
	case strava.ActivityTypes.Walk:
		result = "walked"
	case strava.ActivityTypes.Ride:
		result = "rode"
	}
	return strings.ToLower(result)
}
