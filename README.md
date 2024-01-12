# proteng-user-mgmt

## Running in local

### 1. get `.env.dev` file from notion

### 2. install packages

```
go mod tidy
```

### 3. run development

- setting local environmental variable `ENV`, should be `dev`
- note that if you put set `ENV` to `<environment>` the app will load env vars from `.env.<environment>` file

MacOS

```
ENV=dev go run main.go
```

Windows - CMD

```
set ENV=dev && go run main.go
```

Windows - Powershell

```
$Env:ENV = "dev" && go run main.go
$Env:ENV = "dev" ; go run main.go
```

## Building

building with docker will not bring the env file to the image. Instead, you will have to specify in during the run time

```
docker build -t proteng-user-mgmt .
docker run -d --name proteng-user-mgmt  --env-file .env.dev --network proteng-net -p 8081:8080 proteng-user-mgmt

```
