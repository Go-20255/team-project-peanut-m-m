#!/bin/bash
source ../.internal.env
docker stop monopoly-postgres-e
docker rm monopoly-postgres-e
echo "port: $POSTGRES_PORT, Pass: $POSTGRES_PASSWORD"
docker run --name monopoly-postgres-e -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD -d -p $POSTGRES_PORT:5432 postgres

