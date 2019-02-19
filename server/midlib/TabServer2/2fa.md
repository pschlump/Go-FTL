Processing of 2fa in TabServer2
==

un/pw/2fa all at once.
--

This means that the user must re-type the 2fa and push-button-to-login is not possible.

1. A request to s_login takes place.
2. If the login is successful with un/pw, then
	1. Post processing takes place on val2fa
	2. If this is a valid 2fa value then
	3. "success", else it is an error

With a 2 step form process un/pw are processed and validated, this data is saved to
Redis for 4 minutes (twice the 2fa time).  A UUID is generated in the Go code
as a token and passed back to the client.

The client is on the 2fa input page.  Cases

1. There is only 1 login request happening and the  2fa-app sees this.  It puts up
  a message for 2fa login to take place and sends that to the server.  The server
  sends(wss://push) a message to the client that login is done.
	This is only if wss:// is active and this is configured.
  All the login data is pushed to the client via wss:// and the login is complete.

2. More than one is trying to login - the 2fa-app displays a number 333-333 and the
  user is asked to type it in - or push-to-login can not work and a number is 
  displayed.  The user types in the number and part-2-of-2 of login takes place.
	This part will use the token from (1) to pull the part 1 config, check the 2fa
	token - then - if match - push bck a successful login.

  


