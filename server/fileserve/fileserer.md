FileServer: Extended File Server
================================
``` JSON
	{
		"Section": "Middleware"
	,	"SubSection": "Overview"
	,	"SubSectionGroup": "File Server"
	,	"SubSectionTitle": "Make requests slow to test latency in network."
	,	"SubSectionTooltip": "Use this  as a tool when testing your web application.  Slows it way down"
	, 	"MultiSection":2
	}
```


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


