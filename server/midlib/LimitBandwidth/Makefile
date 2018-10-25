:q!
all:
	go build

test:
	go test

doc:
	markdown-cli -i LimitBandwidth.md -o LimitBandwidth.html -c ../markdown-cfg.json


mkdocs:
	cat LimitBandwidth.md  >>../tmp2
	echo "LimitBandwidth/LimitBandwidth.md" >>../m.list

in:
	watson start go-ftl +limit-bandwidth

out:
	watson stop
