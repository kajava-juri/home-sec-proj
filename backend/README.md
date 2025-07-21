# Home Security Project Backend
REST API and MQTT message handler written in Go.

## Installing dependencies
``` bash
cd backend
go mod tidy
```

## Start the postgres database

Ensure you have PostgreSQL installed and running. I decided to not use Docker for the database to keep it simple.
Used PostgreSQL 17.5

Check your .env file in /backend/.env

``` bash
cd database
./init_database.sh
```