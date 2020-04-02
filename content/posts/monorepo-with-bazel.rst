---
title: Bazelとモノレポ
date: 2019-05-07
tags: ["Bazel", "Go", "Monorepo"]
isCJKLanguage: true
---

本エントリは `オリジナル <https://medium.com/mixi-developers/bazel%E3%81%A8%E3%83%A2%E3%83%8E%E3%83%AC%E3%83%9D-b901ffba61ce>`_ の一部を再編集して掲載しています。（2020/03/31）

.. section-numbering::
.. contents::
    :local:

モノレポのメリット
===================

gitでこのようなリポジトリ運用をされている方はそんなに多くはないのではないでしょうか？
むしろこのような運用は嫌われる傾向にあるかもしれません。
それでもモノレポを使うのには理由があります。

#. リファクタリングが楽
#. コードレビューが楽
#. コンフリクト地獄に陥らない
#. 一つのリポジトリですべて手に入る

そして

* リポジトリに関わる人にすべてを触ってほしい

という思いもあります。

一方、モノレポを実現するためにはいくつか工夫をする必要があります。
今までモノレポで運用をしていなかった場合ビルドツール等はモノレポに耐えられないものが多いと思います。
モノレポの運用をする場合はその点を考え直す必要があります。

モノレポを運用していくにあたって大事なのはビルドツールだと考えています。
複数のソフトウェアが一つのリポジトリに入ってお互いが依存している状態であるため、それらをうまく扱えるビルドツールでなければいけません。
また規模も大きくなるのでビルドが高速であった方が嬉しいなどビルドツールに求める水準がそうじゃない場合に比べて高くなります。

