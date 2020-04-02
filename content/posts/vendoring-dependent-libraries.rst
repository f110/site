---
title: 依存ライブラリは常にvendoringする
date: 2020-03-31
tags: ["Bazel", "Go"]
isCJKLanguage: true
---

BazelでGoの依存ライブラリを管理する方法は2つあります。

#. go.mod からWORKSPACEファイルを生成する
#. vendorディレクトリを使う

以前はリポジトリに ``go.mod`` ファイルが1つしかない（単一モジュール構成のリポジトリ）の場合は前者を、
複数の ``go.mod`` ファイルがある場合は後者の方法を取るのがいいと思っていました。

しかし最近は常に後者を選択する方がいいと思っています。

見えない依存
================

Bazel で Go をビルドする際は rules_go [#rulesgo]_ と gazelle [#gazelle]_ を利用していることでしょう。

rules_go 自身が `依存しているライブラリ <https://github.com/bazelbuild/rules_go/blob/4a42b4092abdc60d14419a79afaec3659fbceb26/go/workspace.rst#go-rules-dependencies>`_ もBazelによって管理されるため同じライブラリに自分のソフトウェアが依存している場合は競合します。
うまく競合しなかった場合は問題ありません。その時は幸せに利用できます。

ですが、競合して2つのバージョンを保持しないといけなくなった場合、 gazelle でそれを扱うのは突然難しくなりますし、そうなっていることを gazelle に伝える必要があります。
本来であれば ``go.mod`` ファイルを ``go get`` コマンドなどで更新してそこから依存ライブラリの定義ファイルを自動生成するだけで使えるはずですがこの場合においてはそれはできません。

増える依存
-------------

rules_go で単純な Go のアプリケーションをビルドするだけではなく、ProtobufやgRPCを扱っている場合はさらに依存が増えます。

`gRPCの依存ライブラリ <https://github.com/bazelbuild/rules_go/blob/4a42b4092abdc60d14419a79afaec3659fbceb26/go/workspace.rst#grpc-dependencies>`_ を追加すると準公式ライブラリへの依存がさらに増えます。

他にも ``github.com/golang/protobuf`` への依存もあったりと、自分のソフトウェアと共有したくないものがいくつもあります。

余談: golang.org/x/ の立ち位置
++++++++++++++++++++++++++++++++++

``golang.org/x/net`` や ``golang.org/x/text`` は果たして準公式ライブラリなのでしょうか？
これらの ``golang.org/x/...`` のライブラリは実はGoのコアと一緒に `配布されています。 <https://github.com/golang/go/tree/master/src/vendor/golang.org/x>`_

これを見ると実質標準ライブラリと言えるかもしれません。ただこの依存はGoのコア用なので使うことを期待してはいけません。

なぜ共有したくないのか
--------------------------

rules_go のバージョンアップでこれらのライブラリも更新されます。

すると場合によっては自分のソフトウェアのビルドが壊れます。

自分が使いたいバージョンのライブラリを使えないということも起こります。
go コマンドで実行している場合はうまく動くのにBazelでビルドしようとするとビルドができないといった事態になります。

vendorしちゃう
================

この問題もvendoringすると起きません。

gazelleは適切にvendorディレクトリへの依存をハンドリングできるので、自分のソフトウェアが使うライブラリは全てvendorディレクトリに入れておくと確実にビルドできます。
また rules_go のアップデートでビルドを壊すことがありません。

ということで Polyrepo 構成のリポジトリであっても最近はvendorを使うようになりました。
vendorするルールを非常によく使うようになったのでそれぞれのリポジトリで同じルールを書かずに `共通ルールとして切り出して <https://github.com/f110/rules_extras/blob/master/go/vendor.bzl>`_ 利用しています。

リンク
=======

.. [#rulesgo] Go rules for Bazel https://github.com/bazelbuild/rules_go
.. [#gazelle] Bazel build file generator for Bazel Project https://github.com/bazelbuild/bazel-gazelle
