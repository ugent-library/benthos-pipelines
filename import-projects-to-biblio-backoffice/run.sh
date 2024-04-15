#!/usr/bin/env bash

export BIBLIO_BACKOFFICE_API_URL="http://localhost:3001/api/v2"
export BIBLIO_BACKOFFICE_API_KEY="xyz"

cat /tmp/projects.json | go run main.go -c config.yaml