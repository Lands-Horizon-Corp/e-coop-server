#!/bin/bash

nats-server -c config/nats.conf &

sleep 2

./app db:reset
./app db:seed

./app server