package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	ClientSecretFileName = "client_secret.json"
	CachedTokentFileName = "gmail-access.json"
)

// Based on https://developers.google.com/gmail/api/quickstart/go
type GmailAuthenticator struct {
	credentialsDir string
	ctx            context.Context
	config         *oauth2.Config
}

func NewGmailAuthenticator(appDir string) *GmailAuthenticator {
	credentialsDir := filepath.Join(appDir, ".credentials")

	ctx := context.Background()

	b, err := ioutil.ReadFile(filepath.Join(credentialsDir, ClientSecretFileName))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return &GmailAuthenticator{credentialsDir: credentialsDir, ctx: ctx, config: config}
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func (a *GmailAuthenticator) getClient() *http.Client {
	cacheFilePath := filepath.Join(a.credentialsDir, CachedTokentFileName)
	token, err := a.tokenFromFile(cacheFilePath)
	if err != nil {
		token = a.getTokenFromWeb()
		saveToken(cacheFilePath, token)
	}
	return a.config.Client(a.ctx, token)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func (a *GmailAuthenticator) getTokenFromWeb() *oauth2.Token {
	authURL := a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := a.config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func (a *GmailAuthenticator) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(filePath string, token *oauth2.Token) {
	fmt.Printf("Saving credential filePath to: %s\n", filePath)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
