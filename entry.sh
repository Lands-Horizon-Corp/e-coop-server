#!/bin/sh

# Start NATS server in the background
nats-server -c config/nats.conf &

# Wait a moment to ensure NATS is up
sleep 2

# Start your Go app (it will load .env from current directory)
./app server