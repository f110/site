---
title: "Ubuntu 20.04へアップデート"
date: 2020-05-04T03:10:36+09:00
isCJKLanguage: true
tags: ["Ubuntu"]
---

普段使っているデスクトップPC（Ryzen Threadripper 2950X）にはSSDが物理的に2枚刺さってて1つ目にはLinuxが、2つ目にはWindowsがインストールされています。

LinuxはずっとUbuntuを使っていて18.04(Bionic Beaver)を使っています。
20.04(Focal Fossa)がリリースされたのでアップデートしました。

18.04（以下bionic）の時にあった主な不具合は以下。

1. gnome-shell のCPU使用率が高い
1. sensorsが表示するCPUコア温度が実際の温度から+27℃されている


1 はCPUのパワーに任せて無視、2 も実使用上は問題にならないので無視していました。

20.04（以下focal）へのアップデートは `do-release-upgrade` で行います。

```console
$ sudo apt update && sudo apt upgrade
$ sudo do-release-upgrade -d
```

これで必要なパッケージは全てアップデートされるので後は再起動するだけです。

GPUにnVidiaのものを使っているので `do-release-upgrade` 後にnvidiaのドライバは再度インストールする必要があります。
ドライバをインストールするまで解像度がすごく低く作業しづらいですが頑張ってインストールします。

focalにアップデートしてbionicにあった2つの大きな問題は解決しました。

以下、アップデート後にそれに伴って個別に対応した点です。

1. nvidiaドライバーのインストール
1. vimプラグインの再インストール
1. Go 1.14でビルドしていたものを1.14.1で再ビルド

現時点のfocalのカーネルは

```
$ uname --kernel-release
5.4.0-28-generic
```

5.4.0-28 ですが、このkernelには 5.4.30 までが [取り込まれています。](https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1870571)

なのでGo 1.14.1にすることでも [Signal Vector問題](https://github.com/golang/go/issues/37436) は緩和されます。

focalのOpenSSHは8.2なので FIDO/U2F 対応が [入っています。](https://www.openssh.com/txt/release-8.2)
これについては後ほど試してみようと思います。
