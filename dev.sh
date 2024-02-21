#!/bin/bash
wgo -file=.go -file=.templ -xfile=_templ.go -xfile=_.css ./templ fmt . :: ./templ generate :: go run ./cmd/server/main.go &
./tailwindcss -i ./public/input.css -o ./public/output.css --watch