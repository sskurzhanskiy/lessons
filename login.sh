#!/bin/bash

LOGIN="joe_don@test.com"
PASSWORD="TestPassword"

if [ $# -eq 2 ]; then
	LOGIN=$1
	PASSWORD=$2

fi

JSON_DATA=$(cat <<EOF
{
	"email":"$LOGIN",
	"password":"$PASSWORD"
}
EOF
)

curl -v -X POST -H "Content-Type:application/json" -H "Accept:application/json" -d "$JSON_DATA" http://127.0.0.1:8080/auth/login
