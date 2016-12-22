Common - why this little package
================================


This package exists for the sole purpose of removing linking errors in Go-FTL.  These errors are cause by loops
when a topological sort is used on the dependencies in the link process.   Items are moved to this package
so that they can be successfully shared between different packages.  This package should have no outside
imports.

