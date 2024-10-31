#!/usr/bin/bash
set -e

if [ $1 != "upgrade" ]; then
    # remove user, if not an upgrade
    deluser wali
fi
