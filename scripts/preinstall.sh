#!/usr/bin/bash

# add wali user and group
useradd --system --no-create-home --user-group wali

# create directory where database is stored
mkdir -p /var/lib/wali
chmod 0640 /var/lib/wali
chown root:wali /var/lib/wali
