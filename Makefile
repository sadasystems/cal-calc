.PHONY: toc deps run

MARKDOWN_FILE?=README.md

toc:
	markdown-toc --prepend'' --indent "    " -i $(MARKDOWN_FILE)

run:
	@go run main.go
