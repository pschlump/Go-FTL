
all:
	go build

test:
	go test

doc:
	markdown-cli -i RewriteProxy.md -o RewriteProxy.html -c ../markdown-cfg.json


mkdocs:
	cat RewriteProxy.md  >>../tmp2
	echo "RewriteProxy/RewriteProxy.md" >>../m.list

