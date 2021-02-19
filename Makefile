.PHONY: toc deps run

MARKDOWN_FILE?=README.md

toc:
	markdown-toc --prepend'' --indent "    " -i $(MARKDOWN_FILE)

deps:
	go get -u golang.org/x/oauth2/google
	go get -u google.golang.org/api/calendar/v3

run:
	go run quickstart.go
