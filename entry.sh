#!/bin/sh
# filepath: /home/zalven/Desktop/horizon-engine/entrypoint.sh

# Start NATS server in the background
nats-server -c nats.conf &

# Start your Go app in the foreground
./app server