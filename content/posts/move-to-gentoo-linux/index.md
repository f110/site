---
title: Gentoo Linuxに移行した
date: 2021-04-29T13:17:00+09:00
isCJKLanguage: true
tags: [Gentoo]
---

メインで使っているマシンは長らく Ubuntu を使っていたのだが今回 Gentoo Linux に移行した。

現時点でデフォルトで起動する OS が Gentoo に切り替えられるくらいにはなったので適当に雑感を書いておく。

# なぜ移行したのか

Ubuntu を使っていてなにか問題があった・できないことがあった、というわけではない。
むしろ Gentoo Linux 移行時に Ubuntu の完成度の高さを改めて感じた。

ディストリビューションを移行したのは偶然いくつかのタイミングが重なったから。

Ubuntu をインストールしていたディスクはディスク全体を2つに分けてそのうちの片方をルートにしていた。
このディスクは 512GB だったので 256GB の容量があったのだがこれを使い切りそうになってしまったのでもう一つのパーティションをマウントしてディスク枯渇を回避したのだが、これは使い勝手が悪かった。

これを解決するためにより大きなディスクを刺して移行するかということで 1TB のディスクを購入して移行することを決めた。

どうせ移行するなら別に Ubuntu じゃなくてもいいなと思ったのと、ちょうどお仕事で Chromium OS について調べてて Portage を使ってみたくなったので Gentoo Linux に移行することにした。

# 移行時に苦労した点

公式のドキュメントが充実しているので基本的にそれに従えばそこまで苦労しないはず。

## 5.10 カーネルが使えなかった

今回ドキュメントになく少し手こずったのが、デフォルトの 5.10 カーネルが使えなかった点。（2021年4月のデフォルトカーネルが 5.10 ）

カーネル自体はちゃんと動作するが、 NIC が完全に動作しなかった。

どうもフラグメントしてる際にパケットの末尾の数バイトがヌル文字になってしまうようで、フラグメントしてしまうと正しく通信ができない。

試したところ 5.4 カーネルは問題なく動作したので今はあえて 5.4 カーネルを使っている。

## Portage の概念

長らくパッケージマネージャは apt を使っていたので Portage の概念を把握できるまでは自分のやりたいことを実現するためにどうしたらいいか分からず調べる必要があるので多少苦労する。

移行時に様々な環境のセットアップをしたので今は概念の把握が進みそこまでストレスではなくなってきた。

# Gentoo の構成

いくつかデフォルトから変更したものがある。

まずは systemd の採用。デフォルトは OpenRC だったので systemd に変更した。

ネットワークのマネジメントには NetworkManager を使うようにした。これは Ubuntu でも使っていたので採用したというところが大きい。

ルートのファイルシステムは btrfs にした。今までずっと ext3 なり ext4 を使ってきたのだが、もうルートのファイルシステムとして使えるくらいにはなってるだろうということで btrfs にするというチャレンジ（自分の中では）をしてみた。
まだ btrfs のより先進的な機能は使ってないので ext4 と btrfs の差には気がついてないが、これは追々使えるところでちゃんと使っていきたい。

Gentoo の公式インストールドキュメントだと最後の方で root のパスワードを設定するというステップがあるが、ここは Ubuntu と同様に root ユーザーのパスワードは設定していない。
インストール時は別の OS からディスクをマウントし chroot しているのでその間に sudo を設定しておけばきっと問題ないだろう。
今のところそれによる問題は起きてない。

# 独自オーバーレイ

Portage は簡単な仕組みで独自のパッケージを加えることができる。

これは早速活用しており、オフィシャルのツリーにないがよく使うソフトウェアなどは自分で ebuild を書いてインストールしている。

例えば [JetBrains Toolbox](https://www.jetbrains.com/toolbox-app/) は IDE を管理する上で必須になっているのでこれをインストールする ebuild を既存のものを参考に書いた。
JetBrains Toolbox の ebuild は第三者が書いているものがいくつかあり、それを使えばインストールできるわけだが第三者のツリーに依存するのも嫌だなと思い、それを参考に自分でメンテすることにした。

他にも [bazelisk](https://github.com/bazelbuild/bazelisk) [kustomize](https://kustomize.io) [kind](https://kind.sigs.k8s.io) なども自作している。

いずれもビルド時に機能を切り替えたりする必要がないので配布されているバイナリをそのままインストールする形にしている。

# 新しいものを覚えるとき

新しいものを覚えようとするときに「前はこうやってやっていたのはどうやったらできるんだろう？」という思考は **捨てたほうが良いと思う。**

そういう考えをしていると前の体験が足かせとなり新しいことを覚えづらくなる。

まさしく「郷に入っては郷に従え」で考え方をリセットし全く違うものとして学習し直すのが良い。
