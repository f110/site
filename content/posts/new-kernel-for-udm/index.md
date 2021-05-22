---
title: UniFi Dream Machineのカーネルがアップデートされそう
date: 2021-05-22T04:44:00+09:00
isCJKLanguage: true
tags: [Home Network, UniFi]
---

UniFi Dream Machine/Pro のファームウェアのカーネルは長らく 4.1 で変わることはなかった。

この Linux 4.1 は個人的に結構曲者でアップデートされるのを首を長くして待っていたのだが、どうやら次のファームウェア（1.10）で 4.19 にアップデートされるらしい。

現在はまだ Beta 版なのでベータプログラムにサインアップしている人しか手に入らないがしばらくすれば Stable としてリリースされるだろう。

Linux 4.1 をアップデートしたい主な理由はこの頃のカーネルに実装されている ECMP が per-packet だったため。
per-packet ECMP だとその上でロードバランサーを構築する際にそれなりの実装が必要で面倒だった。

Linux 4.19 では 3-tuple や 5-tuple が実装されているはずなので期待している。
