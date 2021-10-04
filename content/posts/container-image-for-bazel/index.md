---
title: Bazel のコンテナを作った
date: 2021-10-04T09:00:00+09:00
isCJKLanguage: true
tags: [Bazel, Docker]
---

Bazel のコンテナイメージは公式のドキュメントでは `[l.gcr.io/google/bazel](http://l.gcr.io/google/bazel)` のものが使われている。

しかしこのイメージは更新が止まっており 3.5.0 以降更新されていない。

更に良くないことにこのイメージは Ubuntu 16.04 をベースにしており、Let's Encrypt の CA 切り替えの影響を受け証明書の検証が失敗するようになってしまった。

元々 `[l.gcr.io/google/bazel](http://l.gcr.io/google/bazel)` は [Google Cloud の人たちがメンテナンス](https://github.com/GoogleCloudPlatform/container-definitions/blob/12508ac50e5a1f18ddb88c3dd70f5aa6de7ab3a7/ubuntu1604_bazel/BUILD#L52-L57) していたようで Bazel のチームが直接メンテしていたわけではない様子。
今は [Bazel のチームがメンテしようとしている](https://github.com/bazelbuild/continuous-integration/issues/1060) ようだがそれほど進展がなく特に新しいコンテナがリリースされているわけでもない。

自分用の CI などで Bazel のコンテナイメージに依存しているのでこの状況は好ましくなく、自分でコントロールすべく Bazel のコンテナイメージを作った。

[https://github.com/f110/bazel-container](https://github.com/f110/bazel-container)

いずれ Bazel からコンテナイメージが出てくるだろうと思われるので長いこと使ったりすることはあまり考えておらず、主に自分用のコンテナイメージとなっている。

コンテナイメージ自体は GHCR で公開をしているので誰でも利用することができるし、なにか不具合の報告があれば対応するつもりではいる。
