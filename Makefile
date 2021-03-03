.PHONY: toc dependencies build run

MARKDOWN_FILE?=README.md

toc:
	markdown-toc --prepend'' --indent "    " -i $(MARKDOWN_FILE)

dependencies:
	go mod vendor

build:
	go build -mod=vendor -a -o cal-calc

run:
	./cal-calc
