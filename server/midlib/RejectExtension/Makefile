
all:
	go build

test:
	go test

doc:
	markdown-cli -i RejectExtension.md -o RejectExtension.html -c ../markdown-cfg.json


mkdocs:
	cat RejectExtension.md  >>../tmp2
	echo "RejectExtension/RejectExtension.md" >>../m.list

