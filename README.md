# Prerequisites

1. Install Go
2. Download **dep** tool from https://github.com/golang/dep/releases
3. Clone code from git into $GOPATH/src/github.com/mreider/petebot
 

# Build and Run

1. Change current directory to $GOPATH/src/github.com/mreider/petebot
2. Pull fresh code from git
3. Execute **dep ensure** command
4. Execute **go build** command
5. Modify **settings.local.json** if need (if it doesn't exist copy it from **settings.default.json**)
6. Ensure DB is running
7. Run the server **./petebot**
8. Open http://localhost:5000

# Update Slack Notification URL 
  Go to your app in https://api.slack.com/apps/  
  Update the correct url in event subscription section  
  It will be http://<your-domain>/bot-command 