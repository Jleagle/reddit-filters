#!/bin/sh

echo "### Pulling"
git fetch origin
git reset --hard origin/master

echo "### Building"
dep ensure
go build

echo "### Talking to Rollbar"
curl https://api.rollbar.com/api/1/deploy/ \
  -F access_token=${REDDIT_ROLLBAR_PRIVATE} \
  -F environment=${ENV} \
  -F revision=$(git log -n 1 --pretty=format:"%H") \
  -F local_username=${REDDIT_ROLLBAR_USER} \
  --silent > /dev/null

echo "### Restarting"
/etc/init.d/site.reddit.sh restart
