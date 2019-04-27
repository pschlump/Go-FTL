SrpAesAuth: Strong Authentication for RESTful Requests
======================================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "SRP Login"
	,	"SubSectionGroup": "Authentication"
	,	"SubSectionTitle": "Strong Authentication"
	,	"SubSectionTooltip": "Strong authentication using Secure Remote Password (SRP), Two Factor Authrization (2FA) and encryption of messages with Advanced Encryption Standard (AES)"
	, 	"MultiSection":2
	}
```

``` warning-box
	<h2> Genral Rant </h2>
	I see in the news that 642 million passwords have been compromized.  This is because the passwords	
	were on the server.  If you don't have the passwords then you can't compromize them.  With SRP
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

1. Register with email confirmation and setup of Two Factor Authentication (2FA).
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
SRP's early development; without it, SRP would not have received
anywhere near the amount of analysis and feedback that it has gotten
since it was first proposed and refined. It is thus fitting that
the Internet at large can benefit from the fruits of this endeavor.
Since SRP is specifically designed to work around existing patents
in the area, it gives everybody access to strong, unencumbered
password authentication technology that can be put to a wide variety
of uses."

"The SRP distribution is available under Open Source-friendly
licensing terms (for the net.savvy reader, it's a "BSD-style"
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

For these reasons this protocol should be run over TLS (https) and the security provided should be
in addition to standard TLS encryption.


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

##### Invalid 'v' Value

If SendStatusOnerror is true then,  Status = 400, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9005","msg":"Invalid 'v' verifier value.","LineFile":"File: /.../aessrp_ext.go LineNo:1288",
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
`t`      | SRP 't' value - see references on SRP.		See Reference 1, 2.
`r`      | SRP 'r' value - see references on SRP.		See Reference 1, 2.
`bits`   | By default 2048 - but this can be configured to be 3K, 4k, 6k or 8k.
`B`      | SRP 'B' value - see references on SRP.		See Refernece 1, 2.
`f2`     | "y" If two factor authentication will be required (default), "n" otherwise.

#### Possible Output Errors.

##### Inability to generate SRP values 't' or 'r'. 

If SendStatusOnerror is true then,  Status = 500, Internal server error.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"9007","msg":"Internal error - unable to generate 't' key.","LineNo":0000,"FileName":".../aessrp_ext.go"}
```

or

``` JSON
	{"status":"error","code":"9008","msg":"Internal error - unable to generate 'r' key.","LineNo":0000,"FileName":".../aessrp_ext.go"}
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
`A` | SRP 'A' value
`r` | SRP 'r' value

#### Output.

On Success A JSON hash with:
	// io.WriteString(www, fmt.Sprintf(`{"status":"success","B":"%s","Bits":%d,"M1":%q}`, sss.XB_s, hdlr.Bits, sss.XM1_s))

Name     | Description
|---    | --- 
`status` | "success"
`B`      | Numeric salt value from calculation.
`Bits`   | Bits in computation.  Same as initialization bits.
`M1`     | SRP 'M1' value - See references on SRP.		See Reference 1, 2.

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
`M1`           | SRP 'M1' value as calculated by the client. - This will be validated by the server now.
`r`            | SRP 'r' value.
`stayLoggedIn` | True False flag for staying logged in.  True/Flase, 1/0, y/n etc.   Requries call to /api/setFingerprint to work.

#### Output.

On Success A JSON hash with:

Name             | Description
|---            | ---------------------------------------------------------------------------------
`status`         | "Success"
`M2`             | SRP M2 value. - Calculated.
`FirstLogin`     | Flag if this is the first login that has happed after registration.
`MoreBackupKeys` | True/False flag. - If the user should get more one-time backup keys.  True when less than 5 keys left.  
`UserRole`       | One of the security roles like, 'user', 'admin' etc.  This is the role for this user.
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
`g`                 | SRP 'g' value - See references on SRP.		See Reference 1, 2.
`N`                 | SRP 'N' value - See references on SRP.		See Reference 1, 2.
`TwoFactorRequired` | Enabled by default, 'y'.  'n' if not enabled.
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
to be called with a new `salt` and `v` to reset the user's password to a new value.

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
5. Save 'prw:'||token with "email" and a 24hr timeout to d.b.
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
`PwSendEmailOnRecoverPw` | If 'y' then send email to user when attempt is made to reset password.  Email is 'pw-reset-attempt.tmpl.'
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
1. muxEnc.HandleFunc("/api/confirm-registration", respHandlerConfirmRegistration).Method("GET")            // Redirect-Link: To 1st Login /#/login, confirmed <- 'y'  -- Needs to have templates.




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


##### Invalid 'v' verifier value.

If SendStatusOnerror is true, then  Status = 400, Bad Request.
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"8005","msg":"Invalid 'v' verifier value.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
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


##### fmt.sprintf("unable to find account with email '%s'.", email)).

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"63","msg":"fmt.Sprintf("Unable to find account with email '%s'.", email))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
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
an acceptable source for resetting this user's password.

On error redirects to other error appropriate pages.

#### Possible Output Errors.

See Above.

### Example.

As a GET Request - Split on Multiple Lines:

	 http://www.example.com/api/pwrecov2?email_auth=UUID



## /api/getDeviceId - Get a new DeviceId.

Get a new DeviceId - and update the DeviceId in the users configuration.

This is normally called from a user's client page when they want to register a new device.
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


##### fmt.sprintf(`Unable to find account with email '%s`, email))

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"37","msg":"fmt.Sprintf(`Unable to find account with email '%s`, email))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
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

This is the normal `/api` call to change the user's password.   A new 'salt' and a new 'v' are generated
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
`salt`  | The new 'salt' value.
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

##### Invalid 'v' verifier value.

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"20","msg":"Invalid 'v' verifier value.","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
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
as a part of the user's first login.    After they have used up 75% of these keys (usually 15)
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


##### fmt.sprintf(`unable to find account with email '%s`, email))

If SendStatusOnerror is true, then  Status = 400, Bad Request,
else Status = 200 and JSON response is:

``` JSON
	{"status":"error","code":"58","msg":"fmt.Sprintf(`Unable to find account with email '%s`, email))","LineFile":"File: /.../aessrp_ext.go LineNo:1248"}
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

You can not get the attributes on another 'admin' user. - To do that requires the 'root' account.

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

You can not set the attributes on another 'admin' user. - To do that requires the 'root' account.

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

Sets the password, 'salt', 'v' for some other user.

You can not set the password on another 'admin' user. - To do that requires the 'root' account.

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



