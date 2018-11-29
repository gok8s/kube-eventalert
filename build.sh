#!/usr/bin/env bash
GOOS=linux GOARCH=amd64 go build -o ealert .
chmod +x ealert
docker build -t xx.xx.com/eventalert:v0.2 .
docker push xx.xx.com/eventalert:v0.2
