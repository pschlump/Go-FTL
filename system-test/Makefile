
all:
	go build

# Linux Ubuntu 14.04
HOST=http://192.168.0.182:12000

# Windows 8.1
HOST=http://192.168.0.159:12000

# Windows 7
HOST=http://192.168.0.150:12000

# Mac Os X - 10.9.5

# Mac Os X - 10.10.5

# Mac Os X - 10.11.x ( Old Gray )

# Raspberry Pi - Debian 32 bit

.PHONEY: test test1

test: test1	test2
	@echo "\nPASS"

OF=2

# ---------------------------------------------------------- tests --------------------------------------------------------

# pull back a simple file for index.html
test1:
	@echo "1/${OF}"
	@mkdir -p out ref
	@wget -o out/r1.out -O out/t1.out ${HOST}/index.html
	@grep "200 OK" out/r1.out >/dev/null
	@diff out/t1.out ref

# remote execute command and very get "success"
test2:
	@echo "2/${OF}"
	@mkdir -p out ref
	@wget -o out/r2.out -O out/t2.out '${HOST}/api/run-cmd?cli={%22Cmd%22:[%22ls%22,%22-ltro%22]}&ApiKey=bob-bob-bob&_x_=112'
	@grep "200 OK" out/r2.out >/dev/null
	@grep "success" out/t2.out >/dev/null

# 
# Status - /api/status
# Non-Configured item
# Name Resolved - dev1.test1.com, dev2.test1.com
# Test on 1.7 of Go
# 
# Test on Windows 10
# Test on Windows 8 - -VM- -
# Raw data + Template
# Test HTTP5Path 
# 
# Test *** all the middleware **
# Test *** themes and file_server ***
# Test *** TabServer and with PosgreSQL ***
# Test *** TabServer and with Oracle ***
# Test *** TabServer and with T-SQL/odbc ***
# Test *** TabServer and with MySQL ***
# 

