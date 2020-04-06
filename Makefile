.PHONY: build
build:
	docker run --rm -v $(CURDIR):/site -w /site --user $(shell id -u) hugo:latest hugo -t pickles
