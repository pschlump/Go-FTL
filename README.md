Easy-to-configure Web server and Reverse Proxy Server in Go (golang).
=====================================================================


[![Build Status](https://semaphoreci.com/api/v1/go-ftl/go-ftl/branches/master/badge.svg)](https://semaphoreci.com/go-ftl/go-ftl)

## Index

* [Overview](#overview)
* [Install](#install)
* [Usage](#usage)
* [FAQ](#faq)

## Overview

Go-FTL is a complete web server and reverse proxy.  It supports name based virtual
servers with both HTTP and HTTPS (tls/sni) name resolution.  A simple interface
for developing your own middleware is provided.  A large collection of pre-built
middleware is available.   

[Go-FTL documentation](http://www.go-ftl.com/docs/index.html)

[Go-FTL download](http://www.go-ftl.com/docs//index.html/doc-Download-Compiled-Binaries)

Go-FTL is MIT licensed.  All of the middleware that is built in is MIT licensed.  The libraries that are used are MIT style licensed.
For details see the documentation section on licenses.

Go-FTL was developed because another web server had buried in it a required
download that was a proprietary license.  Given that the other web server claimed to be open source - this hit a nerve and I 
built Go-FTL.

The documentation for Go-FTL is a Creative Commons Attribution License v. 4.0.  The source for the
documentation is in the ./doc directory.   The website for the server is in the ./website
directory and carries the same Creative Commons Attribution license for content (.md and .html files)
and MIT license for code (.tc, .js, .go etc).

The Code is MIT licensed, see the LICENSE file.   

Go-FTL can be easily built from source or binaries can be downloaded.  The
pre-compiled binaries are tested on Windows 7, Mac OS X 10.9.5 and Ubuntu Linux
14.04.

## Install

To download and install

```bash

	$ git clone https://github.com/pschlup/Go-FTL.git
	$ cd Go-FTL/server/goftl
	$ go get 
	$ go build

```

Works with Go v1.6.2+.  You can download a tar ball with the source in it.  Also
pre-compiled versions for Windows, Mac and Linux are available.  Docker images
are planned for the very near future.

## Usage

Create a configuration in JSON like the following example:

```json

	{
		"unique-name": {
			"listen_to":[ "http://localhost:3111" ],
			"plugins":[
				{ "file_server": { "Root":"./static", "Paths":"/"  } }
			]
		}
	}

```

Will serve static files from the path `./static` as `http://localhost:3111/`.

And start the server:

```bash

	$ ./goftl -c config.json

```

## FAQ

### How is in/memory on/disk caching implemented?

Several of the middleware modules use sha256 of the data as the ETag header to guarantee cache consistency.  By doing this
the Go-FTL server can check that the cached file matches what is in the browsers cache and respond with a 304 Not Modified instead of
sending the file.  inMemoryCache works in this way.  You can specify a set of acceptable and reject URLs and a duration.
Each file is sha256 hashed and cached.  If the file changes at all then a new one will be provided.  Otherwise the
304 Not Modified response is sent.

###	Is this open-source?

Yes! 100% raw, unadulterated, Free Open Source Software.
The code is MIT licensed.  The Documentation is Creative Commons Attribution licensed.

#### The "Too Many Open Files" Error

Go-FTL creates a lot of files on `/proc/$pid/fd`. In case you see ftl0 crashing, you can see how many files are open by;

```bash
	$ sudo ls -l /proc/`pgrep ftl0`/fd | wc -l
```

To find out your personal limit:

```bash
	$ ulimit -n
```

To change it:

```bash
	$ ulimit -n 64000
```

You can change soft - hard limits by editing `/etc/security/limits.conf`.

