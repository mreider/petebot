kill $(ps aux | grep '[p]ete-go-bot' | awk '{print $2}')
kill $(ps aux | grep '[m]ail-checker' | awk '{print $2}')
nohup ./petebot  > /dev/null 2>&1 &
cd mail-checker
nohup ./mail-checker  > /dev/null 2>&1 &
