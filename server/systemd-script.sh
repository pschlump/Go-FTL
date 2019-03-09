#!/bin/bash

PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
HOME=/home/pschlump/Projects/go-ftl
DAEMON=$HOME/go-ftl

cd $HOME

./go-ftl -c ./dot206.json >,new-log 2>&1

exit 0



#! /bin/bash
### BEGIN INIT INFO
# Provides:		go-ftl
# Required-Start:	$syslog $remote_fs
# Required-Stop:	$syslog $remote_fs
# Should-Start:		$local_fs
# Should-Stop:		$local_fs
# Default-Start:	2 3 4 5
# Default-Stop:		0 1 6
# Short-Description:	go-ftl - HTTP server
# Description:		go-ftl - HTTP server
### END INIT INFO

NAME=go-ftl
DESC=go-ftl

RUNDIR=/home/pschlump/Projects/go-ftl
PIDFILE=$RUNDIR/go-ftl.pid

test -x $DAEMON || exit 0

if [ -r /etc/default/$NAME ]
then
	. /etc/default/$NAME
fi

. /lib/lsb/init-functions

set -e

case "$1" in
  start)
	echo -n "Starting $DESC: "
	mkdir -p $RUNDIR
	touch $PIDFILE

	cd $RUNDIR

	$DAEMON -c ./dot206.json > ,log 2>&1 &
	THE_PID=$!
	echo "$THE_PID" >$PIDFILE
	;;

  stop)
	echo -n "Stopping $DESC: "
	if [ -f $PIDFILE ] ; then
		kill $( cat $PIDFILE )
		rm -f $PIDFILE
	fi
	sleep 1
	;;

  restart|force-reload)
	${0} stop
	${0} start
	;;

  status)
	# echo "Unknown:TBD"
	wget -O - 'http://127.0.0.1:80/api/status?fmt=text'
	;;

  *)
	echo "Usage: /etc/init.d/$NAME {start|stop|restart|force-reload|status}" >&2
	exit 1
	;;
esac

exit 0
