#!/bin/sh

# Clean cache first
# ./app cache clean

# Reset database (drops all tables)
# ./app db reset

# Migrate database (recreate tables)
# ./app db migrate

# Seed database (insert initial data)
# ./app db seed

# Start your Go app (it will load .env from current directory)
./app server

# fly deploy; fly machine stop 148e4d55f36278; fly machine restart 148e4d55f36278; fly machine restart 90802d3ea0ed38; fly logs
