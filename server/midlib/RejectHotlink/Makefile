
all:
	go build

test:
	go test

doc:
	markdown-cli -i RejectHotLink.md -o RejectHotLink.html -c ../markdown-cfg.json


mkdocs:
	cat RejectHotLink.md  >>../tmp2
	echo "RejectPath/RejectHotLink.md" >>../m.list

