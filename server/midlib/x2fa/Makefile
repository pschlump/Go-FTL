
all: 
	go build


ddl.sql: ddl.m4.sql
	m4 -P ddl.m4.sql >ddl.sql

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


## Manual tests ## ## Manual tests ## ## Manual tests ## ## Manual tests ##
## Manual tests ## ## Manual tests ## ## Manual tests ## ## Manual tests ##
## Manual tests ## ## Manual tests ## ## Manual tests ## ## Manual tests ##

test2minHash:
	wget -o ,a -O ,b http://127.0.0.1:9019/api/x2fa/pull-2-min-hash
	grep OK ,a >/dev/null
	echo Output 
	cat ,b 

get1timeKey:
	wget -o ,a -O ,b 'http://127.0.0.1:9019/api/x2fa/gen-1-time-codes?auth_key=8181.2121&user_id=3290ce1d-14fa-414d-8759-4a323e40ad32&t2faId=901efa49-0484-4411-7433-f25c586d4274'
	grep OK ,a >/dev/null 
	echo Output 
	cat ,b 

#	wget -o ,a -O ,b 'http://127.0.0.1:9019/api/x2fa/gen-1-time-codes?auth_key=8181.2121&user_id=faea31c0-c4d8-4f34-87f1-e16d352d371f&t2faId=d683924e-1a8b-41ee-5008-c053a7114ab5'

validateANumber:
	wget -o ,a -O ,b 'http://127.0.0.1:9019/api/x2fa/is-valid-2fa?auth_key=8181.2121&user_id=3290ce1d-14fa-414d-8759-4a323e40ad32&val2fa=008513'
	grep OK ,a >/dev/null 
	echo Output 
	cat ,b 

# a9ed4ec4-05ec-4469-90ae-b372d6382396
