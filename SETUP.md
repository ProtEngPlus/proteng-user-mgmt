# Setup

## Run locally

1. **Copy the env file**
   ```sh
   cp .env.example .env.local
   ```
   Fill in real values. Done when: `.env.local` exists with real values (not the empty template).

2. **Install dependencies**
   ```sh
   go mod tidy
   ```
   Done when: exits 0, no errors.

3. **Run** (must run from the repo root — email templates load via a relative glob path). `ENV` picks which `.env.<ENV>` file loads (there is no `.env.dev` anymore):
   - macOS/Linux: `ENV=local go run main.go`
   - Windows CMD: `set ENV=local && go run main.go`
   - Windows PowerShell: `$Env:ENV = "local"; go run main.go`

   Done when: log shows the server listening on `HTTP_PORT` with no crash.

## Format

`gofmt` autofixes on save/commit. Run manually against the whole repo:

```sh
gofmt -l -w .
```

## Lint

`go vet` reports issues but does not autofix — fix them by hand:

```sh
go vet ./...
```

## Pre-commit hooks

Format + lint above run automatically via [pre-commit](https://pre-commit.com/) on `git commit`; `go build` + `go test` additionally run on `git push`. See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

Install once per clone:

```sh
pip install pre-commit
pre-commit install --hook-type pre-commit --hook-type pre-push --hook-type commit-msg
```

Run everything manually: `pre-commit run --all-files`

## Build (optional, for deployment testing)

Env vars are not baked into the image — pass them at run time:

```sh
docker build -t proteng-user-mgmt .
docker run -d --name proteng-user-mgmt --env-file .env.local --network proteng-net -p 8081:8080 proteng-user-mgmt
```
