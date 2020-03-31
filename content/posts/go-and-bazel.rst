---
title: GoとBazel
date: 2019-10-23
tags: ["Bazel", "Go"]
isCJKLanguage: true
---

本エントリは `オリジナル <https://medium.com/mixi-developers/go-project-with-bazel-ad807ba19f5c>`_ のコピーです。

以前、 `モノレポ構成にしてビルドツールとしてBazelを利用している <https://medium.com/mixi-developers/bazelとモノレポ-b901ffba61ce>`_ ことを紹介しました。

そのBazelは10月10日にとうとう1.0を `リリースしました！ <https://blog.bazel.build/2019/10/10/bazel-1.0.html>`_
バージョン1.0に到達したというニュースは日本語のニュースサイトでも掲載されるなど、多少注目を浴びたようです。

そこで今回はGoのプロジェクトのビルドツールとしてBazelを利用する例をご紹介します。
GoのプロジェクトではGNU makeをビルドツールとして使われていることが多いと思いますが、一度Bazelに慣れてしまうと手放せないツールになります（なっています）。
導入を検討する際のなにかの参考になりそうなTipsをいくつかご紹介します。

BazelをGoのプロジェクトで使うことのメリット
=============================================

#. Bazel がすべてを用意してくれるので最悪 ``go`` コマンドが入ってなくてもビルドできる。
#. コンテナとしてshipする際に非常に軽量なイメージをdockerを使わずにビルドできる
#. debやrpmパッケージ自体もクロスコンパイルできる（macOSからbazelだけでdebパッケージが作れます）

BUILDファイルの生成
=======================

BazelはBUILDファイルにビルドに必要な情報がすべて書かれています。
依存関係や必要なファイルもすべて記述する必要があります。
これは一見面倒に見えますが、Goのプロジェクトの場合はソースコードを解析して自動生成することが可能です。
（実際、全てを手で書こうと思うとかなり面倒です。なので基本的には自動生成してしまうのがいいと思います。）

これには `gazelle <https://github.com/bazelbuild/bazel-gazelle>`_ を使います。
導入方法はシンプルなのでREADMEの通りにWORKSPACEファイルを作るだけです。

以下にWORKSPACEファイルの例を示しますが、 Bazel も gazelle も rules_go もアップデートが早いプロジェクトですので最新のREADMEを確認してください。

.. code:: python

    load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

    http_archive(
        name = "io_bazel_rules_go",
        urls = [
            "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/v0.20.1/rules_go-v0.20.1.tar.gz",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.20.1/rules_go-v0.20.1.tar.gz",
        ],
        sha256 = "842ec0e6b4fbfdd3de6150b61af92901eeb73681fd4d185746644c338f51d4c0",
    )

    http_archive(
        name = "bazel_gazelle",
        urls = [
            "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.0/bazel-gazelle-v0.19.0.tar.gz",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.0/bazel-gazelle-v0.19.0.tar.gz",
        ],
        sha256 = "41bff2a0b32b02f20c227d234aa25ef3783998e5453f7eade929704dcff7cd4b",
    )

    load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
    load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

    go_rules_dependencies()

    go_register_toolchains()

    gazelle_dependencies()

依存ライブラリ
=================

依存ライブラリを管理する方法は2つあります。
一つは依存ライブラリの管理まで Bazel に任せてしまう方法、もう一つは vendor として一緒にコミットしてこれまでと同じ方法でビルドする方法です。

自分は後者のvendorとしてしまう方法が好みでよくこちらの方法を使っています。

前者の方法は ``go.mod`` ファイルからBazelによる依存関係の定義を生成します。
これも gazelle でできますが、エディタからテストを走らせる際とBazelでビルドする際で2つのキャッシュを保存することになってしまいます。（Bazelのキャッシュディレクトリの中とGoのキャッシュの2つ）
``go.mod`` ファイルがベースになっているのでこの2つは同じではありますが、少し気持ち悪いと感じてしまいます。

``go.mod`` からBazelの依存関係の定義へ変換する場合は ``gazelle`` の ``update-repos`` を使います。

.. code:: python

    load("@bazel_gazelle//:def.bzl", "gazelle")

    # gazelle:prefix github.com/example/project
    gazelle(name = "gazelle")

このようなBUILD.bazelをWORKSPACEと同じ階層に置いておけば

.. code:: shell

    $ bazel run //:gazelle -- update-repos -from_file=go.mod

``update-repos`` を引数に ``run`` することでWORKSPACEファイルを自動で更新することができます。

一方 vendoring する場合はこの特殊操作をすることなく

.. code:: shell

    $ go mod vendor
    $ bazel run //:gazelle -- update

でBUILD.bazelファイルの更新のみを行うだけで済みます。

ただし ``go mod vendor`` コマンドはBUILD.bazelを削除しないため、依存しているライブラリでもBazelを使っている場合はコンフリクトするかもしれません。
（例えば ``grpc-gateway`` 等はコンフリクトするので ``gazelle update`` する前に BUILD.bazel を一掃しておいた方がいいです）

vendorディレクトリを作るGoのランタイム
=======================================

Goでプログラムを書かれるほぼすべての方がGoの処理系を何らかの方法でインストールしていると思います。
なので ``go`` コマンドがどのバージョンなのかは個人の環境によって違う可能性が高いです。
vendorディレクトリを作る時などにこのバージョン違いの影響を受けることを避けるため、vendorディレクトリは Bazel がダウンロードしてきたランタイムを使うようにしています。

