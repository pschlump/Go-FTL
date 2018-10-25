
all:
	go build

test:
	go test

doc:
	markdown-cli -i BasicAuthPgSQL.md -o BasicAuthPgSQL.html -c ../markdown-cfg.json

mkdocs:
	cat BasicAuthPgSQL.md  >>../tmp2
	echo "BasicAuthPgSQL/BasicAuthPgSQL.md" >>../m.list

