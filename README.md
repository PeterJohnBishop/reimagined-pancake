# Gin Server with Postgres database and JWT user authentication

<!-- launch postgres Docker container -->
docker compose up 

<!-- Configure .env with connection variables -->
    ## Gin Server
    PORT=

    ## Database Credentials
    PG_HOST=
    PG_PORT=
    PG_USER=
    PG_PASSWORD=
    PG_DB=

<!-- Launch server -->
go run main.go

<!-- Test CRUD operations and JWT authentication (/server/user_test.go)-->
go test 

# open routes

POST /register
POST /login

# protected routes

GET /api/user/:id
GET /api/users
PUT /api/user
PUT /api/user/password
DELETE /api/user

