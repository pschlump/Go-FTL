
x: load_config

all:
	go build

run: file_output

test: all test1

test1:
	./mon-cli load --file ./cfg0.json
	./mon-cli dump --file ,tmp.out
	diff -w ,tmp.out ./cfg0.json

install:
	cp mon-cli ~/bin

linux:
	( cd .. ; tar -cf ~/x.tar . )
	scp ~/x.tar build_mon_cli.sh pschlump@192.168.0.182:/home/pschlump
	ssh pschlump@192.168.0.182 chmod +x ./build_mon_cli.sh
	ssh pschlump@192.168.0.182 ./build_mon_cli.sh
	scp pschlump@192.168.0.182:/home/pschlump/x.tar.gz .
	tar -xzf x.tar.gz
	
#load_config:
#	./mon-cli load --file ./cfg0.json
load_config:
	./mon-cli.linux load --file ./cfg0.json
	
dump_config:
	./mon-cli dump --file -

up_down:
	./mon-cli status -P 30

new_up_down:
	./mon-cli live-monitor

file_output:
	./mon-cli live-monitor --quiet --file /Users/corwin/go/src/github.com/pschlump/Go-FTL/server/goftl/testdata/9001/www/lm/data.json 2>,log.out &

cp_to_linux_sys:
	scp mon-cli.linux *.tmpl pschlump@192.168.0.140:/home/pschlump/bin

