test_db_url="postgres://postgres:qwerty@localhost:5432/habit_test?sslmode=disable"
test_path="./internal/habit/..."
if [ $# -eq 1 ]; then
	test_path=$1
fi

TEST_DATABASE_URL=$test_db_url go test -v $test_path
