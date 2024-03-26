Contributing To This Project
============================
``` JSON
	{
		"Section": "Overview"
	,	"SubSection": "Contribute"
	,	"SubSectionGroup": "Overview"
	,	"SubSectionTitle": "Contribute to this project"
	,	"SubSectionTooltip": "Requirements/Opportunities to become a contributor to this project"
	, 	"MultiSection":2
	}
```


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


