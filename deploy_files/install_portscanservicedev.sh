#!/bin/sh

APP_PATH="/opt/scanner/"
BIN="portscanservicedev"
SERVICE="portscanservicedev"
FULL_PATH="${APP_PATH}${BIN}"

# prepare dir and service user
if [ -f $FULLPATH ]; then
    sudo systemctl stop $BIN 
fi

sudo mkdir -p ${APP_PATH}dev/
sudo useradd $SERVICE -G scanner,certaccess -s /sbin/nologin -M -N

# add as a proper service with logging
sudo cp ${BIN}.service /lib/systemd/system/${BIN}.service
sudo cp 30-${BIN}.conf /etc/rsyslog.d/30-${BIN}.conf
sudo chmod 755 /lib/systemd/system/${BIN}.service

sudo cp ${BIN} $FULL_PATH
sudo chown $SERVICE $FULL_PATH

# start her up.
sudo systemctl enable ${BIN}.service
sudo systemctl restart rsyslog 
sudo systemctl start ${BIN}