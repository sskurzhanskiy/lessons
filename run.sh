#!/bin/bash

DATABASE_URL="postgres://postgres:qwerty@localhost:5432/habit?sslmode=disable"
JWT_SECRET="prod_jwt_secret_salt_Qwer4!yufdnms,;;fdfdfd"
HABBIT_CONFIG_NAME="prod mode"

RUN_COMMAND="go run ./cmd/api/..."

if [ $1 = "-t" ]; then
	DATABASE_URL="postgres://postgres:qwerty@localhost:5432/habit_test?sslmode=disable"
	JWT_SECRET="test_jwt_secret_salt_JlkSnlknJSNkjn768090S!"
	HABBIT_CONFIG_NAME="test mode"
fi

HABBIT_CONFIG_NAME="$HABBIT_CONFIG_NAME" DATABASE_URL="$DATABASE_URL" JWT_SECRET="$JWT_SECRET" go run ./cmd/api/...
