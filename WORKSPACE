load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "bfc5ce70b9d1634ae54f4e7b495657a18a04e0d596785f672d35d5f505ab491a",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.40.0/rules_go-v0.40.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.40.0/rules_go-v0.40.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "727f3e4edd96ea20c29e8c2ca9e8d2af724d8c7778e7923a854b2c80952bc405",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.30.0/bazel-gazelle-v0.30.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.30.0/bazel-gazelle-v0.30.0.tar.gz",
    ],
)

http_archive(
    name = "build_stack_rules_hugo",
    sha256 = "f076f8098d95e4d3636918eed0b8f09c252f62ac992ba5e396f10c6cf2137849",
    strip_prefix = "rules_hugo-2927451ff7fff708292eb7eb68ca278457c5dd78",
    url = "https://github.com/stackb/rules_hugo/archive/2927451ff7fff708292eb7eb68ca278457c5dd78.zip",
)

http_archive(
    name = "rules_pkg",
    sha256 = "aeca78988341a2ee1ba097641056d168320ecc51372ef7ff8e64b139516a4937",
    urls = [
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.2.6-1/rules_pkg-0.2.6.tar.gz",
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.2.6/rules_pkg-0.2.6.tar.gz",
    ],
)

git_repository(
    name = "dev_f110_rules_extras",
    commit = "23175122db6205204ff30291dc9a2d62752a1862",
    remote = "https://github.com/f110/rules_extras",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.17.2")

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

load("@build_stack_rules_hugo//hugo:rules.bzl", "hugo_repository")

hugo_repository(
    name = "hugo",
    sha256 = "38f1d92fb8219168e684f0b82faef3aea0f3d1bd89752ec2179b41fb9eceea17",
    version = "0.76.5",
)

hugo_repository(
    name = "hugo_darwin_arm64",
    sha256 = "50f7ce43657bf7cfb549c492d43edcfebf05098a23dda14b7dc9fee12711b4ac",
    version = "0.76.5",
)
