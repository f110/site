---
title: UniFi Dream MachineでWireGuardをカーネルで動かすプロジェクト
date: 2021-05-22T04:45:00+09:00
isCJKLanguage: true
tags: [Home Network, Wireguard, UniFi]
---

[UniFi Security Gateway の時](https://f110.jp/posts/vpn-with-wireguard/)は [WireGuard をカーネルモジュール](https://github.com/WireGuard/wireguard-vyatta-ubnt)として動かしていた。

UniFi Dream Machine Pro では OS が新しくなっておりこのモジュールが使えないため、ユーザーランドで動かしていた。
ユーザー空間で動かすとパフォーマンスはかなり良くないが、繋がりはするので仕方なかった。

カーネルが 5.6 以上になってカーネルに WireGuard が組み込まれるかファームウェアでサポートされるまで待つかと思っていたが Dream Machine でもカーネルモジュールとして WireGuard を動かしているプロジェクトを発見した。

[https://github.com/tusc/wireguard-kmod](https://github.com/tusc/wireguard-kmod)

自分の Dream Machine Pro ではまだ試してないが（ユーザーランドの WireGuard が動いているのでまずはそれを止めたりしないといけない）近々使ってみたいと思っている。
