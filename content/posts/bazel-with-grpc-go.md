---
title: Bazelとgrpc-goを仲良くさせる
date: 2019-12-23
lastmod: 2020-03-30
tags: ["Bazel", "Go", "gRPC"]
isCJKLanguage: true
---

[Bazel](https://bazel.build) をGoのプロジェクトで使っている際に多くの人が悩むのがProtocol Buffersの扱い。
gRPCを使っているリポジトリでBazelとgrpc-goを仲良くさせる方法をご紹介します。

# Bazelとは

[前の記事](../integration-test-with-bazel) を参照してください。

Googleが中心となって開発している

* ビルドを独自のサンドボックス環境の中で行う
* ビルドの再現性が高い（サンドボックスの中で行われるので）
* 高速
* 複数の言語に対応できる
* 拡張性が高い

というような特徴を持ったビルドツールです。

# Protocol Buffersをどう扱うか

[以前書いた](https://medium.com/mixi-developers/bazel%E3%81%A8%E3%83%A2%E3%83%8E%E3%83%AC%E3%83%9D-b901ffba61ce) 方針から変わっていません。
protoファイルは生成されたソースコードもリポジトリに含めてしまいます。
生成されたソースコードをリポジトリに入れてしまうのは主にエディタのコード補完のためです。

生成されたソースコードはサンドボックスの中に閉じ込められてしまうのでサンドボックスから救出してあげる必要があります。
Bazelとは別に ``protoc`` や ``protoc-gen-go`` を用意してそれでソースコードを生成することもできますが、複数の開発者がいる場合はバージョンの差異で余計なことに悩むことになるはずです。
なのでできるだけBazelが用意するProtocol Buffersのコンパイラを使うようにして全員の環境を揃えたいところです。

# 方針

1. Bazelが用意する `protoc` でコンパイルする
1. コンパイル結果をサンドボックスの中から救う

# Bazelにコンパイルさせる

Bazelにコンパイルさせるとサンドボックスの中に閉じ込められてしまいますが、まずはなにはともあれコンパイルさせます。
コンパイルした結果をサンドボックスから救出するという方針でやります。

`gazelle:proto disable_proto` を [指定している](https://github.com/f110/bazel-example/blob/24d674c020ca4895247bd614785fd3d728c33fe6/build/root/BUILD.bazel#L4) のでproto関連のルールは手動で各必要があります。

```python
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "helloworld_proto",
    srcs = ["helloworld.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "helloworld_go_proto",
    compilers = ["@io_bazel_rules_go//proto:go_grpc"],
    importpath = "github.com/f110/bazel-example/tools/rpc/helloworld",
    proto = ":helloworld_proto",
    visibility = ["//visibility:private"],
)
```

ファイルが数個であれば手動で書けるレベルだとは思います。この手のルールを書くことに不慣れな人は gazelle でルールを生成するようにして1度だけ実行してみるといいと思います。

# カスタムルールを作る

サンドボックスから生成されたファイルを救出するためのビルドルールを用意します。 [（全体）](https://github.com/f110/bazel-example/blob/21a356c81341a3bb0cdab41ffe3064e18ecca03d/build/rules/go/proto.bzl)

```python
load("@bazel_skylib//lib:shell.bzl", "shell")

def _proto_gen_impl(ctx):
    generated = ctx.attr.src[OutputGroupInfo].go_generated_srcs.to_list()
    substitutions = {
        "@@FROM@@": shell.quote(generated[0].path),
        "@@TO@@": shell.quote(ctx.attr.dir),
    }
    out = ctx.actions.declare_file(ctx.label.name + ".sh")
    ctx.actions.expand_template(
        template = ctx.file._template,
        output = out,
        substitutions = substitutions,
        is_executable = True,
    )
    runfiles = ctx.runfiles(files = [generated[0]])
    return [
        DefaultInfo(
            runfiles = runfiles,
            executable = out,
        ),
    ]

_proto_gen = rule(
    implementation = _proto_gen_impl,
    executable = True,
    attrs = {
        "dir": attr.string(),
        "src": attr.label(),
        "_template": attr.label(
            default = "//build/rules/go:move-into-workspace.bash",
            allow_single_file = True,
        ),
    },
)

def proto_gen(name, **kwargs):
    if not "dir" in kwargs:
        dir = native.package_name()
        kwargs["dir"] = dir

    _proto_gen(name = name, **kwargs)
```

ファイルを救出するだけなのでやっていることは非常に単純でファイルをコピーするだけです。
ただし救出するファイル自体もサンドボックスに閉じ込める必要があります。
どういうことかと言うと、Bazelはターゲットごとに別のサンドボックスを用意します。
つまりファイルをコピーするターゲットに生成されたソースコードを含めなければいけません。
そのために少しビルド定義を書く必要があります。

[サンプルリポジトリ](https://github.com/f110/bazel-example) で上のルールを実行した時のサンドボックス内は以下のようになります

```
bazel-bin/tools/rpc/helloworld/gen.sh.runfiles/__main__
└── tools
    └── rpc
        └── helloworld
            ├── gen.sh
            └── linux_amd64_stripped
                └── helloworld_go_proto%
                    └── github.com
                        └── f110
                            └── bazel-example
                                └── tools
                                    └── rpc
                                        └── helloworld
                                            └── helloworld.pb.go
```

`helloworld.pb.go` のパスは `ctx.attr.src[OutputGroupInfo].go_generated_srcs.to_list()[0].path` に入っています。
コピー先は `native.package_name()` を取ることでWORKSPACEからターゲットまでのパスを手に入れることができます。
あとはこれらを組み合わせてファイルをコピーします。簡単ですね！

```python
load("//build/rules/go:proto.bzl", "proto_gen")

proto_gen(
    name = "gen",
    src = ":helloworld_go_proto",
    visibility = ["//visibility:public"],
)
```

`src` に `go_proto_library` のターゲットを指定するだけです。
サンプルリポジトリであれば `bazel run //tools/rpc/helloworld:gen` で `helloworld.pb.go` がリポジトリ内にコピーされてきます。

# ワンライナーを用意して仕上げ

protoが複数のパッケージに分散している場合など、いちいち `bazel run` していくのは面倒なのでワンライナーを用意しておきましょう。

```console
$ bazel query 'attr(generator_function, proto_gen, //...)' | xargs -n1 bazel run
```

これでリポジトリ内の `proto_gen` を全て実行することができます。
このクエリ言語の強さもBazelの特徴です。

# 全員が同じprotocを使える

最初に書いたようにここまで作ると `protoc` までBazelが用意します。
Goの場合はGoのランタイムもBazelが用意するので、なんとサンプルリポジトリのフルビルドに必要なのはBazelだけです。
BazelがGoのランタイムもprotocもそのプラグインも全て用意します。

全てBazelが用意してくれるというのは非常に楽で、Bazelをインストールしリポジトリを持ってくればビルドできます。
この程度であれば環境構築に悩む必要がなく、その上全員の環境を統一できるメリットもあります。（MySQLなどを使うプロジェクトだとさらに工夫が必要になりそうですが）
リポジトリに長く関わっている人はソフトウェアの設計も把握しているので環境構築を一からやるのも難しくないでしょう。むしろ簡単というような感想になるはずです。
しかし初めてそれに触れる人は何も分からないので先人たちが書いた手順書に従って環境構築をするしかないのです。
Bazelはそういったコストも抽象化したプログラムに落とし込めるツールです。

# まとめ

Bazelでの扱いに少し苦労するProtocol Buffersの扱い方を紹介しました。
サンドボックス内でファイルを生成し、それを救出することでリポジトリへソースコードをコピーしています。これによりIDEで補完を行えるようになりますし、 `go test` でテストを実行することもできるようになります。
`go test` でテストの実行ができるとGoLandから簡単にテストが実行できたりと開発効率の向上につながることでしょう。

今回サンドボックス内で生成したファイルをサンドボックスから取り出すビルドルールを書いたので多少の修正でProtocol Buffers以外にも使えるかもしれません。

Goのコア側でこのようなサンドボックスに入ったソースコードの情報をexposeするツールが開発中のようです。
そちらが動作するようになりIDEがサポートしてくれればそちらへ移行する方がスマートではありますが現状はリポジトリへソースコードをコピーしてくる他ありません。

（ビルドルールの詳細な書き方については解説しません。特にインターフェースについてはどんどん変わっていくためその時のバージョンに合わせてオフィシャルドキュメントを見る方がいいと思います。
Bazel独特のフェーズの動作についてはいつか解説します。）
