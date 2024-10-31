#!/usr/bin/bash
set -e

# add wali user and group
useradd --system --no-create-home --user-group wali

# create directory where database is stored
mkdir -p /var/lib/wali
chmod 0770 /var/lib/wali
chown root:wali /var/lib/wali
