---
title: Bazelでインテグレーションテストをする
date: 2019-12-22
lastmod: 2020-03-30
tags: ["Bazel", "Go"]
isCJKLanguage: true
---

GoのプロジェクトをビルドするツールとしてGNU MakeではなくBazelを利用し、インテグレーションテストを実現する方法について紹介します。
Bazelでテストを実行する場合、Go以外の言語でもユニットテストを実行するのは比較的容易です。
（ユニットテスト=テストコードだけで完結し外部のミドルウェアに依存しないもの）

Kubernetes Operatorを `kubebuilder <https://github.com/kubernetes-sigs/kubebuilder>`_ で実装すると付属してくるテストがetcdとkube-apiserverを要求します。
実行バイナリが必要なのでこれをBazelで実行したいというのがモチベーションです。

Bazelとは
===========

Bazelを知らないという人のために簡単にBazelを説明します。（そもそもBazelをまったく知らない人はこれを読んでない気もしますが）

BazelとはGoogleが中心となって開発しているOSSのビルドツールです。
BazelはGoogle社内で使われているBlazeというビルドツールが源流にあるそうです。
このBlazeはなかなか使い勝手が良かったらしく、ex-Googlerが各所で同様なコンセプトのビルドツールを作っているとかいないとか…
Bazelと似ているビルドツールとしては

* Facebookの `Buck <https://buck.build/>`_
* Twitterを中心に数社で使われている `Pants <https://www.pantsbuild.org/index.html>`_
* Through Machineが作っている `Please <https://please.build/index.html>`_

などがあります。

ツールとしては以下のような特徴を持っています。

* ビルドを独自のサンドボックス環境の中で行う
* ビルドの再現性が高い（サンドボックスの中で行われるので）
* 高速
* 複数の言語に対応できる
* 拡張性が高い

ビルドルールの定義に `Starlark <https://github.com/bazelbuild/starlark>`_ というPythonのサブセット言語が使われます。（StarlarkはBazelのためにデザインされた言語です）

さらに詳しい情報や使い方等は以下の拙稿をご覧ください。

* `Bazelとモノレポ <../monorepo-with-bazel>`_
* `Bazelの使い方詰め合わせ <../tips-on-bazel>`_
* `GoとBazel <../go-and-bazel>`_

（これまでの記事は会社のmediumに書いており、今回は諸事情でqiitaに書いています。mediumでの更新はもうありません）

依存してるミドルウェアの実行ファイルを取得する
==================================================

今回の完全なソースコードを `f110/bazel-example <https://github.com/f110/bazel-example>`_ に用意しました。
実際に動作する状態を見たい方はそちらを見てください。
また以下の説明では一部のソースコードを省略しています。そのためリポジトリの方を見て適宜補完してください。

今回のターゲットはetcdとkube-apiserverです。
幸いなことに両方共バイナリ1つで動作するため非常に扱いやすいです。
まずはこれらのバイナリをBazelで取得してくる必要があります。

お手軽な方法としては ``http_archive`` で取得してくることですがせっかくなのでマルチプラットフォームでテストが実行できることを目標にもう少し頑張ります。
（kube-apiserverのバイナリがLinux向けしか提供されていないので今回は生かされないのですが…）

.. code:: python

    ETCD_URLS = {
        "3.4.3": {
            "linux_amd64": (
                "https://github.com/etcd-io/etcd/releases/download/v3.4.3/etcd-v3.4.3-linux-amd64.tar.gz",
                "6c642b723a86941b99753dff6c00b26d3b033209b15ee33325dc8e7f4cd68f07",
            ),
            "darwin_amd64": (
                "https://github.com/etcd-io/etcd/releases/download/v3.4.3/etcd-v3.4.3-darwin-amd64.zip",
                "9e530371ac2a0b10ee7d5cf1230b493a18c9ff909c6f034d609994728de276f7",
            ),
        },
    }

    def _etcd_impl(ctx):
        version = ctx.attr.version
        os, arch = _detect_os_and_arch(ctx)

        url, checksum = ETCD_URLS[version][os + "_" + arch]

        ctx.file("WORKSPACE", "workspace(name = \"{name}\")".format(name = ctx.name))
        ctx.file("BUILD", "filegroup(name = \"bin\", srcs = [\"etcd\"], visibility = [\"//visibility:public\"])")
        ctx.download_and_extract(
            url = url,
            sha256 = checksum,
            stripPrefix = "etcd-v" + version + "-" + os + "-" + arch,
        )

    etcd = repository_rule(
        implementation = _etcd_impl,
        attrs = {
            "version": attr.string(),
        },
    )

    def _detect_os_and_arch(ctx):
        os = "linux"
        if ctx.os.name == "mac os x":
            os = "darwin"
        arch = "amd64"
        return os, arch

このようなリポジトリルールを定義します。 `（全体） <https://github.com/f110/bazel-example/blob/24d674c020ca4895247bd614785fd3d728c33fe6/build/rules/k8s_testing/deps.bzl>`_
やっていることは非常に簡単で、ホストOSからダウンロードしなければいけないアーカイブを決定しダウンロードします。
ダウンロードしたアーカイブを展開し、WORKSPACEファイルとBUILDファイルを流し込むだけです。
このルールをWORKSPACEから以下のように使います。

.. code:: python

    load("//build/rules/k8s_testing:deps.bzl", "etcd", "kube_apiserver")

    etcd(
        name = "io_etcd",
        version = "3.4.3",
    )

実行ファイルを流し込む
==========================

