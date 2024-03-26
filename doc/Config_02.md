Config - Per Server  Configuraiton
==================================
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Per-Server Config"
	,	"SubSectionGroup": "Server Config"
	,	"SubSectionTitle": "Go-FTL - Configuration of a server"
	,	"SubSectionTooltip": "Go-FTL Configure the set of middleware that the server will use."
	, 	"MultiSection":2
	}
```


Go-FTL is all about middleware.  Other than top level routing of host names all the work is handled by middleware.
This means that there is lots of middleware.  Also it is easy to create your own.

If you find that you are in need of some sort of middleware that has not yet been written it is reasonably likely
that you can get it by creating an "issue" and explaining what you need.   

It is also simple to create your own.   Middleware is included by adding a line to `Go-FTL/server/goftl/inc.go`.
Usually I create a new middleware component by copying an existing one, renaming key things and then including
it.  My most likely target for copying is `Go-FTL/server/midlib/DumpRequest`.

Overview of Existing Middleware
-------------------------------

Each of these is configured in the `plugins` section.

For example the `file_server` middleware sets the Root directory and the http path that the Root directory will match with.

``` json

	{
		"demo1": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:8080", "http://dev2.test1.com:8080" ],
			"plugins":[
				{ "file_server": { "LineNo":__LINE__, "Root":"/www", "Paths":"/"  } }
			]
		}
	}

