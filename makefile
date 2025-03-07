#!/bin/bash

# this is for test on you mac
run:
	go run ./cmd/main.go

# this is for test on you mac
build-mac:
	# GOOS=darwin GOARCH=arm64 go build -o ePhotoPlayer_mac_arm64 ./cmd/main.go

## 這個 --platform linux/arm64 是我運行的環境是 M1 mac
build-image-raspberry:
	docker buildx build --platform linux/arm64 -t my/fyne-cross-raspberry:test ./raspberry

build-raspberry:
	fyne-cross linux -arch arm -image my/fyne-cross-raspberry:test -debug -name ePhotoPlayer_pi ./
