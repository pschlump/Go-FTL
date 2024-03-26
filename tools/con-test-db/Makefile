
all:
	go build

linux: 
	GOOS=linux go build -o con-test-db.linux .

test: test_setup test1 test2
	@echo PASS

test_setup:
	echo "You need to use psql to run ./test.sql before this test will succede in every database"
	go build

# password is from environment variable
#
# 	$ export DB_PASS=whatever
#
test1:
	./con-test-db -C "user=pschlump password=$(DB_PASS) dbname=so port=5432 host=192.168.0.139"

test2:
	./con-test-db -C "user=pschlump password=$(DB_PASS) dbname=test7 port=5432 host=192.168.0.139"

linux:
	GOOS=linux GOARCH=amd64 go build -o con-test-db_linux

deploy:
	scp con-test-db_linux philip@45.79.53.54:/home/philip/tmp

