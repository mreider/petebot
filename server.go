package main

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/nlopes/slack"
	"github.com/strava/go.strava"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

// HTTP Server.
type Server struct {
	settings          *Settings
	repo              *Repository
	appDir            string
	templatesRegistry *TemplatesRegistry
	strava            *Strava
	slack             *Slack
	cookieStore       *sessions.CookieStore
	poster            *ActivitiesPoster
	jobsExecutor      *JobsExecutor
	commandExecutor   *CommandExecutor
}

// Create new HTTP server.
func NewServer(settings *Settings, repo *Repository, appDir string) *Server {
	cookieStore := sessions.NewCookieStore([]byte("something-very-secret")) // We don't store secret information (tokens) permanently.
	cookieStore.Options.HttpOnly = true
	return &Server{
		settings:    settings,
		repo:        repo,
		appDir:      appDir,
		cookieStore: cookieStore,
	}
}

// Run the server.
func (s *Server) Run() {
	s.initialize()

	fs := http.FileServer(http.Dir(filepath.Join(s.appDir, "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", s.home)
	http.HandleFunc("/login", s.login)
	http.HandleFunc("/slack-auth", s.slackAuth)
	http.HandleFunc("/bot-command", s.slack.BotCommand)
	http.HandleFunc("/privacy", s.privacy)
	http.HandleFunc("/privacy/", s.privacy)
	s.strava.SetCallbackHandler(s.oAuthStravaSuccess, s.oAuthStravaFailure)

	log.Fatal(http.ListenAndServe(s.settings.ServerInternalAddress, context.ClearHandler(http.DefaultServeMux)))
}

// Initialize the server.
func (s *Server) initialize() {
	s.templatesRegistry = NewTemplatesRegistry(filepath.Join(s.appDir, "templates"))
	err := s.templatesRegistry.LoadTemplates()
	if err != nil {
		log.Fatalf("Can't load templates: %v\n", err)
	}

	s.strava = NewStrava(s.repo, s.settings.StravaOauthClientId, s.settings.StravaOauthClientSecret, fmt.Sprintf("%s/authcallback", s.settings.ServerPublicAddress))
	s.slack = NewSlack(s.settings.SlackOauthClientId, s.settings.SlackOauthClientSecret, fmt.Sprintf("%s/slack-auth", s.settings.ServerPublicAddress))
	go s.slack.PostOutgoingMessages()

	activityTemplateEngine := NewActivityTemplateEngine(s.appDir)

	s.poster = NewActivitiesPoster(s.slack, s.repo, activityTemplateEngine)

	stravaCommonMonitoringInterval := time.Duration(s.settings.StravaCommonMonitoringIntervalInMinutes) * time.Minute
	stravaUnlockedMonitoringInterval := time.Duration(s.settings.StravaUnlockedMonitoringIntervalInMinutes) * time.Minute

	unlockedTeams, err := s.repo.JobDetails.ListTeamsWithUnlockedUsers()
	if err != nil {
		log.Fatalf("Can't get teams with unlocked users from the database: %v\n", err)
	}
	teamsUnlockInfo := NewTeamsUnlockInfo(unlockedTeams)
	s.jobsExecutor = NewJobsExecutor(s.strava, stravaCommonMonitoringInterval, stravaUnlockedMonitoringInterval, s.poster, teamsUnlockInfo, s.repo)
	go s.jobsExecutor.Run()

	s.commandExecutor = NewCommandExecutor(s.settings, s.repo, s.strava, s.poster, s.jobsExecutor, s.slack, activityTemplateEngine, s.poster, teamsUnlockInfo)
	go s.commandExecutor.Run()
}

// Home page.
func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	session := s.session(r)
	if session.Values["strava_user"] == nil {
		s.templatesRegistry.RenderTemplate(w, "home.html", &TemplateData{})
	} else {
		templateData := &TemplateData{
			SlackRedirectUrl: s.slack.AuthorizationUrl(),
			Session:          make(map[string]interface{}),
		}

		if stravaUser, ok := session.Values["strava_user"]; ok && stravaUser != "" {
			templateData.Session["connected_to_strava"] = true
		} else if stravaToken, ok := session.Values["strava_access_token"]; ok && stravaToken != "" {
			templateData.Session["connected_to_strava"] = true
		}

		if slackUser, ok := session.Values["slack_user"]; ok && slackUser != "" {
			templateData.Session["connected_to_slack"] = true
		} else if slackToken, ok := session.Values["slack_access_token"]; ok && slackToken != "" {
			templateData.Session["connected_to_slack"] = true
		}

		if slackError, ok := session.Values["slack_error"]; ok && slackError != "" {
			templateData.Session["slack_error"] = slackError
		}

		s.templatesRegistry.RenderTemplate(w, "userhome.html", templateData)
	}
}

// Strava login.
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	authorizationUrl := s.strava.authenticator.AuthorizationURL("mystate", "public", true)
	http.Redirect(w, r, authorizationUrl, http.StatusFound)
}

// Strava successful authorization callback.
func (s *Server) oAuthStravaSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	session := s.session(r)
	session.Values["strava_access_token"] = auth.AccessToken
	session.Values["strava_user"] = auth.Athlete.FirstName + " " + auth.Athlete.LastName
	session.Values["strava_user_id"] = int(auth.Athlete.Id)
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

