#!/usr/bin/env bash

# Username as the first argument.
username=$1

# Password from STDIN.
password=$(cat -)

# Run the curl command and dump to jq to grab the token.
TOKEN=$(curl --header "Content-Type: application/json" \
  --silent \
  --request POST \
  --data "{\"username\":\"$username\",\"password\":\"$password\",\"grant_type\":\"password\",\"scope\":\"manage\",\"expires_in\":43200,\"_source\":\"dashboard\"}" \
  https://auth-api-master.aptible-staging.com/tokens | jq -r '.access_token')

# If no token then exit with error
if [[ -z $TOKEN || $TOKEN == "null" ]]; then
  exit 1
fi

echo $TOKEN
