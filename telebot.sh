#!/bin/sh
#
#   /etc/init.d/telebot
#
# chkconfig: - 65 35
# description: Mail copier Milter 25
# processname: /usr/local/twofive/telebot/telebot
# config: /usr/local/twofive/telebot/telebot.conf
# pidfile: /var/lock/subsys/telebot

PATH=/sbin:/bin:/usr/bin:/usr/sbin

# Source function library.
. /etc/rc.d/init.d/functions

pidfile='/var/run/telebot/telebot.pid'
logfile='/tmp/telebot.log'
pidfolder='/var/run/telebot'

start_telebot() {
  if [ -f $1 ];
  then
    echo "FAILED"
    return 1
  else
    cd /usr/local/twofive/telebot/
    nohup ./telebot -p=$1 -l=$2 &>/dev/null &
    sleep 1
    if [ -f $1 ];
    then
      echo "OK"
    else
      echo "FAILED"
      return 1
    fi
  fi
  return 0
}

stop_telebot() {
  if [ ! -f $1 ]; then
      echo "Failed: pidfile: $1 not found"
      return 1
  fi
  kill `cat $1`
  rm -f $1
  echo "OK"
  return 0
}

reload_telebot() {
  if [ ! -f $1 ]; then
      echo "Failed: pidfile: $1 not found"
      return 1
  fi
  kill -HUP `cat $1`
  echo "OK"
  return 0
}

case "$1" in
  start)
      echo -n "Starting telebot: $pidfile $conffile -> "
      start_telebot $pidfile $logfile
    ;;
  stop)
      echo -n "Shutting telebot: "
      stop_telebot $pidfile
    ;;
  status)
      status -p $pidfile
    ;;
  restart)
      echo -n "Restarting telebot: "
      stop_telebot $pidfile
      start_telebot $pidfile $conffile $licfile
    ;;
  reload)
      echo -n "Sending -HUP to telebot to reload config: "
      reload_telebot $pidfile
    ;;
  *)
    echo "Usage: telebot {start|stop|status|restart|reload}"
    exit 1
    ;;
esac
exit $?

exit $RETVAL

