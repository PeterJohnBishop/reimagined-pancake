# Gin Server with Postgres database and JWT user authentication

<!-- launch postgres Docker container -->
docker compose up 

<!-- Configure .env with connection variables -->
    Gin Server
    PORT=

    Database Credentials
    PG_HOST=
    PG_PORT=
    PG_USER=
    PG_PASSWORD=
    PG_DB=

<!-- Launch server -->
go run main.go

<!-- Test CRUD operations and JWT authentication (/server/user_test.go)-->
cd tests && go test 

![test screenshot](https://github.com/PeterJohnBishop/reimagined-pancake/blob/main/assets/screenshot-2026-05-06_01-59-36.png?raw=true)

# Open Routes

POST /register
POST /login

# Protected Routes

GET /api/user/:id
GET /api/users
PUT /api/user
PUT /api/user/password
DELETE /api/user

