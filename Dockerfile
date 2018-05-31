from golang:1.10-alpine

run apk update
run apk add git

run go get github.com/femnad/stuff/...
