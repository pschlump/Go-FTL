1. GenerateXSRF add on - postfix to .html with pattern
	1. if Path matches	( index.html, app2.html etc ) -- look at HTML5Path
		1. Random generated values for CSS
		2. Save hash in Redis for later verification
		3. set as cookie to return with future requests
2. ValidateXSRF 
	1. if Path matches	- /api etc
		Pull out g_xsrf_token - from cookie -
		Look up value in Reis - if valid then ok - else 401















From: https://docs.angularjs.org/api/ng/service/$http
From: https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)_Prevention_Cheat_Sheet

Cross Site Request Forgery (XSRF) Protection
--------------------------------------------

XSRF is an attack technique by which the attacker can trick an
authenticated user into unknowingly executing actions on your
website. Angular provides a mechanism to counter XSRF. When performing
XHR requests, the $http service reads a token from a cookie (by
default, XSRF-TOKEN) and sets it as an HTTP header (X-XSRF-TOKEN).
Since only JavaScript that runs on your domain could read the cookie,
your server can be assured that the XHR came from JavaScript running
on your domain. The header will not be set for cross-domain requests.

To take advantage of this, your server needs to set a token in a
JavaScript readable session cookie called XSRF-TOKEN on the first
HTTP GET request. On subsequent XHR requests the server can verify
that the cookie matches X-XSRF-TOKEN HTTP header, and therefore be
sure that only JavaScript running on your domain could have sent
the request. The token must be unique for each user and must be
verifiable by the server (to prevent the JavaScript from making up
its own tokens). We recommend that the token is a digest of your
site's authentication cookie with a salt for added security.

The name of the headers can be specified using the xsrfHeaderName
and xsrfCookieName properties of either $httpProvider.defaults at
config-time, $http.defaults at run-time, or the per-request config
object.

In order to prevent collisions in environments where multiple Angular
apps share the same domain or subdomain, we recommend that each
application uses unique cookie name.


1. Set Cookie on 1st get that matches a Root index.html file
	- Save hash of Cookie value -> Redis with timeout
	- Have timout specified in seconds
	- Send back "salt" as a variable or cookie in request
	- Provide a facility to get XSRF information ( app, userId, cookie, salt )

2. Validate Request has X-XSRF-TOKEN header on request.
	- 1. JS on Client takes sha1("app":"userId":"cookie":"salt") -> TOKEN to return in X-XSRF-TOKEN
	- 2. Verify this header on request of .html, .js etc.
	- 3. A verify pattern of what to check like `/api/.*` must have it
	- 4. Validate Origin/Referer
	- 5. Set Header on all requests

SetXsrfToken
	Path: /index.html	
	Timeout: 24hrs
	TimeExtendedIfUsed: true
	IdentifyingMarker: UserId, UserName, DeviceID	-- Will be set as a "cookie": XsrfID
	ApplicationName: "string"						-- Will be set as a JS global variable: g_xsrf_app_name
	UseSalt: true									-- Will be set as a JS global varialbe: g_xsrf_salt 
	CoookieName:									-- Cookie Name - Can be set to $random$ for a random name -, g_cookie_name will be set then...
	-- Calculation will set --						-- g_xsrf_token

ValidateXsrfToken
	Path: /api/
	Looks up X-XSRF-TOKEN in redis to see if exits
	Takes value and compares that to ???

0. As a cookie on DeviceID etc.
1. As fetched CSS in file									## + cached forever ## - wait for it to arrive ##
2. As in-line style											## -/- caching is same as index.html file ## + no wait ## ## - requries modification of index.html file/template ##
3. As in-line CSS / with hidden <span> 						## -/- caching is same as index.html file ## + no wait ## ## - requries modification of index.html file/template ##
4. As innerHtml of inline text in .html file (IE 8? mode)	## -/- caching is same as index.html file ## + no wait ## ## - requries modification of index.html file/template ##
5. As a chunk of JS code that is loaded (similar to css)

X-XSRF-TOKEN-id1: Value -> sha1(id1:id2)

/css/baseCss2211441.css - > Generate
	.backgroundd-color	{
		color: #1f2b12;
	}
	.foregroundd-color	{
		color: #912a31;
	}
	-- colors are 32bit random strong values
	-- the class can be accessed in JS on client
	-- Caching can be set to keep this forever!
	-- 64bit unique user ID

http://stackoverflow.com/questions/20377835/how-to-get-css-class-property-in-javascript
http://www.quirksmode.org/dom/getstyles.html
http://caniuse.com/#feat=getcomputedstyle


1. Can get body of CSS - just take sha1(body|remove extra chars)
2. Can generate a quic response to an API call /baseCss22114411.css - generate CSS with headers
3. Can add CSS inline

// * If cookies are valid, then do not set JS code at all -- just send headers to update TTL on cookies.
// This meas that the only time JS code is sent is when no cookies or invalid cookies.

// What if I already have the tokens/cookies on client side - must not generate new ones
// Should this be above or below the inMemoryCache - what are implications of both -- A: above - must be above
// What about caching? 304 response from file - no new stuff!  Is new if must no tokens!
// XyzzySession - add TTL extend on session data keys - if using sessions in Redis
