
all:

# pwSupplied := ps.ByNameDflt("api_table_key", "")

post_t1:
	curl \
		-d '{"api_table_key":"kip.philip","category_id":"0004","desc":"Test+001+Description","title":"Test+001","tags":"a|b"}' \
		-H "Content-Type: application/json" \
		-X POST \
		'http://192.168.0.157:9018/api/table/dt_document?api_table_key=kip.philip'

post_t3:
	curl \
		-d 'api_table_key=kip.philip&category_id=0004&desc=Test+001+Description&title=Test+001&tags=a|b' \
		-H "Content-Type: application/x-www-form-urlencoded" \
		-X POST \
		'http://192.168.0.157:9018/api/table/dt_document?api_table_key=kip.philip'

post_t4:
	curl \
		-d 'api_table_key=kip.philip&category_id=0004&desc=Test+001+Description&title=Test+001&tags=a|b' \
		-H "Content-Type: application/x-www-form-urlencoded" \
		-X POST \
		'http://192.168.0.157:9018/api/table/dt_document'


post_t2:
	curl \
		-d 'api_table_key=kip.philip&category_id=0004&desc=Test+001+Description&title=Test+001' \
		-H "Content-Type: application/x-www-form-urlencoded" \
		-X POST \
		'http://192.168.0.157:9018/api/table/dt_document?api_table_key=kip.philip&desc=D1&title=t1&category_id=0004&tags=a|b|c'



x1:
	curl -X POST 'http://192.168.0.157:9018/api/table/dt_document?api_table_key=kip.philip&category_id=0004&desc=Test+001+Description&title=Test+001'