```

Note: the `file_server` middleware has lots of other configurable features also.

<ul>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-SrpAesAuth-Strong-Authentication-for-RESTful-Requests">
<div class="doc-title">SrpAesAuth: Strong Authentication for RESTful Requests</div>
<div class="doc-subtitle"> Strong authentication using Secure Remote Password (SRP), Two Factor Authrization (2FA) and encryption of messages with Advanced Encryption Standard (AES)</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-BasicAuth-Implement-Basic-Authentication-Using-a-htaccess-File">
<div class="doc-title">BasicAuth: Implement Basic Authentication Using a .htaccess File</div>
<div class="doc-subtitle"> Basic Auth implemented with a flat file for hashed usernames/passwords</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-BasicAuthPgSql-Basic-Auth-Using-PostgreSQL">
<div class="doc-title">BasicAuthPgSql: Basic Auth Using PostgreSQL</div>
<div class="doc-subtitle"> Basic Auth implemented with data stored in PostgreSQL</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-BasicAuthRedis-Basic-Auth-using-Redis">
<div class="doc-title">BasicAuthRedis: Basic Auth using Redis</div>
<div class="doc-subtitle"> Basic Auth implemented with data stored in Redis</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Cookie-Set-Delete-Cookies">
<div class="doc-title">Cookie: Set/Delete Cookies</div>
<div class="doc-subtitle"> Manipulation of cookies</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-DirectoryBrowse-Use-Template-for-Directory-Browsing">
<div class="doc-title">DirectoryBrowse: Use Template for Directory Browsing</div>
<div class="doc-subtitle"> Control layout and availabity of directory browsing with Go templates</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-DumpReq-Dump-Request-with-Message-to-Output-File-Development-Tool">
<div class="doc-title">DumpReq: Dump Request with Message to Output File - Development Tool</div>
<div class="doc-subtitle"> Dump out the contents of the reqeust at ths point in the middlware stack.</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-DumpResponse-Dump-Request-with-Message-to-Output-File-Development-Tool">
<div class="doc-title">DumpResponse: Dump Request with Message to Output File - Development Tool</div>
<div class="doc-subtitle"> Dump out the contents of the response at ths point in the middlware stack.</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Echo-Output-a-Message-When-End-Point-Reached">
<div class="doc-title">Echo: Output a Message When End Point Reached</div>
<div class="doc-subtitle"> Output a message to the log</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Else-Return-a-Page-for-a-Failed-Virtual-Host-Name-or-SNI-Match">
<div class="doc-title">Else: Return a Page for a Failed Virtual Host Name or SNI Match</div>
<div class="doc-subtitle"> This may not be working yet.  Under Construction</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-ErrorTemplate-Convert-Errors-to-Pages">
<div class="doc-title">ErrorTemplate: Convert Errors to Pages</div>
<div class="doc-subtitle"> Extended loggin with additional attributes via a template stubstitution</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-GeoIpFilter-Filer-Requests-Based-on-Geographic-Mapping-of-IP-Address">
<div class="doc-title">GeoIpFilter: Filer Requests Based on Geographic Mapping of IP Address</div>
<div class="doc-subtitle"> Use IP address to filter to a set of geograpic regions</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-GoTemplate-Template-using-Go-s-Buit-in-Templates">
<div class="doc-title">GoTemplate: Template using Go's Buit in Templates</div>
<div class="doc-subtitle"> Use Go Templates to format data</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Gzip-Ban-Certain-IP-Addresses">
<div class="doc-title">Gzip: Ban Certain IP Addresses</div>
<div class="doc-subtitle"> Gzip compresses output before it is returned.  Interacts with caching so 'zip' process only happens if file changed</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-HTML5Path-Redirect-404-Errors-to-Index-html-for-AngularJS-Router">
<div class="doc-title">HTML5Path: Redirect 404 Errors to Index.html for AngularJS Router</div>
<div class="doc-subtitle"> Angular 1.x, 2.x and other HTML5 single pages applications uses multiple URLs that all need to direct to a single .html page.</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Header-Set-Delete-Headers">
<div class="doc-title">Header: Set/Delete Headers</div>
<div class="doc-subtitle"> Manipulation of response heades</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-InMemoryCache-Ban-Certain-IP-Address">
<div class="doc-title">InMemoryCache: Ban Certain IP Address</div>
<div class="doc-subtitle"> Implements in memory caching of hot resources and on disk caching for other pages</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-JSONToTable-Convert-JSON-to-Internal-Table-Data">
<div class="doc-title">JSONToTable: Convert JSON to Internal Table Data</div>
<div class="doc-subtitle"> Format data into JSON</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-JSONp-Implement-JSONp-requests">
<div class="doc-title">JSONp: Implement JSONp requests</div>
<div class="doc-subtitle"> Transorm get reqeusts into JSONp if they have a callback parameter</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-LimitExtensionTo-Limit-Requests-Based-on-File-Extension">
<div class="doc-title">LimitExtensionTo: Limit Requests Based on File Extension</div>
<div class="doc-subtitle"> Prevent access to non authorized paths</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-LimitRePathTo-Limit-Requests-Based-on-File-Extension">
<div class="doc-title">LimitRePathTo: Limit Requests Based on File Extension</div>
<div class="doc-subtitle"> Prevent access to non authorized file extensiosn by limiting to a set of valid extensions</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-LimitPathTo-Limit-Requests-Based-on-File-Extension">
<div class="doc-title">LimitPathTo: Limit Requests Based on File Extension</div>
<div class="doc-subtitle"> Prevent access to non authorized directories</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Logging-Output-a-Log-Message-for-Every-Request">
<div class="doc-title">Logging: Output a Log Message for Every Request</div>
<div class="doc-subtitle"> Add or remove loggin information using templates for log messages</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-LoginRequired-Middleware-After-this-Require-Login">
<div class="doc-title">LoginRequired: Middleware After this Require Login</div>
<div class="doc-subtitle"> Require login before allowing access to the specified paths below this in the middleware stack</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Minify-Compress-Minify-Files-Before-Serving-Them">
<div class="doc-title">Minify: Compress/Minify Files Before Serving Them</div>
<div class="doc-subtitle"> Shrink output using minification techniques.  Compress CSS, JavaScript, HTML, SVG, XML and JSON data.</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RedirectToHttps-Redirect-One-Request-to-Another-Location">
<div class="doc-title">RedirectToHttps: Redirect One Request to Another Location</div>
<div class="doc-subtitle"> Client side (307) response redirects to HTTPS</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Redirect-Redirect-One-Request-to-Another-Location">
<div class="doc-title">Redirect: Redirect One Request to Another Location</div>
<div class="doc-subtitle"> Client side (307) response redirects</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RedisList-Return-Data-from-Redis">
<div class="doc-title">RedisList: Return Data from Redis</div>
<div class="doc-subtitle"> Provide limited access to data in Redis based on prefixes to a set of keys</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RedisListRaw-Return-Data-from-Redis">
<div class="doc-title">RedisListRaw: Return Data from Redis</div>
<div class="doc-subtitle"> Provide limited access to data in Redis based on prefixes to a set of keys.  Return data in an unformated form so that other middlware can easliy access it.</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RejectDirectory-Prevent-Browsing-of-a-Set-of-Directories">
<div class="doc-title">RejectDirectory: Prevent Browsing of a Set of Directories</div>
<div class="doc-subtitle"> Limit all access to a set of directories</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RejectExtension-Reject-Requests-Based-on-File-Extension">
<div class="doc-title">RejectExtension: Reject Requests Based on File Extension</div>
<div class="doc-subtitle"> Prevent to a set of file extensions by banning them</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RejectHotlink-Reject-requests-based-on-invalid-referer-header">
<div class="doc-title">RejectHotlink: Reject requests based on invalid referer header</div>
<div class="doc-subtitle"> Prevent access to images and other files if a valid referer header is not set.</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RejectIPAddress-Ban-Certain-IP-Address">
<div class="doc-title">RejectIPAddress: Ban Certain IP Address</div>
<div class="doc-subtitle"> Prevent access to site based on IP address</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RejectPath-Reject-Requests-Based-on-the-Path">
<div class="doc-title">RejectPath: Reject Requests Based on the Path</div>
<div class="doc-subtitle"> Prevent access to a set of paths</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RejectRePath-Reject-Requests-Based-on-a-Regular-Expression-Path-Match">
<div class="doc-title">RejectRePath: Reject Requests Based on a Regular Expression Path Match</div>
<div class="doc-subtitle"> Prevent access to paths based on a regular expression pattern match</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Rewrite-Rewrite-One-Request-to-Another-Location">
<div class="doc-title">Rewrite: Rewrite One Request to Another Location</div>
<div class="doc-subtitle"> Rewrite of request URLs</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-RewriteProxy-Rewrite-Reqeust-and-Proxy-It-to-a-Different-Server">
<div class="doc-title">RewriteProxy: Rewrite Reqeust and Proxy It to a Different Server</div>
<div class="doc-subtitle"> Combined rewrite of request and proxy request to a different server</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Status-Echo-a-Request-as-JSON-Data">
<div class="doc-title">Status: Echo a Request as JSON Data</div>
<div class="doc-subtitle"> Output request in JSON format to aid in debugging middleware stack</div>
</a>
</li>
</ul>


File: ./doc/Config_02.md
