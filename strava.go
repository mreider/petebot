package main

import (
	"github.com/strava/go.strava"
	"log"
	"net/http"
	"sort"
	"time"
)

// Strava service.
type Strava struct {
	authenticator *strava.OAuthAuthenticator
	repo          *Repository
}

// Create new Strava service.
func NewStrava(repo *Repository, clientId int, clientSecret string, callbackUrl string) *Strava {
	strava.ClientId = clientId
	strava.ClientSecret = clientSecret
	return &Strava{
		repo: repo,
		authenticator: &strava.OAuthAuthenticator{
			CallbackURL: callbackUrl,
		},
	}
}

// Get Strava authorization URL.
func (s *Strava) AuthorizationUrl() string {
	return s.authenticator.AuthorizationURL("mystate", "write", true)
}

// Set Strava authorization callbacks.
func (s *Strava) SetCallbackHandler(
	success func(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request),
	failure func(err error, w http.ResponseWriter, r *http.Request)) {
	path, _ := s.authenticator.CallbackPath()
	http.HandleFunc(path, s.authenticator.HandlerFunc(success, failure))
}

// Get Strava club activities (sorted by activity id).
func (s *Strava) GetClubActivities(accessDetails *AccessDetails, clubId int64) ([]*strava.ActivitySummary, error) {
	client := strava.NewClient(accessDetails.StravaToken)
	service := strava.NewClubsService(client)
	activities, err := service.ListActivities(clubId).PerPage(200).Do()
	if err != nil {
		log.Printf("Failed to get club activities: %v\n", err)
		return nil, err
	}

	sort.Slice(activities, func(i, j int) bool { return activities[i].Id < activities[j].Id })

	return activities, nil
}

// Get Strava rate limits.
func (s *Strava) GetRateLimits() (requestTime time.Time, limitShort, limitLong, usageShort, usageLong int) {
	rateLimiting := strava.RateLimiting
	return rateLimiting.RequestTime, rateLimiting.LimitShort, rateLimiting.LimitLong, rateLimiting.UsageShort, rateLimiting.UsageLong
}
