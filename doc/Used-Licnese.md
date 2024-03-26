Used License
============
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Component Licenses"
	,	"SubSectionGroup": "Licenses"
	,	"SubSectionTitle": "Go-FTL License"
	,	"SubSectionTooltip": "Go-FTL how it came to be"
	, 	"MultiSection":2
	}
```

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


