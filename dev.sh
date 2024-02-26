#!/bin/bash
wgo -file=.go -file=.templ -xfile=_.css -xfile=_templ.go ./templ generate :: go run ./cmd/server/main.go :: ./tailwindcss -i ./public/input.css -o ./public/output.css --watch