#!/usr/bin/env bash

export MONGODB_URL="mongodb://127.0.0.1:27017"
export MONGODB_DB="authority"
export MONGODB_COLLECTION="project"
export PROJECTS_API_KEY="xyz"
export PROJECTS_API_ADD_PROJECT="http://localhost:3000/api/v1/add-project"
export PROJECTS_API_DELETE_PROJECT="http://localhost:3000/api/v1/delete-project"

benthos -c config.yaml
