# AuthService

For generating swagger docs:
```
go install github.com/swaggo/swag/cmd/swag@latest
go get -u github.com/swaggo/echo-swagger
swag init -output docs --parseInternal --parseDependency
```