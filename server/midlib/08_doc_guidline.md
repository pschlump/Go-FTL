How to document a new module
============================

Write the documentation in Markdown (.md) file in the same directory as the code for the module. 
Place the title of the document on the first line.

Run

	$ make doc

in this directory ( .../Go-FTL/server/midlib ) and it will collect all the .md files and convert
them to .html, combine the templates and build the final document.

Each .md file in this directory will become an overview section in the document.
