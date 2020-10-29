---
title: Homecluster basic
date: 2020-04-02
isCJKLanguage: true
tags: ["Kubernetes", "homecluster"]
---

自宅の Kubernetes クラスタをどうやって構成するかという話。

自宅クラスタ関連は [タグ](/tags/homecluster) で一覧できるようにしています。

<!--
.. section-numbering::
.. contents::
    :local:
-->

# minikubeではなく物理クラスタという選択

minikubeではなく物理マシンを複数台用意してクラスタを構成しています。

デスクトップマシンでは Kubernetes のAPIを使った開発をするために minikube を利用していますが、それとは別に24時間稼働する物理クラスタがあります。

物理クラスタを構築したのは UniFi Controller を動かそうと思ったからです。

自宅のネットワークは UniFi の製品を利用していますが、これにはコントローラが必要です
（常時コントローラを起動しておく必要はありませんが何かの変更をする際にはコントローラを使うほうが楽です）
都度ローカルマシンでコントローラを起動する運用でもそれほど不便ではないのですが、やはり常時起動しておきたい。
ふと「もしかしてUniFi Controllerのhelm chartがあるのでは？」と思いリポジトリを見てみるとありがたいことにメンテしてくれている人がいました。
であればこれを利用させてもらうのが一番簡単だろうと考えたからです。

# ハードウェア

