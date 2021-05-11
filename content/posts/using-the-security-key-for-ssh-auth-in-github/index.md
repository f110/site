---
title: GitHubのSSH認証でFIDO2が使えるようになった
date: 2021-05-11T11:07:00+09:00
isCJKLanguage: true
toc: true
tags: [Security]
---

[GitHub の SSH で FIDO2 による認証が使えるようになった](https://github.blog/2021-05-10-security-keys-supported-ssh-git-operations/) とのことなので実際に使ってみてしばらくセキュリティキーを使って認証してみることにした。

openssh 8.2 で FIDO2 対応が実装されたのは当時から知っていたが、自分の用途だとあまりそれを使う機会はなく実際のシーンで使ったことはなかった。
（ FIDO2 対応は openssh の HEAD にコミットされた時点で知った。）

せっかくなので手持ちの YubiKey 5C でやってみることにした。 

# macOS で使う場合

macOS を開発用のマシンとして使っている人でも ssh クライアントは macOS のものを使っていると思う。

しかし macOS の ssh クライアントは OpenSSH 8.1 なのでセキュリティキーを使うことができない。まずは最新の openssh を別にインストールする必要がある。

    $ sudo port install openssh +fido2

新し目の openssh が入ればそれだけで準備は完了。

# 鍵を生成する

まずセキュリティキー用の鍵を生成する必要がある。

既存の鍵をロードしたりすることはできず、セキュリティキーごとに鍵を生成してそれをリモートホストに登録しなければいけない。

    $ ssh-keygen -t ecdsa-sk

セキュリティキーを使う場合は末尾が `-sk` となっているタイプを指定する。
現時点で指定できるものは `ecdsa-sk` か `ed25519-sk` のどちらかである。

`ssh-keygen` を実行するとセキュリティキーをタップせよと言われるのでタップする。

通常の秘密鍵と扱いが異なるのはここだけでそれ以外はほぼ同じように扱うことができる。

# GitHub に公開鍵を登録する

    $ cat ~/.ssh/id_ecdsa_sk.pub | pbcopy

公開鍵も普段と同じように扱うことができるので、いつもどおり公開鍵を GitHub の UI から登録すれば良い。

    $ ssh -T git@github.com

で鍵が使えるかどうか試すことができる。

秘密鍵を利用する際はセキュリティキーをタップする。

# Resident key

`ssh-keygen` に特にオプションをつけずにセキュリティキーを使って秘密鍵を生成するとデバイスごとに秘密鍵を生成しリモートホストに登録することになる。

結局セキュリティキーがなければ使えないのであればセキュリティキー = 秘密鍵という状態（もしくはそれに近い状態）にし、鍵ごと可搬できても良いのではないだろうかと思う。

それに近い状態になるのが `-O resident` オプションである。

これをつけるとセキュリティキーに秘密鍵の "seed" を保存するので最初の鍵を生成したホスト以外でもセキュリティキーさえあれば同じ秘密鍵を生成することができる。

    $ ssh-keygen -t ecdsa-sk -O resident 

`resident` オプションをつけるだけである。（ PIN の入力が必要になる）

別のホストでこのセキュリティキーから秘密鍵を生成するには

    $ ssh-keygen -K

とすることで行うことができる。

秘密鍵が手に入れば公開鍵も簡単に生成できる。

    $ ssh-keygen -y -f ~/.ssh/id_ecdsa_sk > ~/.ssh/id_ecdsa_sk.pub

当然この秘密鍵を使用する際もセキュリティキーのタップが必要になる。

# 注意点

## 使用するたびにタップしなければいけない

pull と push のたびにセキュリティキーをタップしなければいけない。

タップしなくてもいいように `no-touch-required` をつけた鍵を生成すると GitHub の認証ができなくなる。
GitHub 側の sshd で `no-touch-required` がついていないようだ。

なので今のところ `no-touch-required` をすることはできない。

## Resident key を使う場合は新し目の YubiKey が必要

新し目といっても2019年9月以降に購入したものであれば良いのでそこまでではないかもしれない。（ YubiKey 5C は2018年に出たのでそれに飛びついた人じゃなければ新し目である可能性はまぁまぁ高い）

対応しているかどうかを見極めるにはファームウェアのバージョンをチェックする。

[5.2.3 以上](https://www.yubico.com/blog/whats-new-in-yubikey-firmware-5-2-3/) が必要で、ファームウェアのバージョンを見るには [YubiKey Manager](https://www.yubico.com/support/download/yubikey-manager/) を使うと簡単である。

## macOS の ssh-agent と相性が悪い

macOS の ssh-agent はキーチェーンアクセスに秘密鍵を保存できるものになっている。

ただこれは macOS のものなので当然 FIDO2 対応はされていない。

ssh-agent も新しい openssh のものであれば Resident key を使って ssh-agent に秘密鍵を登録できるので認証のたびにタップすることを回避できるとは思う。

ssh-agent と組み合わせて認証のたびのタップを回避しようと思ったがこれはあえなく失敗となった。

# 総評

まだ初日なので今後も運用していけるかどうかはもうしばらくしないと判断できない。

特に push や pull のたびにタップが必要なのが面倒に感じるようになるかもしれない。
とはいえ、現時点での感触としては毎回タップするコストは払えるといった感じである。

前からセキュリティキーを適度にマシンに刺したりする口実が欲しかった。

この SSH の認証にセキュリティキーを用いるというのはその口実として良さそうだという直感がある。

マシンにセキュリティキーを刺しっぱなしにはしたくないので使わないときは抜いて保管しているのだがそうすると本当に必要なときにしか刺さない。
今度は刺すのが面倒で本当に必要な時しか刺さなくなる。

例えば GitHub の sudo モードにはセキュリティキーで入れるがパスワードでも入れる。そういう場合、わざわざセキュリティキーを刺すのが面倒なのでパスワードを使ってしまう。

マシンを使っている時だけセキュリティキーが刺さっているという状態を作れればそういったシーンでもセキュリティキーの方が使いやすくなるし、WebAuthn が使えるならそっちを使おうというモチベーションになる。
（今のところ WebAuthn でログインできるまともなサイトは見たことがないが…）

# 参考

- [Security keys are now supported for SSH Git operations](https://github.blog/2021-05-10-security-keys-supported-ssh-git-operations/)
- [GitHub now supports SSH security keys](https://www.yubico.com/blog/github-now-supports-ssh-security-keys/)
- [[Server Login] OpenSSHでのU2F/FIDO2認証を検証してみる](https://blog.nicopun.com/post/2019-12-28-ssh_with_u2f_fido2/)
