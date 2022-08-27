---
title: GitOpsでシークレットを扱う
date: 2022-08-27T09:00:00+09:00
isCJKLanguage: true
tags: [Homecluster, Kubernetes]
---

自宅の Kubernetes クラスタは GitOps をしている。マニフェストを git で管理するのであれば GitOps をするようにしておいた方が良い。
GitOps をせずに手動での適用で運用しようと思うと必ず適用を忘れたりする。

しかし GitOps をしようとすると一つ Secret をどのように管理するかという問題がある。

リポジトリに秘密情報をそのまま含めたくなければなにか工夫をしなければいけない。

そこで Secret はマニフェストとしてリポジトリに含めるがファイル自体は暗号化しておき、それを復号化できるのは自分だけという状態にしていた。
これに利用していたのが [git-crypt](https://github.com/AGWA/git-crypt) である。

すべてのファイルを暗号化してしまうと [github.com](http://github.com) でファイルが全く見れなくなってしまうので秘密情報を含むファイルのみを暗号化する運用にしていた。

そして秘密情報を含むファイルのみ手動で適用してた。
ArgoCD で適用時に復号化し適用しても良いのだがそれをセットアップするのが面倒といった理由でやってこなかった。

更に git-crypt には別の問題があり、秘密情報を含むファイルを事前に .gitattributes に列挙しなければいけない。
事前にファイル名のルールを決めておくという手もあるが、必要に応じて列挙する運用をしていた。

秘密情報を含んだファイルをコミットした後に .gitattributes の編集忘れに気がつくということは一度や二度ではないくらいあり、これは非常に煩わしい。

そこでこれらを解決するために [argocd-vault-plugin](https://github.com/argoproj-labs/argocd-vault-plugin) を使っている。

# argocd-vault-plugin

argocd-vault-plugin は ArgoCD のプラグインで、マニフェストに書かれた情報を元に [Vault](https://www.vaultproject.io/) にアクセスしデータを埋め込んでくれる。
これにより Vault にあるデータを Secret オブジェクトにすることができ、かつそのデータをマニフェストに直接書く必要がない。

つまり Secret オブジェクトのマニフェストも ArgoCD で GitOps をすることができる。
マニフェストには秘密情報が書かれていないので何も注意せずにコミットすることができる。

    apiVersion: v1
    kind: Secret
    metadata:
      name: minio-token
      annotations:
        avp_path: "cluster/data/storage/token"
    type: Opaque
    stringData:
      accesskey: <accesskey>
      secretkey: <secretkey>

というようなマニフェストを書いておき、これを適用する ArgoCD のアプリケーションで argocd-vault-plugin を有効にするだけで良い。

GitOps をしていて Secret 管理に悩んでいる人はぜひやってみてほしい。

（なお ArgoCD でプラグインを使うのに若干の手間はある。v2.4.0 からサイドカーコンテナでプラグインを入れることができるらしいがまだ試していない）
