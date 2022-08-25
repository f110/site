---
title: KubernetesのCRDをProtocol Buffersで定義する
date: 2022-08-25T21:07:00+09:00
isCJKLanguage: true
tags: [Kubernetes]
---

自宅の Kubernetes クラスタでは自分のワークロードに合わせたコントローラを書いていて [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) を定義している。

コントローラを書くフレームワークはいくつかあり、それぞれで API の定義方法が少しずつ異なる。
自作コントローラの API の定義方法はその進化に合わせて少しずつ変わっていき最終的には自作のツールで定義しコード生成するようになった。

ここではなぜ[自作ツール](https://github.com/f110/kubeproto)にたどり着いたのかについて紹介する。

# client-go 時代

コントローラを書き始めた最初の頃は [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) を使っていたがこれは早々に [client-go](https://github.com/kubernetes/client-go) を使ったものに置き換えられた。
置き換えたのはちょうどその頃コントローラを書くことが仕事になり、そうした方が都合が良かったのと直接キューを扱って細かい制御をしたかったため。

API の定義は Go のコードで行われていて [code-generator](https://github.com/kubernetes/code-generator) でクライアントや informer を生成していた。
CRD の yaml は kubebuilder をそのまま使っていて、code-generator と kubebuilder を Bazel のサンドボックス下で実行するためのルールも作成して使っていた。

# 自作ツールへ

前述の code-generator & kubebuilder on Bazel は少し遅いながらもしばらく使っていたがある時から使えなくなってきたので最終的に自作するしかないなという結論になった。

code-generator はジェネレータとしてかなりナイーブな実装になっており思いがけず大掛かりな計算をしている。
API が定義されたコードとそれが依存している **ソースコード** をすべて読み込もうとする点が一番使いづらかった。

API の定義では必ず `[k8s.io/apimachinery/pkg/apis/meta/v1](http://k8s.io/apimachinery/pkg/apis/meta/v1)` や `[k8s.io/api/core/v1](http://k8s.io/api/core/v1)` をインポートすることになるが code-generator は更にこれらが依存しているソースコードもすべて読み込もうとする。
最終的には Go 本体のコード（ `GOROOT`以下のコード）も必要になるというところが扱いづらい。

そのため、Bazel のサンドボックス下で code-generator を実行する時は Go 本体のコードもすべてサンドボックスに入れざる負えなかった。

しばらくはそういうワークアラウンドで問題がなかったのだが、Go が embed を実装した頃からこれが難しくなってきた。
Go 本体のコードでも embed を使うようになると code-generator が読み込むコードでも embed をしなければいけなくなる。
これは通常の環境では特に問題がないのだが Bazel のサンドボックス下では結構面倒で、通常サンドボックス下ではファイルの書き込みができない。

つまり実質的に code-generator on Bazel が使えなくなったので自作ツールに移行した。

# 自作ツールはどうなっているのか

API の定義は Protocol Buffers で行うようになっている。
Protocol Buffers はまさに API の定義なども行うことができる DSL なので非常に向いている。

protoc（Protocol Buffersのコンパイラ）のプラグイン機構を利用して Kubernetes 用のいくつかのファイルを生成するプラグインを書き、それを Bazel から実行している。

.proto ファイルから以下のものを生成している。

1. Go の struct の定義
2. struct の DeepCopy 関数
3. Client
4. Informer
5. Lister
6. CRD の yaml
7. テスト用のモッククライアント

struct の定義だけではなくクライアントや Informer など一式を生成している。

Informer や Lister は code-generator と同様のインターフェースを生成すれば入れ替えが楽になると思われるがあえてインターフェースにも若干の変更を加えている。

code-generator が生成する Informer は

    client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})

というようなインターフェースになっている。

このようなインターフェースだと interface を使ってモックをしたりするのが綺麗にできないのとパッケージから公開されているものが interface （ここで言えば `Pods()` の返り値が `type PodInterface interface`）になっているのも全く好きじゃなかった。

interface をエクスポートするのはそうでなければ実現できない強い理由がある時にだけ行うべきでほぼすべてのケースではそれは必要ないと考えている。

なのでこういった好きじゃない点を直した自分好みなクライアント・Informer や Lister を自作ツールで生成している。
