#!/bin/bash

set -e

DATABASE_URL="postgres://go_user:strongpassword@localhost:5432/go_app_db?sslmode=disable"

migrate -path ./migrations -database $DATABASE_URL up
