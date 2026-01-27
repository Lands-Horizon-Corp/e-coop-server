#!/bin/sh

./app cache-clean

./app security-enforce-blocklist

./app db-reset

./app db-migrate

./app db-seed

./app server