// Strava failed authorization callback.
func (s *Server) oAuthStravaFailure(err error, w http.ResponseWriter, r *http.Request) {
	log.Printf("Strava OAuth failure: %v\n", err)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Slack authorization callback.
func (s *Server) slackAuth(w http.ResponseWriter, r *http.Request) {
	session := s.session(r)
	resp, err := s.slack.Authorize(w, r)
	if err != nil {
		log.Printf("Slack OAuth error: %v\n", err)
		session.Values["slack_error"] = "Sorry, there was a problem."
	} else {
		session.Values["slack_user"] = resp.UserID
	}
	session.Save(r, w)

	if err == nil {
		stravaToken, _ := session.Values["strava_access_token"].(string)
		stravaUserId, _ := session.Values["strava_user_id"].(int)
		stravaUserName, _ := session.Values["strava_user"].(string)
		_, err := s.createOrUpdateAccessDetails(resp, stravaToken, stravaUserId, stravaUserName)
		delete(session.Values, "strava_access_token")
		session.Save(r, w)

		bot, err := s.createOrUpdateBotDetails(resp)
		if err == nil {
			s.commandExecutor.AddBot(bot)
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// Get http cookie session.
func (s *Server) session(r *http.Request) *sessions.Session {
	session, _ := s.cookieStore.Get(r, "session")
	return session
}

// Create or update access details in the database.
func (s *Server) createOrUpdateAccessDetails(resp *slack.OAuthResponse, stravaToken string, stravaUserId int, stravaUserName string) (*AccessDetails, error) {
	userDetails, err := s.repo.UserDetails.Get(stravaUserId)
	if err != nil {
		log.Printf("Can't get user details from the database: %v\n", err)
		return nil, err
	}

	if userDetails != nil {
		err := s.repo.UserDetails.Update(userDetails, map[string]interface{}{"UserName": stravaUserName})
		if err != nil {
			log.Printf("Can't update user details in the database: %v\n", err)
			return nil, err
		}
	} else {
		userDetails = &UserDetails{StravaUserId: stravaUserId, UserName: stravaUserName}
		err := s.repo.UserDetails.Create(userDetails)
		if err != nil {
			log.Printf("Can't insert user details to the database: %v\n", err)
			return nil, err
		}
	}

	access, err := s.repo.AccessDetails.GetForUser(resp.TeamID, resp.UserID)
	if err != nil {
		log.Printf("Can't get access details from the database: %v\n", err)
		return nil, err
	}

	if access != nil {
		err := s.repo.AccessDetails.Update(access, map[string]interface{}{"SlackToken": resp.AccessToken, "StravaToken": stravaToken, "StravaUserId": stravaUserId})
		if err != nil {
			log.Printf("Can't update access details in the database: %v\n", err)
			return nil, err
		}
	} else {
		access = &AccessDetails{TeamId: resp.TeamID, SlackToken: resp.AccessToken, StravaToken: stravaToken, SlackUserId: resp.UserID, StravaUserId: stravaUserId}
		err := s.repo.AccessDetails.Create(access)
		if err != nil {
			log.Printf("Can't insert access details to the database: %v\n", err)
			return nil, err
		}

		pendingLocks, err := s.repo.PendingUnlocks.Get(resp.UserID)
		if err != nil {
			log.Printf("Can't get pending locks from the database: %v\n", err)
		}
		if pendingLocks == nil {
			pendingLocks, err = s.repo.PendingUnlocks.Get(strconv.Itoa(stravaUserId))
			if err != nil {
				log.Printf("Can't get pending locks from the database: %v\n", err)
			}
		}
		if pendingLocks != nil {
			err := s.commandExecutor.unlockUser(stravaUserId, "ADMIN", resp.TeamID)
			if err != nil {
				log.Printf("Can't insert/update user details in the database: %v\n", err)
			} else {
				err := s.repo.PendingUnlocks.Delete(pendingLocks)
				if err != nil {
					log.Printf("Can't delete pending locks from the database: %v\n", err)
				}
			}
		}
	}
	return access, nil
}

// Create or update bot details in the database.
func (s *Server) createOrUpdateBotDetails(resp *slack.OAuthResponse) (*BotDetails, error) {
	bot, err := s.repo.BotDetails.Get(resp.TeamID)
	if err != nil {
		log.Printf("Can't get bot details from the database: %v\n", err)
		return nil, err
	}

	if bot != nil {
		if bot.BotId != resp.Bot.BotUserID || bot.BotAccessToken != resp.Bot.BotAccessToken {
			err := s.repo.BotDetails.Update(bot, map[string]interface{}{"BotId": resp.Bot.BotUserID, "BotAccessToken": resp.Bot.BotAccessToken})
			if err != nil {
				log.Printf("Can't update bot details in the database: %v\n", err)
				return nil, err
			}
		}
	} else {
		bot = &BotDetails{TeamId: resp.TeamID, BotId: resp.Bot.BotUserID, BotAccessToken: resp.Bot.BotAccessToken}
		s.repo.BotDetails.Create(bot)
		if err != nil {
			log.Printf("Can't insert bot details to the database: %v\n", err)
			return nil, err
		}
	}
	return bot, nil
}

func (s *Server) privacy(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.appDir, "templates", "privacy.html"))
}
