load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "build_stack_rules_hugo",
    url = "https://github.com/f110/rules_hugo/archive/741c3abf624f2061c4118b265d34ecd9b75dd09b.zip",
    sha256 = "28aea3459aa9c88f065ee3e851b9571ecc6cb2f504d4ef63b1244eaf3d704873",
    strip_prefix = "rules_hugo-741c3abf624f2061c4118b265d34ecd9b75dd09b",
)

http_archive(
    name = "rules_pkg",
    urls = [
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.2.6-1/rules_pkg-0.2.6.tar.gz",
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.2.6/rules_pkg-0.2.6.tar.gz",
    ],
    sha256 = "aeca78988341a2ee1ba097641056d168320ecc51372ef7ff8e64b139516a4937",
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

load("@build_stack_rules_hugo//hugo:rules.bzl", "hugo_repository")

hugo_repository(
    name = "hugo",
    version = "0.76.5",
    sha256 = "38f1d92fb8219168e684f0b82faef3aea0f3d1bd89752ec2179b41fb9eceea17",
)
