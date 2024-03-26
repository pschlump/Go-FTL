# Why Choose to use Go-FTL

## Go-FTL v.s. Caddy

1. The license is better - if you are in business with any kind of hardware the Apache 2.0 license is a non-starter.  Check the provision on patents in it.
2. The middle-ware implementation mode is simpler.  If you are going to extend the server - this is an easier server to extend.
3. Go-FTL has a wider selection of middle-ware (At least the last time I checked).
4. Go-FTL supports remote on-the-fly configuration changes.
5. Go-FTL checks the syntax of it's configuration files and only accepts valid input.   Caddy uses a set of ad-hock scanning and parsing that has huge syntactical gaps in it.  An example.  A redirect with "redirect https://localhost:8104/ code 307" actually results in a 301 permanent
redirect without a syntax error.  Some browsers cache the 301 forever! That means that you can cause a user to redirect and never ever get
that user back!
6. Go-FTL has some complete usage examples - in jQuery and AngularJS - that show how to use the advanced security features.
7. Go-FTL is free - and - has the availability of commercial support.

## Go-FTL v.s. Ngnix

1. Go-FTL is in Go and that is way easier to extend and work with than C code (A lot safter too).
1. Go-FTL is about the same speed as Ngnix.
1. Go-FTL supports on-line remote reconfiguration.  Ngnix uses a file based configuration.  This makes Go-FTL a more data-center friendly tool.

There are some reasons to choose Ngnix over Go-FTL too.  Ngnix has been around for a while!

## Go-FTL v.s. Apache

1. Apache is ancient.   If you are going to pick a server at least use something more modern.
2. Apache is truly hideous to write extensions for.
3. Apache is slow.
4. Apache fails to match the kind of server hardware that is commonly in use today.
5. Apache uses the Apache 2.0 license.  That is a problem.

## MOST IMPORTANT

1. You can configure Go-FTL to only include the middle-ware that you need.  That means that you are not trying to protect
yourself from 20+ years of stuff that has nothing to do with you you need.






