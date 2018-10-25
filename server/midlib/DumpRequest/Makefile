
all:
	go build

.PHONY: test

test:
	go test


doc:
	markdown-cli -i DumpRequest.md -o DumpRequest.html -c ../markdown-cfg.json


mkdocs:
	cat DumpRequest.md  >>../tmp2
	echo "DumpRequest/DumpRequest.md" >>../m.list

