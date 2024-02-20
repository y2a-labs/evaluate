#!/bin/bash

wgo -file=.go -file=.templ -xfile=_templ.go -xfile=_.css ./templ generate :: ./tailwindcss -i ./public/input.css -o ./public/output.css :: ./templ fmt . :: go run ./cmd/server/main.go