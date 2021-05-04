#!/bin/bash

if [ $EUID != 0 ]; then
    sudo "$0" "$@"
    exit $?
fi

go build
sudo cp nxutil /usr/bin/nxutil