---
title: facebook/saplingをビルドする
date: 2022-11-17T16:23:00+09:00
isCJKLanguage: true
tags: [Productivity, Monorepo]
---

Meta が社内で使っている Version Control System の一部（？）を公開してドキュメントなどが整備されたので早速試してみた。
試してみるといってもコンパイルからで、ドキュメントに沿ってコマンドを実行すればビルドが成功するものではなかったのでビルド方法を記事にする。

なお Homebrew を使っている人は [公式のドキュメント](https://sapling-scm.com/docs/introduction/installation#macos) に従った方が良いと思われる。
自分の環境は MacPorts なのでソースコードからコンパイルするしかないが Homebrew を使っていればもっと簡単に利用できるはずである。

ソースコードの入手についてはどんな方法で行っても構わない。

## ビルドに必要なソフトウェアをインストールする

yarn や Rust、Python が必要なのでインストールする。

    $ port install yarn

Rust については MacPorts でインストールしない。これは現時点で MacPorts でインストールされる Rust が古くて sapling をコンパイルできないからである。
[Rustの公式のドキュメント](https://www.rust-lang.org/learn/get-started) を参考にインストールし、 `rustc` と `cargo` にパスを通しておく。

Python についてはインストール済みということにする。

## build.rs を修正する

MacPorts で libiconv がインストールされているとそちらへリンクしようとしてしまう。

MacPorts の libiconv はシンボルが変更されており OS のものとコンフリクトしないようになっている。したがって MacPorts の libiconv とリンクしようとしてしまうとリンクで失敗する。

そこで次のような変更をしてシステムに入っている libiconv とリンクするようにする。

    diff --git a/eden/scm/exec/hgmain/build.rs b/eden/scm/exec/hgmain/build.rs
    index f0e5d0a570..db562c9456 100644
    --- a/eden/scm/exec/hgmain/build.rs
    +++ b/eden/scm/exec/hgmain/build.rs
    @@ -8,6 +8,7 @@
     use std::env;
    
     fn main() {
    +    println!("cargo:rustc-link-search=/usr/lib");
         println!("cargo:rerun-if-changed=build.rs");
         if let Some(lib_dirs) = env::var_os("LIB_DIRS") {
             for lib_dir in std::env::split_paths(&lib_dirs) {

## ビルド

以下は `$HOME/local/sapling` にインストールする例である。

    $ CARGO_NET_GIT_FETCH_WITH_CLI=true make install-oss PREFIX=$HOME/local/sapling

まだまだ変更が入っていきそうなので他と混じらないように全く別のディレクトリを PREFIX にしている。

## 実行スクリプト

上記のコマンドでビルドしていれば `$HOME/local/sapling/bin/sl` に実行ファイルができているがこれをそのまま実行することができなかった。

これを実行するためにラッパースクリプトを用意してそれを経由して実行している。

    #!/usr/bin/env bash
    
    export PYTHONPATH=$HOME/local/sapling/lib/python3.10/site-packages
    exec $HOME/local/sapling/bin/sl $@

もっといい方法があるのかもしれないがこれで動くのでとりあえずは満足している。
