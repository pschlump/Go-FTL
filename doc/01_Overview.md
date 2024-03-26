Overview
========
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Overview"
	,	"SubSectionGroup": "Overview"
	,	"SubSectionTitle": "Overview of the Go-FTL Server - Major Features"
	,	"SubSectionTooltip": "Go-FTL is a scalable server and forward proxy designed for development and scalable deployment of web applications."
	, 	"MultiSection":2
	}
```

Go-FTL is a scalable server and forward proxy designed for development and scalable deployment of web applications.

This documentation is being served with a Go-FTL server with a single custom middleware (it's open source) that allows search engines to correctly index the documentation.
The documentation is a single page application written with jQuery and Bootstrap.   By reading this you are helping to make the authors of Go-FTL happy - you are using
the server that they wrote!  So... Give yourself a pat on the back for a good deed that you have done today!

Major Features
--------------

Features in detail are listed in a separate document.

1. Strong authentication.  The strong authentication combines SRP with AES so that the server never has the user's password - but it is verified and all of the RESTful calls are 256-bit AES encrypted.  Two Factor Authentication (2fa) is a part of this.  Clients for iOS/iPhone, Android and other mobile platforms are provided.   A full example in Angular 1.x is provided.  Other examples in Angular 2.0, jQuery and React are in the works.  This is a *full* example including Login, Logout, Password Validation via Email, Lost Password recovery, password change etc.

2. Configuration based RESTful database server (TabServer2).  This allows complete applications to be built on the front end without code changes in the Go server.   Usually 90% or more of the application can be built with some simple configuration of the TabServer2.  Specialized business logic can be added by writing stored procedures in the database and exposing these as RESTful calls with simple additional configuration.  A full e-commerce system has been built this way.  A performance tracking system was ported from T-SQL (Microsoft's database) and a PHP back end to this with conversion of all data in a single day.   The system has security backed right in.
Emails use a templating system and are fully configurable.

3. Detailed tracing. One of my ongoing frustrations with servers is that they are black boxes that either work or don't.   You have no way to debug them.   With Go-FTL the exact opposite is true.  The tracing package allows you to see:
<div>
<ul>
<li> What middleware matched a request </li>
<li> What actions where taken </li>
<li> What errors may have occurred </li>
<li> Warnings if any along the way </li>
<li> How long it took for each section of code to run </li>
<li> If you are using the built-in micro-services - then what happens in them (and this is extensible to your code also) </li>
<li> With TabServer2 - what tables in the database where accessed, the queries/updates that were built, the bind variables in the queries, the data returned and how long it took the database to do this </li>
<li> With strong authentication when/and/if the user authenticated and what happened under the covers in the authentication process </li>
<li> With the file server - how the file was resolved to a final file and what file got served </li>
<li> Tracing of what is in the cache and why a request either did or did not get satisfied by the cache </li>
</ul>
</div>

4. Name based resolution of server configuration.  This allows for multiple virtual clients on a single server.  When you read this the server is on a single machine with at least 5 other name-resolved sites - all running in Go-FTL.

5. HTTP 2.0 support.  Much faster than HTTP 1.1 and supported by most browsers.  HTTP 2.0 provides a significant boost to performance.

6. Socket.IO and Go combine to allow pushing of content to clients.  This layer on top of web-sockets provides a cross-browser/cross-platform way of full bi-directional communication with a browser.  If you try the chat example it is built using this.  The tracing uses this to push content from the server to a browser.

7. Server Farm Ready.  Instead of saving context information inside the
server it is always saved in Redis.  This includes session states.  This means that you can run Go-FTL on more than one machine with the same 
configuration and it will work.

8. Written in Go (golang) for performance, ease of modification and stability.  Realistic examples of using a Go Server to work with Angular 1.x, 2.0, React and other front end systems.

9. Logging with Logrus.  This allows your server logs to interact with a myriad of other systems.

10. A full featured file server that supports dependency analysis.   For example, if you have TypeScript (.ts) files that need to be compiled into JavaScript (.js) the server can take a request for a .js file and use the TypeScript compiler to build the .js file on the fly.  It checks to see if the file needs to be rebuilt and will only build the .js file when the .ts has changed.   The results of this can be cached in the caching layer.   The file server supports a file-system based inheritance system allowing single page applications to be fully themed.

Planned Features
----------------

1. Lots of improvements to documentation.
1. Additional examples in React, React Native and Angular 2.0.
2. Integration with Let's Encrypt using the Lego library - Automatic updates of HTTPS/TLS certificates.
3. Configuration changes on-the-fly without a server restart.  A complete server management interface for use with multiple servers in a data center.
4. Improvements to Caching
5. A more advanced password-less strong authentication system.
6. Payed hosting for instant spin up of Go-FTL.









