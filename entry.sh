#!/bin/sh

# Start NATS server in the background
nats-server -c config/nats.conf &

# Wait a moment to ensure NATS is up
sleep 2

# Reset and seed the database
./app cache:clean
./app db:migrate
./app db:reset
./app db:seed

# Start your Go app (it will load .env from current directory)
./app server