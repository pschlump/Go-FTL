
Passed as URL/POST value or cookie
	api_table_key, default "324d4b9f-00dc-4ea9-7a6c-e5f125207759" 
set in config as ApiTableKey - if set to "" then turned off (for login type siturations)

Can be set to "$XSRF_TOKEN$" - this will cause an $XSRF_TOKEN$ to be returned on login, and required
for every request - the token is a GUID that is generated and unique to that client.  It can be passed back
as api_table_key or xsrf_token or as the header X-XSRF-Token.

/Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/TabServer2/ts2_ftl.go:600

