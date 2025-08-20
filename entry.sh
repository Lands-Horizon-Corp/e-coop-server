#!/bin/sh

# Reset and seed the database
./app cache clean
./app db migrate
./app db reset
./app db seed

# Start your Go app (it will load .env from current directory)
./app server