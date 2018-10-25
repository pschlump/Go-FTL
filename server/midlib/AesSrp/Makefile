
all:
	go build


update_doc_file:
	perl upd_doc_file.pl *.go

doc:
	markdown-cli -i AesSrp.md -o AesSrp.html -c ../markdown-cfg.json

mkdocs:
	cat AesSrp.md  >>../tmp2
	echo "AesSrp/AesSrp.md" >>../m.list

