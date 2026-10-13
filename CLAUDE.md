# proteng-user-mgmt

Accounts service of ProtEngPlus (Go, Gin, MongoDB): sign up, email verification, login that issues RS256 JWTs, forgot and change password, admin accounts, and job notification email consumed from the RabbitMQ queue `job_status_email_notification`. It is called only through proteng-bff.

**This repo is public.** Never add IPs, hostnames of the deploy machines, NodePorts, runbooks or any credential, including in comments and commit messages. Those live in the private repos manual-guides-2023 and devops-infra.

## Commands

- `make check` before finishing any change: gofmt check, `go vet`, `go test`. CI runs the same target.
- `make fmt`, `make run` (starts Mailhog, then runs from the repo root), `make mailhog-up`, `make build`.
- On Windows a CRLF working tree makes `gofmt -l` list every file; that is a checkout problem, not a code problem.

## Code layout

- `apis/routes`, `apis/controllers`: `/auth/*`, `/users/*`, `/admins/*`. `models/`, `repositories/` (`EnsureIndexes` builds the unique email index at startup).
- `internal/rabbitmq/consumer`: job notification email. `utils/email.go` (`SendEmail`), `utils/token.go`, `utils/password.go`, `utils/email_normalize.go` (`NormalizeEmail`, with tests).
- `templates/*.html` are loaded with a relative glob, so the service must run from the repo root.
- `cmd/backfill-lowercase-emails`: one-off data migration, dry run by default.

## Things that break

- `SMTP_PORT` is an int: an empty value makes envconfig stop the service at startup. Local defaults point at Mailhog (`localhost:1025`).
- `MONGO_DB` defaults to `proteng-dev` in code. Local runs must set their own database name; never point a local run at `proteng-dev` or `proteng-production`.
- The signing key must pair with bff's `ACCESS_TOKEN_PUBLIC_KEY` of the same environment.
- Security: never log email bodies, tokens, reset or verification links, or SMTP settings. `SendEmail` currently prints the message and the dialer; that is a known issue to remove, not a pattern to follow. `/admins/*` has no auth guard in this service.
- Always normalize emails with `NormalizeEmail` before storing or querying.
- New Go functions need a test in a `_test.go` file of the same package, in the same commit.
- A push to `dev` deploys dev and a push to `main` deploys production, even for docs-only changes.

## Team workflow

Issue first with a commit plan, branch from `dev`, Conventional Commits in English with one topic per commit, PR into `dev` using the template in Thai, no emoji anywhere. Full rules: `CONTRIBUTING.md` of manual-guides-2023.
