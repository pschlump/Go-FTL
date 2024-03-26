Overview - Features in Detail
=============================
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Features"
	,	"SubSectionGroup": "Overview"
	,	"SubSectionTitle": "Overview - Go-FTL - Features in Detail"
	,	"SubSectionTooltip": "Go-FTL Features in Detail"
	, 	"MultiSection":2
	}
```

In no particular order...

1. Name based server -- differentiate between http://www.go-ftl.com/ and http://docs.go-ftl.com/ and http://www.2c-why.com/ etc.
2. Configurable stack of middleware services.  Lots of middleware already built:
<ul>
<li> SrpAesAuth: Strong Authentication for RESTful Requests </li>
<li> BasicAuth: Implement Basic Authentication Using a .htaccess File </li>
<li> BasicAuthPgSql: Basic Auth Using PostgreSQL </li>
<li> BasicAuthRedis: Basic Auth using Redis </li>
<li> Cookie: Set/Delete Cookies </li>
<li> DirectoryBrowse: Use Template for Directory Browsing </li>
<li> DumpReq: Dump Request with Message to Output File - Development Tool </li>
<li> DumpResponse: Dump Request with Message to Output File - Development Tool </li>
<li> Echo: Output a Message When End Point Reached </li>
<li> Else: Return a Page for a Failed Virtual Host Name or SNI Match </li>
<li> ErrorTemplate: Convert Errors to Pages </li>
<li> GeoIpFilter: Filer Requests Based on Geographic Mapping of IP Address </li>
<li> GoTemplate: Template using Go's Buit in Templates </li>
<li> Gzip: Ban Certain IP Addresses </li>
<li> HTML5Path: Redirect 404 Errors to Index.html for AngularJS Router </li>
<li> Header: Set/Delete Headers </li>
<li> InMemoryCache: Ban Certain IP Address </li>
<li> JSONToTable: Convert JSON to Internal Table Data </li>
<li> JSONp: Implement JSONp requests </li>
<li> LimitExtensionTo: Limit Requests Based on File Extension </li>
<li> LimitRePathTo: Limit Requests Based on File Extension </li>
<li> LimitPathTo: Limit Requests Based on File Extension </li>
<li> Logging: Output a Log Message for Every Request </li>
<li> LoginRequired: Middleware After this Require Login </li>
<li> Minify: Compress/Minify Files Before Serving Them </li>
<li> RedirectToHttps: Redirect One Request to Another Location </li>
<li> Redirect: Redirect One Request to Another Location </li>
<li> RedisList: Return Data from Redis </li>
<li> RedisListRaw: Return Data from Redis </li>
<li> RejectDirectory: Prevent Browsing of a Set of Directories </li>
<li> RejectExtension: Reject Requests Based on File Extension </li>
<li> RejectHotlink: Reject requests based on invalid referer header </li>
<li> RejectIPAddress: Ban Certain IP Address </li>
<li> RejectPath: Reject Requests Based on the Path </li>
<li> RejectRePath: Reject Requests Based on a Regular Expression Path Match </li>
<li> Rewrite: Rewrite One Request to Another Location </li>
<li> RewriteProxy: Rewrite Reqeust and Proxy It to a Different Server </li>
<li> Status: Echo a Request as JSON Data </li>
</ul>
3. You can easily build your own middleware.
4. Syntax and other checks on all configuration.
5. MIT or similar licensed -- this allows for embedding the server into devices.
6. Strong Authentication with Two Factor
7. xyzzy

