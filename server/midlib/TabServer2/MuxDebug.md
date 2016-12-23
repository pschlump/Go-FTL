Mux Debuing
===========

Two additional featrues were added to the mux to aid in debuging.  First you can add a "Comment" field to 
each path.  This gets printed out to stderr when the path matches if the "DebugMatchOn" flag is true.
The flag can be set with calls to

	func (r *MuxRouter) DebugMatch(tf bool) 

In tab server it is turned on by default.

TODO
----
	
1. May want to add "params" in output when DebugMatchOn is true.


