
```
root@peach:/etc/systemd# cd system
root@peach:/etc/systemd/system# cat go_ftl.service
[Unit]
Description=Run Go-FTL server on 206 - port 80
DefaultDependencies=no
After=systemd-sysctl.service
Before=sysinit.target

[Service]
Type=simple
ExecStart=/etc/go_ftl/go_ftl.sh /etc/go_ftl/go_ftl.cfg start
ExecReload=/etc/go_ftl/go_ftl.sh /etc/go_ftl/go_ftl.cfg reload
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
root@peach:/etc/systemd/system# cat /etc/go_ftl/go_ftl.sh
#!/bin/bash

. $1

mkdir -p $LOG_DIR

echo "At 1st boot" >>$LOG_DIR/$LOG_FILE
date >>$LOG_DIR/$LOG_FILE

cd $HOME_DIR

while true ; do
	pwd >>$LOG_DIR/$LOG_FILE
	ls -l ./go-ftl >>$LOG_DIR/$LOG_FILE
	ls -l $CFG_FILE >>$LOG_DIR/$LOG_FILE
	$HOME_DIR/go-ftl -c $CFG_FILE >$LOG_DIR/$LOG_FILE 2>&1
	echo "exit=$?" >>$LOG_DIR/$LOG_FILE
	date >>$LOG_DIR/$LOG_FILE
	echo "Crash Recovery - sleep 60" >>$LOG_DIR/$LOG_FILE
	sleep 60
done

exit 0

root@peach:/etc/systemd/system#
root@peach:/etc/systemd/system# cat /etc/go_ftl/go_ftl.cfg

LOG_DIR=/home/pschlump/Projects/go-ftl/log
HOME_DIR=/home/pschlump/Projects/go-ftl

CFG_FILE="dot206.json"
LOG_FILE=",newLog.log"


```

```
root@peach:~/Projects/go-ftl# cat set-port-80.sh
#!/bin/bash

setcap 'cap_net_bind_service=+ep' go-ftl
```
