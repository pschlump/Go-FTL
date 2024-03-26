
all: index.html

index.html: index.m4.html config.m4.html
	m4 -P index.m4.html >index.html

prod:
	cp prod.m4.html config.m4.html
	( cd . ; make all )


dev:
	cp dev.m4.html config.m4.html
	( cd . ; make all )


