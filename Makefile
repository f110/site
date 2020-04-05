.PHONY: build
build:
	docker run --rm -v $(CURDIR):/site -w /site hugo:latest hugo -t pickles