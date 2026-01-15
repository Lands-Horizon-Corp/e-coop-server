#!/bin/sh

# Clean cache first
# ./app cache clean

# Enforce security blocklist
# ./app security enforce

# Reset database (drops all tables)
# ./app db reset

# Migrate database (recreate tables)
# ./app db migrate

# Seed database (insert initial data)
# ./app db seed

# Start your Go app (it will load .env from current directory)
./app server