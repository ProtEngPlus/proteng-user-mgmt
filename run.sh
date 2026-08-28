#!/usr/bin/env bash
# Starts a local SMTP catcher (Mailhog) if not already running
# Web inbox: http://localhost:8025 - matches SMTP_HOST=localhost / SMTP_PORT=1025 in .env.local.
if [ -z "$(docker ps -a -q -f name=^mailhog$)" ]; then
  docker run -d --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog >/dev/null
elif [ -z "$(docker ps -q -f name=^mailhog$)" ]; then
  docker start mailhog >/dev/null
fi

ENV=local go run main.go
