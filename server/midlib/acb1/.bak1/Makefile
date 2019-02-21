
all:
	go build


test:
	go test

# * 1. JSON Data
# * 2. Curl Post - to what URL
# * 3. Config Server to use /api/acb1 URL path
# * 4. Restart server ( $aa/pvp/
# * 5. Get Output
# + 6. Validate
test001:
	mkdir -p ./out
	./test001.sh >out/test001.out 2>&1

# 7. Add in other fields
#   7a. seq
#   7b. Timestamp
# # site_id -> presmis_id
#	look up using premis_id and premis_code
# # Add in a table for Premis "Demo Ranch", "Demo Wyoming".
#
# 8. Add a insert/update capability
# 9. Setup data for X=15 cows
# 
# QR generation capablity -
#	1. Have a set of QR codes already built.
#   2. Use one
#   3. Wtite out the data
#   4. Update the URL shortner with new data
#

doc:
	markdown-cli -i Acb1.md -o Acb1.html -c ../markdown-cfg.json


mkdocs:
	cat Acb1.md >>../tmp2
	echo "Acb1/Acb1.md" >>../m.list