モノレポとして扱うビルドツールは Bazel [#bazel]_ というものがあります。

Bazel
==========

Googleが社内で使っているBlazeというビルドツールのOSS実装です。（だそうです。私はGoogleの社内を直接見たことがないので真偽は分かりません）

ビルドルールの定義をStarlark [#starlark]_ と呼ばれるPython3のサブセット言語で行うことが特徴です。
ビルドツール自体の挙動もStarlarkで定義されるため新しい言語や独自の方法でのビルドもそれ自体を書けば対応することができます。

Bazelと同様なコンセプトのビルドツールに Pants [#pants]_ 、 Buck [#buck]_ や Please [#please]_ があります。
いずれもやはりBlazeのコンセプトを参考に実装されているもののようです。

Bazelの細かい使い方などは公式のドキュメント [#bazeldocs]_ を参照してください。
ここでは細かい使い方までは解説しません。

特徴
------

Bazelはビルドルールをサブセット言語で行うだけではなく色々な特徴を持っています。

#. ビルドを独自のサンドボックス環境の中で行う
#. ビルドの再現性が高い
#. ビルドが高速
#. 複数の言語に対応できる
#. 拡張性が高い

サンドボックス環境でビルドを行うためビルドの再現性が高いです。（各言語のビルドルールも再現性が高くなるようにされています）
同一のソースコードからであれば基本的に同じ結果が得られます。
そのため一度コンパイルした結果などはキャッシュされ2回目以降はキャッシュを使います。

このキャッシュを全く別のマシンと共有することもできます。

自分の使い方
--------------

自分のリポジトリではBazelを使って

* Go で書かれたツールのビルド
* Go のテスト
* GitHub Releaseで配布されているソフトウェアのパッケージ化
* 設定ファイルの自作テストスイートのビルドと実行
* コンテナの作成

を行っています。

なぜBazelを使うのか
======================

モノレポのメリットを最大限享受したいのでBazelを利用しています。

例えばいくつかのツールが依存するライブラリのコードをリファクタリングしたくなったとしましょう。
この時、インターフェースを変えないようなリファクタリングであればそんなに問題ないかもしれません。
ですが往々にしてインターフェースは後から変えたくなります。
最初からそれを前提にインターフェースを作るのもあまり綺麗とは言えません。

Goのような言語の場合、インターフェースを大きく変えるようなリファクタリングも行いやすいのでこのメリットを常に受けられる状態にしておきたいです。
リポジトリがライブラリと利用側で分かれているとこのようなリファクタリングが行いにくくなりますし、その変更をレビューする側も大変です。
複数のリポジトリへPRを作り、それらの整合性を維持し、またマージのタイミングも考慮しないといけないかもしれません。
PRが3つくらいならまだそれらを把握して維持できるでしょう。ですがPRが数十個になったらどうでしょうか？
少なくとも自分はそもそもそんな数のPRを作りたいと思わないのでリファクタリングを諦めるでしょう。

他にもライブラリ側のコードを修正したらそれを利用している側のソフトウェアのテストを実行したいでしょう。
このような場合でも非常に力を発揮できます。

Bazelのビルドルールにはソフトウェアの依存関係が記述されています。
そのため依存先が変更された場合は依存元のテストも行われます。

コミットごとに全ツールのテストを行えばこのような問題は気になりません。
しかしツールが大量にあり、テストに時間がかかるようになったらコミットごとに全テストを実行してられるでしょうか。
そもそも変更されてない部分が大半なのに毎回それらもテストするのは時間と計算資源の無駄遣いでしかないのではないでしょうか。

Bazelはテスト結果もキャッシュされます。
キャッシュの範囲内が変更されていなければそのテストは実行されずに前の結果が使われます。

BazelとGo
============

上でも若干触れていますがBazelはビルドルールに依存関係が書かれています。

.. code:: python
    :number-lines: 1

    load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

    go_library(
        name = "go_default_library",
        srcs = ["hello.go"],
        importpath = "github.com/f110/bazel-example/lib/hello",
        visibility = ["//visibility:public"],
    )

    go_test(
        name = "go_default_test",
        srcs = ["hello_test.go"],
        embed = [":go_default_library"],
    )

Goの場合はimport文から生成することができます。
これは gazelle [#gazelle]_ で行っています。

モノレポの場合は少し工夫する必要があるかもしれません。

各ツールが依存しているライブラリのバージョンがそれぞれで別です。
つまり各ツールごとに ``go.mod`` ファイルが存在します。

BazelはWorkspaceという単位で外部のリポジトリに依存を定義することができるので単一の ``go.mod`` ファイルが存在する場合はそちらの方法で依存を定義しておくでしょう。
（WORKSPACEファイルをgazelleでアップデートしていく。 ``update-repos`` を使う方法）
しかし複数の ``go.mod`` ファイルがある場合はこれはうまく機能しません。もしくは機能させるために工夫が必要です。

そこで我々はvendoringをしています。
Go Modulesでもvendoringは使えるので各ツールはvendorディレクトリを持っていて依存しているソースコードも **全部コミットしています** 。

新たに依存モジュールを追加する場合は以下のように行っています。（go 1.12の場合）

.. code:: shell

    $ GO111MODULE=on go get github.com/google/go-github/v25/github
    $ GO111MODULE=on go mod vendor
    $ bazel run //:gazelle -- update

vendoringをしているのでリポジトリをCloneしてくれば依存ライブラリをダウンロードしてこなくてもビルドできます。
更にコンパイルに使われるGolangはBazelがダウンロードしてきます。
つまりリポジトリのCloneとBazelのインストールさえ行えばBazel管理下のツールはすべてビルドできます。

vendoringをしているとリポジトリの容量が気になるかもしれません。
確かにClone時はちょっと転送量が多いかもしれません。それでも ``.git`` ディレクトリはいまのところ100MB程度ですので現代のインターネット回線であればそれほどストレスはないでしょう。

一方PRの差分が大きくなってしまうという問題はあります。
ですがこれは差分を見るツール側の問題であるのでこの場では無視します。

基本的なファイル構成
----------------------

Bazelを初めて使った時はなかなかサンプルも少なくちょっと悩んだりもしました。

そこでサンプルのリポジトリを用意しました。

https://github.com/f110/bazel-example

.. code::

    ├── build
    │   └── root
    ├── debian_packages （debianパッケージのビルドルール）
    │   └── mysqld_exporter
    ├── lib （ライブラリ用のディレクトリ）
    │   └── hello
    ├── tools （大小さまざまなツール）
    │   ├── helloworld1
    │   ├── helloworld2
    │   └── helloworld3
    ├── BUILD.bazel -> build/root/BUILD.bazel
    └── WORKSPACE -> build/root/WORKSPACE

helloworld1は何にも依存していないツールです。
helloworld2は ``lib/hello`` に依存しています。
helloworld3は外部のライブラリに依存しておりvendoringされています。

リポジトリのrootに ``WORKSPACE`` と ``BUILD.bazel`` の2つのファイルを置きます。
この2つにはリポジトリ全体で使われるルールなどが書かれています。

具体的には ``WORKSPACE`` ファイルには ``rules_go`` や ``gazelle`` などの依存がかかれています。

最低限、この2つのファイルを準備すれば後は通常通りファイルを配置していくだけです。
自分でファイルを作ったり依存を増やした時に ``bazel run //:gazelle -- update`` を実行すれば各ファイルのimport文をパースし適切なビルドファイルを生成してくれます。

実際の動作
-----------

まずはmasterブランチでテストを実行してみてください。
初回は依存しているツールなどをダウンロードするため多少時間がかかります。

.. code:: shell

    $ bazel test //...
    INFO: Analysed 11 targets (56 packages loaded, 6879 targets configured).
    INFO: Found 8 targets and 3 test targets...
    INFO: Elapsed time: 2.022s, Critical Path: 1.41s
    INFO: 39 processes: 39 linux-sandbox.
    INFO: Build completed successfully, 71 total actions
    //lib/hello:go_default_test                              PASSED in 0.1s
    //tools/helloworld1:go_default_test                      PASSED in 0.1s
    //tools/helloworld2:go_default_test                      PASSED in 0.1s

テストの中身は空なのでこれは成功します。

次に `このような <https://github.com/f110/bazel-example/commit/3331200a8809587f7f8a7c1a74f5a92ae8030f85>`_ リファクタリングを行ったとしましょう。
この ``Println`` 関数は helloworld2 が使用しています。なのでこれだけでは当然helloworld2のビルドに失敗する状況です。
（この依存関係もBazelのQueryで取り出すことができます）

refactoringブランチに切り替えて同様にテストを実行しようとするとビルドができずテストに失敗する様子をみることができます。

.. code:: shell

    $ git checkout refactoring
    $ bazel test //...
    INFO: Analysed 11 targets (0 packages loaded, 0 targets configured).
    INFO: Found 8 targets and 3 test targets...

    Use --sandbox_debug to see verbose messages from the sandbox
    compile: error running compiler: exit status 2
    4f4e60651d05cfbd821556564b8b40e6/sandbox/linux-sandbox/4/execroot/__main__/tools/helloworld2/main.go:6:15: not enough arguments in call to hello.Println
            have (number)
            want (int, int)
    INFO: Elapsed time: 0.378s, Critical Path: 0.19s
    INFO: 4 processes: 4 linux-sandbox.
    FAILED: Build did NOT complete successfully
    //tools/helloworld1:go_default_test                  (cached) PASSED in 0.1s
    //lib/hello:go_default_test                                NO STATUS
    //tools/helloworld2:go_default_test                        NO STATUS

    Executed 0 out of 3 tests: 1 test passes and 2 were skipped.
    FAILED: Build did NOT complete successfully

helloworld2はビルドに失敗したログが出ているのがわかるかと思います。
helloworld1はライブラリに依存していないのでテスト結果はキャッシュされたものが利用されます。

リポジトリに入っているソフトウェアにちゃんとテストが書かれていればテストを実行するだけでリファクタリングの確かさをある程度は確認することができます。
この例ではビルドが失敗する例でしたが、ロジックの変更でも同じようにテストで問題を発見することができると思います。

BazelとProtocol Buffers
=========================

ツールの中にはIDLとしてProtocol Buffersを使っているものもあります。
``.proto`` ファイルからGoのソースコードを生成しているものもありますし、生成していないものもあります。

BazelはProtocol Buffersをサポートしているのでコンパイルを行うこともできます。
ですがこれは **使ってません** 。

これはコンパイルされたファイルがサンドボックスの中に閉じ込められてしまいIDEから参照できないためです。
将来的にはIDEから参照できるようになるような気配もありますが現在はできません。
そのためprotoファイルのコンパイルはそれぞれツールをインストールしてもらいコンパイルした結果も **コミットしています。**

生成物をリポジトリに入れたくないという人もいるかと思いますがこれらは **入れてしまった方が楽です。**

.. code:: python
    :number-lines: 1

    load("@bazel_gazelle//:def.bzl", "gazelle")

    # gazelle:proto disable_global

リポジトリのルートに上記のようなBUILD.bazelファイルを置いてリポジトリ全体でprotoファイルのコンパイルを行わないようにしています。
（gazelleでprotoファイルをコンパイルするようなルールを生成しないようにしています）

IntelliJ IDEAプラグイン
=========================

個人的には最近コーディングをする時はIDEを使うようにしていますし、周りにもIDEを使うことをお勧めしています。

Bazelのルールファイルを書くためのプラグイン [#intellijplugin]_ が存在するためそれは入れておいた方が便利です。
ファイルのフォーマットなどが行われます。

ただし最新のIntelliJ IDEAにすぐ対応されずちょっと間があります。
このプラグインのために最新のIDEAではなく一つ前を使ったりすることもあるので最新への追従が速いとありがたいのですがこればかりはしょうがありません。

課題
======

vendoring
-----------

上述のように各ツールで依存しているライブラリのバージョンが別でvendorディレクトリが散在している状況です。

これを統一してリポジトリ全体で一つの依存にできると素敵だなと思っています。
ただツールといっても色々な性質のものがあり、それらをすべて統一するのは得策ではないかもしれません。

悩ましいところでまだ結論が出ていません。

まとめ
=======

* モノレポは楽
* ビルドツールにBazelを使うことで更に楽
* 依存はvendoringしてリポジトリに取り込む
* 今のところProtocol BuffersのコンパイルはBazel外で行っている

モノレポだったり生成物をリポジトリに含めていたりとそういうのに抵抗がある方もそれなりにいらっしゃると思います。
そんな方もここまで読んでいただいてありがとうございます。
でもきっと有用なことはなかったことでしょう。ごめんなさい、この記事のことは忘れてください。

抵抗がないよ！という方はぜひどこかで試してみてください。
この便利さ・楽さを経験してしまうと抜け出せないかもしれません。

リンク
=========

.. [#bazel] Bazel a fast, scalable, multi-language and extensible build system https://bazel.build
.. [#starlark] Starlark https://github.com/bazelbuild/starlark
.. [#pants] Pants: A fast, scalable build system https://www.pantsbuild.org/index.html
.. [#buck] Buck A high-performance build tool https://buckbuild.com/
.. [#please] Please https://please.build/index.html
.. [#bazeldocs] Bazel Overview - Bazel https://docs.bazel.build/versions/master/bazel-overview.html
.. [#gazelle] Gazelle is a Bazel build file generator for Go projects. https://github.com/bazelbuild/bazel-gazelle
.. [#intellijplugin] IntelliJ plugin for Bazel projects https://github.com/bazelbuild/intellij

Git Logo by `Jason Long <https://twitter.com/jasonlong>`_ is licensed under the `Creative Commons Attribution 3.0 Unported License <https://creativecommons.org/licenses/by/3.0/>`_ .
