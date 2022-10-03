This is api server of to-do list application by go.
Use:

- net/http
- MongoDB(Docker container)
- MongoDB official go driver ( go.mongodb.org/mongo-driver )

## Get started

```
# Create .env
$ vim .env
MONGO_URI=mongodb://root:example@localhost:27010/?maxPoolSize=20&w=majority

# Start DB container
$ docker compose up -d

# Stop DB container
$ docker compose stop

# Build and start server
$ docker build
$ ./go-todo-app-api
```

## Send request to server

```
# Read
$ curl http://localhost:8080/tasks/ -w '%{http_code}\n'

# Create
$ curl -X POST -d '{"title": "running"}' -H "Content-Type: application/json" http://localhost:8080/tasks -w '%{http_code}\n'

# Update
$ curl -X PUT -d '{"title": "swimming", "completedAt": "2022-10-02T07:04:25.000Z"}' -H "Content-Type: application/json" http://localhost:8080/tasks/{task mongoId} -w '%{http_code}\n'

# Delete
$ curl -X DELETE http://localhost:8080/tasks/{task mongoId} -w '%{http_code}\n'
```
