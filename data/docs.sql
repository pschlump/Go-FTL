

delete from "p_document" where "group" = 'go-ftl';


insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













The extended file server provides a number of useful capabilities that are not found in the standard static file server.

1. The ability to compile/process input files and serve output files.  For example a Markdown file (.md or .markdown) can be translated into a HTML (.html) file automatically.   This also works a little like Make in that the translation from input to output will only occur if the output is out of date.
2. A systematic way of templating and per-user templating pages.
3. Integrates properly with the in-memory/on-disk cache.
4. Extensible - allowing for processing and handling of input to output on paths on demand.  Transpile .ts into .js when the .js file is requested.
5. Extended logging
6. A tool (in the ../../tools/PreBuild directory) that can process the log file(s) and pre-build all generated files.
7. Integration with the tracer tool to report on how a path gets processed into a final file.
8. Templated directory browse.
9. Report a sample of available files to the log - this is tremendously useful for verifying that you have the correct directory and permissions.


', 'FileServer-Extended-File-Server-100000.html'
	, 'FileServer: Extended File Server' , 'Use this  as a tool when testing your web application.  Slows it way down', '/doc-FileServer-Extended-File-Server', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












``` warning-box
	<h2> Genral Rant </h2>
	I see in the news that 642 million passwords have been compromized.  This is because the passwords	
	were on the server.  If you don''t have the passwords then you can''t compromize them.  With SRP
	you have a large verification number (2k in size) stored on the server.  For it to be compromized
	requres factoring of a <b>BIG</b> number.   The number is big enough that it is basically imposible
	to factor.  This technology has been around for over a decade.  That means that breaches that
	leek passwords are unaceptable.  The server never should have had the password to start off with. 
	<br> &nbsp; <br>
```

