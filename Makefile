.PHONY: build
build:
	docker run --rm -v $(CURDIR):/site -w /site --user $(shell id -u) hugo:latest hugo -t pickles
	find ./public -name BUILD.bazel -delete

.PHONY: update-deps
update-deps:
	bazel run //:vendor

CONTENT_DATABASE_ID = 36a84de31af5484ba98d04ac40944d04
CONTENT_DIR = $(CURDIR)/content/posts

.PHONY: update-contents
update-contents:
	bazel run //cmd/site -- update-content --id $(CONTENT_DATABASE_ID) --dir $(CONTENT_DIR)
