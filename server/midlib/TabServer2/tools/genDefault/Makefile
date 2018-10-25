
all:
	go build

example1:
	echo "creating table"
	psql -a -P pager=off -h 127.0.0.1 -U pschlump <../../example/feedback.sql
	echo "creating example"
	go run ./main.go feedback >sample.out