ここからが少しややこしいところで基本的に ``go_test`` は外部依存も含めてテストを実行することができません。
ビルドルールを拡張し対応した場合はルールのメンテナンスコストが高そうです。
（ ``go_test`` の実装のどこかに割り込むことができないので全てをコピーしてくることになります。つまりアップストリームに追従していく必要があります）
アップストリームにそのような変更をいれるように頑張ってもよいのですがこれが最善の方法とは思えないのでなんとかハックして解決します。
（このようなインテグレーションテストをBazelで上手に実行する方法はまだないように見え、もしかしたらBlaze側にはあるのではないかなと期待しています）

実行ファイルを組み合わせてテストを実行するためにはまず実行ファイルをサンドボックスの中に入れる必要があります。
そこで以下のように ``go_test`` のdata attrを利用します。

.. code:: python

    go_test(
        name = "go_default_test",
        srcs = [
            "suite_test.go",
            "utils_test.go",
        ],
        data = [
            "@io_etcd//:bin",
            "@io_k8s_kube_apiserver//:bin",
        ],  # keep
        embed = [":go_default_library"],
        deps = [
            "//operator/api/v1:go_default_library",
            "//operator/vendor/github.com/onsi/ginkgo:go_default_library",
            "//operator/vendor/github.com/onsi/gomega:go_default_library",
            "//operator/vendor/k8s.io/client-go/kubernetes/scheme:go_default_library",
            "//operator/vendor/k8s.io/client-go/rest:go_default_library",
            "//operator/vendor/sigs.k8s.io/controller-runtime/pkg/client:go_default_library",
            "//operator/vendor/sigs.k8s.io/controller-runtime/pkg/envtest:go_default_library",
            "//operator/vendor/sigs.k8s.io/controller-runtime/pkg/log:go_default_library",
            "//operator/vendor/sigs.k8s.io/controller-runtime/pkg/log/zap:go_default_library",
        ],
    )

本来 data はtestdataをサンドボックスに入れるためのattrです。
なのでtestdataディレクトリを含むパッケージのビルドファイルをgazelleで生成するとdata attrは利用されます。
今回はここに実行ファイルを指定することでサンドボックスに閉じ込めるというハックをします。

これだけだとファイルがサンドボックスに同梱されるだけなので `テスト側 <https://github.com/f110/bazel-example/blob/24d674c020ca4895247bd614785fd3d728c33fe6/operator/controllers/utils_test.go>`_ も少し直します。

.. code:: go

    func FindEtcd() (string, error) {
        wd, err := os.Getwd()
        if err != nil {
            return "", err
        }

        e, err := findExternal(wd)
        if err != nil {
            return "", err
        }
        path := filepath.Join(e, "io_etcd/etcd")
        if _, err := os.Stat(path); os.IsNotExist(err) {
            return "", errors.New("can't find etcd binary")
        }

        return path, nil
    }

    func findExternal(start string) (string, error) {
        p := start
        for {
            files, err := ioutil.ReadDir(p)
            if err != nil {
                return "", err
            }
            for _, v := range files {
                if strings.HasSuffix(filepath.Join(p, v.Name()), "__main__/external") {
                    return filepath.Join(p, v.Name()), nil
                }
            }
            p = filepath.Dir(p)
            if p == "/" {
                break
            }
        }

        return "", errors.New("can't find external")
    }

カレントディレクトリから上の遡っていき実行ファイルを探します。
サンドボックスの中は下のような構造になっているので少し上に遡れば実行ファイルを見つけることができます。

.. code::

    __main__
    ├── external
    │   ├── io_etcd
    │   │   └── etcd
    │   └── io_k8s_kube_apiserver
    │       └── kube-apiserver
    └── operator
        └── controllers
            └── linux_amd64_stripped <-- working directory
                └── go_default_test

最後に ``suite_test.go`` （テストの本体）から環境変数に実行ファイルのパスを設定してあげれば完成です。

.. code:: go

    if path, err := FindEtcd(); err == nil {
        os.Setenv("TEST_ASSET_ETCD", path)
    }

この方法の致命的な欠点
=========================

定義された名前に依存します。
これがうまく動作するのはetcdが ``@io_etcd`` 、 kube-apiserverが ``@io_k8s_kube_apiserver`` という名前で定義されているときだけです。
違う名前で定義するとサンドボックス内のパスが変わるのでテスト側もそれに対応するか名前のゆらぎに対応できるようにする必要があります。

他に考えられる方法
=====================

ソースコードごと持ってくる
-----------------------------

etcdもkube-apiserverもオープンソースなプロジェクトなのでソースコードを手に入れることができます。
それを利用してテストにそれらそのものをテストに埋め込んでしまうこともできます。

ですがこれはおそらくテスト側の記述量が増えメンテしていくコストも高いはずです。
またバージョンアップ時のコストもそれなりに発生するはずです。

テストの中でビルドする
------------------------

テストの中でビルドして別のプロセスとして実行するということもできなくもないかもしれません。
が、これも現実的ではないように思います。

特にマルチプラットフォームをサポートしようと思った場合、相当な量の作業が発生することでしょう。

まとめ
========

kubebuilderを題材にBazelでインテグレーションテストを実行する方法の一例を紹介しました。
rules_go を改変したりせず最小のコーディングと少々のハックでインテグレーションテストの実行を実現しています。
data attrを使った依存ファイルの流し込みハックは見かけたことがなかったので紹介しました。

インテグレーションテストを実行する方法はこれ1つではなく依存するミドルウェアなどでも変わってくるかもしれません。
完璧なソリューションというのは今のところ存在しません。

ここまで読んだ多くの人の感想は「Bazelめんどくせえ」でしょう。
はい、実際ここまでやらないといけないのは面倒くさいです。
ですがこれをやると再現性も高くなりますし将来への投資だと思って試行錯誤している段階です。
