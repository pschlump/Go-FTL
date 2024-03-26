
all:
	go build

test: pre test1 test2 test3 test3 test4 test5 test6 test7 post

pre: 
	mkdir -p test 
	go build

post: 
	@echo PASS

test1:
	./email-cli echo a b c >test/test1.out
	diff ref/test1.out test/test1.out

test2:
	./email-cli show-apps >test/test2.out
	diff ref/test2.out test/test2.out

test3:
	-rmdir ../tmp/xyzzy
	./email-cli add-apps -app xyzzy  >test/test3.out
	./chk-dir.sh ../tmpl/xyzzy
	grep xyzzy ../cfg.json
	diff ref/test3.out test/test3.out

test4:
	./email-cli del-apps -app xyzzy  >test/test4.out
	./ungrep.sh xyzzy ../cfg.json
	diff ref/test4.out test/test4.out

test5:
	./email-cli show-tmpl -app qr-today.com  >test/test5.out
	diff ref/test5.out test/test5.out

test6:
	echo "Add new tempalte"

test7:
	echo "Remove tempalte"

test8:
	@echo "1st Email - just send one - simulated"
	go build
	./email-cli send --app qr-today.com --tmpl debug8.tmpl --to pschlump@gmail.com subject "Hi Bob" bodyhtml "Ya Ya Ya..." bodytext "Ya Ya Ya Text!..."
	
test9:
	./email-cli send --app qr-today.com --tmpl debug8.tmpl --to pschlump@uwyo.edu subject "Hi Bob" bodyhtml "Ya Ya Ya..." bodytext "Ya Ya Ya Text!..."
	

xx:
	cat test/test4.out
	cp test/test4.out ref