.. code:: python

    load("@io_bazel_rules_go//go:def.bzl", "go_context", "go_rule")
    load("@bazel_skylib//lib:shell.bzl", "shell")

    def _go_vendor(ctx):
        go = go_context(ctx)
        out = ctx.actions.declare_file(ctx.label.name + ".sh")
        substitutions = {
            "@@GO@@": shell.quote(go.go.path),
            "@@GAZELLE@@": shell.quote(ctx.executable._gazelle.short_path),
        }
        ctx.actions.expand_template(
            template = ctx.file._template,
            output = out,
            substitutions = substitutions,
            is_executable = True,
        )
        runfiles = ctx.runfiles(files = [go.go, ctx.executable._gazelle])
        return [
            DefaultInfo(
                runfiles = runfiles,
                executable = out,
            ),
        ]

    go_vendor = go_rule(
        implementation = _go_vendor,
        executable = True,
        attrs = {
            "_template": attr.label(
                default = "//build/rules/go:vendor.bash",
                allow_single_file = True,
            ),
            "_gazelle": attr.label(
                default = "@bazel_gazelle//cmd/gazelle",
                executable = True,
                cfg = "host",
            ),
        },
    )
    
これがビルドルールで

.. code:: bash

    #!/usr/bin/env bash

    GO=@@GO@@
    GAZELLE_PATH=@@GAZELLE@@

    RUNFILES=$(pwd)
    GO_RUNTIME="$RUNFILES"/"$GO"
    GAZELLE="$RUNFILES"/"$GAZELLE_PATH"

    cd "$BUILD_WORKSPACE_DIRECTORY"
    "$GO_RUNTIME" mod tidy
    "$GO_RUNTIME" mod vendor
    find vendor -name BUILD.bazel -delete
    "$GAZELLE" update

これがルールから実行されるシェルスクリプトのテンプレートです。
Bazelのビルドルールを書くことに慣れてない場合、よくわからないかもしれません。
しかし内部でやっていることはそんなに難しくありません。

#. ``go`` コマンドのパスを手に入れる（bazelのキャッシュディレクトリのどこかに入っている ``go`` コマンドです）
#. gazelleのパスを手に入れる
#. それらのパスをテンプレートに埋め込み、 ``go`` と ``gazelle`` を実行する

これだけです。
``go`` と ``gazelle`` が生成したシェルスクリプトの実行に必要だと定義されているため、Bazelはもしこれらのバイナリがまだなければコンパイルやインターネットからの取得を行います。

通常、Bazelからシェルスクリプトを実行した場合、サンドボックスの中で実行されます。
そのままだと期待通り動作しないため、コマンドを実行する前にサンドボックスからescapeしています。

これらのファイルをよしなに配置しておき、それをロードすれば

.. code:: python

    load("//build/rules/go:vendor.bzl", "go_vendor")

    go_vendor(name = "vendor")

このような定義を書いておくだけで ``bazel run //:vendor`` で ``go mod vendor`` が実行されBazelのビルドファイルもアップデートされます。

すべてをBazelでやろうとしない
================================

ですが、すべてをBazelでやろうとするとうまくできないことがありフラストレーションがたまるかもしれません。

Goのソースコードを編集するにはIDEなりが便利だったりしますが、エディタとBazelの連携はまだまだだと思います。
GoLand等ではテストを書いてる際、ワンクリックで編集してる部分のテストだけを実行できると思いますがここに更にBazelを組み合わせるのは現状は難しいように思います。

Protocol BuffersのコンパイルもBazelでできますが一工夫した方が現状はより扱いやすくなります。
例えばprotocの生成したファイルはサンドボックスの中に閉じ込められるためIDEからは発見できません。
ですのでこういうコード生成が必要だった場合は生成されたファイルもリポジトリに含めてしまい、エディタにはそのファイルを発見させてます。

一方生成物をリポジトリに含めずにBazelがビルドする際に定義ファイルから生成される状態にしておけば、最新ビルドが一つ前のコミットの定義ファイルを使っていたという事故は防ぐことができます。

生成物をリポジトリに含めると、リポジトリに入っているファイルの更新を忘れるという問題が起きますがこれはリポジトリへのPushをフックしてファイル生成を自動化し勝手にコミットするようにすることで解決できます。
（決してローカルリポジトリにフックを設定することを強要してはいけません。必ずリモート側で勝手にやってしまうようにしましょう）

テストの書き方
================

Bazelはユニットテストを実行するのは得意です。
ですが、インテグレーションテストを実行するには工夫が必要です。
なので導入初期はなるべくユニットテストのみを書くようにした方がいいです。
Bazelのビルドルールなどを書くことに慣れてきた・抵抗がなくなってきた時に初めてインテグレーションテストの実行を考えるといいと思います。

Bazel流のインテグレーションテストの実行は必要なミドルウェアもBazelを通して準備することだと考えています。
これを実現するには

#. 依存ミドルウェアの取得とビルド（各プラットフォームと各バージョンにも対応できるとより良い）
#. ミドルウェアのパスをテストスイートに何らかの方法で渡す
#. テストスイート内でミドルウェアを実行して使い終わったら終了する

というようなことが必要になるでしょう。

一度は使ってみてほしいBazel
==============================

Bazelは導入するために多くのことを学ぶ必要があります。

もし、会社などですでにBazelを使われているリポジトリがあればそういうものを参考にして個人的なリポジトリに導入してみたりするのがいいでしょう。
会社であればそれを導入した人に直接話しを聞いたりできるかと思うのでそうやって学ぶというのも一つの手です。

OSSにもBazelを採用しているプロジェクトがいっぱいあるのでgithubには参考にできるようなものがたくさんあります。
Bazelにやらせたいと思っていることを実現しているプロジェクトを探して真似してみたりするとルールの書き方が分かってくると思います。

The Go gopher was designed by `Renee French <https://reneefrench.blogspot.com>`_
