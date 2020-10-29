---
title: Bazelの使い方詰め合わせ
date: 2019-05-13
tags: ["Bazel", "Go"]
isCJKLanguage: true
---

本エントリは [オリジナル](https://medium.com/mixi-developers/bazel%E3%81%AE%E4%BD%BF%E3%81%84%E6%96%B9%E8%A9%B0%E3%82%81%E5%90%88%E3%82%8F%E3%81%9B-f6784c7bb874) の一部を再編集して掲載しています。 (2020/03/31)

[前の記事](../monorepo-with-bazel) ではBazelについて簡単に紹介しました。
ここでは更に1歩、Bazelの使い方に踏み込んでみたいと思います。

自分のリポジトリに含まれている色々なツール等をビルドしてshipするにあたって分かりづらかったりした点を中心に説明したいと思います。
すべてを網羅できているわけではありませんし、あくまで自分のリポジトリの頻出パターンなので皆さんのリポジトリでは違った点で悩むかもしれません。
Bazelを使い始めようという時に思い出して見ていただけるとよいかもしれません。

今回もサンプルコードは前回と同じリポジトリに置いてあります。

https://github.com/f110/bazel-example

# コンテナを作る

サンプルリポジトリの `helloworld1` はコンテナの作成もできるようになっています。

```python
load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_image(
    name = "image",
    binary = ":helloworld1",
    pure = "on",
    visibility = ["//visibility:public"],
)
```

このようなルールを書いて（ `WORKSPACE` でrules_dockerの定義をしておく必要はあります）

```console
$ bazel build //tools/helloworld1:image.tar
```

とビルドすればコンテナのtarファイルができあがります。 **ターゲット名に.tarをつけるのがポイントです。**

手元のdockerなどで実行したい場合は作成されたtarファイルをロードすればいいだけです。

```console
$ docker load -i ./bazel-bin/tools/helloworld1/image.tar
```

このコンテナはコンパイルされたバイナリしか入っていません。
busyboxすら入っていないため直接デバッグするのは難しいです。

busyboxを入れたコンテナもビルドすることは可能なので動いてるコンテナを直接デバッグする必要があればbusybox入りのイメージをビルドするのがいいと思います。

# コンテナのPush

コンテナを作成できればそのままPushも行いたいですよね？

```python
load("@io_bazel_rules_docker//container:container.bzl", "container_push")

container_push(
    name = "push",
    format = "Docker",
    image = ":image",
    registry = "asia.gcr.io",
    repository = "example/example",
    tag = "{BUILD_TIMESTAMP}",
)
```

と BUILD.bazel に書いておけば

```console
$ bazel run //tools/helloworld1:push
```

だけでコンテナの作成からPushまで一度に行えます。

ご想像通り、もしコンテナに同梱されるバイナリがビルドされていなければビルドも行われます。

イメージのタグがタイムスタンプになるため場合によっては使いづらいかもしれません。（実際ちょっと使いづらいと思っています）
その場合はイメージをsha256のハッシュで指定すると良いでしょう。

ビルドごとにちゃんと同一のソースコードから同一のバイナリが生成されるようになっていれば何度コンテナを作成してもソースコードが変更されていなければイメージのハッシュは変わらないはずです。

# 複数のファイルを一つのコンテナにする

複数の実行ファイルを一つのコンテナに入れ実行時にコマンドを渡したり、実行ファイルの動作に必要なファイルをコンテナに含めることもよくあることでしょう。
そんなコンテナもそれほど難しくなく作れます。

```python
go_binary(
    name = "helloworld1",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_binary(
    name = "helloworld2",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
```

このようにバイナリが2つあったとしましょう。

```python
pkg_tar(
    name = "bin",
    deps = [
        ":helloworld1",
        ":helloworld2",
    ],
)

container_image(
    name = "image",
    base = "@com_google_distroless_base//image",
    tars = [
        ":bin",
    ],
)
```

`pkg_tar` で一つのtarにまとめてからそれを `container_image` でコンテナにすれば2つのバイナリが一つのコンテナに入ります。
この例だと `helloworld1` と `helloworld2` というファイルがルート直下にできてしまうのでパスを変えたいこともあるかもしれません。
（実際大体パスは変えます）

その場合は

```python
pkg_tar(
    name = "bin",
    deps = [
        ":helloworld1",
        ":helloworld2",
    ],
    package_dir = "/usr/local/bin",
)
```

`package_dir` を指定すればそのパス以下になります。

複数のパスに同時に別々のファイルを置くことは出来ません。
つまりディレクトリごとに `pkg_tar` を定義していく必要があります。

これは最初とっつきにくいかもしれません。
しかしBazelの特徴であるサンドボックス化とキャッシュはここでも効くので再ビルドする際などは変更があるディレクトリのtarファイルだけ再作成されます。

# 複雑なコンテナを作る

複雑といってもバイナリだけではなくその動作に必要なライブラリを含める場合です。

[Distroless](https://github.com/GoogleContainerTools/distroless) のbaseイメージを使ってそれに必要なパッケージを追加していくような形でコンテナを作ります。

```python
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

git_repository(
    name = "com_google_distroless",
    remote = "https://github.com/GoogleContainerTools/distroless.git",
    commit = "432c6f934f6c615142489650d22250c34dc88ebd"
)
```

を `WORKSPACE` ファイルに記述してリポジトリを取り込みます。

```python
load("@com_google_distroless//package_manager:package_manager.bzl", "dpkg_list", "dpkg_src", "package_manager_repositories")

package_manager_repositories()

dpkg_src(
   name = "debian_stretch",
   arch = "amd64",
   distro = "stretch",
   sha256 = "9e7870c3c3b5b0a7f8322c323a3fa641193b1eee792ee7e2eedb6eeebf9969f3",
   snapshot = "20181019T145930Z",
   url = "https://snapshot.debian.org/archive",
)

dpkg_src(
   name = "debian_stretch_backports",
   arch = "amd64",
   distro = "stretch-backports",
   sha256 = "3ddd744c8560dcc03dcd339bc043af54547201780a51fa541916ee083ccbdac4",
   snapshot = "20181019T145930Z",
   url = "http://snapshot.debian.org/archive",
)

dpkg_src(
   name = "debian_stretch_security",
   package_prefix = "https://snapshot.debian.org/archive/debian-security/20181019T145930Z/",
   packages_gz_url = "https://snapshot.debian.org/archive/debian-security/20181019T145930Z/dists/stretch/updates/main/binary-amd64/Packages.gz",
   sha256 = "c212bcbde4e22d243d0238faed7b9f3eb05c708f7ba7937e2bed562c8de71cc9",
)

dpkg_list(
   name = "package_bundle",
   packages = [
       "busybox-static",
       "rsync",
       # ここに必要なパッケージを列挙する
   ],
   sources = [
       "@debian_stretch_security//file:Packages.json",
       "@debian_stretch_backports//file:Packages.json",
       "@debian_stretch//file:Packages.json",
   ],
)
```

というような感じで必要なパッケージを取得してきます。
これも `WORKSPACE` ファイルに記述しておきます。

これで下準備は完了しているので次に実際にコンテナを作ります。
以下は `BUILD` ファイルに記述します。

```python
load("@io_bazel_rules_docker//container:container.bzl", "container_image")
load("@package_bundle//file:packages.bzl", "packages")

container_image(
   name = "image",
   base = "@com_google_distroless_base//image",
   debs = [
       packages["busybox-static"],
       packages["rsync"],
       # 他にコンテナに含めたいファイルがある場合はここに追記する
   ],
   visibility = ["//visibility:public"],
)
```

## Dockerfileの代替になるか

このようにパッケージを追加したイメージも作れますが `Dockerfile` の代替とするのは難しいなと感じています。

いくつも依存パッケージがあるようなコンテナの場合、非常に定義が煩雑になります。
`Dockerfile` であれば一行 `RUN` を書けば済むところが何十倍も定義を書かないといけないです。

それでもBazelでコンテナを作った方が確実ではあると思います。
Bazelは誰のローカルで実行してもCI上で実行しても同じ定義ファイルからは同じコンテナイメージができます。
またコンテナの作成にはdockerは不要です。
一方docker buildした場合は実行した時間によって結果は違いますし、dockerの状態によっても結果が大きく変わってしまいます。

そんなデメリットがあってもBazelですべてのイメージを作るのは大変で `Dockerfile` を使っているものがいくつもあります。

# パッケージを作る

サンプルリポジトリには `mysqld_exporter` のdebパッケージを作るルールも同梱しています。

このようにバイナリがリリースされていてそのバイナリをdebパッケージとして詰め込むだけであれば非常に簡単です。

サンプルリポジトリでは

```console
$ bazel build //debian_packages/mysqld_exporter:package
```

で `bazel-bin/debian_packages/mysqld_exporter/mysqld-exporter_0.11.0-1_amd64.deb` ができあがります。

パッケージとして使うにはこれ以外のファイルも入れたくなるでしょう。
もちろんそれも上述の方法を応用して可能です。

# テストスイートを自作してテストをする

自分のリポジトリでは設定ファイルのテストスイートを自作してそれを利用してCIもしています。
CIではテストスイートのビルドから対象となる設定ファイルを食わせて実行するところまでのすべてをBazelで行います。

これを実現するために必要なものは

1. テストスイートの実装
1. テストスイートを実行するための定義
1. テスト対象の設定ファイルを定義する

の3つです。

テストスイートの実装は引数で設定ファイルを受け取り、テストした結果に応じてexit codeが変わるものであればいいです。

テストスイートを実行するための定義が通常のビルド用の `BUILD.bazel` ファイルを書いたりする時とは違います。

以下のようなファイルを `def.bzl` として作ります。

（このコードも https://github.com/f110/bazel-example の `test-suite` と `config` ディレクトリのそれぞれに入っています）

```python
def _example_config_test_impl(ctx):
    src = ctx.file.src
    kicker = ctx.actions.declare_file("%s_kicker.sh" % ctx.label.name)
    ctx.actions.expand_template(
        template = ctx.file._wrapper_template,
        output = kicker,
        substitutions = {
            "{executable_binary}": ctx.executable._test_suite.short_path,
            "{config_file}": src.short_path,
        },
        is_executable = True,
    )
    runfiles = ctx.runfiles(files = [kicker, ctx.executable._test_suite, src])
    return [DefaultInfo(executable = kicker, runfiles = runfiles)]

example_config_test = rule(
    implementation = _example_config_test_impl,
    test = True,
    attrs = {
        "src": attr.label(allow_single_file = True),
        "_test_suite": attr.label(
            default = Label("//test-suite/example-suite"),
            executable = True,
            cfg = "target",
        ),
        "_wrapper_template": attr.label(
            allow_single_file = True,
            default = "kicker.tpl",
        )
    },
)
```

`exmaple_config_test` を定義します。
これには実際に実行される際に何が行われるか関数として定義され `example_config_test` に指定できる属性値も定義されます。

テストスイートを実行する際には直接バイナリを実行するのではなく、シェルスクリプトを作りそれを実行するようにしています。
これは引数を渡してバイナリを実行するというのが提供されていなかったのでこのようにしています。

`runfile` という変数に動作に必要なファイルのリストを作り `DefaultInfo()` の引数で渡しています。
ここでは `kicker` シェルスクリプト、 `ctx.executable._test_suite` テストスイートのバイナリ、 `src` テスト対象のファイル、の3つを渡します。
これらのファイルがサンドボックスの中に入るためここで必要なファイルを列挙しておく必要があります。

シェルスクリプトのテンプレートとなるkicker.tplは以下のように単純にバイナリに引数をつけて実行するだけです。

```bash
#!/bin/bash
exec {executable_binary} {config_file}
```

最後にこの2つのファイルの `BUILD.bazel` を用意します。

```python
filegroup(
    name = "all_rules",
    visibility = ["//visibility:public"],
)

exports_files(
    ["kicker.tpl"],
    visibility = ["//visibility:public"],
)
```

ここまで用意できれば、最後はテスト対象の設定ファイルを定義するだけです。

```python
load("//test-suite/example-suite/rules:def.bzl", "example_config_test")

example_config_test(
    name = "test",
    src = "success.conf"
)
```

サンプルのリポジトリであれば

```console
$ bazel test //config/...
```

でテストができることを確認できます。
`config` ディレクトリの `success.conf` を `[]` に編集して再度テストするとテストが失敗する様子も見ることが出来ます。

# Go の静的ファイル埋め込み

Go でバイナリにリソースを埋め込むというのもBazelで出来ます。
ただしBazelはリソースを埋め込んだソースファイルを生成するだけですので、実際にそれを利用したソフトウェアを構築するにはそれ以外の部分でも工夫が必要なこともあります。

例えばChat Botにも休日を与えるため祝日のデータがCSV形式で埋め込まれています。（なおこのCSVは内閣府大臣官房総務課が配布しているものを利用しています）

```python
load("@io_bazel_rules_go//extras:embed_data.bzl", "go_embed_data")

go_embed_data(
    name = "embed",
    srcs = [
        "//bot/data/holiday",
    ],
    visibility = ["//visibility:public"],
)

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "embed_data",
    srcs = [":embed"],
    importpath = "github.com/f110/bazel-example/bot/assets",
    visibility = ["//visibility:private"],
)

go_library(
    name = "go_default_library",
    srcs = ["dummy.go"],  # keep
    embed = [":embed_data"],  # keep
    importpath = "github.com/f110/bazel-example/bot/assets",
    visibility = ["//visibility:public"],
)
```

go_embed_data で go-bindata を用いて任意のファイルを埋め込んだソースコードが生成できます。
ただしそれだけだとサンドボックス内で生成されてしまい、IDE等ではコンパイルエラーになります。
そこで定義だけの `no_bazel.go` を作っておきます。

```go
var Data = map[string][]byte{}
```

しかしこの `no_bazel.go` はビルドからは除外したいです。
ですがgazelleを使って `BUILD.bazel` を生成していると勝手にビルド対象に含まれてしまいます。
そこでgazelleの **# keep がある行は変更しない** という機能をうまく利用します。

上記のサンプルコードの場合は2箇所利用しています。
一度gazelleで `BUILD.bazel` を生成した後、no_bazel.goを削除し # keep をつけたりするといいでしょう。

## テンプレートファイルの場合

前述の方法でテンプレートファイル自体はバイナリに埋め込めますがそれでは開発時に不便です。
開発時は多少遅くてもリクエストのたびにテンプレートをコンパイルしてレンダリングしてくれたほうがテンプレートの修正が簡単に行えるので大変便利です。

そこで基本的に開発時は埋め込まれたテンプレートを使わず、埋め込むべきファイルのパスを設定ファイルで指定してリクエストのたびにファイルを読み込むようにしています。

これはBazelを使わない場合も同様に似たようなものを実装する必要があるので皆さんも何度か実装されたことがあるかもしれません。

# 番外：Remote Cache

サンドボックスでビルドやテストが実行されるBazelは生成物をキャッシュすることができます。
ローカルで2回目以降のビルドにかかる時間がすごく短いのはこのキャッシュのおかげでもあります。

そしてこのキャッシュはリモートサーバーで共有することができます。

```console
$ bazel test --remote_http_cache=http://bazel-cache //...
```

と引数にキャッシュサーバーのアドレスを渡すだけなので簡単に始められます。

キャッシュサーバーはNginxで提供されるWebDAVとキャッシュされたオブジェクトのクローラーの2つから成り立ちます。

クローラーはキャッシュの領域が減ると古いオブジェクトから目標とする容量まで削除します。
このクローラーは動作が複雑ではなかったのでサクッと自作しました。

# まとめ

Bazelを使ってコンテナやパッケージを作ったり、テストスイートとそれによるテストの実行方法について簡単に紹介しました。
ソフトウェアのビルドにとどまらずテストやパッケージングも一つのツールで完結することができます。

Bazelは実際に自分のリポジトリで使ってみると拡張性の高さを体感することができますし、ここで紹介したものはBazelの力のほんの一部でしかないことに気がつくかもしれません。
今までの方法と大きく違う面も多く慣れないうちは戸惑うこともあるかと思いますが、Bazel流のやり方の理解が進むと良さが分かってくるかと思います。
勘所が掴めると `rules_*` のSkylarkを理解するのも簡単になってきます。
自分のやりたいことを実現する直接的な方法が分からなくても `rules_*` を参考にすることができます。

Bazelをリポジトリ全体のビルドツールとして利用していくには主要なコミッターが最低一人はBazel流のやり方に合わせていける人である必要があると思います。
そういった点で導入の障壁の高さや学習コストの問題はあるかもしれません。
そのため、まだ誰にでもおすすめできるツールではないですが興味がある人はチャレンジしてみてください！
