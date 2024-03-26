# How I was snookered

I saw an open source project that looked really good.   Suddenly I am thinking that this will
solve a bunch of my problems.   It's open source so I can get right in and add in just the code
for my application and I don't have to write all of it.

I am excited.  I even contributed to the project.  I added the JSONP module that I needed.
Luckily for me the process of adding the module didn't go as easily as it should have.
I had to hunt down some problems and gain some understanding of how all the code worked
together.  That led me into the dark underbelly of the project.  The software that I
had just contributed to was actually proprietary!   It had includes that pulled in
a non-open source library - from the *SAME* author!  



Proof:
[The propriatary license](http://github.com/pschlump/Go-FTL/img/Screen-Shot-2015-10-04-at-3.39.39-PM.png)  
[the readme](http://github.com/pschlump/Go-FTL/img/Screen-Shot-2015-10-04-at-3.41.45-PM.png) 

Also the documentation for the proprietary tool is proprietary.   Notice the (C) at the bottom
without any other information.  I checked and that is a proprietary license.

[the copyright](http://github.com/pschlump/Go-FTL/img/Screen-Shot-2015-10-06-at-8.02.34-PM.png)

I had just been conned.  Just goes to show you that "there is one born every minute" and 
in this case I was the one.

So...

I decided to write my own web server that completely replace the proprietary one.   Also I decided
to develop a license check utility that will walk down all the GO dependencies and verify that
what you are looking at is actually consistently licensed.  The web server will be first.

That is how [Go-FTL](http://github.com/pschlump/Go-FTL) was born.



