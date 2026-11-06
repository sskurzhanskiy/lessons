#!/bin/bash

login="joe_don@test.com"
password="TestPassword"

if [ $# -eq 2 ]; then
  login=$1
  password=$2
fi

json_data=$(cat <<EOF
{
  "name":"Joe",
  "email":"$login",
  "password":"$password",
  "age":27
}
EOF
)

curl -X POST -H "Content-Type:application/json" -H "Accept:application/json" -d "$json_data" http://127.0.0.1:8080/auth/register
