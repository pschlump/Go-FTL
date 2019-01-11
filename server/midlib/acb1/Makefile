
all:
	go build


test:
	go test

test001_plan:
	echo "*" 1. JSON Data
	echo "++" 2. Curl Post - to what URL
	echo "*" 3. Config Server to use /api/acb1 URL path
	echo 4. Restart server ( $aa/pvp/
	echo 5. Get Output
	echo 6. Validate

test001:
	curl -X POST http://localhost:9019/api/acb1/track_add?bulk=

doc:
	markdown-cli -i Acb1.md -o Acb1.html -c ../markdown-cfg.json


mkdocs:
	cat Acb1.md >>../tmp2
	echo "Acb1/Acb1.md" >>../m.list