基本的には [Intel NUC](https://www.intel.com/content/www/us/en/products/boards-kits/nuc.html) を使っています。

HDDを搭載するつもりは一切ないのでそれを搭載できないモデルを選んでいます。
具体的には [NUC8I3BEK](https://www.intel.com/content/www/us/en/products/boards-kits/nuc/kits/nuc8i3bek.html) を使っています。

| パーツ | 型番 | スペック | 備考 |
| --- | --- | --- | --- |
| 本体 | Intel NUC8I3BEK | Core i3 |
| メモリー | SO-DIMM DDR4 | 8GB x 2 |
| ディスク | SSD | 480GB | 詳細な型番は忘れました |

1台あたりおおよそ5万5千円です。
これを2台持っています。

クラスタ運用中にマシンの買い増しを行い今は3台構成のクラスタになっています。

```
$ k get nodes
NAME     STATUS   ROLES    AGE    VERSION
whale1   Ready    master   162d   v1.17.4
whale2   Ready    <none>   162d   v1.17.4
whale3   Ready    <none>   16d    v1.17.4
```

# ソフトウェア

クラスタを構築するためのソフトウェアは以下を利用しています。

| 種類 | ソフトウェア | 備考 |
| --- | --- | --- |
| OS | Ubuntu 19.10 | |
| 管理ツール | kubeadm | |
| コンテナランタイム | docker / contaienrd | containerdに移行中 |
| ネットワーク | Calico | BGPモード |
| ロードバランサ | MetalLB | ARPモード |

最低限これらを用いればPodをスケジュールするところまではできます。

この構成であればコンシューマ向けルータ配下でも構築できるでしょう。
自宅はルータがUniFi Security Gatewayなのでもっと踏み込んだ使い方もできます。(例えばMetalLBをBGPモードで動かすとか）

# ブートストラップ

クラスタを動作させるまでを簡単に書きます。
記憶を頼りに書いているため、この手順を踏んでも構築できるとは限りません。

基本的には kubeadm の [オフィシャルドキュメント](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/) を進めます。
台数を確保出来る場合はマスターもHA構成を取ることをおすすめします。

## OSのインストール

台数が少ないので手動インストールをしています。

UbuntuのブートUSBを作成し、インストーラを用いて通常通りインストールします。

インストール中にパッケージを選択できますがそこではopenssh-serverのみインストールする最小構成にします。

## Swapを切る

kubeadm によるセットアップを行う前にswapを切っておく必要があります。

```
$ swapoff
```

でswapを切っておきます。
再起動してswapが元に戻ってしまうのも面倒なので合わせて `/etc/fstab` からも消しておきましょう。

## kubeadm によるセットアップ

このステップはオフィシャルドキュメントの通りに行ってください。

`kubeadm init` と `kubeadm join` を繰り返し実行するだけになります。

## Calicoのインストール

ネットワークプラグインには Calico を利用しますが配布されているマニフェストはIPIPが有効になっています。
特にこだわりがなければそのまま動かすのがいいでしょう。

自宅クラスタではIPIPは不要なのでoffにしておきます。

マニフェストをダウンロードし編集します。

```console
$ curl -O https://docs.projectcalico.org/manifests/calico.yaml
```

ここで行う編集は3つ。

1. PodのIPレンジを決める
1. IPIPを切る
1. XDPをdisableにする

```diff
--- calico.yaml 2020-04-02 00:10:56.197222351 +0900
+++ calico.a.yaml       2020-04-02 00:10:15.728870190 +0900
@@ -614,7 +614,7 @@
               value: "autodetect"
             # Enable IPIP
             - name: CALICO_IPV4POOL_IPIP
-              value: "Always"
+              value: "off"
             # Set MTU for tunnel device used if ipip is enabled
             - name: FELIX_IPINIPMTU
               valueFrom:
@@ -624,8 +624,8 @@
             # The default IPv4 pool to create on startup if none exists. Pod IPs will be
             # chosen from this range. Changing this value after installation will have
             # no effect. This should fall within `--cluster-cidr`.
-             - name: CALICO_IPV4POOL_CIDR
-               value: "192.168.0.0/16"
+            - name: CALICO_IPV4POOL_CIDR
+              value: "192.168.0.1/16"
             # Disable file logging so `kubectl logs` works.
             - name: CALICO_DISABLE_FILE_LOGGING
               value: "true"
@@ -640,6 +640,8 @@
               value: "info"
             - name: FELIX_HEALTHENABLED
               value: "true"
+            - name: FELIX_XDPENABLED
+              value: "false"
           securityContext:
             privileged: true
           resources:
```

XDPをdisableにするのはCalicoのバージョンに依存するはずです。
enableのままでcalico-nodeのPodが正常に動作しているようであればenableのままでいいでしょう。

あとは

```console
$ k apply -f calico.yaml
```

とするだけです。

`kube-system` で必要なPodが動作するので動作を確認しましょう。

```console
$ k -n kube-system get pod
```

## MetalLBのインストール

これも [MetalLBのオフィシャルドキュメント](https://metallb.universe.tf/installation/) の通りで出来ます。

作業を行う前に MetalLB が使うIPのレンジを決めましょう。
上位のルータとServiceを使う予定の数でレンジの幅は調整してください。

ちなみに現時点でクラスタ内には60以上のServiceがありますが Type=LoadBalancer でIPアドレスを割り当てられているものは6個程度しかありません。
あまり広いレンジを確保する必要はないでしょう。

```console
$ k apply -f https://raw.githubusercontent.com/google/metallb/v0.9.3/manifests/metallb.yaml
```

[Layer2のConfigration](https://metallb.universe.tf/configuration/#layer-2-configuration) も忘れずに行ってください。

MetalLBにBGPを喋らせる場合はそれに応じた設定をしましょう。

# 動作確認

後は適当にPodをデプロイして動作確認をしてください。

# ユーザー認証

ユーザー認証の設定をしていないのでクラスタの操作は `kubernetes-admin` で行うことになります。

[認証方法についてのオフィシャルドキュメント](https://kubernetes.io/docs/reference/access-authn-authz/authentication/) を参照して自分にあった認証方法を選択してください。

我が家のクラスターは [クライアント証明書認証](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#x509-client-certs) を採用しています。
複数台あるマシンでそれぞれ秘密鍵とCSRを生成し Control plane のマシンにあるCAで署名するようにしています。
このプロセスはopensslコマンドで実施されており、何らかのPKIエンジンを使っているわけではありません。
