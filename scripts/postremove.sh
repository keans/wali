#!/usr/bin/bash
set -e

if [ $1 != "upgrade" ]; then
    # stop service before deleting the user
    systemctl stop wali

    # remove user, if not an upgrade
    deluser wali
fi
