.PHONY: build
build:
	bazel clean
	bazel build //:site_tar
	rm -rf public/*
	tar xf bazel-bin/site_tar.tar -C public --strip-components=2
	chmod -R 755 public/*

.PHONY: update-deps
update-deps:
	bazel run //:vendor

CONTENT_DATABASE_ID = 36a84de31af5484ba98d04ac40944d04
CONTENT_DIR = $(CURDIR)/content/posts

.PHONY: update-contents
update-contents:
	bazel run //cmd/site -- update-content --id $(CONTENT_DATABASE_ID) --dir $(CONTENT_DIR)
