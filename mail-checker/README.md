# Prerequisites

1. Install Go
2. Download **dep** tool from https://github.com/golang/dep/releases
3. Clone code from git into $GOPATH/src/github.com/mreider/petebot
 

# Build

1. Change current directory to $GOPATH/src/github.com/mreider/petebot
2. Pull fresh code from git
3. Execute **dep ensure** command
4. Change current directory to subdirectory mail-checker
5. Execute **go build** command

# Setup 
0. Modify **settings.local.json** if need (if it doesn't exist copy it from **settings.default.json**) - setup the same DATABASE_URI as for bot service 
1. https://cloud.google.com/pubsub/docs/quickstart-console
   Do "Before you begin", Step 1 - create a project
   Do "Create a topic", Steps 1-3
2. Modify **settings.local.json** - set TOPIC_NAME to full topic name (like projects/email-confirm-1506360046231/topics/TestTopic)
3. Go to developer console: https://console.developers.google.com/
4. Select the created project
5. Enable GMail API for the project
6. Fill "Product name shown to users" on "Credentials"/"OAuth consent screen"
7. Create OAuth 2.0 client ID (Application Type: other)
8. Download generated client secret json file and put its content into a new file '.credentials/client_secret.json'
9. Ensure DB is running
10. Run the service **./mail-checker**
11. On the first run it should ask you to go to the following link in your browser
12. Do authentication using Gmail account that will be used for mail confirmation.
13. Type (paste) the authorization code to the console
14. The service should print the account email and keep running (if not, look at service.log).
15. Go to "Credentials"/"Domain verification" and verify your domain
16. Return to https://console.cloud.google.com/cloudpubsub/topicList and add new PUSH subscription using endpoint URL https://<your-domain>/push
17. Grant publish rights (Pub/Sub Editor) on your topic to gmail-api-push@system.gserviceaccount.com. Details: section "Grant publish rights on your topic" https://developers.google.com/gmail/api/guides/push
18. Restart the service


#Run
1. Ensure DB is running
2. Run the service **./mail-checker**
3. The service should print the account email and keep running (if not, look at service.log).


#Test
1. If you send test letter from some account to the service account that contains words "rye brook" or "choice-in-dying", you should get a response letter with dummy confirmation code