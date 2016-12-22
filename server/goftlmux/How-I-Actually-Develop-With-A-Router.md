# How I Actually Develop With A Router

If I loved the design decisions that people made with routers I would not have implement this one.
The biggest problem is how development actually occurs with a router.

Most important - it is not some clean emaciate process.  It is a dirty grungy process.

I start out with some routes

```

	/api/test
	/api/simple
	/*filename

```

This is all good.  I test.  I check that I can call /api/test and /api/simple.
I check that files will display.  So cool!  I have a router.

Then the router grows to a bunch of paths.  The problem is that I have all the
paths and then the exceptions.  That looks like:

```

	/api/oops1
	/api/oops2
	/api/oops3
	/api/:expected
	/*filename

```

*:expected* is what I was expecting.  That is what was in the original design.
That is the easy case.  /api/oops1..3 are the unexpected.  They are the paths
that I really need to look like /api/<something> but they don't return the same
something as :expected.

Now some time passes.  Usually a few weeks.  Business requirements change.
Software is deployed.  It gets really hard to change the deployed software.
Just think that the software is now installed in the mountains of Peru and
you are in the US.   You don't really want to book a flight to change the
software.

Some new routes are needed.


```

	/api/oops1
	/api/oops2
	/api/oops3
	/api/whatever/domore
	/api/whatever/again
	/api/whatever/yep
	/api/whatever/*unexpected
	/api/:expected
	/*filename

```

This is where having some nice clean, "Only Explicit Matches", really 
starts to produce a lot of pain.

This is where some sort of magic mod_rewrite takes over.  You end up with
a set of rewrite rules in Apache or NgNix that nobody will every understand.

