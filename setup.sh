#!/bin/bash

if [ ! -f ".internal.env" ] ; then
    printf "Missing .internal.env in backend/env. Generating secure password and JWT secret...\n"

    postgres_pass=$(openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 24)

    touch .internal.env
    set -o noclobber
    echo "POSTGRES_PASSWORD=$postgres_pass" >> .internal.env
    echo "POSTGRES_PORT=1357" >> .internal.env

    printf "\n.internal.env file created and populated!\n"
    printf "You can change these values manually by going into the .internal.env file 
manually and editing the file"
fi


printf "\nCompleted setup! :)\n"