This middleware implements strong authentication based on a key exchange using Secure Remote Password
[SRP](http://srp.stanford.edu/) and once the keys are exchanged every request is
then encrypted using the Advanced Encryption Standard
[AES-256](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard)
256 bit encryption.  Each request is signed and validated using 64 bit Counter with CBC Mac known as 
[CCM](https://en.wikipedia.org/wiki/CCM_mode).   
As a part of the login authentication a two factor authentication is required.

1. A series of standard, and tested, libraries are used to perform all of this.
2. The users password is never transmitted to the server.  
3. Every request that is protected requires encryption with the unique key that is generated on a per-login basis.  
4. Examples are provide in AngularJS 1.3, React and React Native, and soon in AngularJS 2.0.
5. This has been designed and tested to work across a server farm.
6. This can work on top of TLS (HTTPS requests) to provide additional security.
7. With React Native this has been tested on iOS (iPhone) and on Android.
8. Simple clients for iOS and Android for the two factor authentication (2fa) are provided in source.

A full cycle of user authentication is implemented. 

1. Register with email confirmation.
2. Login.
3. Logout.
3. Automatic Logout based on timeout.
4. Password Change.
5. Forgotten Password Recover.
6. Stay Logged In.
7. Administrative password change.
8. Account expiration.
9. Per-Account privileges.
9. Administrative API for setting and updating account attributes.

This integrates directly with the TabServer2 middleware for interaction with a relational or non-relational database.

There is a full section of references at the bottom of this document.

Examples (cookbooks) for AngularJS 1.4+, Angular JS 2.x, React and jQuery
are included in the source code.  

What is SRP?
------------

This is taken from
[http://srp.standord.edu](http://srp.standord.edu)

"SRP is a secure password-based authentication and key-exchange
protocol. It solves the problem of authenticating clients to servers
securely, in cases where the user of the client software must
memorize a small secret (like a password) and carries no other
secret information, and where the server carries a verifier for
each user, which allows it to authenticate the client but which,
if compromised, would not allow the attacker to impersonate the
client. In addition, SRP exchanges a cryptographically-strong secret
as a byproduct of successful authentication, which enables the two
parties to communicate securely.  Many password authentication
solutions claim to solve this exact problem, and new ones are
constantly being proposed. Although one can claim security by
devising a protocol that avoids sending the plaintext password
unencrypted, it is much more difficult to devise a protocol that
remains secure when:

1. Attackers have complete knowledge of the protocol.
2. Attackers have access to a large dictionary of commonly used passwords.
3. Attackers can eavesdrop on all communications between client and server.
4. Attackers can intercept, modify, and forge arbitrary messages between client and server.
5. A mutually trusted third party is not available."

"The idea behind SRP first appeared on USENET in late 1996, and
subsequent discussion led to refined proposals in 1997 to address
these security properties. This lead to the development of one of
the variants of the protocol still in use today, known as SRP-3,
which was published in 1998 after several rounds of discussion and
refinement on cryptography-related newsgroups and mailing lists,
and has withstood considerable public analysis and scrutiny since
then. The technology evolved into a newer variant known as SRP-6,
which maintains the security of SRP-3 but has refinements that make
it more flexible and easier to incorporate into existing systems.
Technical details of the actual protocol design are available from
this site."

"SRP is available to commercial and non-commercial users under a
royalty-free license. The Internet played a significant role in
SRP''s early development; without it, SRP would not have received
anywhere near the amount of analysis and feedback that it has gotten
since it was first proposed and refined. It is thus fitting that
the Internet at large can benefit from the fruits of this endeavor.
Since SRP is specifically designed to work around existing patents
in the area, it gives everybody access to strong, unencumbered
password authentication technology that can be put to a wide variety
of uses."

"The SRP distribution is available under Open Source-friendly
licensing terms (for the net.savvy reader, it''s a "BSD-style"
license). More information about the SRP project is available at
this site, and a reference implementation, which includes versions
of Telnet and FTP that incorporate SRP support, can be downloaded
as well. The links page has pointers to a wide range of projects
and products, both commercial and non-commercial, that use SRP, as
well as related work and papers that cover strong password
authentication."

This middleware implements SRP version 6a.



Configuration
-------------

For the paths that you want to protect and require authentication.

``` JSON
	{ "SrpAesAuth": {
		"Paths": [ "/api/", "/private/" ],
		More...
	} },
```

There are extensive configuration options - most of which can be ignored - use the default values.  These are:


Name                      | Type   | Default              | Description
|---                      | ---    | ---                  | ---
`UserNameForRegister`     | bool   | false                | By default the email address is the username.   If this is set to true then a different value can be used for the Username.  Usernames must be 7 characters starting with an alphabetic.  
`SendStatusOnerror`       | bool   | false                | By default a 200 status and a JSON message is returned. If true then a non-200 status code (4xx) will be returned on errors.
`FailedLoginThreshold`    | int    | 10                   | Number of failed logins before a 5 minute delay is inserted between login attempts.  (Delay should be configurable: TODO_4001)
`TwoFactorRequired`       | string | y                    | If "y", then two factor authentication (2fa) will be required, else just username/password.
`Bits`                    | int    | 2048                 | Length of the encryption prime numbers.  Legitimate values are 2048, 3072, 4096, 6144 and 8192.
`SendEmail`               | bool   | true                 | On registration send email for account confirmation.
`NewUserPrivs`            | string | user                 | The name of the role that a new user will be by default created with.  "DeviceId" is reserved for internal use.
`NGData`                  | hash   | rfc5054 2k values    | Allows setting of "N", "bits" and "g" values for SRP.
`AdminPassword`           | string | n/a                  | If the server is compiled in test mode, then the is the password for simulating email confirm on accounts.
`EmailRelayIP`            | string | `n/a`                | IP address of an [email relay](https://github.com/pschlump/email-relay) to send templated email to a person.
`EmailAuthToken`          | string | `n/a`                | Authorization token for an [email relay](https://github.com/pschlump/email-relay) to send templated email to a person.
`SupportEmailTo`          | string | `pschlump@gmail.com` | Address to send support emails to.
`StayLoggedInExpire`      | int    | 86400 seconds        | Amount of time that an account with Stay Logged In is retained.  Default is 1 day.
`PwRecoverTemplate1`      | string | `{{.HTTPS}}{{.HOST}}/unable-to-pwrecov1.html` | Template used to display message to user when unable to reset password due to timeout.
`PwRecoverTemplate2`      | string | `{{.HTTPS}}{{.HOST}}/unable-to-pwrecov2.html` | Template used to display message to user when unable to reset password due to other error.
`PwRecoverTemplate3`      | string | `{{.HTTPS}}{{.HOST}}/#/pwrecov2` | Template used to display message to user with link and token for password reset.
`RegTemplate1`            | string | `{{.HTTPS}}{{.HOST}}/unable-to-register1.html` | Error displayed to users.
`RegTemplate2`            | string | `{{.HTTPS}}{{.HOST}}/unable-to-register2.html` | Error displayed to users.
`RegTemplate3`            | string | `{{.HTTPS}}{{.HOST}}/unable-to-register3.html` | Error displayed to users.
`RegTemplate4`            | string | `{{.HTTPS}}{{.HOST}}/#/login` | Successful registration - user login location.
`AllowReregisterDeviceId` | bool   | false                | If true, then 2fa devices can re-register as many times as they want. - True should *only* be used in test/development mode.
`EncReqPaths`             | string | "/"                  | Paths that require encrypted login before they are accessable, "/api/private" for example.
`MatchPaths`              | string | "/"                  | Paths that will not require any login, static fiels like "/js/" for example.
`SecurityAccessLevelsName` | hash   | has a default        | See section on Security, Roles and Privileges.
`SecurityPrivilages`      | hash   | has a default        | See section on Security, Roles and Privileges.
`SecurityConfig`          | hash   | has a default        | See section on Security, Roles and Privileges.

Items that have defaults and you probably will never need to change:

Name                   | Type   | Default           | Description
|--------------------  | ----   | ---               | ---
`BackupKeyIter`        | int    | 1000              | Number of iterations for rehash of key.
`KeyIter`              | int    | 1000              | Number of iterations for rehash of key.
`BackupKeySizeBytes`   | int    | 16                | Key size.
`CookieExpireInXDays`  | int    | 1                 | Number of days before auth cookies expire.
`CookieExpireInXDays2` | int    | 2                 | Number of days before auth cocookies expire.
`SessionLife`          | int    | 86400             | Life of a login session in seconds, 1 day = 86,400 seconds.
`KeySessionLife`       | int    | 300               | Life of a key.
`CookieSessionLife`    | int    | 172800            | Life of a session in seconds, 3 days default.
`TwoFactorLife`        | int    | 360               | Life of a two factor key in seconds, 5 minutes + 1 minute grace default.
`PreEau`               | string | eau:              | Prefix for keys in Redis.
`PreKey`               | string | ses:              | Prefix for keys in Redis.
`PreAuth`              | string | aut:              | Prefix for keys in Redis.
`Pre2Factor`           | string | p2f:              | Prefix for keys in Redis.
`TestModeInject`       | string | ""                | If compiled in test mode, a set of error conditions to return when no error exists.
`PasswordSV`           | string | ""                | If compiled in test mode, turns on the generation of static simulated accounts in the development sandbox.
`SandBoxExpreTime`     | int    | 7200              | Duration in seconds that a sandbox will persist before it is deleted.



Full Example
------------

``` JSON
	{
		"working_test_for_aes_srp": { "LineNo":2,
			"listen_to":[ "http://localhost:3118" ],
			"plugins":[
				{ "DumpResponse": { "LineNo":5, "Msg":"At Top" } },
				{ "DumpRequest": { "LineNo":6, "Msg":"Request at top At Top", "Paths":"/api/status", "Final":"yes" } },
				{ "Redirect": { "LineNo":7,
					"Paths": [ "/api/ios-app",                    "/app/android-app" ],
					"To": [
						{ "To":"http://localhost:3118/ios-app.html" },
						{ "To":"http://localhost:3118/android-app.html" }
					]
				} },
				{ "JSONp": { "LineNo":14, "Paths":[ "/api/" ] } },
				{ "SrpAesAuth": { "LineNo":15,
					"Paths": "/api/" ,
					"MatchPaths": [ "/" ],
					"AllowReregisterDeviceId": true
					} },
				{ "DumpResponse": { "LineNo":19, "Msg":"After Proxy" } },
				{ "Status": { "LineNo":20, "Paths":"/api/status" } },
				{ "RedisList": { "LineNo": 21,
					"Paths":             "/api/list/user",
					"Prefix":            "srp:U:",
					"UserRoles":         [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]

				} },
				{ "file_server": { "LineNo":27, "Root":"./angular-login-example/build", "Paths":"/"  } }
			]
		}
	}
```

API
---

### Overview of process

Registration of new users is handled by `/api/srp_register` with email confirmation by sending a
conformation token to `/api/srp_email_confirm.` For a hand entered token or
the pair of `/api/confirm-registration` and `/api/pwrecov2` that allow an email, click
link with an underlying AngularJS single page application to perform the confirmation.

Login has 3 steps due to the SRP-6 implementation.  These are `/api/srp_login`, 
`/api/srp_challenge` and `/api/srp_validate`.  

Logout is simple. - It cleans up but is not required.  It is `/api/srp_logout`.

Password recovery is performed by a pair of calls.  The first one is
`/api/srp_recover_password_pt1.` - This sends the password recovery email.  The
second call is the one that will take the user to a "change password" screen.
This is `/api/srp_recover_password_pt2`.

You can get back version information on this middleware with `/api/version`.

Once the SRP-6 is complete then AES encryption is used to encrypt every 
request.  All future requests are sent to `/api/cipher` as POST or GET requests.  It then
decrypts the request and sends it on to the handler for the request.  Normally this will
only apply to RESTful requests. 

If two factor authentication (2fa) is enabled, then the one time key will be required
before the login process is completed.  The iOS(iPhone) or Android device with
the application can be used or one of the pre-arranged one-time keys can be used
that was generated (and hopefully printed out) during the registration process.
The Call to get a 2fa "key" is, `/api/get2Factor`.
The client attempting to login then takes the user-supplied key and validates
it with a call to `/api/valid2Factor`.  This completes the 2fa login process.

A logged in user can change a password with, `/api/srp_change_password`.

An administrator with "admin" privilege can change a password with `/api/admin_set_password`.
The administrator can also set other attributes of a user with, `/api/admin_set_attrs` like
changing the "disabled" state of an account or resetting the number of failed login 
attempts.  Any of the non SRP attributes of a user can be set with this.

Once logged in a user can generate a new set of printable one-time keys with a call to
`/api/genTempKeys`.  This will erase the previous set of one-time keys.

At registration the user, if 2fa is enabled, can get the unique ID for an Android or iOS
device with, `/api/getDeviceId`.   You can have more than one device associated with
an account.

If "stay logged in" is used, then a session must be marked with a browser fingerprint
and a pair of cookies.  This is a weaker form of authentication than a full login
but it is convenient.  The browser fingerprint is set with `/api/markPage`.
An example of this is in the Angular 1.x code.   When login is successful, a call to
`/api/setupStayLoggedIn` is made.   This updates and creates the necessary server configuration
for staying logged in.  The next time the page is loaded a call to `/api/resumeSession`
recovers the previous login.   The resumed login is easily identifiable on the
server side and limited to a set of APIs with the security configuration.  Normally
a full login would be required before allowing operations requiring a credit card
or changes to shipping of products.

If you use other than 2k standard random numbers (rfc5054), then you will need to
set these in your configuration and perform this call at the beginning of your client code
to get the random numbers.  `/api/srp_getNg`.  This call will  return the
SRP "N" and "g" values and 2fa on/off  flag.













## /api/cipher - Generate a new set of backup one-time-keys

This is the central point of the AesSrp authentication.

Lookup the key that was generated via SRP using the `t` value.  Decrypt the replacement request and replace the current request with
a decrypted request.

When the request returns, reverse the process and re-encrypt the
response.

The examples show how to perform the client encryption and decryption for jQuery/AJAX and for AngularJS 1.x.

Each encrypted request is signed and has a random initialization vector added to it.
This provides a very grantee that: 

1. The request is from the client that it is supposed to be from.
2. The request has not been tampered with in the middle.
3. The only recipient of the response is the interned client.
4. Nobody can listen in to the middle of a set of messages and get anything.
4. Nobody can tamper with a message.
4. Nobody can replay any messages.

SRP also provides perfect future security since each key from each login is generated at login time.

#### Weeknesses

1. Cookies and headers are not encrypted.  
2. This only encrypts the RESTful requests and responses.

#### File: aessrp_ext.go Line No: 3089, 528, 3566, 4118


``` green-bar
	func respHandlerCipher(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input

Supports GET and POST.  Post is usually better for this.

Name | Description
|--- | --- 
`t` | Session identifier.
`data` | The encrypted data.


#### Output

On Success an encrypted response in JSON.  On an error a non-encrypted
JSON response with "status" == "error".   

Note: You only get errors back from this when the error is at the encryption/decryption level.
Any errors from a lower level are sent back as an encrypted response.


#### Possible Output Errors

See above. - Output errors are handled with redirect to user supplied pages.

### Example

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/cipher?
		t=UUID&
		data=EncryptedData&
		_ran_=0.12332321123




## /api/srp_register - Start Registration Process 

Unencrypted - This is the call to register a new, unverified, user.   This is also the call that is used to
register a new 2fa device and associate it with an existing verified user.

#### File: aessrp_ext.go Line No: 1467, 512, 547


	func respHandlerSRPRegister(www http.ResponseWriter, req *http.Request) {
		...
	}


#### Input

Supports GET and POST

Name         | Description
|---        | --- 
`email`      | Email address of user. - Used as the username - or `DeviceId` - one of the 2 is required.
`salt`       | The generated random salt to use with this user.
`v`          | The password verification number generated as a part of the SRP process.
`DeviceId`   | Supplied when registering a 2fa verification device instead of email.  Additional data will be returned to the device.
`UserName`   | Configuration option. - Use this username for login instead of email.  Set `UserNameForRegister` to true to use.
`RealName`   | Optional. - If supplied then associated with user data.
`FirstName`  | Optional. - If supplied then associated with user data.
`MidName`    | Optional. - If supplied then associated with user data.
`LastName`   | Optional. - If supplied then associated with user data.

#### Output

For normal user registration with `Email` as a parameter A JSON hash with:

Name | Description
|--- | --- 
`status` | "success" | "error"
`msg` | Only supplied if an error occurs.

The data is saved in Redis for the user.

A confirmation email is sent to the user.  This can be templated.  Please see the section on email templates.

#### Possible Output Errors

##### Account Already Exists

If SendStatusOnerror is true, then Status = 400, Internal server error.
Else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9001","msg":"Account is already registered with this email but has not been confirmed.  A new email confirmation has been sent.",
	"LineFile":"File: /.../aessrp_ext.go LineNo:1346","URI":"/api/srp_register?email=t1@example.com&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...number...&_ran_=2323232323232323232"}
```

or

``` JSON
	{"status":"error","code":"9003","msg":"Account is already registered with this email.",
	"LineFile":"File: /.../aessrp_ext.go LineNo:1346","URI":"/api/srp_register?email=t1@example.com&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...number...&_ran_=2323232323232323232"}
```

##### Invalid Salt

If SendStatusOnerror is true, then Status = 400, Internal server error.
Else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9004","msg":"Invalid Salt.",
	"LineFile":"File: /.../aessrp_ext.go LineNo:1283","URI":"/api/srp_register?email=t3@example.com&salt=1&v=51...big...number...&_ran_=2323232323232323232"}
```

##### Invalid ''v'' Value

If SendStatusOnerror is true then,  Status = 400, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9005","msg":"Invalid ''v'' verifier value.","LineFile":"File: /.../aessrp_ext.go LineNo:1288",
	"URI":"/api/srp_register?email=t1@example.com&salt=42ce852b31aa2beb5e2f89872f944d4b&validate=51...big...nnumber...&_ran_=2323232323232323232"}
```

##### Invalid Username (When configured to use username.)

If SendStatusOnerror is true then,  Status = 400, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9103","msg":"UserName can not be a UUID.","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```

##### When a device is registered no Username or Email may be supplied.

If SendStatusOnerror is true then,  Status = 400, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9100","msg": "DeviceId registrations can not have UserName or Email attributes.","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```

##### When a device is registered no Invalid DeviceId.

If SendStatusOnerror is true then,  Status = 400, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9105","msg": "DeviceId are all 8 digits.","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```

or

``` JSON
	{"status":"error","code":"9106","msg": "DeviceId are all digits.","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```


##### Username Invalid.

If SendStatusOnerror is true then,  Status = 400, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9101","msg": "UserName must be at least 7 characters long.","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```

or

``` JSON
	{"status":"error","code":"9102","msg": "UserName must start with a letter, a-z or A-Z.","LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```

or

``` JSON
	{"status":"error","code":"9104","msg": "UserName can not be used for registration." ,"LineFile":"File: /.../aessrp_ext.go LineNo:1248",
	"URI":"/api/srp_register?email=t1@example.com&UserName=fredFred&salt=42ce852b31aa2beb5e2f89872f944d4b&v=51...big...nnumber...&_ran_=2323232323232323232"}
```




#### Example.

Note: a `_ran_` catch buster has been appended.  Any additional parameters will be ignored.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/srp_register?	
		salt=72837283728327328728378273827382738237&
		email=myEmailAddr@example.com&
		verifier=343434343434434343434343434343434343433434f23232&
		_ran_=0.12232321123

Currently you will receive a confirmation email from my personal email address - pschlump@gmail.com.
Soon this will be configurable to other addresses.

#### Test.

Tested On: Mon Mar 21 17:25:28 MDT 2016.

Test: aes_1_test.go






## /api/srp_login - Start Login Process.

Login has 3 or 5 parts. `/api/srp_login`, `/api/srp_challenge`, `/api/srp_validate`.  This is the first
part.  If you are using two-factor-authentication (2fa), then the 4th part is `/api/get2FactorFromDeviceId`
and it usually is from a different device than the original login process.  This gets the one-time key
that is then used in the 5th part, `/api/valid2Factor`.

This is Step 1 of 3 for the login/srp process.

#### File: aessrp_ext.go Line No: 2477, 124, 514, 548


``` green-bar
	func respHandlerSRPLogin(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST - POST is preferred.

Name | Description
|--- | --- 
`Email` | Use as the Username if Username is not supplied.
`Username` | The Username to start login process with.

#### Output.

On Success A JSON hash with:

Name | Description
|---    | --- 
`status` | "success"
`salt`   | Numeric salt value from registration.
`t`      | SRP ''t'' value - see references on SRP.		See Reference 1, 2.
`r`      | SRP ''r'' value - see references on SRP.		See Reference 1, 2.
`bits`   | By default 2048 - but this can be configured to be 3K, 4k, 6k or 8k.
`B`      | SRP ''B'' value - see references on SRP.		See Refernece 1, 2.
`f2`     | "y" If two factor authentication will be required (default), "n" otherwise.

#### Possible Output Errors.

##### Inability to generate SRP values ''t'' or ''r''. 

If SendStatusOnerror is true then,  Status = 500, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9007","msg":"Internal error - unable to generate ''t'' key.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

or

``` JSON
	{"status":"error","code":"9008","msg":"Internal error - unable to generate ''r'' key.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

##### Invalid email address.

If SendStatusOnerror is true, then Status = 400, Bad request
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"25","msg":"Invalid email address.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```


##### Account has been disabled.

If SendStatusOnerror is true, then Status = 401, Unauthorized
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9013","msg":"The account has been disabled.  Please contact customer support (call them).","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

or (Only the code and line numbers changed.)

``` JSON
	{"status":"error","code":"9015","msg":"The account has been disabled.  Please contact customer support (call them).","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

Code 9015 is caused by exceeding the limit on the number of invalid logins.  The default is 10 and is set by, "FailedLoginThreshold" configuration parameter.

Code 9013 is caused by a disabled account in the per-user account information.  If the `disabled` item in the per-user information is other than "y", then you get this.

Both of these can be set/changed by an administrator in the user management interface.  The 10 failed attempts, code 9015, will automatically reset after a five
minute delay.  The delay will be configurable in the next version (item TODO_4001). 

##### Account not confirmed.

If SendStatusOnerror is true, then  Status = 401, Unauthorized
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9003","msg":"The account has not been confirmed.  Please confirm or register again and get a new confirmation email.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

This indicates that the person has not replied to the email confirmation.

### Example.

Note: a `_ran_` catch buster has been appended.  Any additional parameters will be ignored.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/srp_login?	
		Email=myEmailAddr@example.com&
		_ran_=0.12232321123








## /api/srp_challenge - Start Login Process.

Step 2 of 3 for the login/srp process.


#### File: aessrp_ext.go Line No: 2607, 131, 515, 549


``` green-bar
	func respHandlerSRPChallenge(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST - POST is preferred.

Name | Description
|--- | --- 
`A` | SRP ''A'' value
`r` | SRP ''r'' value

#### Output.

On Success A JSON hash with:
	// io.WriteString(www, fmt.Sprintf(`{"status":"success","B":"%s","Bits":%d,"M1":%q}`, sss.XB_s, hdlr.Bits, sss.XM1_s))

Name     | Description
|---    | --- 
`status` | "success"
`B`      | Numeric salt value from calculation.
`Bits`   | Bits in computation.  Same as initialization bits.
`M1`     | SRP ''M1'' value - See references on SRP.		See Reference 1, 2.

#### Possible Output Errors.

##### Temporary data lost.

If SendStatusOnerror is true, then  Status = 500, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9008","msg":"Temporary data lost.  Please try login again.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

or

``` JSON
	{"status":"error","code":"9009","msg":"Temporary data corrupted.  Please try login again.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```


### Example.

Please look in the AngularJS 1.x, React, React Native or AngularJS 2.x code for an example.









## /api/srp_validate - Start Login Process.

Step 3 of 3 for the login/srp process.

If this step passes and two-factor-auth (2fa) is enabled, then the user is set to a "pending" state waiting
for the 2fa one time key to be validated.   If 2fa is not enabled then the user is logged in.

Some return data is specific to a first login.

#### File: aessrp_ext.go Line No: 2682, 132, 141, 152, 516, 550, 933


``` green-bar
	func respHandlerSRPValidate(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST - POST is preferred.

Name           | Description
|---         | ---------------------------------------------------------------------------------------------------------
`M1`           | SRP ''M1'' value as calculated by the client. - This will be validated by the server now.
`r`            | SRP ''r'' value.
`stayLoggedIn` | True False flag for staying logged in.  True/Flase, 1/0, y/n etc.   Requries call to /api/setFingerprint to work.

#### Output.

On Success A JSON hash with:

Name             | Description
|---            | ---------------------------------------------------------------------------------
`status`         | "Success"
`M2`             | SRP M2 value. - Calculated.
`FirstLogin`     | Flag if this is the first login that has happed after registration.
`MoreBackupKeys` | True/False flag. - If the user should get more one-time backup keys.  True when less than 5 keys left.  
`UserRole`       | One of the security roles like, ''user'', ''admin'' etc.  This is the role for this user.
`DeviceId`       | If first login and two-factor-auth is enabled, then the DeviceId for 2fa.
`BackupKeys`     | One time backup keys to be printed. - If 2fa and first login.
`OwnerEmail`     | Email of the user.
`Attrib`		 | Other user attributes that are configured to be returned if any (optional field).

On a successful login the number of failed login attempts is zeroed and the last login time is set to the current time stamp.

#### Possible Output Errors.

##### Failed to login. - Invalid username/password.

If SendStatusOnerror is true, then  Status = 401, Unauthorized
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9100","msg":"Failed to login.  Incorrect username/email or password.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

Upon failure to login the number of login attempts that have failed is incremented.  Also the last login time is set.
The first step in this process will check these values and if too many failed login attempts are made it rebuffs any
new attempts for a pre-determined (5 minute) time period.

##### Temporary data lost.

If SendStatusOnerror is true, then  Status = 500, Internal server error
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9010","msg":"Temporary data lost.  Please try login again.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

or

``` JSON
	{"status":"error","code":"9011","msg":"Temporary data lost.  Please try login again.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```


### Example.

Please look in the AngularJS 1.x, React, React Native or AngularJS 2.x code for an example.











## /api/srp_getNg - Initialize - Get SRP `n`, `G`, `bits`, security configuration and two factor enabled.

At the very beginning of the process you will need to initialize the security model and the SRP
authentication.  This provides all of the initialization data in 2 formats.  Either as a 
JavaScript file that declares some globals or as a JSON response.

JavaScript is more convenient in a .html file.  For example:

``` HTML
	<script type="text/javascript" src="http://www.example.com/api/srp_getNg"></script>
```

JSON is easier inside a Go or Swift program.

#### File: aessrp_ext.go Line No: 1408, 517, 551.


``` green-bar
	func respHandlerSRPGetNg(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST - POST is preferred.

Name | Description
|--- | --- 
`fmt` | use `js` for JavaScript or `json` for JSON formated output.

#### Output.

On Success A JSON hash with:

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "success"
`bits`              | 2048 is default, may be, 3072, 4096, 6144, 8192.
`g`                 | SRP ''g'' value - See references on SRP.		See Reference 1, 2.
`N`                 | SRP ''N'' value - See references on SRP.		See Reference 1, 2.
`TwoFactorRequired` | Enabled by default, ''y''.  ''n'' if not enabled.
`SecurityData`      | The roles and privileges data model. 

#### Possible Output Errors.

None - Service could be down.

### Example.

Note: a `_ran_` catch buster has been appended.  Any additional parameters will be ignored.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/srp_getNg?	
		fmt=JSON&
		_ran_=0.12332321123

Output:

``` JSON
	{
		"status": "success",
		"TwoFactorRequired": "y",
		"bits": 2048,
		"g": "2",
		"N": "AC6BDB41324A...default...see...rfc5054...",
		"SecurityData": {
			"AccessLevels": {
				"admin": [ "admin" ],
				"anon": [ "public" ],
				"public": [ "*" ],
				"root": [ "root", "admin", "user", "public" ],
				"user": [ "user", "admin" ]
			},
			"MayAccessApi": {
				"DeviceId": [
					"/api/srp_register",
					"/api/srp_login",
					"/api/srp_challenge",
					"/api/srp_validate",
					"/api/srp_getNg",
					"/api/send_support_message",
					"/api/version",
					"/api/srp_logout",
					"/api/cipher",
					"/api/get2FactorFromDeviceId"
				],
				"admin":  [ "*" ],
				"anon":   [ "*" ],
				"public": [ "*" ],
				"root":   [ "*" ],
				"user":   [ "*" ]
			},
			"Privilages": {
				"admin": [ "MayChangeOtherPassword", "MayCreateAdminAccounts", "MayChangeOtherAttributes" ]
			},
			"Roles": [ "public", "user", "admin", "root" ]
		}
	}
```

The `bits`, `N` and `g` values can be set as a part of the AesSrp configuration with `NGData` option.  You do not need to use
the values from RFC5054.  You can use a key bigger than 2k.


## /api/version - Return a version string for AesSrp.

This just returns the version of the code that is implemented as JSON data.

There is an encrypted call to this.  It is useful for checking to see if you are logged in.  It is `/api/enc_version`.

#### File: aessrp_ext.go Line No: 754, 526, 560, 582.


``` green-bar
	func respHandlerVersion(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.

#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"
`msg`               | SRP 6a, AES-256 RESTful encryption (part of Go-FTL 0.5.9)
`version`           | 1.0.1
`BuildData`         | Fri Apr  8 08:03:32 MDT 2016

#### Possible Output Errors.

None - Service could be down.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/version

Output:

``` JSON
	{"status":"success","msg":"SRP 6a, AES-256 RESTful encryption (part of Go-FTL 0.5.9)","version":"1.0.1","BuildDate":"Thu Mar 31 14:55:27 MDT 2016"}
```


## /api/enc_version - Return a version string for AesSrp.

This just returns the version of the code that is implemented as JSON data.

There is a non-encrypted call to this.    It is `/api/version`.

#### File: aessrp_ext.go Line No: 

#### Input.

Supports GET and POST.

#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"
`msg`               | SRP 6a, AES-256 RESTful encryption (part of Go-FTL 0.5.9)
`version`           | 1.0.1
`BuildData`         | Fri Apr  8 08:03:32 MDT 2016

#### Possible Output Errors.

None - Service could be down.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/version

Output:

``` JSON
	{"status":"success","msg":"SRP 6a, AES-256 RESTful encryption (part of Go-FTL 0.5.9)","version":"1.0.1","BuildDate":"Thu Mar 31 14:55:27 MDT 2016"}
```



## /api/srp_logout - End a login session.

This is an optional call to indicate that the user has intentionally logged out.

#### File: aessrp_ext.go Line No: 2332, 527, 561, 2331.


``` green-bar
	func respHandlerSRPLogout(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.

#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.

None - Service could be down.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/srp_logout

Output:

``` JSON
	{"status":"success"}
```



## /api/srp_recover_password_pt1 - Start password recovery process.

This call will send a templated  email to the user to start password recovery.  The email contains
a link to `/api/pwrecov2?auth_token=UUID` that will redirect to a configurable destination to show a password recovery
page.   Also the email should display a *token* that can be entered to complete the password
recovery process.  The email can be sent multiple times.  The *token* lasts for 24 hours.
`/api/srp_recover_password_pt2?...` is the second part of the password recovery.  It is what needs
to be called with a new `salt` and `v` to reset the user''s password to a new value.

Note: `/api/srp_recover_password_pt2` also resets all backup one time keys, and creates a new
`DeviceId` for two factor authentication.   This means that password recovery invalidates any
existing devices.

The email template is `password-reset.tmpl`.  It is passed the users `email` address and `email_auth_token`.

This uses the `PwResetKey` (`pwr:`) in Redis.

### Overview of process.

1. Recovered page is marked with an acceptable token "page." - Page token is fetched/generated. 
 <ol>
 <li> Call /api/getPageToken to get cookie that lasts for 24 hrs. The cookie is used in /api/srp_recover_password_pt1.
 </ol>
2. Generate token. 
3. Mark account with token and timestamp of recovery start.
4. On request to /api/srp_recover-password_pt1 - Save the "page" cookie, associate the cookie with the email address.
5. Save ''prw:''||token with "email" and a 24hr timeout to d.b.
6. Send Email with `auth_token`.
 <ol>
 <li>If user clicks on link,
  <ol>
  <li> Link directs to /api/pwrecov2 that will do a server-temporary (307) redirect to page to enter Password.
  <li> A "link" cookey is created in the response as hash(Salt:Email:Token) - This will get checked later.
  <li> Creates a data-pass "cookie" that passes the email/username to the client for display in form. -- Add in other user values (Full Name etc.)
  <li> Client displays form - and deletes data-pass cookie cookie.
  <li> User enters new password + token. - Hits submit.
  </ol>
 <li> If user re-enters token into form,  - the form already has email from the request to reset password. -
  <ol>
   <li> User enters new password + token. - Hits submit.
  </ol>
 <li>Call is made with Token/Usernmae=email/Salt/V + cookies to /api/srp_recover_password_pt2. - This will allow setting of password for user.
  <ol>
  <li> Email is sent (optionally PwSendEmailOnRecoverPw) to user to notify them.
  <li> Verify either the "page" token or the more secure "link" token.
  <li> New salt/v for verifying paswords.
  <li> Resets invalid login count and last login dates.
  <li> Create new DeviceId, one-timekeys - and returns these to user.
  </ol>
 </ol>
7. Happy user logged in. - Enters new DeviceId into 2fa device.  Probably clicks *get one time key* key button.
 <ol>
 <li> On next click of *get one time key* button - When device is connected to network.
  <ol>
  <li> Will know that device is not registed.
  <li> Will register and login. 
  <li> Will get new backup offline one time keys.  
  <li> Will get a one-time-key for the user to login.
  </ol>
 </ol>


#### File: aessrp_ext.go Line No: 1924, 518, 552.


``` green-bar
	func respHandlerRecoverPasswordPt1(www http.ResponseWriter, req *http.Request) {
		...
	}


	func respHandlerRecoverPasswordPt1(www http.ResponseWriter, req *http.Request) {
		...
	}
```

#### Configuration Options.

Name                     | Description
|---                    | ------------------------------------------------------------------------------------
`PwResetKey`             | Prefix used in Redis for temporary storage of auth_token.  Temporary is determined by value in seconds of `PwExpireIn` and defaults to one day (86400 seconds).
`PwExpireIn`             | Temporary key is saved for this amount of time in seconds. Defaults to one day (86400 seconds).
`PwSendEmailOnRecoverPw` | If ''y'' then send email to user when attempt is made to reset password.  Email is ''pw-reset-attempt.tmpl.''
`PwRecoverTemplate1`     | Used in /api/pwrecov2.  Default: `{{.HTTPS}}{{.HOST}}/unable-to-pwrecov1.html`   Error message to user for invalid email.
`PwRecoverTemplate2`     | Used in /api/pwrecov2.  Default: `{{.HTTPS}}{{.HOST}}/unable-to-pwrecov2.html`   Error message to user for invalid email.
`PwRecoverTemplate3`     | Used in /api/pwrecov2.  Default: `{{.HTTPS}}{{.HOST}}/#/pwrecov2`  Success redirect link.  Client page to display to user collect new password.


#### Input.

Supports GET and POST.  Requires a valid email address and valid email configuration.

#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.

None - Service could be down.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/srp_recover_password_pt1

Output:

``` JSON
	{"status":"success"}
```

### TODO / In Progress Now.

1. This needs to be passed other selected user attributes so the email can put in the persons real name.  See: /api/set_user_attrs and /api/get_user_attrs.
1. Need to create token to map to user email and other info. - Instead of "cookie" pass redirect process.
1. Add ability to send email to user that an attempt was made to reset password on account.
1. muxEnc.HandleFunc("/api/confirm-registration", respHandlerConfirmRegistration).Method("GET")            // Redirect-Link: To 1st Login /#/login, confirmed <- ''y''  -- Needs to have templates.




## /api/srp_recover_password_pt2 - Start password recovery process.

This resets the users password and authentication inform ant. 
/api/srp_recover_password_pt1 needs to be called first to create an `auth_token`.

The *auth_token* lasts for 24 hours.
/api/srp_recover_password_pt2 is the second part of the password recovery.  It is what needs
to be called with a new `salt` and `v` to reset the users password to a new value.

Note: /api/srp_recover_password_pt2 also resets all backup one time keys, and creates a new
`DeviceId` for two factor authentication.   This means that password recovery invalidates any
existing devices.

#### File: aessrp_ext.go Line No: 1980, 121, 519, 553.


	func respHandlerRecoverPasswordPt2(www http.ResponseWriter, req *http.Request) {
		...
	}


#### Configuration Options.

See `/api/srp_recover_password_pt1.`

#### Input.

Supports GET and POST.  Requires a valid email address and valid email configuration.

Name               | Description
|---              | --- 
`salt`             | Salt for creating new password. - If empty then this will set a cookie and return.
`v`                | Vector for creating new password. - If empty then this will set a cookie and return.
`email_auth_token` | Token from email that allows for password recovery.


#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"
`DeviceId`          | A new DeviceId to be entered into the 2fa device.
`BackupKeys`        | A set of one-time backup keys that the user should print out and save.

#### Possible Output Errors.

##### Invalid token.

If SendStatusOnerror is true then,  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"8001","msg":"Invalid token.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid token or email.

If SendStatusOnerror is true, then  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"8002","msg":"Invalid token or email.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid token.

If SendStatusOnerror is true, then  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"8003","msg":"Invalid token.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid salt.

If SendStatusOnerror is true, then  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"8004","msg":"Invalid salt.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid ''v'' verifier value.

If SendStatusOnerror is true, then  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"8005","msg":"Invalid ''v'' verifier value.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```



### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/srp_recover_password_pt1?via=link&
		email_auth_token=UUID&
		email=pschlump@example.com

Output:

``` JSON
	{"status":"success"}
```



## /api/valid2Factor - Validate a two factor (2fa) one time key.

If two factor authentication is enabled then a 2fa key will be required to complete
the login process.  This key is generated/retrieved using one of the 2fa client
programs.  The key is then passed to this `/api` call to validate it.   The
account that is being logged into is in a "pending" state until validation
occurs and the set of `/api` calls that it will allow is limited to:

	"/api/1x1.gif"
	"/api/cipher"
	"/api/confirm-registration"
	"/api/enc_version"
	"/api/get2FactorFromDeviceId"
	"/api/getPageToken"
	"/api/pwrecov2"
	"/api/send_support_message"
	"/api/srp_challenge"
	"/api/srp_email_confirm"
	"/api/srp_getNg"
	"/api/srp_login"
	"/api/srp_logout"
	"/api/srp_recover_password_pt1"
	"/api/srp_recover_password_pt2"
	"/api/srp_register"
	"/api/srp_simulate_email_confirm"
	"/api/srp_validate"
	"/api/valid2Factor"
	"/api/version"

#### File: aessrp_ext.go Line No: 969, 570, 2846.


``` green-bar
	func respHandlerValid2Factor(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  Requires prior Aes/Srp login to complete successfully.  This
call can only be accessed via an encrypted call.

#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.


##### Invalid one time key.

If SendStatusOnerror is true, then  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"68","msg":"Invalid one time key","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/valid2Factor?OneTime=8121412

Output:

``` JSON
	{"status":"success"}
```




## /api/get2FactorFromDeviceId - Use DeviceId to fetch a two factor (2fa) one time key.

If two factor authentication is enabled then a 2fa key will be required to complete
the login process.  This call is used to retrieve a key from the server using the
DeviceId.  The key is then passed to this `/api/valid2Factor` call to validate it.  

The account that is being logged into is in a "pending" state until validation
occurs and the set of `/api` calls that it will allow is limited to:

	"/api/1x1.gif"
	"/api/cipher"
	"/api/confirm-registration"
	"/api/enc_version"
	"/api/get2FactorFromDeviceId"
	"/api/getPageToken"
	"/api/pwrecov2"
	"/api/send_support_message"
	"/api/srp_challenge"
	"/api/srp_email_confirm"
	"/api/srp_getNg"
	"/api/srp_login"
	"/api/srp_logout"
	"/api/srp_recover_password_pt1"
	"/api/srp_recover_password_pt2"
	"/api/srp_register"
	"/api/srp_simulate_email_confirm"
	"/api/srp_validate"
	"/api/get2FactorFromDeviceId"
	"/api/version"

#### File: aessrp_ext.go Line No: 859, 533, 577, 876, 2845.


``` green-bar
	func respHandlerGet2FactorFromDeviceId(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  Requires prior Aes/Srp login to complete successfully.  This
call can only be accessed via an encrypted call.

#### Output.

On Success A JSON hash with (these values were correct the day this document was written):

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.


##### Invalid one time key.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"68","msg":"Invalid one time key","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid one time key.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"68","msg":"Invalid one time key","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### fmt.sprintf("unable to find account with email ''%s''.", email)).

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"63","msg":"fmt.Sprintf("Unable to find account with email ''%s''.", email))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### fmt.sprintf("secuirty error: not a legitimae end point for a %s type account\n\n", mdata["acct_type"])).

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"463","msg":"fmt.Sprintf("Secuirty Error: not a legitimae end point for a %s type account\n\n", mdata["acct_type"]))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Unable to retrieve a key based on this deviceid.

If SendStatusOnerror is true then,  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"84","msg":"`Unable to retrieve a key based on this DeviceId`)","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/get2FactorFromDeviceId?DeviceId=8121412

Output:

``` JSON
	{"status":"success","version":1,"OneTimeKey":"82112121"}
```



## /api/send_support_message - Use DeviceId to fetch a two factor (2fa) one time key.

This is used to send an email to the configured email support address.  The email
support address is set in the web site configuration with the `SupportEmailTo`
option.

This is basically so that people can tell me if there is a problem.

#### File: aessrp_ext.go Line No: 811, 525, 559, 809.


``` green-bar
	func respHandlerSendSupportMessage(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  Requires prior Aes/Srp login to complete successfully.  

Name            | Description
|---           | ------------------------------------------------------------------------------------
`email_fr`      | Address that the email is from.  Must be non-blank and an email address.
`subject`       | Subject of the message.
`body`          | The body.

#### Output.

On Success A JSON hash:

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.


##### Invalid input email.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"251","msg":"Invalid input email.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid input email body.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"253","msg":"Invalid input email body.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid input email body.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"253","msg":"Invalid input email body.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/send_support_message?email_fr=me@example.com&subject=not+working&body=please+help

Output:

``` JSON
	{"status":"success"}
```




## /api/pwrecov2 - Intermediary in click-link password recovery.

This is an intermediary step in the password recovery process.  When an email is sent
to a user with a recovery `auth_token` in it the user can re-enter the token in the
original form.  That bypasses this entirely.   If on the other hand they want to
click on a link, then that link needs to bring up the application with the correct
page for resetting the password.  This is the `/api` to do that.   It checks a
number of chunks of data and then redirects (307) to a client page that should
prompt the user for the new password and new password again.

The client page then calls `/api/srp_recover_password_pt2` to complete the password reset process.

This is configured with 3 go templates:

Configuration Item   | Default                                        | Description
|---                | ---------------------------------------------- | -----------------------------------------------------------------------------
`PwRecoverTemplate1` | `{{.HTTPS}}{{.HOST}}/unable-to-pwrecov1.html`  | Error when auth_token is invalid because it is not a correct token.
`PwRecoverTemplate2` | `{{.HTTPS}}{{.HOST}}/unable-to-pwrecov2.html`  | Error when auth_token is invalid because it has expired or never existed.
`PwRecoverTemplate3` | `{{.HTTPS}}{{.HOST}}/#/pwrecov2`				  | Client page to redirect to on success. - Default matches with AngularJS 1.x demo.


#### File: aessrp_ext.go Line No: 1237, 119, 164, 524, 558.


``` green-bar
	func respHandlerRecoverPw2(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  Requires prior Aes/Srp login to complete successfully.  

Name            | Description
|---           | ------------------------------------------------------------------------------------
`auth_token`    | The token from the email address.

#### Output.

On Success a 307 redirect to the password reset page.  Also a cookie is set marking this as
an acceptable source for resetting this user''s password.

On error redirects to other error appropriate pages.

#### Possible Output Errors.

See Above.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/pwrecov2?email_auth=UUID



## /api/getDeviceId - Get a new DeviceId.

Get a new DeviceId - and update the DeviceId in the users configuration.

This is normally called from a user''s client page when they want to register a new device.
The client may have used the old device to login or a backup one-time-key or they may have
already been logged in.

If you need to recover from a lost device, then the normal password recovery process will create
and return to you a new DeviceId and a new set of backup one-time-keys.

#### File: aessrp_ext.go Line No: 650, 572.


``` green-bar
	func respHandlerGetDeviceId(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

You must be successfully logged in to a "user" account with AesSrp encryption to make this call.

#### Output.

On Success A JSON hash with:

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"
`DeviceId`            | The new id.

#### Possible Output Errors.


##### Invalid input data.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"35","msg":"Invalid input data.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Failed to find user. Invalid input email.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"36","msg":"Failed to find user. Invalid input email.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### fmt.sprintf(`Unable to find account with email ''%s`, email))

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"37","msg":"fmt.Sprintf(`Unable to find account with email ''%s`, email))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/getDeviceId?	
		fmt=JSON&
		_ran_=0.12332321123

Output:

``` JSON
	{"status":"success","DeviceId":"8121223"}
```



## /api/srp_change_password - Allow a user to change a password.

This is the normal `/api` call to change the user''s password.   A new ''salt'' and a new ''v'' are generated
from the password on the client side.

You must be successfully logged in to a "user" or "admin" account with AesSrp encryption to make this call.


#### File: aessrp_ext.go Line No: 1771, 120, 569.


``` green-bar
	func respHandlerChangePassword(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.  
Supports GET and POST.

Name    | Description
|---   | --- 
`email` | The users email address.
`salt`  | The new ''salt'' value.
`v`     | The new verifier.


#### Output.

On Success A JSON hash with:

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.


##### `Invalid input data.`)

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"18","msg":"`Invalid input data`)","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### fmt.sprintf("Fatal error - Did not get passed a bufferhtml.midbuffer - at: %s\n", godebug.lf()))

If SendStatusOnerror is true, then  Status = 500, Internal Server Error,
else Status = 200 and JSON response is:

	{"status":"error","code":"5","msg":"fmt.Sprintf("Fatal Error - Did not get passed a bufferhtml.MidBuffer - AT: %s\n", godebug.LF()))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}


##### Unable to find account with specified email.  If SendStatusOnerror is true, then  Status = 400, Bad Request.  else Status = 200 and JSON response is: {"status":"error","code":"21","msg":"Unable to find account with specified email.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"} 

##### Invalid ''v'' verifier value.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"20","msg":"Invalid ''v'' verifier value.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Unable to find account with specified email.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"21","msg":"Unable to find account with specified email.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/srp_change_password?	
		email=me@example.com&
		salt=32232323232323232323232323232323223&
		v=444444444444444...big.number...4444444444444444&
		_ran_=0.12332321123

Output:

``` JSON
	{"status":"success"}
```


### TODO.

1. Require that the page have a pageMarker cookie that is set and current.



## /api/genTempKeys - Generate a new set of backup one-time-keys.

Upon login the user is presented with a set of backup one-time-keys.  These are generated
as a part of the user''s first login.    After they have used up 75% of these keys (usually 15)
the sample application will prompt them with a new set of 20 more backup keys.  This is
the underlying call that generates these keys.

You must be logged into with AesSrp to make this call.

If this is called from the 2fa device on a DeviceId account, then the backup keys are
for the device to work off-line.

The current DeviceId is also returned.  This may be removed in the future.

#### File: aessrp_ext.go Line No: 764, 571, 760.


``` green-bar
	func respHandlerGenTempKeys(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.

No input required.


#### Output.

On Success A JSON hash with:

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"
`BackupKeys`        | The set of backup keys - usually 20 of them.

#### Possible Output Errors.


##### Invalid input data.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"38","msg":"Invalid input data.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Failed to find user. Invalid input email.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"39","msg":"Failed to find user. Invalid input email.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### fmt.sprintf(`unable to find account with email ''%s`, email))

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"58","msg":"fmt.Sprintf(`Unable to find account with email ''%s`, email))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/genTempKeys?
		_ran_=0.12332321123

Output:

``` JSON
	{"status":"success","BackupKeys":"...","DeviceId":"92112121"}
```




## /api/srp_email_confirm - Generate a new set of backup one-time-keys.

Finish the registration process by authorizing the account.   Upon registration
the user is sent an email with an authorization token in it.   This is the call
to finish the registration.

#### File: aessrp_ext.go Line No: 1721, 522, 556.


``` green-bar
	func respHandlerEmailConfirm(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.

Name | Description
|--- | --- 
`email_auth_token` | The authorization token from the email that was sent to the user.


#### Output.

On Success A JSON hash with:

Name                | Description
|---               | ------------------------------------------------------------------------------------
`status`            | "Success"

#### Possible Output Errors.


##### fmt.sprintf("Fatal error. - Did not get passed a bufferhtml.midbuffer - at: %s\n", godebug.lf()))

If SendStatusOnerror is true, then  Status = 500, Internal Server Error,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"5","msg":"fmt.Sprintf("Fatal Error. - Did not get passed a bufferhtml.MidBuffer - AT: %s\n", godebug.LF()))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid email address.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"16","msg":"Invalid email address.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


##### Invalid auth_token address.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"17","msg":"Invalid auth_token address.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
```


### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/srp_email_confirm?
		email_auth_token=UUID&
		_ran_=0.12332321123

Output:

``` JSON
	{"status":"success"}
```






## /api/get_user_attrs - Get user attributes associated with a logged in user.

Get the user attributes for the currently logged in user.

#### File: aessrp_ext.go Line No: 3751, 566.


``` green-bar
	func respHandlerGetUserAttrs(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  

Requires an encrypted login.

#### Output.

Varies.  This depends on what user attributes have been set.


#### Possible Output Errors.

If logged in - probably none.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/get_user_attrs?
		_ran_=0.12332321123




## /api/admin_get_user_attrs - Get user attributes associated with a specified user.

Get the user attributes for another user.  Only an admin can do this.

You can not get the attributes on another ''admin'' user. - To do that requires the ''root'' account.

#### File: aessrp_ext.go Line No: 3918, 568.


``` green-bar
	func respHandlerAdminGetUserAttrs(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  

Requires an encrypted login to an admin acount.

Name | Description
|--- | --- 
`email` | The user to get the attributes for.



#### Output.

Varies.  This depends on what user attributes have been set.


#### Possible Output Errors.

If logged in - probably none.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/admin_get_user_attrs?
		_ran_=0.12332321123




## /api/set_user_attrs - Get user attributes associated with a user.

Sets the user attributes for the currently logged in user.

#### File: aessrp_ext.go Line No: 3673, 565.


``` green-bar
	func respHandlerSetUserAttrs(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  

Requires an encrypted login to an admin account.

Name | Description
|--- | --- 
`attribute` | A specified attribute that you want to set.



#### Output.

Success if this is a valid user and this is a valid admin logged in.


#### Possible Output Errors.

If logged in - probably none.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/set_user_attrs?
		_ran_=0.12332321123





## /api/admin_set_user_attrs - Get user attributes associated with a specified user.

Sets the user attributes for another user.  Only an admin can do this.

You can not set the attributes on another ''admin'' user. - To do that requires the ''root'' account.

#### File: aessrp_ext.go Line No: 2240, 123, 579, 580.


``` green-bar
	func respHandlerAdminSetAttributes(www http.ResponseWriter, req *http.Request) {
		...
	}
```


#### Input.

Supports GET and POST.  

Requires an encrypted login to an admin acount.

Name | Description
|--- | --- 
`email` | The user to get the attributes for.
`attribute` | A specified attribute that you want to set.



#### Output.

Sucess if this is a valid user and this is a valid admin logged in.


#### Possible Output Errors.

If logged in - probably none.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/admin_set_user_attrs?
		_ran_=0.12332321123




## /api/admin_set_password - Set the password for another user.

Sets the password, ''salt'', ''v'' for some other user.

You can not set the password on another ''admin'' user. - To do that requires the ''root'' account.

#### File: aessrp_ext.go Line No: 2149, 122, 578.


	func respHandlerAdminSetPassword(www http.ResponseWriter, req *http.Request) {
		...
	}


#### Input.

Supports GET and POST.  

Requires an encrypted login to an admin account.

Name | Description
|--- | --- 
`email` | The user to set the password on.
`salt`             | Salt for creating new password. - If empty, then this will set a cookie and return.
`v`                | Vector for creating new password. - If empty, then this will set a cookie and return.



#### Output.

Success if this is a valid user and this is a valid admin logged in.


#### Possible Output Errors.

If logged in - probably none.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://localhost:8000/api/admin_set_password?
		_ran_=0.12332321123























### TODO - Document.

#### Delayed due to re-write of Stay Logged In feature.

	muxEnc.HandleFunc("/api/resumeSession", respHandlerResumeSession).Method("GET", "POST")       // ENC:	Allow for resumption of a session.






### Tested.

Tested On: Thu Dec 17 14:24:25 MST 2015, Version 0.5.8 of Go-FTL.

Tested With: Redis 2.8 and PostgreSQL 9.4.

Tested With: Redis 2.8 and PostgreSQL 9.5.

Tested On: Wed Mar 30 09:48:28 MDT 2016 -- AngularJS 1.x test + unit tests ./aes_1_test.go. -- Passed.

Tested On: Fri Apr  6 06:22:08 MDT 2016 with a React 0.14 and React 15 front end. Passed.

Tested On: Fri Apr  8 08:09:43 MDT 2016 with a jQuery and jQuery mobile front end. Passed.

Tested On: Mon Apr 18 14:03:15 MDT 2016 with a ReactNative iOS front end. Passed.

### TODO.

1. Configurable time delay.  TODO_4001 - "TimeDelayForInvalidLogin" to be added.
2. /api/resumeSession, /api/setupStayLoggedIn.
3. Implement for Socket.IO.

References.
----------

1. [SRP Standard, RFC 5054](http://www.ietf.org/rfc/rfc5054.txt)
2. [Standford STP library and information](http://srp.stanford.edu/)



', 'SrpAesAuth-Strong-Authentication-for-RESTful-Requests-100001.html'
	, 'SrpAesAuth: Strong Authentication for RESTful Requests' , 'Strong authentication using Secure Remote Password (SRP), Two Factor Authrization (2FA) and encryption of messages with Advanced Encryption Standard (AES)', '/doc-SrpAesAuth-Strong-Authentication-for-RESTful-Requests', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This middleware implements HTTP basic auth with the authorization stored in a flat file.
If you need to use a database for the storage of usernames/passwords, then you should look
at one of the other two basic-auth middlware.   If you are looking for an example of how
to use a relational database, or how to use a non-relational database, the other basic-auth
middlware are recomended.

Basic auth should only be used in conjunction with TLS (https).  If you need to use
an authentication scheme with http, or you want a better authentication scheme
take a look at the auth_srp.go  middleware.  There are examples of using it with
jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   

Also this is "basic auth" with the ugly browser popup of username/password and no
real error reporting to the user.  If you want something better, switch to the SRP/AES
solution.

Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
of the time.  So... this is only "basic" auth - with low security.

So what is "basic" auth really good for?  Simple answer.  If you need just a
touch of secruity - and no more.   Example:  You took a video of your children
and you want to send it to Grandma.  It is too big for her email so
you need to send a link.  So do a quick copy of it up to your server and set basic
auth on the directory/path.  Send her the link and the username and password.
This keeps googlebot and other nosy folks out of it - but it is not really
secure.  Then a couple of days later you delete the video.   Works like a
champ!

There is a command line tool in ./cli-tools/htaccess to maintain the .htaccess
file with the usernames and hashed passwords.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "BasicAuth": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "myserver.com"
	} },
``` 

With the "AuthName" you can set the name of the authorization file.  It defaults to .htaccess in the current directory.  

``` JSON
	{ "BasicAuth": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "myserver.com",
		"AuthName": "/etc/go-ftl-cfg/htaccess.conf"
	} },
``` 

If you use this middleware it will also ban fetching .htaccess or whatever you have set for AuthName as a file.

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "BasicAuth": {
					"Paths": [ "/private1", "/private2" ],
					"Realm": "zepher.com",
					"AuthName": "/Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/basicauth/htaccess.conf"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Thu Dec 17 14:24:25 MST 2015, Version 0.5.8 of Go-FTL

Tested On: Sat Feb 27 07:30:27 MST 2016

### TODO

1. Add check that .htaccess becomes un-fetchable
', 'BasicAuth-Implement-Basic-Authentication-Using-a-htaccess-File-100002.html'
	, 'BasicAuth: Implement Basic Authentication Using a .htaccess File' , 'Basic Auth implemented with a flat file for hashed usernames/passwords', '/doc-BasicAuth-Implement-Basic-Authentication-Using-a-htaccess-File', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This middleware implements HTTP basic auth with the authorization stored in PostgreSQL.

The PG package used to access the database is:

	https://github.com/jackc/pgx

Pbkdf2 is used to help prevent cracking via rainbow tables.  Each hashed password
is strengthened by using salt and 5,000 iterations of Pbkdf2 with a sha256 hash.

Basic auth should only be used in conjunction with TLS (https).  If you need to use
an authentication scheme with http, or you want a better authentication scheme,
take a look at the aessrp.go  middleware.  There are examples of using it with
jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   

Also this is "basic auth" with the ugly browser popup of username/password and no
real error reporting to the user.  If you want something better switch to the SRP/AES
solution.

Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
of the time.  So... this is only "basic" auth - with low security.

So what is "basic" auth really good for?  Simple answer.  If you need just a
touch of secruity - and no more.   Example:  You took a video of your children
 and you want to send it to Grandma.  It is too big for her email so
you need to send a link.  So quick copy it up to your server and set basic
auth on the directory/path.  Send her the link and the username and password.
This keeps googlebot and other nosy folks out of it - but it is not really
secure.  Then a couple of days later you delete the video.   Works like a
champ!

There is a command line tool in ../../../tools/user-pgsql/user-pgsql.go to maintain the data
in the PostgreSQL database.  You can create/update/delete users from the database.  Also the
tool is useful for verifying that you can connect to the database.

The database connection information is in the global-cfg.json file.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "BasicAuthPgSql": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "example.com"
	} },
``` 

SQL Configuration Script

The setup script to create the table in the database is in .../Go-FTL/server/midlib/basicpgsql/user-setup.sql.
You will need to modify this file and run this before using the middleware.  The realm in the "username" field
is "example.com".  That will need to match the realm you are using in your configuration.

``` SQL
	-- drop TABLE "basic_auth" ;
	CREATE TABLE "basic_auth" (
		  "username"				char varying (200) not null primary key
		, "salt"					char varying (100) not null
		, "password"				char varying (180) not null 
	);

	delete from "basic_auth" where "username" = ''example.com:testme'';
	insert into "basic_auth" ( "username", "salt", "password" ) values ( ''example.com:testme'', ''salt'', 
		''9b6095510e3e1c0ea568c3faf29e545c364265d017b16614b1a2de3efe96bc6313cb9e1d221134a46fd5faa8499ebb8568a2ec489e32fa4c4adcd89c05394292''
	);

	\q
``` 
	
### Tested
		
Tested on : Thu Mar 10 16:25:37 MST 2016, Version 0.5.8 of Go-FTL with Version 9.4 of PostgreSQL.

', 'BasicAuthPgSql-Basic-Auth-Using-PostgreSQL-100003.html'
	, 'BasicAuthPgSql: Basic Auth Using PostgreSQL' , 'Basic Auth implemented with data stored in PostgreSQL', '/doc-BasicAuthPgSql-Basic-Auth-Using-PostgreSQL', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This middleware implements HTTP basic auth with the authorization stored in Redis.

The package used to access the Redis database is:

	https://github.com/garyburd/redigo/redis

Pbkdf2 is used to help prevent cracking via rainbow tables.  Each hashed password
is strengthened by using salt and 5,000 iterations of Pbkdf2 with a sha256 hash.

Basic auth should only be used in conjunction with TLS (https).  If you need to use
an authentication scheme with http, or you want a better authentication scheme,
take a look at the aessrp.go  middleware.  There are examples of using it with
jQuery and AngularJS 1.3 (2.0 of AngularJS coming soon).   

Also this is "basic auth" with the ugly browser popup of username/password and no
real error reporting to the user.  If you want something better switch to the SRP/AES
solution.

Remember that rainbow tables can crack MD5 hashes in less than 30 seconds 95%
of the time.  So... this is only "basic" auth - with low security.

So what is "basic" auth really good for?  Simple answer.  If you need just a
touch of secruity - and no more.   Example:  You took a video of your children
and you want to send it to Grandma.  It is too big for her email so
you need to send a link. So quick copy it up to your server and set basic
auth on the directory/path.  Send her the link and the username and password.
This keeps googlebot and other nosy folks out of it - but it is not really
secure.  Then a couple of days later you delete the video.   Works like a
champ!

There is a command line tool in ../../tools/user-redis/user-redis.go to maintain the data
in the Redis database.  You can create/update/delete users from the database.  Also the
tool is useful for verifying that you can connect to the database.

The database connection information is in the global-cfg.json file.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth.  In the server configuration file:

``` JSON
	{ "BasicAuthRedis": {
		"Paths": [ "/video/children", "/family/pictures" ],
		"Realm": "example.com"
	} },
``` 

A sample setup for Redis is in: `redis-setup.redis`.  To run

``` Bash
	$ redis-cli <redis-setup.redis
``` 

To run this you must have valid connection info in ../test_redis.json.

### Tested
		
Tested on : Thu Mar 10 16:00:44 MST 2016, Version 0.5.8 of Go-FTL with Version 2.8 of Redis.

', 'BasicAuthRedis-Basic-Auth-using-Redis-100004.html'
	, 'BasicAuthRedis: Basic Auth using Redis' , 'Basic Auth implemented with data stored in Redis', '/doc-BasicAuthRedis-Basic-Auth-using-Redis', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Create headers to set or delete cookies.

Configuration
-------------

Name and Value are required.  Other configuration options for the cookie are optional.  Normally Domain will
also need to be set.  If you want your cookie to be available to `www.example.com` and `cdn.example.com,` then use
`.example.com`.  

Use only one of `MaxAge` and `Expires`.  To delete a cookie set the value to an empty `Value`, `""` and `MaxAge` to `-1`.

In this example the path `/somepath` will get a cookie named `testcookie` with a value of `1234`.  The cookie 
expires in a very confusing `12001` seconds or in 2018 (not good, but this is an example).  This is not
a secure cookie.

Secure cookies can only be set when using HTTPS.


``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Cookie": { 
					"Paths":    "/somepath",
					"Name":     "testcookie",
					"Value":    "1234",
					"Domain":   "www.example.com",
					"Expires":  "Thu, 18 Dec 2018 12:00:00 UTC",
					"MaxAge":   "12001",
					"Secure":   false,
					"HttpOnly": false
				} },
			...
		]
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Cookie": { 
					"Paths":    "/somepath",
					"Name":     "testcookie",
					"Value":    "1234",
					"Domain":   ".zepher.com",
					"Expires":  "Thu, 18 Dec 2018 12:00:00 UTC"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

### Tested

Thu, Mar 10, 13:11:43 MST, 2016

### TODO

Use template substitution on the cookie name and value.

Add a "Delete" flag that correctly sets the values for a delete with a single flag.

', 'Cookie-Set-Delete-Cookies-100005.html'
	, 'Cookie: Set/Delete Cookies' , 'Manipulation of cookies', '/doc-Cookie-Set-Delete-Cookies', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Implements templated directory browsing.  

You provide a template, (see example below), and place that in one of the directories specified by "Root" option.
If a *directory* is browsed inside the set of "Paths," then the template will be applied to the file names.

If the template fails to parse, or if no template is supplied, then this is logged to the log file.
An error will be returned.

If the tempalte root is not specified, then the root directory for serving files will be searched
for the specified template name.

This is implemeted inside the "file_serve" - this middlware just sets configuration for 
"file_serve".

Configuration
-------------

Specify template name and the location to find it.  The default template name is "index.tmpl".

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "DirectoryBrowse": { 
					"Paths": [ "/static", "/www" ],
					"TemplateName": "dir-template.tmpl",
					"Root": [ "/static/tmpl" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "DirectoryBrowse": { "LineNo":5, 
					"Paths":   "/static",
					"TemplateName": "index.tmpl"
				} },
				{ "DirectoryLimit": { "LineNo":5, 
					"Paths":   "/static",
					"Disalow": [ "/static/templates" ],
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


Example template, Put in index.tmpl

``` HTML
	{{define "content_type"}}text/html; charset=utf-8{{end}}
	{{define "page"}}<!DOCTYPE html>
	<html lang="en">
	<body>
		<ul>
		{{range $ii, $ee := .files}}
			<li><a href="{{$ee.name}}">{{$ee.name}}</a></li>
		{{end}}
		</ul>
	</body>
	</html>
	{{end}}
``` 

### Tested

Wed, Mar 2, 10:05:04 MST, 2016

', 'DirectoryBrowse-Use-Template-for-Directory-Browsing-100006.html'
	, 'DirectoryBrowse: Use Template for Directory Browsing' , 'Control layout and availabity of directory browsing with Go templates', '/doc-DirectoryBrowse-Use-Template-for-Directory-Browsing', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This is a simple middleware that allows the dumping of the HTTP or HTTPS request.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths, so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the requests/responses are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.

Configuration
-------------

If the `FileName` is not specified, then standard output will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "DumpReq": { 
					"Paths":   "/api",
					"FileName": "./log/out.log",
					"Msg": "At beginning of request",
					"SaveBodyFlag": true
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "DumpReq": { "LineNo":5, 
					"Paths":   "/api",
					"Msg": "At beginning of request"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Wed Mar  2 14:19:00 MST 2016

', 'DumpReq-Dump-Request-with-Message-to-Output-File-Development-Tool-100007.html'
	, 'DumpReq: Dump Request with Message to Output File - Development Tool' , 'Dump out the contents of the reqeust at ths point in the middlware stack.', '/doc-DumpReq-Dump-Request-with-Message-to-Output-File-Development-Tool', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This is a simple middleware that allows the dumping of response to an output file. 

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  It can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  It can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the request/response are at this point in the stack.
`Else` |   A catch all for handling requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.

Configuration
-------------

If the `FileName` is not specified, then standard output will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "DumpResponse": { 
					"Paths":   "/api",
					"FileName": "./log/out.log",
					"Msg": "At beginning of request",
					"SaveBodyFlag": true
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "DumpResponse": { "LineNo":5, 
					"Paths":   "/api",
					"Msg": "At beginning of request"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Wed Mar  2 12:03:48 MST 2016

', 'DumpResponse-Dump-Request-with-Message-to-Output-File-Development-Tool-100008.html'
	, 'DumpResponse: Dump Request with Message to Output File - Development Tool' , 'Dump out the contents of the response at ths point in the middlware stack.', '/doc-DumpResponse-Dump-Request-with-Message-to-Output-File-Development-Tool', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This is a simple middleware that allows echoing of a message.

This can be used as an end-point to test other items.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths. So it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the requests/responses are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.


Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Echo": { 
					"Paths":   "/api/echo",
					"Msg": "Yes I reaced this point"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Echo": { "LineNo":5, 
					"Paths":   "/api/echo",
					"Msg": "Yes I reaced this point"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Wed Mar  2 15:11:25 MST 2016

', 'Echo-Output-a-Message-When-End-Point-Reached-100009.html'
	, 'Echo: Output a Message When End Point Reached' , 'Output a message to the log', '/doc-Echo-Output-a-Message-When-End-Point-Reached', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













It is possible that a server will receive requests that do not match any name or IP address.
An example would be a new, or unconfigured DNS name, that maps to the current IP address of
the Go-FTL server.  If an `Else` middleware is used, then a list of resolvable names
will be displayed to the user and the user can click on one of the links.

There is only 1 `Else` for all named servers.  Usually it is placed at the bottom of the
configuration file.

Configuration
-------------

Not much configuration.  The only option is to have a message that displays before the
list of configured servers.

``` JSON
	{
		...
		...
		...
		...
		"elseServer": { 
			"listen_to":[ "*" ],
			"plugins":[
				{ "Else": { 
					"Paths":   "/",
					"Msg": "<h1> This is the Go-FTL server for: </h1>"
				} }
	}
``` 

### Tested

Tested On: Fri Mar 11 07:46:02 MST 2016

', 'Else-Return-a-Page-for-a-Failed-Virtual-Host-Name-or-SNI-Match-100010.html'
	, 'Else: Return a Page for a Failed Virtual Host Name or SNI Match' , 'This may not be working yet.  Under Construction', '/doc-Else-Return-a-Page-for-a-Failed-Virtual-Host-Name-or-SNI-Match', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Map error codes from lower level calls onto template files.

The following items can be used in the template file, `{{.IP}}`. For example:

Item         | Description
|---        | --- 
`IP`         | Remote IP address
`URI`        | Remote URI
`delta_t`    | How long this has taken to process
`host`       | Host name
`ERROR`      | Text error message if any
`method`     | Request method
`now`        | Current time stamp
`path`       | Path from request
`port`       | Port request was made on
`query`      | The request query string
`scheme`     | http or https
`start_time` | Time request was started at
`StatusCode` | Status code, 200 ... 5xx
`StatusText` | Text description of status code

Configuration
-------------

You provide a list of errors that you want to have mapped, with a template, onto 
a page.  You can provide a directory where the templates are for custom error
templates.   If you do not, then the directory `./errorTemplates/` will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "ErrorTemplate": { 
					"Paths":   "/",
					"Errors": [ "404", "500" ]
				} },
			...
	}
``` 


Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "ErrorTemplate": { "LineNo":5, 
					"Paths":   "/",
					"Errors": [ "404", "500" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Wed, Mar 30, 06:03:59 MDT, 2016

### TODO

1. Way to configure "application" or "home-page" for template.
2. Logging of errors.
3. Possibility of a "form" for errors to contact user when error is fixed.
4. Contact Support info.
5. ./errorTempaltes relative to "root" of application.
6. For users that are logged in - a different template that reflects name/time etc for logged in user.
7. Match "4xx" as an error to a 4xx.tmpl file and a 400 error so you don''t have to have zillions of files.

', 'ErrorTemplate-Convert-Errors-to-Pages-100011.html'
	, 'ErrorTemplate: Convert Errors to Pages' , 'Extended loggin with additional attributes via a template stubstitution', '/doc-ErrorTemplate-Convert-Errors-to-Pages', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of content to geographic regions based on mapping of IP addresses to these regions.
This works on a per-country basis most of the time.  The data is not 100% accurate.

The data is based on the freely available GetLite2 database.  You need to download your own copy
of this data - the data that is in the ./cfg directory is terribly out of date and should only
be used for testing of this middleware.

Also note: The data changes periodically.   Hopefully one day this module will automatically
update the data - but for the moment you have to update it by hand.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "GeoIpFilter": { 
					"Paths":   "/",
					"Action":  "reject",
					"CountryCodes":  [ "JP", "VN" ],
					"DBFileName":    "./cfg/GeoLite2-Country.mmdb",
					"PageIfBlocked": "not-avaiable-in-your-country.html"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "GeoIpFilter": { "LineNo":5, 
					"Paths":   "/",
					"Action":  "reject",
					"CountryCodes":  [ "JP", "VN", "CN" ],
					"DBFileName":    "./cfg/GeoLite2-Country.mmdb",
					"PageIfBlocked": "not-avaiable-in-your-country.html"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

### Tested

Fri, Mar 11, 09:15:38 MST, 2016

### TODO

1. Add automatic update of underlying data.
1. Improve data quality.


', 'GeoIpFilter-Filer-Requests-Based-on-Geographic-Mapping-of-IP-Address-100012.html'
	, 'GeoIpFilter: Filer Requests Based on Geographic Mapping of IP Address' , 'Use IP address to filter to a set of geograpic regions', '/doc-GeoIpFilter-Filer-Requests-Based-on-Geographic-Mapping-of-IP-Address', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












GoTemplate implements a middleware that combines templates with underlying data.

Basic usage of Go templates is also supported.  You can build a page with a header
template, a footer template and a body template.

A more powerful way to use this is to combine data with templates to render a
final text.  Examples of each of these will show how this can be used.

Configuration
-------------

Specify a path for templates and the location of the template library.

Parameter | Description
|--- | --- 
`TemplateParamName` | The name on the URL of the template that is to be rendered with this data.
`TemplateName` | The name of the template if __template__ has an empty value.
`TemplateLibraryName` | An array of file names or a single file that has the set of templates for rendering the page.
`TemplateRoot` | The path to search for the template libraries.  If this is not specified, then it will be searched for in `Root`.
`Root` | The root for the set of web pages.  It should be the same root as the `file_server` `Root`.


``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "GoTemplate": { 
					"Paths": [ "/data" ],
					"TemplateParamName":     "__template__",
					"TemplateName":          "render_body",
					"TemplateLibraryName":   "common_library.tmpl",
					"TemplateRoot":          "./tmpl",
					"Root":                  ""
				} },
			...
	}
``` 


Example 1: Simple Page Composition
----------------------------------

You have a website with a common header, footer and each body is different.

The Go-FTL configuration file is:

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "GoTemplate": { 
					"Paths": [ "/twww" ],
					"TemplateParamName":     "__template__",
					"TemplateName":          "body",
					"TemplateLibraryName":   [ "common_library.tmpl", "{{.__tempalte__}}.tmpl" ]
					"TemplateRoot":          "./tmpl",
					"Root":                  ""
				} },
				{ "Echo": { 
					"Paths": [ "/twww" ],
					"Msg": ""
				} }
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

In ./tmpl/common_library.tmpl you have

``` HTML
	{{define "content_type"}}text/html; charset=utf-8{{end}}
	{{define "header"}}<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
	{{end}}
	{{define "footer"}}
		<div> footer </div>
	</body>
	</html>
	{{end}}
	{{define "body"}}
		{{header .}}
		<div> this is my default body - it is a good body 1 </div>
		<div> this is my default body - it is a good body 2 </div>
		{{footer .}}
	{{end}}
``` 

In ./tmpl/main.tmpl you have

``` HTML
	{{define "main"}}
		{{header .}}
		<div> this is my main body </div>
		{{footer .}}
	{{end}}
``` 

A request for `http://www.zepher.com:3210/twww?__template__=main` will do the following:

1. GoTemplate sees the url `/twww` and calls the next function down the stack.
2. Echo sees the url `/twww` and matches - It returns the Msg string as the results.  An empty string.
3. GoTemplate uses the returning data from Echo.  This is actually an empty string.   It reads in the template files in order, common_library.tmpl then substituting the parameter, main.tmpl.  It then calls the template "main" witch calls the "header" and "footer" templates to render.

The returned data is transformed into (with a couple of extra blank lines suppressed)

``` HTML
	<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
		<div> this is my main body </div>
		<div> footer </div>
	</body>
	</html>
``` 

If __template__ had not been specified, then the template "body" would have been called.  It acts as a default body in this case.

The `content_type` template is used to generate the content type for the page.  You can use this to generate XML or SVG, or to transform data
and return it in other mime types.

In this example you may want to use Rewrite first to generate the ugly URL: `http://www.zepher.com:3210/twww?__template__=main`

The documentation for this tool is generated in this fashion.  It is actually a little bit more complicated.  The files are in Markdown (.md) and processed from .md to
.html, then written into templates, .tmpl and combined with headers and footers.

Example 2: Page Composition with Data
-------------------------------------

Combining data with templates is incredibly powerful.  For this example we will combine some static data in a .json file with templates to render it.
You can also use this with the RedisListRaw to pull data out of Redis and combine it with templates to render it.   This turns the templates into a
simple report writer tool.  Complete access to a relational database is also available with the `TabServer2` middleware.  This has been tested with
PostgreSQL, MySQL, Oracle, and Microsoft MS-SQL.

The Go-FTL configuration file is:

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "GoTemplate": { 
					"Paths": [ "/data/" ],
					"TemplateParamName":     "__template__",
					"TemplateName":          "body",
					"TemplateLibraryName":   [ "data_library.tmpl" ]
					"TemplateRoot":          "./tmpl",
					"Root":                  ""
				} },
				{ "JSONToTable": { "LineNo":5, 
					"Paths":   "/data/",
					"ConvertRowTo1LongTable": true
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

In ./tmpl/data_library.tmpl you have

``` HTML
	{{define "content_type"}}text/html; charset=utf-8{{end}}
	{{define "header"}}<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
		<ul>
	{{end}}
	{{define "data_render_body"}}
		{{range $ii, $ee := .data}}
			<li><a href="/{{$ee.abc}}.html"> {{$ee.abc}} id:{{$ee.myId}} </a></li>
		{{end}}
	{{end}}
	{{define "footer"}}
		</ul>
		<div> footer </div>
	</body>
	</html>
	{{end}}
```

With data served by the file server in ./data/some_data.json

``` JSON
	[
		{ "abc": "page-1", "myId": 101 },
		{ "abc": "page-2", "myId": 102 },
		{ "abc": "page-3", "myId": 103 }
	]
``` 

A request for `http://www.zepher.com:3210/data/some_data.json?__template__=data_render_body`
will do the following:

1. The request works its way down to the `file_server`.
2. JSONToTable converts the returning text to table data internally.
3. GoTemplate takes the table data and applies the templates.  `data_render_body` creates a header, then iterates over the set of rows, then adds the footer.

The url: `http://www.zepher.com:3210/data/some_data.json?__template__=data_render_body` will produce the following:

``` HTML
	<!DOCTYPE html>
	<html lang="en">
	<body>
		<div> header </div>
		<ul>
			<li><a href="/page-1.html"> page-1.html id:101 </a></li>
			<li><a href="/page-2.html"> page-1.html id:102 </a></li>
			<li><a href="/page-3.html"> page-1.html id:103 </a></li>
		</ul>
		<div> footer </div>
	</body>
	</html>
``` 

In this case any source of table data or a row of data can then be rendered into a final output form.

### Tested

Tested On: Wed Mar  2 10:01:28 MST 2016 - Unit Tests

Tested On: Wed Mar  3 12:40:48 MST 2016 - End to End Tests of Templates.

### TODO

TODO - Have links to Go templates and how to use them.

', 'GoTemplate-Template-using-Go-s-Buit-in-Templates-100013.html'
	, 'GoTemplate: Template using Go''s Buit in Templates' , 'Use Go Templates to format data', '/doc-GoTemplate-Template-using-Go-s-Buit-in-Templates', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Gzip allows for compression of return data.  It may pose a security risk if used in
combination with HTTPS.  The security risk is a timing attack. It is mitigated
by using caching that causes the gzip compression to only run when the file 
changes.

The default is to compress anything that is larger than 500 bytes.

Configuration
-------------

Gizp any data that is larger than 1,000 bytes and is from the /static directory.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Gzip": { 
					"Paths":   "/static",
					"MinLength": 1000
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Gzip": { "LineNo":5, 
					"Paths":   "/static",
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Sat Feb 27 08:04:50 MST 2016

Tue May  3 09:14:29 MDT 2016 -- After changes to work with caching.


', 'Gzip-Ban-Certain-IP-Addresses-100014.html'
	, 'Gzip: Ban Certain IP Addresses' , 'Gzip compresses output before it is returned.  Interacts with caching so ''zip'' process only happens if file changed', '/doc-Gzip-Ban-Certain-IP-Addresses', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












AngularJS 2.0 and AngularJS 1.x have an interesting default routing.  They change the current path.
For example, `http://myapp.com/app.html` becomes `http://myapp.com/app.html/dashboard`  and then `http://myapp.com/app.html/productList`.
When a person bookmarks or refreshes one of these URLs the server has no clue what a "/app.html/dashboard" is and returns
a 404 error.  

One possible solution is to map all 404 errors to `app.html`.   This is *icky* because it breaks all 404 handling.  You end up
returning `app.html` for `/image/nonexistent.jpg` and the browser is not happy with you at all (and it shouldn''t be!)

The solution in this middleware is more nuanced. If the lower levels return a 404 and this is a `GET` request
then if one of the Paths regular expression matches use a regular expression to replace the selected portion of
the URL and retry that.

What should happen is that all of these should be mapped to the single page application.  By default this
is `&lt;some-name&gt;.html`.  You can change this with the `ReplaceWith` option.

After the file server returns a 404 you can limit the set of paths with the `LimitTo` set of options.
If `LimitTo` is not specified, then all 404 errors will be returned as index.html.

At best this should be considered *experimental*. I am still working on what to do with `/` maping to `/index.html`.
Anticipate changes in this middleware in the near future.

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "HTML5Path": { "LineNo":__LINE__,
					"Paths":["(/.*\\.html)/.*"]
				} },
			...
	}
```

Full Example
------------

This example is the server confgiuration that I used for my Angular 2.0 rc 1 documentation development (The page 
you are currently reading)  This also includes the configuration for TabServer2.

``` JSON
	{
		"working_test_AngularJS_20": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16020", "http://dev2.test1.com:16020" ],
			"plugins":[
				{ "HTML5Path": { "LineNo":__LINE__,
					"Paths":["(/.*\\.html)/.*"]
				} },
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "docs.go-ftl.com",
					"AppRoot": "/Users/corwin/Projects/docs-Go-FTL/data/",
					"StatusMessage":"Version 0.0.4 Sun May 22 19:12:43 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/Users/corwin/Projects/docs-Go-FTL", "Paths":"/"  } }
			]
		}
	}
```

### Tested

Tested On: Wed Jun  1 13:05:11 MDT 2016 (Note - Tested by using it in an AngularJS 2.0 application)  An automated test is in-the-works.

### TODO

1. A better name for this middleware.   As soon as I can figure out what to call it I will change this.




', 'HTML5Path-Redirect-404-Errors-to-Index-html-for-AngularJS-Router-100015.html'
	, 'HTML5Path: Redirect 404 Errors to Index.html for AngularJS Router' , 'Angular 1.x, 2.x and other HTML5 single pages applications uses multiple URLs that all need to direct to a single .html page.', '/doc-HTML5Path-Redirect-404-Errors-to-Index-html-for-AngularJS-Router', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Create headers to set or delete cookies.

See a header in the response to a request, or delete a header if it exists.

Configuration
-------------

Create a header.  If you want to set a cookie it is probably better to use `Cookie` middleware instead.
If the header `Name` starts with ''-'' then delete the header if it exists.

Create a header

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Header": { 
					"Paths":    "/somepath",
					"Name":     "X-Header",
					"Value":    "1234"
				} },
			...
	}
``` 

Delete a header

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Header": { 
					"Paths":    "/somepath",
					"Name":     "-X-bad-header"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Header": { 
					"Paths":    "/somepath",
					"Name":     "X-Test-Header1",
					"Value":    "1234"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

### Tested

Tested On: Sat Feb 27 08:02:47 MST 2016

### TODO

1. Use template to allow substitution of header name and values.

', 'Header-Set-Delete-Headers-100016.html'
	, 'Header: Set/Delete Headers' , 'Manipulation of response heades', '/doc-Header-Set-Delete-Headers', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This is primarily intended as an in-memory cache.  It will also, if configured, cache files to disk.
The cleanup time on disk cached items is by default 1 hour.

Configuration
-------------

Lots of configuration items.

Item | Default | Description
|--- | --- | ---
`Extensions`      | no-default  | The set of file extensions that will be cached.
`Duration`        |          60 | How long, in seconds, to cache in memory.
`IgnoreUrls`      | no-default  | Paths to be ignored - and not cached.  For example "/api/".
`SizeLimit`       |      500000 | Limit on size of items to be cached in memory.  Size in bytes.
`DiskCache`       | no-default  | Set of disk locations to place on-disk cached files.  Used round-robin.  If this item is empty then no disk caching will take place.
`DiskSize`        | 200000000   | Maximum amount of disk space to use for on-disk cached files.
`RedisPrefix`     |    "cache:" | The prefix used in Redis for data stored and updated by this middleware.
`DiskSizeLimit`   |    2000000  | The maximum size for disk-cached items.
`DiskCleanupFreq` |  3600       | How long to keep items in the disk cache.  They are discarded after this number of seconds.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "InMemoryCache": { 
					"Paths":   "/",
					"Extensions":       [ ".js", ".css", ".html" ],
					"Duration":         60,
					"IgnoreUrls":       [ "/api/" ],
					"SizeLimit":        500000,
					"DiskCache":        [ "./cache/" ],
					"DiskSize":         200000000,
					"RedisPrefix":      "cache:",
					"DiskSizeLimit":    2000000,
					"DiskCleanupFreq":  3600
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "InMemoryCache": { "LineNo":5, 
					"Paths":   "/api",
					"Extensions":       [ ".js", ".css", ".html" ],
					"IgnoreUrls":       [ "/api/" ],
					"DiskCache":        [ "./cache/" ],
					"DiskSize":         200000000
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri, Mar 11, 09:22:36 MST, 2016

### TODO

1. Extensive testing with multiple components and the InMemoryCache at the same time.  For example verify that TabServer2 can/will correctly set cache timeout when used with this component.
2. Add the set of mime types to cache - instead of file extensions.
3. Make the file extensions consistent across the Go-FTL system.   In other places the extension `.js` is just `js`.

', 'InMemoryCache-Ban-Certain-IP-Address-100017.html'
	, 'InMemoryCache: Ban Certain IP Address' , 'Implements in memory caching of hot resources and on disk caching for other pages', '/doc-InMemoryCache-Ban-Certain-IP-Address', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Convert data in JSON format into internal table data in the response buffer.

By itself this is not very useful.  However when combined with a template
it allows for JSON data to be read from a file and then formatted into a
final set of data.

Configuration
-------------

A number of options are planned. (See TODO below.)

ConvertRowTo1LongTable:  If this is true, then
a single row of data will be converted into an array 1 long.   If the data is empty,
then an empty array will be returned.

Convert1LongTableToRow: If this is true, then
a table that is 1 row long, (or 0), will be converted to a hash.

Both flags can not be true at the same time.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "JSONToTable": { 
					"Paths":   "/api",
					"ConvertRowTo1LongTable": true,
					"Convert1LongTableToRow": false
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "GoTemplate": { "LineNo":5, 
					"Paths":   "/config/initialSetupData.json",
					"TemplateName": "initialSetupData.tmpl",
					"TemplateRoot": "/tmpl/"
				} },
				{ "JSONToTable": { "LineNo":5, 
					"Paths":   "/config/initialSetupData.json",
					"ConvertRowTo1LongTable": true
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Fri, Mar 11, 12:15:38 MST, 2016

', 'JSONToTable-Convert-JSON-to-Internal-Table-Data-100018.html'
	, 'JSONToTable: Convert JSON to Internal Table Data' , 'Format data into JSON', '/doc-JSONToTable-Convert-JSON-to-Internal-Table-Data', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












JSONP allows for remotely accessing an API that is cross domain.  This implements
JSONP for an existing API.  For example if "callback=Func9999" is provide on the URL and the JSON
returned is {"josn":"code"}, will be wrapped in:

	Func9999({"json":"code"});

This converts the original JSON to a JavaScript callback function.   This can be used 
from jQuery with a request type of "jsonp".

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "JSONp": { 
					"Paths":   "/api",
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "JSONp": { "LineNo":5, 
					"Paths":   "/api/status",
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

[//]: # (This may be the most platform independent comment)

Wed Mar  2 10:36:09 MST 2016


', 'JSONp-Implement-JSONp-requests-100019.html'
	, 'JSONp: Implement JSONp requests' , 'Transorm get reqeusts into JSONp if they have a callback parameter', '/doc-JSONp-Implement-JSONp-requests', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This is a simple middleware that allows the dumping of the HTTP or HTTPS request.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths, so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the requests/responses are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.

Configuration
-------------

If the `FileName` is not specified, then standard output will be used.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "DumpReq": { 
					"Paths":   "/api",
					"FileName": "./log/out.log",
					"Msg": "At beginning of request",
					"SaveBodyFlag": true
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "DumpReq": { "LineNo":5, 
					"Paths":   "/api",
					"Msg": "At beginning of request"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Wed Mar  2 14:19:00 MST 2016

', 'DumpReq-Dump-Request-with-Message-to-Output-File-Development-Tool-100020.html'
	, 'DumpReq: Dump Request with Message to Output File - Development Tool' , 'Dump out the contents of the reqeust at ths point in the middlware stack.', '/doc-DumpReq-Dump-Request-with-Message-to-Output-File-Development-Tool', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












This is a simple middleware that allows slowing down requests.  It is intended to test slow networks like mobile and rural.

Good values to use are 50, for a slow rural network or Version.net mobile, 114 for an average mobile network, 240 for a busy
at 3 in the afternoon mobile network and 522 for my remote Wyoming land line.  By the way this is not an endorsement of
Verison.net in anyway - they claim  to have 50ms latency - but my tests indicate 148 is *much* more likely.

Configuration
-------------

If the `SlowDown` is not specified, then 500ms will be used (1/2 second).

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Latency": { 
					"Paths":   "/slowDownPath/",
					"SlowDown": 500
				} },
			...
	}
``` 

Full Example
------------

This full example slows down the results of every request by 114ms.  That is the average that I seed when I test on
ATT''s mobile network in my remote Wyoming locaiton.


``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Latency": { "LineNo":5, 
					"Paths":   "/",
					"SlowDown": 114
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Wed Jun 15 09:26:46 MDT 2016

', 'Latency-Simulate-latency-for-testing-behavior-on-slow-networks-i-e-mobile--100021.html'
	, 'Latency: Simulate latency for testing behavior on slow networks (i.e. mobile)' , 'Use this  as a tool when testing your web application.  Slows it way down', '/doc-Latency-Simulate-latency-for-testing-behavior-on-slow-networks-i-e-mobile', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified
extensions, then reject the request.

Configuration
-------------

You can provide a simple list of extensions that when matched will be served. 
All other extensions return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "LimitExtensionTo": { 
					"Paths":   "/",
					"Extensions": [ ".html" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "LimitExtensionTo": { 
					"Paths":   "/",
					"Extensions": [ ".html", ".json", ".css", ".js" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 10:48:45 MST 2016


', 'LimitExtensionTo-Limit-Requests-Based-on-File-Extension-100022.html'
	, 'LimitExtensionTo: Limit Requests Based on File Extension' , 'Prevent access to non authorized paths', '/doc-LimitExtensionTo-Limit-Requests-Based-on-File-Extension', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of regular expressions.  If the file is not one of the specified
paths, then reject the request.

Configuration
-------------

You can provide a simple list of extensions that when matched will be served. 
All other paths return  a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "LimitRePathTo": { 
					"Paths": [ "^/.*\\.html" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "LimitRePathTo": { 
					"Paths": [ "^/[a-z][a-z]/", ".html$" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 14:52:04 MST 2016


', 'LimitRePathTo-Limit-Requests-Based-on-File-Extension-100023.html'
	, 'LimitRePathTo: Limit Requests Based on File Extension' , 'Prevent access to non authorized file extensiosn by limiting to a set of valid extensions', '/doc-LimitRePathTo-Limit-Requests-Based-on-File-Extension', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Limit serving of files to the specified set of paths.  If the file is not one of the specified
paths, then reject the request with a HTTP Not Found (404) error.

Configuration
-------------

You can provide a simple list of paths that when matched will be served. 
All other paths return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "LimitPathTo": { 
					"Paths":   "/api"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "LimitPathTo": { 
					"Paths":  [ "/blog", "/api" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 10:55:41 MST 2016

<! -- /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/midlib/LimitPathTo/LimitPathTo.md -->

', 'LimitPathTo-Limit-Requests-Based-on-File-Extension-100024.html'
	, 'LimitPathTo: Limit Requests Based on File Extension' , 'Prevent access to non authorized directories', '/doc-LimitPathTo-Limit-Requests-Based-on-File-Extension', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Log all requests to the logger.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the request/response are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.

The format can substitute any of these items:

Item | Description
:---: | --- 
`IP` | IP address of remote client
`URI` | URI 
`delta_t` | How long the request has taken
`host` | Host name
`ERROR` | Error message that is returned by lower level middleware
`method` | Request Method
`now` | Current Time
`path` | Request Path
`port` | Port Number
`query` | Query String
`scheme` | HTTP or HTTPS
`start_time` | Start time of request
`status_code` | Numeric status code
`StatusCode` |  Numeric status code
`StatusText` | Numeric status converted to a description

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Logging": { 
					"Paths":   "/api",
					"Format": "IP: {{.IP}} METHOD: {{.method}}"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Logging": { "LineNo":5, 
					"Paths":   "/api",
					"Format": "IP: {{.IP}} METHOD: {{.method}}"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Wed, Mar 2, 15:18:12 MST, 2016

', 'Logging-Output-a-Log-Message-for-Every-Request-100025.html'
	, 'Logging: Output a Log Message for Every Request' , 'Add or remove loggin information using templates for log messages', '/doc-Logging-Output-a-Log-Message-for-Every-Request', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Each of the middleware after this in the processing stack will require a login via AesSrp.
This middleware also works with the BasicAuth, BasicAuthRedis, BasicAuthPgSQL.

This tests to verify if a successful login has been passed at a previous point in the
processing.  The top level of the processing reserves a set of parameters like `$is_logged_in$`.
During login, if the login is successful, then this parameter will be set to `y`.  That gets
checked by this middleware.

If "StrongLoginReq" is set to  "yes" then the parameter `$is_full_login$` is also checked to
be a `y`.  This is set to `y` when login has happened and if configured for it, two factor
authentication has taken place.

Why this works
--------------

At the top level the server (top) will remove the parameters $is_logged_in$ and $is_full_login$.  If the parameters
are found then they will get converted into "user_param::$is_logged_in$" and "user_param::$is_full_login$".
Then if login occurs it can set the params and this can see them.

Configuration
-------------

For the paths that you want to protect with this turn on basic auth, or use the AesSrp
authentication.  In the server configuration file:

``` JSON
	{ "LoginRequired": {
		"Paths": [ "/PrivateStuff" ],
		"StrongLoginReq":  "yes"
	} },
``` 


Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "LoginRequired": {
					"Paths": [ "/private1", "/private2" ],
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

As a part of the AesSrp login process.

', 'LoginRequired-Middleware-After-this-Require-Login-100026.html'
	, 'LoginRequired: Middleware After this Require Login' , 'Require login before allowing access to the specified paths below this in the middleware stack', '/doc-LoginRequired-Middleware-After-this-Require-Login', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













Middleware is used for most configuration in the Go-FTL server.

The list on the left hand menu has all the different types of middleware that is already written.


', 'Middleware-100027.html'
	, 'Middleware' , 'Go-FTL is a scalable server and forward proxy designed for development and scalable deployment of web applications.', '/doc-Middleware', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This provides on-the-fly compression and minimization of a number of different file types.  Currently all the files are
text based.

If used in combination with InMemoryCache the files will be cached.  The cache will automatically flush if the original
source file is changed.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Minify": { 
					"Paths":   "/api",
					"FileTypes": [ "html", "css", "js", "svg", "json", "xml" ]
				} },
			...
	}
``` 


Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Minify": { "LineNo":5, 
					"Paths":   [ "/www/", "/static/" ],
					"FileTypes": [ "css", "js", "svg", "json", "xml" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tesed On: Fri Mar 11 09:05:10 MST 2016

### TODO and Notes/Caveats

1. Using the node/npm UglifyJS middleware produces better results for minifying JavaScript than the internal Go code in this middleware.  Consider using that (accessible via the file_server middleware) instead of this.
2. Compression of images.
3. Compression of HTML will remove the `<body>` tag.  This can cause some client side JavaScript to break.

', 'Minify-Compress-Minify-Files-Before-Serving-Them-100028.html'
	, 'Minify: Compress/Minify Files Before Serving Them' , 'Shrink output using minification techniques.  Compress CSS, JavaScript, HTML, SVG, XML and JSON data.', '/doc-Minify-Compress-Minify-Files-Before-Serving-Them', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Adding a prefix (like Google or Facebook) helps to prevent the direct execution of JSON
code.  AngularJS supports `)]}'',\n` as a prefix by default.

``` json

	where(1);{"json":"code"}

```

or

``` json

	)]};{"json":"code"}

```

This addresses [a known JSON security vulnerability](http://haacked.com/archive/2008/11/20/anatomy-of-a-subtle-json-vulnerability.aspx/).

Both server and the client must cooperate in order to eliminate these threats.
This implements the server side for mitigating this attack.
Angular comes pre-configured with strategies that address this issue, but for this to work backend server cooperation is required.
Other front end packages will use a different prefix.  You can set the prefix, but the default is for Angular.

JSON Vulnerability Protection
-----------------------------

A JSON vulnerability allows third party website to turn your JSON resource URL into JSONP request under some conditions.
To counter this your server can prefix all JSON requests with following string ")]}'',\n".
The Client must automatically strip the prefix before processing it as JSON.

For example if your server needs to return:

``` json

	[''one'',''two'']

```

which is vulnerable to attack, your server can return:

``` json

	)]}'',
	[''one'',''two'']

```


Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Prefix": { 
					"Paths":  "/api",
					"Prefix": ")]}'',\n"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Prefix": { "LineNo":5, 
					"Paths":   "/api",
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

[//]: # (This may be the most platform independent comment)

Tested On: Tue Jun 21 08:26:53 MDT 2016


', 'Prefix-Allows-configuration-of-a-prefix-before-JSON-responses-100029.html'
	, 'Prefix: Allows configuration of a "prefix" before JSON responses' , 'Transorm get reqeusts into Prefix if they have a callback parameter', '/doc-Prefix-Allows-configuration-of-a-prefix-before-JSON-responses', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Redirect provides the ability to redirect a client to a new location on this or other servers.  If you do
not specify a HTTP status, then 307 temporary redirect will be used.   It is highly recommended that you
avoid 301 redirects.

Configuration
-------------

You can provide a simple list of paths that you want to redirect.  These will get 307 Temporary redirects.
This will take `/api.v2/getData` and redirect it to http://www.example.com/api/getData.
`{{.THE_REST}}` is defined to be any remaining content from the request URI after the Paths match.
 
``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RedirectToHttps": { 
					"Paths":  [ "/api.v2", "/v1.api" ],
					"To":  [ "http://www.example.com/api{{.THE_REST}}", "http://www.example.com/api{{.THE_REST}}" ],
					"Code": [ "MovedTemporary", "MovedPermanent" ],
					"TemplateFileName": "moved.tmpl"
				} },
			...
	}
``` 


Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RedirectToHttps": { "LineNo":5, 
					"To":  [ "http://www.zepher.com:3210/api{{.THE_REST}}", "http://www.zepher.com:3210/api{{.THE_REST}}" ],
					"To":  [ "/api", "/api" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Sat Feb 27 18:26:02 MST 2016

1. Tested with simple redirect - Done
1. Test with template
1. Test with invalid configuration
1. Test with invalid template
1. Test with missing template


TODO
----

What happens with post/del etc.

', 'RedirectToHttps-Redirect-One-Request-to-Another-Location-100030.html'
	, 'RedirectToHttps: Redirect One Request to Another Location' , 'Client side (307) response redirects to HTTPS', '/doc-RedirectToHttps-Redirect-One-Request-to-Another-Location', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Redirect provides the ability to redirect a client to a new location on this or other servers.  If you do
not specify a HTTP status, then 307 temporary redirect will be used.   It is highly recommended that you
avoid 301 redirects.

Configuration
-------------

You can provide a simple list of paths that you want to redirect.  These will get 307 temporary redirects.
This will take `/api.v2/getData` and redirect it to http://www.example.com/api/getData.
`{{.THE_REST}}` is defined to be any remaining content from the request URI after the Paths match.
 
``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Redirect": { 
					"Paths":  [ "/api.v2", "/v1.api" ],
					"To":  [
						{ "To": "http://www.example.com/api{{.THE_REST}}", "Code": "MovedTemporary" },
						{ "To": "http://www.example.com/api{{.THE_REST}}", "Code": "MovedPermanent" },
					}
					"TemplateFileName": "moved.tmpl"
				} },
			...
	}
``` 


Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Redirect": { "LineNo":5, 
					"To":  [
						{ "To":"http://www.zepher.com:3210/api{{.THE_REST}}" }
					]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Sat Feb 27 18:26:02 MST 2016

1. Tested with simple redirect. - Done.
1. Test with template.
1. Test with invalid configuration.
1. Test with invalid template.
1. Test with missing template.


### TODO

What happens with post/del etc.

', 'Redirect-Redirect-One-Request-to-Another-Location-100031.html'
	, 'Redirect: Redirect One Request to Another Location' , 'Client side (307) response redirects', '/doc-Redirect-Redirect-One-Request-to-Another-Location', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This allows for retrieving data from Redis that has a common prefix.

The data is converted to JSON before it is returned.  If you need "raw" data then use RedisListRaw.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RedisList": { 
					"Paths":           "/api",
					"Prefix":          "pf3:",
					"UserRoles":       [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]
					"UserRolesReject": [ "anon-user" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RedisList": { "LineNo":5, 
					"Paths":   "/api",
					"Prefix":          "pf3:",
					"UserRoles":       [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Sat Apr  9 13:10:17 MDT 2016

### TODO

Allow for other Redis types. - Currently only allows for name/value key pair.


', 'RedisList-Return-Data-from-Redis-100032.html'
	, 'RedisList: Return Data from Redis' , 'Provide limited access to data in Redis based on prefixes to a set of keys', '/doc-RedisList-Return-Data-from-Redis', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This allows for retrieving data from Redis that has a common prefix.

The data is returned as "raw" table data - it has not been converted into JSON or other text.   Pre-converted text can be had with RedisList.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RedisListRaw": { 
					"Paths":           "/api",
					"Prefix":          "pf3:",
					"UserRoles":       [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]
					"UserRolesReject": [ "anon-user" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RedisListRaw": { "LineNo":5, 
					"Paths":   "/api",
					"Prefix":          "pf3:",
					"UserRoles":       [ "anon,$key$", "user,$key$,confirmed", "admin,$key$,confirmed,disabled", "root,name,confirmed,disabled,disabled_reason,login_date_time,login_fail_time,n_failed_login,num_login_times,privs,register_date_time" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Sat Apr  9 13:08:03 MDT 2016

### TODO

Allow for other Redis types. - Currently only allows for name/value key pair.


', 'RedisListRaw-Return-Data-from-Redis-100033.html'
	, 'RedisListRaw: Return Data from Redis' , 'Provide limited access to data in Redis based on prefixes to a set of keys.  Return data in an unformated form so that other middlware can easliy access it.', '/doc-RedisListRaw-Return-Data-from-Redis', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified


RejectDirectory allows for a set of directories to be un-browsable.   Files from the directories
can still be served - but the directories themselves would not be browsable.

If you do not want anything served from the directory, then use "LimitRePath".

This is implemeted inside the "file_serve." - This middlware just sets configuration for 
"file_serve".

Configuration
-------------

Specify a path and a set of specific directory to not be browsable.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectDirectory": { 
					"Paths": [ "/static" ],
					"Disalow": [ "/static/templates" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectDirectory": { "LineNo":5, 
					"Paths":   "/static",
					"Disalow": [ "/static/templates" ],
				} },
				{ "DirectoryBrowse": { "LineNo":5, 
					"Paths":   "/static",
					"TemplateName": "index.tmpl"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 

### Tested

Wed, Mar 2, 10:01:28 MST, 2016

', 'RejectDirectory-Prevent-Browsing-of-a-Set-of-Directories-100034.html'
	, 'RejectDirectory: Prevent Browsing of a Set of Directories' , 'Limit all access to a set of directories', '/doc-RejectDirectory-Prevent-Browsing-of-a-Set-of-Directories', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Based on file extension reject requests.  For example, you may want to prevent anybody
accessing any file ending in `*.cfg`.

Configuration
-------------

You can provide a simple list of extensions that when matched will return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectExtension": { 
					"Paths":   "/",
					"Extensions": [ ".cfg" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectExtension": { 
					"Paths":   "/",
					"Extensions": [ ".cfg", ".password_db" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 11:03:27 MST 2016
', 'RejectExtension-Reject-Requests-Based-on-File-Extension-100035.html'
	, 'RejectExtension: Reject Requests Based on File Extension' , 'Prevent to a set of file extensions by banning them', '/doc-RejectExtension-Reject-Requests-Based-on-File-Extension', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

For matching paths, if the file extension for the request matches then only allow the specified set of
`Referer` headers.   This is primarily used to prevent hot linking of images and JavaScript across sites.

Process:

If the path starts with one of the selected paths then:

If the host is in the list of ignored hosts then just pass this request on to the next handler.

If the request has one of the extensions then check the `referer` header. If the header is valid then pass this on.

If the tests fail to pass then either return an error (ReturnError is true) or return an empty clear 1px by 1px GIF image.

Configuration
-------------

You can provide a simple list of paths to match.  

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectHotlink": { 
					"Paths":           [ "/js/", "/css/", "/img/" ],
					"AllowedReferer":  [ "www.example.com", "example.com" ],
					"FileExtensions":  [ ".js", ".css", ".gif", ".png", ".ico", ".jpg", ".jpeg" ],
					"AlloweEmpty":     "false",
					"IgnoreHosts":     [ "localhost", "127.0.0.1" ],
					"ReturnError":     "yes"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectHotlink": { 
					"Paths":           [ "/js/", "/css/", "/img/" ],
					"AllowedReferer":  [ "www.zepher.com", "zepher.com", "cdn0.zepher.com", "cdn1.zepher.com", "img.zepher.com" ],
					"FileExtensions":  [ ".js", ".css", ".gif", ".png", ".ico", ".jpg", ".jpeg", ".otf", ".eot", ".svg", ".xml", ".ttf", ".woff", ".woff2", ".less", ".sccs", ".csv", ".pdf" ],
					"AlloweEmpty":     "false",
					"IgnoreHosts":     [ "localhost", "127.0.0.1", "[::1]", "::1" ],
					"ReturnError":     "no"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Apr 22 12:46:06 MDT 2016 -- Tested only as a part of an entire server.  The automated test is still in the works.



', 'RejectHotlink-Reject-requests-based-on-invalid-referer-header-100036.html'
	, 'RejectHotlink: Reject requests based on invalid referer header' , 'Prevent access to images and other files if a valid referer header is not set.', '/doc-RejectHotlink-Reject-requests-based-on-invalid-referer-header', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

Allows for the banning of specific IP addresses.  If a matching IP address is found, then a
HTTP Status Forbidden (403) error will be returned.

Planned:  Adding ability to match ranges and sets of IP addresses. 

Also you can block based on geographic location using geoIPFilter.

Configuration
-------------

You can provide a simple list of IP addresses, either IPv4 or IPv6 addresses.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectIPAddress": { 
					"Paths":   "/api",
					"IPAddrs": [ "206.22.41.8", "206.22.41.9" ]
				} },
			...
	}
``` 

or you can provide a Redis prefix where a successful lookup will result in a
HTTP Status Forbidden (403) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectIPAddress": { 
					"Paths":            "/api",
					"RedisPrefix": 		"reject-ip|"
				} },
			...
	}
``` 

If both IPAddrs and RedisPrefix are provided, then an error will be logged and the RedisPrefix will be used.  
To apply to all paths use a "Paths" of "/".

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectIPAddress": { "LineNo":5, 
					"Paths":   "/api",
					"IPAddrs": [ "206.22.41.8", "206.22.41.9" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Thu Feb 25 12:37:05 MST 2016

### TODO

Add IP Ranges/Patterns: see /Users/corwin/Projects/IP/ip.go

', 'RejectIPAddress-Ban-Certain-IP-Address-100037.html'
	, 'RejectIPAddress: Ban Certain IP Address' , 'Prevent access to site based on IP address', '/doc-RejectIPAddress-Ban-Certain-IP-Address', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

If the path matches, then reject the requests.

Configuration
-------------

You can provide a simple list of paths to match.  Each match returns a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectPath": { 
					"Paths":   "/SrcCode"
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectPath": { 
					"Paths":   [ "/SrcCode", "/Tests" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 11:05:31 MST 2016


', 'RejectPath-Reject-Requests-Based-on-the-Path-100038.html'
	, 'RejectPath: Reject Requests Based on the Path' , 'Prevent access to a set of paths', '/doc-RejectPath-Reject-Requests-Based-on-the-Path', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

If the path matches - using a regular expression - then reject the requests.

Configuration
-------------

You can provide a simple list of paths to match.  Each match returns a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RejectRePath": { 
					"Paths":   [ "/.*/config" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RejectRePath": { 
					"Paths":   [ "^/.*/config$" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Fri Feb 26 11:26:41 MST 2016


', 'RejectRePath-Reject-Requests-Based-on-a-Regular-Expression-Path-Match-100039.html'
	, 'RejectRePath: Reject Requests Based on a Regular Expression Path Match' , 'Prevent access to paths based on a regular expression pattern match', '/doc-RejectRePath-Reject-Requests-Based-on-a-Regular-Expression-Path-Match', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '











Rewrite provides the ability to rewrite a URL with a new URL for later processing.

The rewrite uses a regular expression match for the URL.   The replacement allows substitution
of matched items into the resulting URL. 

If RestartAtTop is true, then the set of middleware is restarted from the very top with a re-parse
of parameters and rerunning of each of the middleware that preceded the Rewrite.  If it is false,
the processing continues with the next middleware.

A loop with RestartAtTop is limited to LoopLimit rewrites before it fails.  If RestartAtTop is 
true, then the rewritten URL should not match the regular expression.

Either way query parameters are re-parsed after the rewrite.

Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com", "http://localhost:8204/" ],
			"plugins":[
			...
				{ "Rewrite": { 
					"Paths":  [ "/api" ],
					"MatchReplace": [
						{ "Match": "http://(example.com)/(.*)\\?(.*)",
					      "Replace": "http://example.com/rw/process?${2}&name=${1}&${3}"
						}
					],
					"LoopLimit":     50, 
					"RestartAtTop":  true
				} },
			...
	}
``` 


Full Example
------------

``` JSON

	{
		"localhost-13004": { "LineNo":2,
			"listen_to":[ "http://localhost:13004" ],
			"plugins":[
				{ "DumpRequest": { "LineNo":6, "Msg":"Request Before Rewrite", "Paths":"/", "Final":"no" } },
				{ "Rewrite": { "LineNo":6, "Paths":"/",
						"MatchReplace": [
							{ "Match": "http://(localhost:[^/]*)/(.*)\\?(.*)",
							  "Replace": "http://localhost:13004/rw/${2}?rewriten_from=${1}&${3}"
							}
						]
				} },
				{ "DumpRequest": { "LineNo":10, "Msg":"Request After Rewrite", "Paths":"/", "Final":"no" } },
				{ "file_server": { "LineNo":11, "Root":"./www.test1.com", "Paths":"/"  } }
			]
		}
	}

``` 

Example with better regular expressions.  The previous regular expressions require a `?name=value` before 
matching.  This one is a little more realistic.

``` JSON
	{
		"working_test_ReactJS_15": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16020", "http://dev2.test1.com:16020" ],
			"plugins":[
				{ "HTML5Path": { "LineNo":__LINE__,
					"Paths":["(/.*\\.html)/.*"]
				} },
				{ "DumpRequest": { "LineNo":__LINE__, "Msg":"Request at top At Top", "Paths":"/api/status", "Final":"yes" } },
				{ "GoTemplate": { "LineNo":__LINE__,
					"Paths":["/api/table/p_document"],
					"TemplateParamName":     "__template__",
					"TemplateName":          "search-docs.tmpl",
					"TemplateLibraryName":   "./tmpl/library.tmpl",
					"TemplateRoot":          "./tmpl"
				} },
				{ "Rewrite": { 
					"Paths":  [ "/api/comments" ],
					"MatchReplace": [
						{ "Match": "http://([^/]*)/api/comments(\\?)?(.*)",
						  "Replace": "http://${1}/api/table/comments${2}${3}"
						}
					],
					"RestartAtTop":  false
				} },
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "www.go-ftl.com",
					"AppRoot": "/Users/corwin/Projects/www.go-ftl.com_doc/_site/data/",
					"StatusMessage":"Version 0.0.4 Sun May 22 19:12:43 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/Users/corwin/Projects/www.go-ftl.com_doc/_site", "Paths":"/"  } }
			]
		}
	}

```

### Tested

Tested On: Thu, Mar 10, 06:31:05 MST, 2016

Tested On: Sun, Mar 27, 11:48:58 MDT, 2016

### TODO

1. Match on the method also.  GET v.s. POST.   Allow replacement/alteration of method.




', 'Rewrite-Rewrite-One-Request-to-Another-Location-100040.html'
	, 'Rewrite: Rewrite One Request to Another Location' , 'Rewrite of request URLs', '/doc-Rewrite-Rewrite-One-Request-to-Another-Location', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

A single combed rewrite and proxy that allows access to a different server.

Configuration
-------------

You can provide a simple list of extensions that when matched will return a HTTP Not Found (404) error.

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "RewriteProxy": { 
					"Paths":   "/",
					"Extensions": [ ".cfg" ]
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "RewriteProxy": { 
					"Paths":   "/",
					"Extensions": [ ".cfg", ".password_db" ]
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Tested On: Thu, Mar 10, 19:42:36 MST, 2016

', 'RewriteProxy-Rewrite-Reqeust-and-Proxy-It-to-a-Different-Server-100041.html'
	, 'RewriteProxy: Rewrite Reqeust and Proxy It to a Different Server' , 'Combined rewrite of request and proxy request to a different server', '/doc-RewriteProxy-Rewrite-Reqeust-and-Proxy-It-to-a-Different-Server', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Limit serving of files to the specified set of extensions.  If the file is not one of the specified

This is a simple middleware that allows echoing of a request as JSON data.

This can be used as an end-point to test other items or as an "I am Alive" synthetic request.

Status is also useful for debugging the middleware stack.

This is one of a set of tools for looking into the middleware stack.
These include:

Middleware | Description
|--- | --- 
`DumpResponse` | Look at output from a request.  Can be placed at different points in the stack. 
`DumpReq` |   Look at what is in the request.  Can be placed at different points in the stack.
`Status` |   Send back to the client what was in the request.  It returns for all matched paths so it is normally used only once for each path.
`Echo` |   Echo a message to standard output when you reach this point in the stack.
`Logging` |   Log what the request/response are at this point in the stack.
`Else` |   A catch all for handing requests that do not have any name resolution.  It will, by default, list all of the available sites on a server by name.


Configuration
-------------

``` JSON
	{
		"servername": { 
			"listen_to":[ "http://www.example.com" ],
			"plugins":[
			...
				{ "Status": { 
					"Paths":   "/api",
				} },
			...
	}
``` 

Full Example
------------

``` JSON
	{
		 "www.zepher.com": { "LineNo":2,
			"listen_to":[ "http://www.zepher.com:3210/" ],
			"plugins":[
				{ "Status": { "LineNo":5, 
					"Paths":   "/api/status"
				} },
				{ "file_server": { "LineNo":9,
					"Root":"./www.zepher.com__3210",
					"Paths":"/"
				} }
			]
		}
	}
``` 


### Tested

Wed Mar  2 14:44:20 MST 2016

', 'Status-Echo-a-Request-as-JSON-Data-100042.html'
	, 'Status: Echo a Request as JSON Data' , 'Output request in JSON format to aid in debugging middleware stack', '/doc-Status-Echo-a-Request-as-JSON-Data', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Go-FTL is a scalable server and forward proxy designed for development and scalable deployment of web applications.

This documentation is being served with a Go-FTL server with a single custom middleware (it''s open source) that allows search engines to correctly index the documentation.
The documentation is a single page application written with jQuery and Bootstrap.   By reading this you are helping to make the authors of Go-FTL happy - you are using
the server that they wrote!  So... Give yourself a pat on the back for a good dead that you have done today!

Major Features
--------------

Features in detail are listed in a separate document.

1. Strong authentication.  The strong authentication combines SRP with AES so that the server never has the user''s password - but it is verified and all of the RESTful calls are 256-bit AES encrypted.  Two Factor Authentication (2fa) is a part of this.  Clients for iOS/iPhone, Android and other mobile platforms are provided.   A full example in Angular 1.x is provided.  Other examples in Angular 2.0, jQuery and React are in the works.  This is a *full* example including Login, Logout, Password Validation via Email, Lost Password recovery, password change etc.

2. Configuration based RESTful database server (TabServer2).  This allows complete applications to be built on the front end without code changes in the Go server.   Usually 90% or more of the application can be built with some simple configuration of the TabServer2.  Specialized business logic can be added by writing stored procedures in the database and exposing these as RESTful calls with simple additional configuration.  A full e-commerce system system has been built this way.  A performance tracking system was ported from T-SQL (Microsoft''s database) and a PHP back end to this with conversion of all data in a single day.   The system has security backed right in.
Emails use a templating system and are fully configurable.

3. Detailed tracing. One of my ongoing frustration with servers is that they are black boxes that either work or don''t.   You have no way to debug them.   With Go-FTL the exaxt opposite is true.  The tracing package allows you to see:
<div>
<ul>
<li> What middleware matched a request </li>
<li> What actions where taken </li>
<li> What errors may have occurred </li>
<li> Warnings if any along the way </li>
<li> How long it took for each section of code to run </li>
<li> If you are using the built-in micro-services - then what happens in them (and this is extensible to your code also) </li>
<li> With TabServer2 - what tables in the database were accessed, the queries/updates that were built, the bind VA rabbles to the queries, the data returned and how long it took the database to do this </li>
<li> With strong authentication when/and/if the user authenticated and what happed under the covers in the authentication process </li>
<li> With the file server - how the file was resolved to a final file and what file got server </li>
<li> Tracing of what is in the cache and why a request either did or did not get satisfied by the cache </li>
</ul>
</div>

4. Name based resolution of server configuration.  This allows for multiple virtual clients on a single server.  When you read this the server is on a dingle machine with at least 5 other name-resolved sites - all running in Go-FTL.

5. HTTP 2.0 support.  Much faster than HTTP 1.1 and supported by most browsers.  HTTP 2.0 provides a significant boost to performance.

6. Socket.IO and Go combine to allow pushing of content to clients.  This layer on top of web-sockets provides a cross-browser/cross-platform way of full bi-directional communication with a browser.  If you try the chat example it is built using this.  The tracing uses this to push content from the serer to a browser.

7. Server Farm Ready.  Instead of saving context information inside the
server it is always saved in Redis.  This includes session states.  This means that you can run Go-FTL on more than one machine with the same 
configuration and it will work.

8. Written in Go (golang) form performance, ease of modification and stability.  Realistic examples of using a Go Server to work with Angular 1.x, 2.0, React and other front end system.

9. Logging with Logrus.  This allows your server logs to interact with a myriad of other systems.

10. A full featured file server that supports dependency analysis.   For example, if you have TypeScript (.ts) files that need to be compiled into JavaScript (.js) the server can take a request for a .js file and use the TypeScript compiler to build the .js file on the fly.  It checks to see if the file needs to be rebuilt and will only build the .js file when the .ts has changed.   The results of this can be cached in the caching layer.   The file server supports an file-system based inheritance system allowing single page applications to be fully themed.

Planned Features
----------------

1. Lots of improvements to documentation.
1. Additional examples in React, React Native and Angular 2.0.
2. Integration with Let''s Encrypt using the Lego library - Automatic updates of HTTPS/TLS certificates.
3. Configuration changes on-the-fly without a server restart.  A complete server management interface for use with multiple servers in a data center.
4. Improvements to Caching
5. A more advanced password-less strong authentication system.
6. Payed hosting for instant spin up of Go-FTL.









', 'Overview-100043.html'
	, 'Overview' , 'Go-FTL is a scalable server and forward proxy designed for development and scalable deployment of web applications.', '/doc-Overview', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












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
<li> GoTemplate: Template using Go''s Buit in Templates </li>
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

', 'Overview-Features-in-Detail-100044.html'
	, 'Overview - Features in Detail' , 'Go-FTL Features in Detail', '/doc-Overview-Features-in-Detail', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













Please Help!

Yes I would like to have the help of others in building this server.  


A good place to start is with some sort of custom middleware.   This is a partial list.
If you have needs for other capabilities/middleware then let me know so I can update this list.
Also if you see something on this list that you need then let me know.  I have put the list
in some sort of order based on what I perceive as the most important to least important.
Requests from multiple people will help move items up the list.

If you are going to work on something - let me know.  Generally it is better to have only
one person working on one thing at a time.



Item  | Difficulty       | Description
:---: | ---------------- | -----------------------------------------------------------------------------------------------------------------------------
OAuth2 | Very Hard       | Integration with an OAuth2 server so the AesSrp two factor authentication and be used by 3rd party applications.  There are some Authentication servers written in Go that look very promising.  Take one of them and tie it to this for logging in.  [OSIN](https://github.com/RangelReale/osin) might be a possibility.  There are others.   Other systems of authentication besides OAuth2 might be useful.
Origin | Moderately Easy | Middleware that supports the "Origin"/"Access-Control-Allow-Origin" header and configurable allowed origins.  This would include in PostgreSQL a table of allowed origins.




Working on defects is hard.  If you see a defect that you want to fix please get in contact with me first. 
I may have some ideas of how it needs to be fixed and what kind of effort would be required.

Remember that all contributions are welcome.  If you contribute but you name is not listed below then let 
me know and I will add it.

Remember that this server is primarily MIT licensed.  Future contributions should also have that license.
Also, be prepared to support any middleware that you write.  Defects will need to be addressed.  Questions
will need to be answered.







Credit Where Credit is Due
--------------------------

If you contribute to this project we will make our best effort to give you the credit.
This is not just code.  Blog posts, utilities, documentation, tests, defects and 
most other activities that add to this will get you credit.

Emoji key
---------

Emoji | Represents | Links to
:---: | --- | ---
💻 | Code | `https://github.com/${ownerName}/${repoName}/commits?author=${username}`
🔌 | Plugin/utility libraries | the repo home
🔧 | Tools | the repo home
📖 | Documentation
❓ | Answering Questions (in Issues, Stack Overflow, Gitter, Slack, etc.)
⚠️ | Tests | `https://github.com/${ownerName}/${repoName}/commits?author=${username}`
🐛 | Bug reports | `https://github.com/${ownerName}/${repoName}/issues?q=author%3A${username}`
💡 | Examples | the examples
📝 | Blogposts | the blogpost
✅ | Tutorials | the tutorial
📹 | Videos | the video
📢 | Talks | the slides/recording/repo/etc.

All Contributors
----------------

Thanks goes to these wonderful people ([emoji key](https://github.com/kentcdodds/all-contributors#emoji-key)):

Contributor | Contributions
:---: | :---:
[![Philip J. Schlump](https://avatars2.githubusercontent.com/u/543809?v=3&s=130)<br />Philip J. Schlump](http://www.pschlump.com) | [📖](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [💻](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [🔌](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [⚠️ ](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [💡](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [🔧](https://github.com/pschlump/Go-FTL/commits?author=pschlump)
[![Chantelle R. Schlump](https://avatars2.githubusercontent.com/u/543809?v=3&s=130)<br />Chantelle R. Schlump](http://www.crs-studio.com) | [📖](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [💻](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [🔌](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [⚠️ ](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [💡](https://github.com/pschlump/Go-FTL/commits?author=pschlump) [🔧](https://github.com/pschlump/Go-FTL/commits?author=pschlump)
[![Kent C. Dodds](https://avatars1.githubusercontent.com/u/1500684?s=130)<br />Kent C. Dodds](http://kentcdodds.com) | [📖](https://github.com/kentcdodds/all-contributors/commits?author=kentcdodds)
[![Divjot Singh](https://avatars1.githubusercontent.com/u/6177621?s=130)<br />Divjot Singh](http://bogas04.github.io) | [📖](https://github.com/kentcdodds/all-contributors/commits?author=bogas04)


This project follows the [all-contributors](https://github.com/kentcdodds/all-contributors) specification.
Contributions of any kind welcome!

## LICENSE

MIT


', 'Contributing-To-This-Project-100045.html'
	, 'Contributing To This Project' , 'Requirements/Opportunities to become a contributor to this project', '/doc-Contributing-To-This-Project', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












./doc/About.md

Go-FTL is a web server and forward proxy.  It is licensed so that you can use it in embedded systems.  It is written in Go (golang) so that it is easy to modify and easy to deploy.
The ultimate goal is a full stack web-development environment that makes it easy to build and maintain web applications that use a database on the back end.

Go-FTL was started in 2015 but uses code that has been under development for a decade or more.


', 'About-100046.html'
	, 'About' , 'Go-FTL how it came to be', '/doc-About', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













Go-FTL configuration is in JSON files.  When you run the server by default it will look for two files, `ftl-config.json` and `global-config.json`.

`global-config.json` has global configuration in it like the name/type of the server being run and the connection information for how to authenticate
with Redis and PostgreSQL.  You can set a different global configuration file with the `-g` or `--globalCfgFile` command line option.

`ftl-config.json` is the per-server configuration file.  This has a section in it for each named server that will be run.  You can set the file name
with `-c` or `--cdgFile` option.

For Example, the a `glboal-config.json` file that I use has:

``` json

	{
		"debug_flags": [ "server" ],
		"trace_flags": [ "*" ],
		"server_name": "Go-FTL (v 0.5.9)",
		"RedisConnectHost":  "192.168.0.133",
		"RedisConnectAuth":  "lLJSwkwww3e24wAbr4RM4MWIaBM",
		"PGConn": "user=pschlump password=803728230121123 sslmode=disable dbname=pschlump port=5433 host=127.0.0.1",
		"DBType": "postgres",
		"DBName": "pschlump",
		"LoggingConfig": {
			"FileOn": "yes",
			"RedisOn": "yes"
		}
	}

```

Extensive examples of how to configure `ftl-config.json` are in the next section.  Some middleware components will have additional
configuration files, however most of the configuration is in `ftl-config.json`.


TODO
----

The plan is to allow the sections in `ftl-config.json` to be changed on the fly with a web interface.  That is still under development.



File: ./doc/Config.md 
', 'Config-Server-Configuraiton-100047.html'
	, 'Config - Server Configuraiton' , 'Go-FTL Configure a set of servers', '/doc-Config-Server-Configuraiton', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













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
<div class="doc-title">GoTemplate: Template using Go''s Buit in Templates</div>
<div class="doc-subtitle"> Use Go Templates to format data</div>
</a>
</li>
<li>
<a href="http://www.go-ftl.com/boot-docs.html/doc-Gzip-Ban-Certain-IP-Addresses">
<div class="doc-title">Gzip: Ban Certain IP Addresses</div>
<div class="doc-subtitle"> Gzip compresses output before it is returned.  Interacts with caching so ''zip'' process only happens if file changed</div>
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
', 'Config-Per-Server-Configuraiton-100048.html'
	, 'Config - Per Server  Configuraiton' , 'Go-FTL Configure the set of middleware that the server will use.', '/doc-Config-Per-Server-Configuraiton', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Go-FTL has extensive middleware to dedicated to database access.  PostgreSQL is the intended relational database however it is possible to use Oracle, Microsoft T-SQL, or MySQL.
Support for MySQL is still under development. 

The tab at the top, `TabServer` is all about configuration for database access.

A Quick Example.  You want to create a contact form:  First create the table in the database:

``` sql

	create table "p_issue" (
		  "id"						char varying (40) DEFAULT uuid_generate_v4() not null primary key
		, "title"					char varying (250)
		, "desc"					text		
		, "type_group"				char varying (50)		-- webpage / product name etc // [ Notification - ask for help ]
	);

```

Now create a TabServer configuration file with:

``` json

	{
		 "/api/table/contact": { "crud": [ "select", "insert", "update", "delete", "info" ]
			, "Comment": "Save Contact Requests"
			, "TableName": "p_issue"
			, "LineNo":"__LINE__, File:__FILE__"
			, "LoginRequired":false
			, "Method":["GET","POST","PUT","DELETE"]
			, "ReturnMeta":false
			, "ReturnAsHash":true
			, "cols": [
				  { "colName": "id"    		 , "colType": "s",	               "insert":true, "autoGen":true, "isPk":true 							}
				, { "colName": "title"	 	 , "colType": "s",	"update":true, "insert":true						, "DataColName":"subject"		}
				, { "colName": "desc"		 , "colType": "s",	"update":true, "insert":true														}
				, { "colName": "type_group"	 , "colType": "s",	"update":true, "insert":true, "DefaultData":"contact"					 			}
				]
			}
	}

```

We have turned off login on this with, `"LoginRequired":false`.  The API end point is `/api/table/contact`.  It will respond  to 

* GET requests to perform select
* POST requests to do insert
* PUT requests to do updates
* DELETE requests to do deletes

Add in a configuration section in the middleware to use TabServer and...

``` json

	{
		"working_test_AngularJS_20": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16020", "http://dev2.test1.com:16020" ],
			"plugins":[
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "www.go-ftl.com",
					"AppRoot": "/Users/corwin/Projects/docs-Go-FTL/data/",
					"StatusMessage":"Version 0.0.4 Sun May 22 19:12:43 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/Users/corwin/Projects/docs-Go-FTL", "Paths":"/"  } }
			]
		}
	}

```

You now have a working API.

![Database Output](./images/db1.png "Image showing output of API request.")

Things To Note
--------------

* Security is baked in.  Use the AesSrp module to provide strong authentication.
* A separate tracing package provides details of what happens with each request and how they get processed into queries.  It is incredibly useful for debugging your front end.
* You can also access Redis with the TabServer middleware.
* You can build complex data to return with joins.
* You can do full word searches and hierarchal/tree searches.
* Deletes can be configured to update the row with a flag as deleted data.
* You can call stored procedures as easily as performing queries.  This means that you can create complex business logic very quickly.
* You can cache tables or rows in Redis.
* The default format for data to be returned is JSON.  However this works with the GoTempalte middleware so that you can take a query and apply a template to the result.  This makes for some quick HTML results!
* All of this can be cached so that if the data has not changed then the cache can return a reasonable result.
* This is *NOT* an ORM.  This is a secure way to provide front end access and configuration based database access to your application.  This means that if you need to build a front end that updates 500,000 rows at a time it is easy to do and you don''t need to end up with 500,000 update statements.
* You can (and should) validate the input at the level of the TabServer.

File: ./doc/Database_01.md




', 'Database-Features-in-Detail-100049.html'
	, 'Database - Features in Detail' , 'Go-FTL Database Access', '/doc-Database-Features-in-Detail', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Currently the only way to download the server is in source and to build it.   This will change in a few days.

## Source Download

To download and install

```bash

	$ git clone https://github.com/pschlup/Go-FTL.git
	$ cd Go-FTL/server/goftl
	$ go get 
	$ go build

```

## Recompiled Binaries 

You can also download pre-compiled binaries for Mac OS X, Win 7 and Ubuntu Linux 14.04 - See "Download Compiled" on the left hand menu.



<!-- File: ./doc/Downloads.md -->



', 'Download-Source-100050.html'
	, 'Download - Source' , 'Go-FTL Download', '/doc-Download-Source', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












Biaries have been built with Go 1.6.2.

4 executables are included.  `goftl` is the server.

For Mac OS 10.9.5 <a href="http://www.go-ftl.com/download/goftl.MacOS.10.10.5_ver0.5.9.tar.gz">Download .tar.gz</a> file and extract.

```bash

	$ wget http://www.go-ftl.com/download/goftl.MacOS.10.10.5_ver0.5.9.tar.gz
	$ tar -xzf goftl.MacOS.10.10.5_ver0.5.9.tar.gz

```


For Linux Ubuntu 14.04 <a href="http://www.go-ftl.com/download/goftl.Win7_ver0.5.9.tar.gz">Download .tar.gz</a> file and extract.

```bash

	$ wget http://www.go-ftl.com/download/goftl.Win7_ver0.5.9.tar.gz
	$ tar -xzf goftl.Win7_ver0.5.9.tar.gz

```


For Windows 7+ <a href="http://www.go-ftl.com/download/goftl_Linux_ubuntu_14.04_ver0.5.9.tar.gz">Download .tar.gz</a> file and extract.

```bash

	$ wget http://www.go-ftl.com/download/goftl_Linux_ubuntu_14.04_ver0.5.9.tar.gz
	$ tar -xzf goftl_Linux_ubuntu_14.04_ver0.5.9.tar.gz

```

ARM for Raspberry Pi is in the works.


<!-- File: ./doc/DownloadsBinar.md -->


', 'Download-Compiled-Binaries-100051.html'
	, 'Download - Compiled Binaries' , 'Go-FTL Download', '/doc-Download-Compiled-Binaries', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













Yes you can get help with Go-FTL.  If you submit an issue - include your email address so that I can contact you.
I will look into all issues.




', 'Help-100052.html'
	, 'Help' , 'Go-FTL get help with this', '/doc-Help', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













Reprot Issue Go-FTL



', 'Report-Issue-Website-100053.html'
	, 'Report Issue Website' , 'Go-FTL get help -- contact the authors', '/doc-Report-Issue-Website', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













Contact Us



', 'Contact-Us-100054.html'
	, 'Contact Us' , 'Go-FTL get help -- contact the authors', '/doc-Contact-Us', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












License for Go-FTL
------------------

The MIT License (MIT)

Copyright (C) 2015 Philip Schlump, 2014-Present

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.


Licenses Used by Libraries Go-FTL Uses
--------------------------------------

I belove that all the used license are compatible with using Go-FTL as server embedded in a product.

Package                                     | License               | Description
|----------------------------------------   | --------------------- | -------------
"github.com/mgutz/ansi"                     | MIT                   | Out put command line information with colored text
"github.com/mediocregopher/radix.v2"        | MIT                   | Interface library to Redis
"github.com/mitchellh/mapstructure"         | MIT                   | Take data from a map to a structured
"github.com/oleiade/reflections"            | MIT                   | Extend reflection of data to make it easier to use
"github.com/oschwald/maxminddb-golang"      | ISC Licnese           | More permissive than MIT - use for geo-locaiton database.
"github.com/pschlump/filelib"               | MIT                   | Misc. File related functions
"github.com/pschlump/godebug"               | MIT                   | Debugging and tracing
"github.com/pschlump/gosrp"                 | MIT                   | SRP for GO
"github.com/pschlump/gosrp/big"             | Go-Source-License     | Big number implementation - modified for SRP
"github.com/pschlump/json"                  | Go-Source-License     | Extended JSON for Go-SRP
"github.com/pschlump/sizlib"                | MIT                   | Misc Library functions
"github.com/pschlump/templatestrings"       | MIT                   | Go Templates as string operations
"github.com/pschlump/uuid"                  | MIT                   | Generate UUID - faster than original
"github.com/pschlump/verhoeff_algorithm"    | MIT                   | Generate and check checksums for strings of digits
"github.com/russross/blackfriday"           | BSD-2-Clause          | Convert Markdown to HTML
"github.com/tdewolff/minify"                | MIT                   | Pack HTML, JS, CSS, etc.
"golang.org/x/crypto/pbkdf2"                | MIT                   | Encrypt passwords 
"github.com/Sirupsen/logrus"                | MIT                   | Logging to many destinations
"github.com/gorilla/websocket"				| MIT					| Used by Socket.IO - WebSockets
"github.com/shurcooL/sanitized_anchor_name"	| MIT					| See README.md in project - cleans anchor names
"github.com/mgutz/ansi"						| MIT					| Ansi terminal colored output
"github.com/hhkbp2/go-strftime"				| BSD-2-Clause			| Formatting of date/time
"github.com/lib/pq"							| MIT					| Database interface for PostgreSQL

It is simple to build a custom version of the code that includes or removes a certain
middleware.  The only file you need to change is ./server/goftl/inc.go.  It includes
each of the middleware files.

((My apologies if I left somebody out!))


', 'Used-License-100055.html'
	, 'Used License' , 'Go-FTL how it came to be', '/doc-Used-License', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '













The configuration file is in JSON.  It is very easy to configure Go-FTL.   This file has a set of progressivly
more complex examples in it.

## Example 1 - Simple - Listen for one machine

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

Listen on `localhost` port 8080 and on `http://dev2.test1.com:8080` for requests and serve them.   The directory with the files to server is ./www.


## Example 2 - Simple - Listen for one machine and gzip the data being sent back.

``` json

	{
		"http://localhost:8080/": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:8080", "http://dev2.test1.com:8080" ],
			"plugins":[
				{ "Gzip": { "LineNo":__LINE__, 
					"Paths":   "/www/static",
					"MinLength": 500
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"/www", "Paths":"/"  } }
			]
		}
	}

```

Listen on localhost port  8080 for requests and serve them.   The directory with the files to server is ./www.
Add the middleware "gzip" with a 500 byte minimum.  It will gzip any data returned (if the client accepts
gzip) that is over 500 bytes in size.

This shows how to pipe results from the `file_server` through another layer `gzip` before it is returned.

## Example 3 - Simple - Listen for both http and https requests.

``` json

	"demo_server": {
		"listento": [ "http://localhost:8080/", "https://localhost:8081" ],
		"certs": [ "/home/pschlump/certs/cert.pem", "/home/cpschlump/crts/key.pem" ],
		"root": "./www",
		"gzip": {
			"minsize": 500,
			"httponly": true
		}
	}

```

The "name" of the server is "demo_server".
List for http request on 8080 and for https on 8081.  Serve ./www.  Note the limitation on gzip as it
may be a security risk when combined with https.   The certificates are specified with the "certs" 
options. 

## Example 4 - Multiple name resolved servers.

``` json

	"demo_server": {
		"listento": [ "http://localhost:8080/", "https://localhost:8081" ],
		"certs": [ "/home/pschlump/certs/cert.pem", "/home/cpschlump/crts/key.pem" ],
		"root": "./www/demo_server",
		"gzip": {
			"minsize": 500,
			"httponly": true
		}
	}

	"test_server": {
		"listento": [ "http://test.2c-why.com/", "https://test.2c-why.com" ],
		"certs": [ "/home/pschlump/certs/test.2c-why.com/cert.pem", "/home/pschlump/crts/test.2c-why.com/key.pem" ],
		"root": "./www/test.2c-why.com",
		"gzip": {
			"minsize": 500,
			"httponly": true
		}
	}

```

Listen and server two different sets of pages.  The default ports are used for "test_server" with 80 for http
and 443 for https.








', 'Go-FTL-Configuration-100056.html'
	, 'Go-FTL Configuration' , 'Go-FTL configuration examples', '/doc-Go-FTL-Configuration', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












TabServer2 is a Go-FTL middleware that allows the creation of RESTful interfaces to database tables.

A configuration file in JSON determines which tables will have interfaces and what security will be
applied.   Validation can be specified for the parameters.  Parameters are always substituted using
bind variables, never directly into the SQL statements.  This improves efficiency and prevents most
forms of SQL injection attacks.

Stored procedures in the database can be called.  This allows for the creation of business logic at
the level of the database.  Since most data related business logic requires multiple database 
queries the most efficient place to put it is inside the database stored procedures.  Also this
data-centric processing is test and developed quicker when it is in database stored procedures.

Data can be post-joined to produce more complex results.  For example an invoice can be returned
with its invoice details in a single RESTful call.

Searches and updates can use complex where clauses.  The where criteria can be supplied as a parse 
tree from the front end.  The set of columns that can be used in the where can be limited so that
only indexed columns are accessed in the where.

Column names can differ from the named parameters supplied.  By default a 1 to 1 match is assumed.

Every input can be validated.

In PostgreSQL complex keys like document keyword searches and hierarchal searches are directly supported.


', 'TabServer2-A-Go-FTL-middleware-for-building-Restful-Interfaces-to-a-relational-database-100057.html'
	, 'TabServer2 - A Go-FTL middleware for building Restful Interfaces to a relational database' , 'TabServer2 - search for configuration files', '/doc-TabServer2-A-Go-FTL-middleware-for-building-Restful-Interfaces-to-a-relational-database', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '












TabServer searches for its configuration files using a search path.
They are named `sql-cfg[AppName].json` where `[AppName]` is the AppName in the configuration
and SearchPath is `~/cfg:./cfg:.` by default.  

`~` will be substituted with your home directory.  `~name/` is substituted for the home directory of the request user.

Examples:

``` gray-bar

		"AppName": "www.go-ftl.com",
		"AppRoot": "/Users/myuser/Projects/www.go-ftl.com/data/",

```

This will set the application name to `www.go-ftl.com` and search with a top level directory of:
`/Users/myuser/Projects/www.go-ftl.com/data/`.   It can find `sql-cfg-www.go-ftl.com.json`.

The search order is:

``` gray-bar

		"%{path_element%}/%{fileName%}-%{AppName%}-%{HostName%}%{fileExt%}",
		"%{path_element%}/%{fileName%}-%{AppName%}%{fileExt%}",
		"%{path_element%}/%{fileName%}-%{HostName%}%{fileExt%}",
		"%{path_element%}/%{fileName%}%{fileExt%}",

```

Where `path_element` is a path our of the `SearchPath` in the order supplied.

`fileName` is the `sql-cfg` section of the file name.

`AppName` is the specified application name.

`HostName` is the name of your computer.

`fileExt` is .json

You can create host-specific global-cfg.json files by putting them in your ~/cfg directory.
For example, you have `pschlump-dev1` and `pschlump-dev2` machines.  If you are not on one of
these then use the default file.

``` gray-bar

	~/cfg/global-cfg-pschlump-dev1.json
	~/cfg/global-cfg-pschlump-dev2.json
	~/cfg/global-cfg.json

```


', 'Search-Path-for-global-cfg-json-and-sql-cfg-json-100058.html'
	, 'Search Path for global-cfg.json and sql-cfg.json' , 'TabServer2 - search for configuration files', '/doc-Search-Path-for-global-cfg-json-and-sql-cfg-json', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '















## Create the table - or pic a table to use

Create tables in the database.  In this case we are going to create a table called `feedback` in PostgreSQL.
The file for this can be found in ./example/feedback.sql.

``` bash
	$ cd ./example
	$ psql -a -P pager=off -h 127.0.0.1 -U pschlump <feedback.sql
```


## How to serve a table

Now create the example `sql-cfg.json` file:

``` bash
	$ cd ../tools/genDefault
	$ go run ./main.go feedback >sample.out
	$ cat sample.out
```

should give you in sample.out

``` JSON
	{
		"note:generated":"Tables: [feedback]", 

		,"/api/table/feedback": { "crud": [ "select", "insert", "update", "delete", "info" ]
			, "Comment": "Generated by genDefault"
			, "TableName": "feedback"
			, "LineNo":"__LINE__"
			, "Method":["GET","POST","PUT","DELETE"]
			, "cols": [
				  { "colName": "id"      , "colType": "s",	               "insert":true, "autoGen":true, "isPk": true }
				, { "colName": "subject" , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "body"    , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "status"  , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "updated" , "colType": "t",	"update":true, "insert":true		}
				, { "colName": "created" , "colType": "t",	"update":true, "insert":true		}
				]
			}
	}
```

You may want to make the last 2 columns read only as they get generated in the database.  This gives you:

``` JSON
	{
		"note:generated":"Tables: [feedback]", 

		,"/api/table/feedback": { "crud": [ "select", "insert", "update", "delete", "info" ]
			, "Comment": "Generated by genDefault"
			, "TableName": "feedback"
			, "LineNo":"__LINE__"
			, "Method":["GET","POST","PUT","DELETE"]
			, "cols": [
				  { "colName": "id"      , "colType": "s",	               "insert":true, "autoGen":true, "isPk": true }
				, { "colName": "subject" , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "body"    , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "status"  , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "updated" , "colType": "t",                                      }
				, { "colName": "created" , "colType": "t",                                      }
				]
			}
	}
```

Copy the modified file into the `./test-sql-cfg/` as `sql-cfg.json` or modify an existing `sql-cfg.json` file and add the 
lines from `,"/api...` down to the matching `}`.

## Add to your config

This is a sample of the configuration file for the server.  The above file is put in ./test-sql-cfg/ as sql-cfg.json

``` JSON
	{
		"working_test_AngularJS_20": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16000" ],
			"plugins":[
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "www.2c-why.com",
					"AppRoot": "./test-sql-cfg/",
					"StatusMessage":"Version 0.0.1 Thu Apr 21 07:44:42 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"./quickstart", "Paths":"/"  } }
			]
		}
	}
```

Restart the server and then we can test it.

## test it...

To test it use a browser or some other tool that can fetch from localhost (wget, curl for example).
In the browser enter the URL `http://localhost:16000/api/table/feedback` to perform a GET on this end point.

![example output from fetching data](https://github.com/pschlump/TabServer2/raw/master/img/example.png "example output")



## turn on security for table if you need it/want it.

To require login before changing the data you will need to change `, "LoginRequired":false` to `, "LoginRequired":true`
as is shown below:

``` JSON
	{
		"note:generated":"Tables: [feedback]", 

		,"/api/table/feedback": { "crud": [ "select", "insert", "update", "delete", "info" ]
			, "Comment": "Generated by genDefault"
			, "TableName": "feedback"
			, "LineNo":"__LINE__"
			, "LoginRequired":true
			, "Method":["GET","POST","PUT","DELETE"]
			, "cols": [
				  { "colName": "id"      , "colType": "s",	               "insert":true, "autoGen":true, "isPk": true }
				, { "colName": "subject" , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "body"    , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "status"  , "colType": "s",	"update":true, "insert":true		}
				, { "colName": "updated" , "colType": "t",                                      }
				, { "colName": "created" , "colType": "t",                                      }
				]
			}
	}
```

And then include some form of login, for example the `SrpAesAuth` authentication:

``` JSON
	{
		"working_test_AngularJS_20": { "LineNo":__LINE__,
			"listen_to":[ "http://localhost:16000" ],
			"plugins":[
				{ "SrpAesAuth": { "LineNo":__LINE__,
					"Paths": "/api/" ,
					"MatchPaths": [ "/api/table/", "/api/list" ]
				} },
				{ "TabServer2": { "LineNo":__LINE__,
					"Paths":["/api/"],
					"AppName": "www.2c-why.com",
					"AppRoot": "./test-sql-cfg/",
					"StatusMessage":"Version 0.0.1 Thu Apr 21 07:44:42 MDT 2016"
				} },
				{ "file_server": { "LineNo":__LINE__, "Root":"./quickstart", "Paths":"/"  } }
			]
		}
	}
```

Then restart the server or have the server re-load the configuration file.  Login is now required to access
the table.  This means that you will have to authenticate, use two factor authentication and make requests using the
`/api/cipher` interface with fully encrypted messages.  Take a look at one of the authentication/login examples.


## data field names fail to match column names in the table

Suppose that you have data that is incoming called "subject" but the column in the table is called "title".
You can still match it up by using the `DataColName` option.

``` JSON
	...
	,"/api/table/contact": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "Comment": "Save Contact Requests"
		, "TableName": "p_issue"
		, "LineNo":"__LINE__"
		, "LoginRequired":false
		, "Method":["GET","POST","PUT","DELETE"]
		, "ReturnMeta":false
		, "ReturnAsHash":true
		, "cols": [
			  { "colName": "id"    		 , "colType": "s",	               "insert":true, "autoGen":true, "isPk":true 	}
			, { "colName": "title"	 	 , "colType": "s",	"update":true, "insert":true						, "DataColName":"subject"		}
			, { "colName": "desc"		 , "colType": "s",	"update":true, "insert":true								}
			, { "colName": "type_group"	 , "colType": "s",	"update":true, "insert":true, "default":"contact"			}
			]
		}
	...

```



', 'How-to-Serve-a-Single-Table-100059.html'
	, 'How to Serve a Single Table' , 'How to configure to serve and update data in a single table', '/doc-How-to-Serve-a-Single-Table', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '














## Overview

This configuration file allows limits and configures access to the database without
changes to code.   Most RESTful access can be implemented using this file.   You can

1. Determine what tables are to be accessed via the RESTful interface.
2. Required authentication - or not - on tables and stored procedure calls.
3. Limit the columns that are returned.
4. Create Validation for all inputs.
5. Have validation that is required for individual operations.  For example,
you can require that a primary key be supplied, and that it is a UUID/GUID
for performing a delete on a specific table.
6. Configure REST calls that will access stored procedures in the database.
7. Set tables/rows that are to be cached in Redis.
8. Make changes on the fly with a running TabServer2 server or set of servers.
Make these changes without a server restart.

The configuration file can be watched for changes and the TabServer2 server can
be automatically notified when the file changes.   This is very convenient 
during development.

## Example 001

``` JSON
	,"/api/perfTestDB1": { "query": "select * from posts", "p": [ ], "LoginRequired":false
		, "LineNo":"__LINE__"
		, "valid": {
			 "callback": { }
			}
		}
```

"/api/perfTestDB1" is the RESTful GET call.  

"query" is the SQL query that will be returned.

"p":[] indicates that there a NO parameters that will be substituted as bind 
variables in this call.

"LoginRequired":false indicates that no authentication is required to make this call.

"LineNo":"__LINE__" indicates that the line number will be reported if there is an error.
Two items can be substituted, __LINE__ and __FILE__.   LineNo is a string so you can have
both.  For example:  `"LineNo":"File: __FILE__ LineNo: __LINE__"`.

"valid":{...} is the validation for all POST, GET, PUT, DELETE calls.
Since the method was not specified it defaults to just GET.  This is 
appropriate for a database *select* query operation.

"valid":{"callback":{}"}"  The empty callback option indicates that
this can be called via JSONp.   This is by default a non-required
field.  Requests will only have access to fields that are specified
in one of the validation sections.   You can have method-specific
validation fields.  If "callback" had not been specified then JSONP
would not be allowed.   "validDel", "validGet", "validPut", "validPost"
override this section and match with the "DELETE", "GET", "PUT" and "POST"
methods.  There is no validation on "HEAD" requests.

"TableName" was NOT specified.  This means that TraceRest will not
know what table was accessed when this operation is called.






## Exampel 002

``` JSON
	,"/api/saveJs": { "Fx": "e_js_save"
		, "LineNo":"__LINE__"
		, "LineNo":"171"
		, "p": [ "user", "id", "desc", "title" ]
		, "Method":["POST"]
		, "valid": {
			 "user": { "required":true, "type":"string", "min_len":4, "max_len":100 }
			,"id": { "required":true, "type":"string", "max_len":40 }
			,"desc": { "required":false, "type":"string", "max_len":400 }
			,"title": { "required":false, "type":"string", "max_len":400 }
			,"auth_token": { "required":true, "type":"uuid" }
			}
		}
```

"f" specifies a query that calls a stored procedure - no return value is expected from the
stored procedure.

"p":["user","id","desc","title"] specifies the parameters that will be bound to the stored 
procedure.

"Method":["POST"]  Specifies that this call will only respond to "POST" calls.  This is
appropriate for a stored procedure that saves/updates data in the database.   Using a "GET" 
call (other that in a known development/test environment) when calling for a database
change is not an advised activity.

"LineNo":"171" hard codes the line number to 171 for some reason.  Usually better to use __LINE__ or __LINE__ and __FILE__.
 
"user":{"required":true,"type":"string","min_len":4,"max_len":100}  Sets the validation
for this to be a string, with a minimum length of for and not to exceed 100 chars.  It is
a required field.

"id":{"required":true,"type":"string","max_len":40}  Sets a required field.  It has
a maximum length of 40.

"desc":{"required":false,"type":"string","max_len":400} Sets an optional field.   An empty
string will be used if the field is not supplied.  Maximum length is 400.

"auth_token":{"required":true,"type":"uuid"} Is required for all authorization required requests.
Since "noauth":true has not been specified, this is an authorization required operation.
"auth_token" will be required and validated and the type of the "auth_token" is "uuid" or "u".





## Example 003

``` JSON
	,"/api/test/change_password": { "g": "change_password", "p": [ "password", "again", "token","$ip$" ]
		, "LineNo":"__LINE__"
		, "Method":["POST"]
		, "TableList":["t_user","t_ip_ban"]
		, "valid": {
			 "password": 	{ "required":true, "type":"string", "max_len":80, "min_len":4 }
			,"again":	 	{ "required":true, "type":"string", "max_len":80, "min_len":4 }
			,"token": 		{ "required":true, "type":"string", "max_len":40, "min_len":2 }
			,"$ip$": 		{ "required":true, "type":"string", "max_len":40, "min_len":4 }
			}
		, "CallAfter": ["CacheEUser"]
		}
```

"g" Call stored procedure named "change_password"

"p":["password","again","token","$ip$"]  call stored procedure with "password", "again", "token", and "$ip$".
"$ip$" is an injected value.  It is the IP address of the client making the change.    For all of the
injected values see the secion on parameter injection.

"TableList":["t_user","t_ip_ban"]  This stored procedure access the tables "t_user" and "t_ip_ban".

"LineNo":"__LINE__" indicates the line number for any errors.

"Method":["POST"]  Specifies that this call will only respond to "POST" calls.  

"valid" - all 4 values are required.  Length is specified for all of them.

"CallAfter":["CacheEUser"]  A function named CacheEUser is called in the  GO code after this stored
procedure runs.   See the section on *pre/post function calls* for the parameters and details on
this.




## Exampe 004 - Expose a table for selects and updates.  Require authentation.

``` JSON
	,"/api/table/t_email_track": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "LineNo":"__LINE__"
		, "p": [ ]
		, "TableName":"t_email_track"
		, "Method":["GET","POST","PUT","DELETE","HEAD"]
		, "cols": [
				  { "colName": "id" 			, "colType":"u"	,"isPk":true, "insert":true									, "autoGen":true					}
				, { "colName": "user_id" 		, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "auth_token" 	, "colType":"s"				, "insert":true	, "update":false													}
				, { "colName": "to" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "from" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "subject" 		, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "body" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "error" 			, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "status" 		, "colType":"s"				, "insert":true	, "update":true														}
				, { "colName": "ip"	 			, "colType":"s"				, "insert":true	, "update":true, "DataColName": "$ip$"								}
			]
		, "validDel": {
				 "id":		 			{ "required":false, "type":"string", "max_len":40, "min_len":2 }
				,"auth_token": 			{ "required":false, "type":"string", "max_len":40, "min_len":2 }
				,"$ip$": 				{ "required":true, "type":"string", "max_len":40, "min_len":4 }
			}
		, "validPost": {
				 "to": 					{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"subject": 			{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"from": 				{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"auth_token": 			{ "required":false, "type":"s", "max_len":40, "min_len":2 }
				,"user_id": 			{ "required":false, "type":"s", "max_len":40, "min_len":2 }
				,"$ip$": 				{ "required":true, "type":"s", "max_len":40, "min_len":4 }
			}
		, "validPut": {
				 "to": 					{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"subject": 			{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"from": 				{ "required":false, "type":"s", "min_len":2, "max_len":250 }
				,"auth_token": 			{ "required":false, "type":"s", "max_len":40, "min_len":2 }
				,"$ip$": 				{ "required":true, "type":"s", "max_len":40, "min_len":4 }
			}
		, "orderBy": [ { "colName": "4" } ]
		}
```

"/api/table/..." indicates that this is to be handed as a table request.  This means that the select/update/delete/insert
queries will be generated automatically.

"crud" sets the operations that are allowed on this table.

"crud":["select","insert","update","delete","info"]  Allows all operations on this table.  "info" is for HEAD requests.

"Method":["GET","POST","PUT","DELETE","HEAD"]  This sets the methods that are allowed.

"cols" sets the coluns that will be allowed/returned in operations on this table.  The table may have other columns
that are not visible to the outside world.  For example, creation_date, update_date columns are often not returned.

{"colName":"id","colType":"u","isPk":true,"insert":true,"autoGen":true}  Specifies the column named "id".  Its column type
is set with "colType" to "u" - this is a UUID.  It is marked as primary key (or unique key) with the "isPk":true.  "insert":true
says that this field can be inserted.  "autoGen":true says that TabServer2 should generate a UUID for this if it is not
supplied during an insert operaiton. The UUID will be returned.  UUIDs are generated by the TabServer2 server rather than
by the database.  This is so that as little load as possible is placed on the databse.  The *CLIENT* can generate the
UUID/GUID and send it with an insert.  

"update":false on "colName":"auth_token" indicates that this field can ot be updated after it has been created.

{"colName":"ip","colType":"s","insert":true,"update":true,"DataColName":"$ip$"}  The column name in the database is "ip".  The
parameter that is used in the REST call is "$ip$".  This is an injected value for the client IP address.

Validation is performed based on the method of the operation.

"orderBy":[{"colName":"4"}]  On GET/select requests the default sort order is column 4.  



## Exampel 005 -- Expose a table for selects.

This is a good example of what will be generated using the "app-discovery.html" tool.  A no-authentation, select
only configuration.  "app-discovery.html"  allows for quick discovery of tables in the database and generation of
the necessary configuration for them.

``` JSON
	,"/api/table/t_available_test_systems": { "crud": [ "select" ]
		, "TableName": "t_available_test_systems"
		, "LineNo":"__LINE__"
		, "Method":["GET"]
		, "cols": [
				  { "colName": "osNameClass" 			, "colType": "s" 			}
				, { "colName": "browserNameClass" 		, "colType": "s" 			}
				, { "colName": "osMajorClass" 			, "colType": "s" 			}
				, { "colName": "osMinorClass" 			, "colType": "s" 			}
				, { "colName": "browserMajorClass" 		, "colType": "s" 			}
				, { "colName": "browserMinorClass" 		, "colType": "s" 			}
				, { "colName": "browserName" 			, "colType": "s" 			}
				, { "colName": "title"					, "colType": "s" 			}
				, { "colName": "n_clients"	 			, "colType": "s" 			}
				, { "colName": "n_runs"	 				, "colType": "s" 			}
				, { "colName": "useragent_id"	 		, "colType": "s" 	, "NoSort":true		}
				, { "colName": "is_running_now"	 		, "colType": "s" 			}
			]
	}
```

"TableName":"t_available_test_systems"  name of the table being exposed.

"crud":["select"] The only operation allowed.

"Method":["GET"]  The only operation allowed.

"cols" the set of columns that will be retuned in the select.

No validation is specified.  Where clauses on any column in the "cols" are allowed.
Sorting by any of the columns, except "useragent_id", is allowed.



## Exampel 006 -- Validation of int fields.

``` JSON
	,"/api/saveOneNote": { "Fx": "e_save_data_func", "p": [ "url", "top", "left" ]
		, "LineNo":"__LINE__"
		, "valid": {
			 "url": { "required":true, "type":"s" }
			,"top": { "required":false, "type":"i", "min": -4000, "max": 4000 }
			,"left": { "required":false, "type":"i", "min": -4000, "max": 4000 }
			,"auth_token": { "required":true, "type":"u" }
			}
		}
```

"top":{"required":false,"type":"i","min":-4000,"max":4000}
the type is specified to be an integer, with the "type":"i".  Minimum and maximum integer values are set.
"min","max" can also be set for "f"/float data.

"Fx" is the name of a stored procedure to be called.   Data returned from the stored procedure is 
logged and discarded.  You can have a "Query" called after this.  If you need data retuned then use
the "G" instead.



# Exampel 007 -- Confirmation of return values 

``` JSON
	,"/api/table/tblDepartment": { "crud": [ "select", "insert", "update", "delete", "info" ]
		, "TableName": "tblDepartment"
		, "LineNo":"__LINE__"
		, "ReturnGetPKAsHash": true
		, "Method":["GET","POST","PUT","DELETE"]
		, "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "Absent":"1", "Present":"0" }
		, "CustomerIdPart": { "colType":"u", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "name"				, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "description"		, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "isDeleted"			, "colType": "i",	"update":true, "insert":true		}
			]
		, "orderBy": [ { "colName": "2" } ]
		}
```

"ReturnGetPKAsHash":true Sets the return value when the /api/table/tblDepartment/&lt;PK&gt; call is made.   The default
is to return an array 1 long with the data as a has in the array.   Setting this to true just returns the hash
without the array.

"deleteViaUpdate":{"colType":"i","colName":"isDeleted","Absent":"1","Present":"0"}  Sets that deletes are 
handed via performing an update on the row.  The data tyep for the update is an integer, "i".  The column
name is `isDeleted`.   The row is deleted when `isDelted` is set to 1, and not-delted when set to 0.  You can
use strings, booleans and integers for deleted flags.

"CustomerIdPart":{"colType":"u","colName":"customer_id"}  This table is partitioned by customer_id.   This
allows for multi-customer in a singe table data.   The customer id is a UUID/GUID in this case.   If you are
using Oracle the partitioning should be an integer that is generated.  If you are using Microsoft SQL Server
you should not use UUID/GUID for partitioning.  


# Exampel 008 -- Post Join

``` JSON
	,"/api/table/p_cart:GET": { "crud": [ "select" ]
		, "TableName": "p_cart"
		, "LineNo":"__LINE__"
		, "Method":["GET"]
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "state"				, "colType": "s",	}
				, { "colName": "marked"				, "colType": "s",	}
				, { "colName": "user_id"			, "colType": "s",	}
				, { "colName": "cart_name"			, "colType": "s",	}
				, { "colName": "pagecookie"			, "colType": "s" 	}
				, { "colName": "total"				, "colType": "f" 	}
			]
		, "orderBy": [ { "colName": "1" } ]
		, "PostJoin": [
					{ "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "cartItemsOld"
						, "Query": "SELECT \"p_cart_item\".\"id\" as \"item_id\", \"p_cart_item\".\"product_id\", \"p_cart_item\".\"product_inventory_id\", \"p_cart_item\".\"n_in_cart\", \"p_cart_item\".\"state\", \"p_product\".\"prod_name\", \"p_product\".\"desc\", \"p_product\".\"state\", \"p_product\".\"SKU\" , \"p_cart_item\".\"total\", \"p_cart_item\".\"ex_total\", \"p_cart_item\".\"options\" FROM \"p_cart_item\" as \"p_cart_item\" left join \"p_product\" as \"p_product\" on \"p_cart_item\".\"product_id\" = \"p_product\".\"id\" WHERE \"p_cart_item\".\"cart_id\" = $1 "
					}
					, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "imageList"
						, "Query": "SELECT \"p_image_list\".\"id\", \"p_image_list\".\"image_id\", \"p_image_list\".\"seq_no\", \"p_image\".\"file_name\", \"p_image\".\"base_file_name\", \"p_image\".\"h_size\", \"p_image\".\"w_size\", \"p_image\".\"f_size\", \"p_image\".\"img_type\", \"p_cart_item\".\"product_id\" FROM \"p_cart_item\" as \"p_cart_item\" left join \"p_image_list\" as \"p_image_list\" on ( \"p_cart_item\".\"product_id\" = \"p_image_list\".\"fk_id\" ) left join \"p_image\" as \"p_image\" on ( \"p_image_list\".\"image_id\" = \"p_image\".\"id\" ) WHERE \"p_cart_item\".\"cart_id\" = $1 and \"p_image_list\".\"id\" is not null "
					}
					, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "cartItems"
						, "Query": "SELECT * from get_cart_items ( $1 )"
					}
				]
		}
	,"/api/table/p_cart": { "crud": [ "insert", "update", "delete" ]
		, "TableName": "p_cart"
		, "LineNo":"__LINE__"
		, "LoginRequired":true
		, "Method":["POST","PUT","DELETE","HEAD"]
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "state"				, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "marked"				, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "user_id"			, "colType": "s",	"update":true, "insert":true		}
				, { "colName": "cart_name"			, "colType": "s",	"update":true, "insert":true		}
			]
		}
```

# Exampel 009 -- Templates for Queries

``` JSON
	,"/api/table/tblActionPlan:GET": { "crud": [ "select" ]
		, "TableName": "tblActionPlan"
		, "TableList":[ "tblActionPlan", "tblPerson" ]
		, "LineNo":"__LINE__"
		, "ReturnGetPKAsHash": true
		, "Method":["GET"]
		, "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "ColAlias":"tblActionPlan", "Absent":"1", "Present":"0" }
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id", "ColAlias":"tblActionPlan" }
		, "cols": [
				  { "colName": "id" 					, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "cardId" 				, "colType": "s", 	}
				, { "colName": "sequence" 				, "colType": "i", 	}
				, { "colName": "actionPlan" 			, "colType": "s", 	}
				, { "colName": "dateEntered" 			, "colType": "d" 	}
				, { "colName": "targetCompletion" 		, "colType": "d", 	}
				, { "colName": "responsiblePersonId" 	, "colType": "s", 	}
				, { "colName": "notes" 					, "colType": "s", 	}
				, { "colName": "actionCompleted" 		, "colType": "d", 	}
				, { "colName": "isDeleted"	 			, "colType": "i" 	}
				, { "colName": "firstName" 				, "colType": "s" 	}
				, { "colName": "lastName" 				, "colType": "s" 	}
				, { "colName": "email" 					, "colType": "s"	}
				, { "colName": "phone" 					, "colType": "s" 	}
			]
		, "orderBy": [ { "colName": "3" } ]
		, "SetWhereAlias":"tblActionPlan"
		, "SelectPK1Tmpl": " SELECT \"tblActionPlan\".\"id\" ,\"tblActionPlan\".\"cardId\" ,\"tblActionPlan\".\"sequence\" ,\"tblActionPlan\".\"actionPlan\" ,\"tblActionPlan\".\"dateEntered\" ,\"tblActionPlan\".\"targetCompletion\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"notes\" ,\"tblActionPlan\".\"actionCompleted\" ,\"tblActionPlan\".\"isDeleted\" ,\"tblPerson\".\"firstName\" ,\"tblPerson\".\"lastName\" ,\"tblPerson\".\"email\" ,\"tblPerson\".\"phone\" FROM \"tblActionPlan\" as \"tblActionPlan\" left join \"tblPerson\" as \"tblPerson\" on \"tblActionPlan\".\"responsiblePersonId\" = \"tblPerson\".\"id\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
		, "SelectTmpl": " SELECT \"tblActionPlan\".\"id\" ,\"tblActionPlan\".\"cardId\" ,\"tblActionPlan\".\"sequence\" ,\"tblActionPlan\".\"actionPlan\" ,\"tblActionPlan\".\"dateEntered\" ,\"tblActionPlan\".\"targetCompletion\" ,\"tblActionPlan\".\"responsiblePersonId\" ,\"tblActionPlan\".\"notes\" ,\"tblActionPlan\".\"actionCompleted\" ,\"tblActionPlan\".\"isDeleted\" ,\"tblPerson\".\"firstName\" ,\"tblPerson\".\"lastName\" ,\"tblPerson\".\"email\" ,\"tblPerson\".\"phone\" FROM \"tblActionPlan\" as \"tblActionPlan\" left join \"tblPerson\" as \"tblPerson\" on \"tblActionPlan\".\"responsiblePersonId\" = \"tblPerson\".\"id\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
		}
	,"/api/table/tblActionPlan": { "crud": [ "insert", "update", "delete" ]
		, "TableName": "tblActionPlan"
		, "LineNo":"__LINE__"
		, "ReturnGetPKAsHash": true
		, "Method":["POST","PUT","DELETE","HEAD"]
		, "deleteViaUpdate": { "colType":"i", "colName":"isDeleted", "Absent":"1", "Present":"0" }
		, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
		, "cols": [
				  { "colName": "id" 					, "colType": "s" 			, "autoGen": true , "isPk": true }
				, { "colName": "cardId" 				, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "sequence" 				, "colType": "i", 	"update":true, "insert":true		}
				, { "colName": "actionPlan" 			, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "dateEntered" 			, "colType": "d" 	}
				, { "colName": "targetCompletion" 		, "colType": "d", 	"update":true, "insert":true		}
				, { "colName": "responsiblePersonId" 	, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "notes" 					, "colType": "s", 	"update":true, "insert":true		}
				, { "colName": "actionCompleted" 		, "colType": "d", 	"update":true, "insert":true		}
				, { "colName": "isDeleted"	 			, "colType": "i" 	}
			]
		, "orderBy": [ { "colName": "3" } ]
		}
```

# Exampel 010 -- Using PostgreSQL keyword search facility

``` JSON
	{
	  "note:comment": { "f": "(C) Philip Schlump, 2009-2015." }
	, "note:version": { "f": "v1.0.2" }

		,"/api/table/x_product:GET": { "crud": [ "select" ]
			, "TableName": "x_product"
			, "LineNo":"__FILE__ : __LINE__"
			, "ReturnGetPKAsHash": true
			, "ReturnMeta": true
			, "TableList":[ "x_product", "x_attr", "x_product_inventory", "x_product_options", "x_product_options_meta" ]
			, "Method":["GET"]
			, "valid": {
				 "$customer_id$": { "required":true, "type":"s" }
				,"callback": { }
				}
			, "CustomerIdPart": { "colType":"s", "colName":"customer_id" }
			, "cols": [
					  { "colName": "id" 				, "colType": "s" 			, "autoGen": true , "isPk": true }
					, { "colName": "customer_id"		, "colType": "s" }
					, { "colName": "prod_name"			, "colType": "s" }
					, { "colName": "desc"				, "colType": "s" }
					, { "colName": "cart_tmpl"			, "colType": "s" }
					, { "colName": "state"				, "colType": "s" }
					, { "colName": "limit_per_cust"		, "colType": "s" }
					, { "colName": "min_count_of"		, "colType": "i" }
					, { "colName": "max_count_of"		, "colType": "i" }
					, { "colName": "category_id"		, "colType": "s" }
					, { "colName": "valid_attr_id"		, "colType": "s" }
					, { "colName": "inventory_order"	, "colType": "s" }
					, { "colName": "min_inv_level"		, "colType": "i" }
					, { "colName": "price_model"		, "colType": "s" }
					, { "colName": "price_01"			, "colType": "f" }
					, { "colName": "price_02"			, "colType": "f" }
					, { "colName": "price_03"			, "colType": "f" }
					, { "colName": "price_04"			, "colType": "f" }
					, { "colName": "price"				, "colType": "s" }
					, { "colName": "SKU"				, "colType": "s" }
					, { "colName": "product_type"		, "colType": "s" }
					, { "colName": "is_default"			, "colType": "s" }
					, { "colName": "group"				, "colType": "s" }
					, { "colName": "taxable_item"		, "colType": "s" }
					, { "colName": "start_date"			, "colType": "d" }
					, { "colName": "end_date"			, "colType": "d" }
					, { "colName": "prod_start_date"	, "colType": "d" }
					, { "colName": "prod_end_date"		, "colType": "d" }
				]
			, "orderBy": [ { "colName": "3" } , { "colName": "2" } , { "colName": "4" } ]

			, "SelectPK1Tmpl": "SELECT \"id\", \"customer_id\", \"prod_name\", \"desc\", \"cart_tmpl\", \"state\", \"limit_per_cust\", \"min_count_of\", \"max_count_of\", \"category_id\", \"valid_attr_id\", \"inventory_order\", \"min_inv_level\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", p_price_for_product_3x ( \"id\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", \"start_date\", \"end_date\" ) as \"price\", \"SKU\", \"taxable_item\" FROM \"x_product\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"
			, "SelectTmpl": "SELECT \"id\", \"customer_id\", \"prod_name\", \"desc\", \"cart_tmpl\", \"state\", \"limit_per_cust\", \"min_count_of\", \"max_count_of\", \"category_id\", \"valid_attr_id\", \"inventory_order\", \"min_inv_level\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", p_price_for_product_3x ( \"id\", \"price_model\", \"price_01\", \"price_02\", \"price_03\", \"price_04\", \"start_date\", \"end_date\" ) as \"price\", \"SKU\", \"taxable_item\" FROM \"x_product\" %{where_where%} %{where%} %{order_by_order_by%} %{order_by%} %{limit_limit%} %{limit%} %{offset_offset%} %{offset%}"

			, "key_word_col_name": "key_word"
			, "key_word_list_col": "__keyword__"
			, "key_word_tmpl": " %{kw_col%} @@ plainto_tsquery( %{kw_vals%} ) "

			, "category_col_name": "category"
			, "category_col": "category_id"
			, "category_tmpl": " %{cat_col%} in ( select \"id\" from p_get_children_of ( ''%{cat_vals%}''::varchar[] ) ) "

			, "attr_table_name": "x_product"
			, "attr_col": "id"
			, "attr_tmpl":" %{attr_col%} in ( select a1.\"fk_id\" from \"x_attr\" as a1 where a1.\"attr_type\" = ''%{attr_type%}'' and a1.\"attr_name\" = ''%{attr_name%}'' and a1.\"%{ref_col%}\" %{attr_op%} %{attr_vals%} )"

			, "PostJoin": [
						{ "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "x_attrs"
							, "Query": "SELECT \"id\", \"attr_type\", \"attr_name\", \"fk_id\", \"val1s\", \"val2s\", \"val1i\", \"val2i\", \"val1f\", \"val2f\", \"val1d\", \"val2d\" FROM \"x_attr\" WHERE \"fk_id\" = $1 "
						}
						, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "productInventory"
							, "Query": "SELECT \"id\", \"seq_no\", \"is_countable\", \"count_of\", \"reservation_count_of\", \"location_of\", \"start_date\", \"end_date\", \"weight\", \"box_size_h\", \"box_size_w\", \"box_size_d\", \"SKU\" FROM \"x_product_inventory\" WHERE \"product_id\" = $1 "
						}
						, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "imageList"
							, "Query": "SELECT \"x_image_list\".\"id\", \"x_image_list\".\"image_id\", \"x_image_list\".\"seq_no\", \"p_image\".\"file_name\", \"p_image\".\"base_file_name\", \"p_image\".\"h_size\", \"p_image\".\"w_size\", \"p_image\".\"f_size\", \"p_image\".\"img_type\"  FROM \"x_image_list\" as \"x_image_list\" left join \"p_image\" as \"p_image\" on \"x_image_list\".\"image_id\" = \"p_image\".\"id\" WHERE \"x_image_list\".\"fk_id\" = $1 "
						}
						, { "ColName": "id", "ColType":"s", "p":[ "id" ], "SetCol": "optionsList"
							, "Query": "SELECT \"x_product_options\".\"id\" as \"product_options_id\" , \"x_product_options\".\"group\" , \"x_product_options_meta\".\"display_order\" , \"x_product_options\".\"seq_no\" , \"x_product_options\".\"price_01\" , \"x_product_options\".\"price_02\" , \"x_product_options\".\"price_03\" , \"x_product_options\".\"price_04\" , \"x_product_options\".\"start_date\" , \"x_product_options\".\"end_date\" , \"x_product_options\".\"SKU\" , \"x_product_options_meta\".\"id\" as \"product_options_meta_id\" , \"x_product_options_meta\".\"option_type\" , \"x_product_options_meta\".\"required\" , \"x_product_options_meta\".\"count_of_option\" , \"x_product_options_meta\".\"price_model\" , \"p_image\".\"file_name\" , \"p_image\".\"base_file_name\" , \"x_product_options\".\"option_title\" FROM \"x_product_options\" as \"x_product_options\" left join \"x_product_options_meta\" as \"x_product_options_meta\" on ( \"x_product_options_meta\".\"group\" = \"x_product_options\".\"group\" ) left join \"p_image_list\" as \"p_image_list\" on ( \"p_image_list\".\"fk_id\" = \"x_product_options\".\"id\" ) left join \"p_image\" as \"p_image\" on ( \"p_image_list\".\"image_id\" = \"p_image\".\"id\" and \"p_image\".\"img_type\" = ''other'' ) WHERE \"x_product_options\".\"product_id\" = $1 ORDER BY 3 asc, 4 asc "
						}
					]
			}
	}
```




', 'TabServer2-sql-cfg-json-configuraiton-file-100060.html'
	, 'TabServer2 - sql-cfg.json configuraiton file' , 'A set of examples with explanation of what is being configured.', '/doc-TabServer2-sql-cfg-json-configuraiton-file', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '














These interfaces are created in crud.go, lines 68 to 82.

## Insert

HTTP, method POST

``` green-bar
	/api/table/NAME?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.

## Insert - with ID

HTTP, method POST

``` green-bar
	/api/table/NAME/ID?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.

## Update

HTTP, method PUT

``` green-bar
	/api/table/NAME?col=Value&col2=value...
```

The primary key can be a multi-part PK and must be supplied 
in the parameters.   It is possible to do non-PK updates with
this.  There is a flag in `sql-cfg.json` for this.
The ID/PK is specified in the `sql-cfg.json` file.

## Update - with ID

HTTP, method PUT

``` green-bar
	/api/table/NAME/ID?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.


## Delete

HTTP, method DELETE

``` green-bar
	/api/table/NAME?col=Value&col2=value...
```

The primary key can be a multi-part PK and must be supplied 
in the parameters.   It is possible to do non-PK updates with
this.  There is a flag in `sql-cfg.json` for this.
The ID/PK is specified in the `sql-cfg.json` file.

## Delete - with ID

HTTP, method DELETE

``` green-bar
	/api/table/NAME/ID?col=Value&col2=value...
```

If you supply the ID as Primary key it will be used.
The ID/PK is specified in the `sql-cfg.json` file.

## Select 

Select can be performed in both

``` green-bar
	/api/table/NAME/ID
```

and 

``` green-bar
	/api/table/NAME?col=Val1&col2=val2
```

format.   It is assumed that this is for a single row - primary or
unique key select.

Select provides a set of other options.  You can not name columns
with these options and use them in the `?col=val1` format.

``` green-bar
	?orderBy=[{"ColName":"abc","Dir":"asc"}]
```

is a JSON encoded array of columns with "asc" and "desc".
Note lower case on asc/desc.  Also the "Dir" is optional
and assumed to be "asc" if not specified.

``` green-bar
	?where={...}
```

a where clause - as a parse tree subset of a SELECT were
clause.  More on this later.

``` green-bar
	?limit=
	?offset=
```

Are integer values that will subset the query to the
specified range.  These are optional.




', 'Table-API-the-CDUD-100061.html'
	, 'Table API - the CDUD' , 'Details on the API suported in TabServer2', '/doc-Table-API-the-CDUD', 'go-ftl' );

insert into "p_document" ( "doc", "file_name", "title", "desc", "link", "group" ) values ( '














``` green-bar
	?where={...}
```

This is a JSON encoded string with a parse tree in it.
Each of the constants from this will be used as a bind 
variable in the select.

``` green-bar
	?where={"op":"and","List":[...]}
```

A list of and-ed together clauses in the where.

```
	?where={"op":"and","List":[{"op":"between","name":"DateColumName","Val1d":"2006-02-08T14:22:23","Val2d":"***2nd ISO Date/DateTime"},{...}]}
```

Op''s are: `between`, `not between`, `<`, `>`, `==`, `!=`, `<>`, `>=`, `<=`, `like`, `not like`, `in`, `not in`

"name" is the column name to do the operation on.

Values are: Val1d, Val2d - dates or date/time in ISO format, YYYY-MM-DDTHH24:MI:SS.sssssss
Val1s, Val2s - strings.  Val1b, Val2b - boolean.  Val1i, Val2i as integers.
Val1f, val2f as floats.   

``` green-bar
	?where={"op":"and","List":[{"op":"in","name":"ColName","List":[{"Val1s":"abc"},{"Val1s":"def"}] }] }
```

becomes

``` green-bar
	where ColName in ( $1, $2 )
```

with values of $1 == "abc" and $2 == "def".  It is not generally a good idea to use floating point data and
in lists in combination.  Val1? are used for values, so Val1s, Val1d, Val1i, Val1b, Val1f.

Expressions can be inclued as r-values.  To get `where ColName = ( 12 + 14 )` you can

``` green-bar
	?where={"op":"and","List":[{"op":"=","name":"ColName","Expr":[{"op":"+","Expr":[{"Val1i":12},{"Val1i":14}] }] }] }
```

It is my intent to build a SQL to Expression translator - SOON!  That would allow you to put in a string form
of the "where" clause and get the JSON parse tree that is equivalent.  This would be a purely development tool.

', 'Where-Clause-on-Select-Update-Delete-100062.html'
	, 'Where Clause on Select, Update, Delete' , 'Details on the API where clause', '/doc-Where-Clause-on-Select-Update-Delete', 'go-ftl' );


