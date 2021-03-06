---
title: キーボードをKinesisにした
date: 2020-12-19T17:06:00+09:00
isCJKLanguage: true
tags: []
---

ここ1年くらいは [HHKB Professional HYBRID Type-S](https://happyhackingkb.com/jp/products/hybrid_types/) を使っていた。
1年くらい前にHHKBを買ったのは前からHHKBは欲しかったんだけど、日本語配列+黒+静音という組み合わせがなかったから。
HHKBが新しくなってこの組み合わせのモデルがあるということで迷わず買った。

このキーボードはかなり気に入っているので2台持ってて、1台は自宅用もう1台は会社用として買った。
が、3月以降はずっとフルリモートなので会社用は必要なく家で眠ってる状態になってる。

で、HHKBを今年はずっと使ってきたんだけども久しぶりにKinesis Advantage2を引っ張り出し使い始めた。

### なぜHHKBを使っていたのか

これには2つ理由があり

1. 打ち心地が大変好み
2. PC2台に簡単に接続できる

打ち心地は静電容量式のスイッチと勝負できるものがなく、大変気持ちが良い。
タイピングが気持ちいいというのは重要で、キーボードを打たないと仕事にならないのでそこが気持ちいいとストレスにならない。

PC2台と接続できるのもかなり便利である。
仕事をする時は会社のPCで、仕事が終われば自分のマシンを使うわけなので切り替える必要がある。机の上にそれぞれのPCに対応したキーボードを置いておくのは邪魔なのでやりたくない。
また机の上に2つキーボードあると姿勢が中途半端になってしまい、腰や背中を痛める原因になる。

HHKBでキーボードの切り替えは、片方をBluetooth接続、もう片方を有線接続していてショートカットで切り替えていた。
自分のマシンはデスクトップだがBluetooth接続はできるのでどちらもBluetooth接続にすれば机の上からケーブルを1本減らせるので良かったのだが自分のマシンはほぼLinuxで使っているので無線接続は不安なので有線接続にしていた。

### HHKBの有線接続はまだバグってる

有線と無線接続を切り替えて使っていたので有線接続のバグを発見した。

これはPFUに報告して、向こうでも再現したようだが現時点のファームウェアではまだ修正されていない（なお報告して再現したのは3月。もう9ヶ月も直せていないよう）

USBから給電が途切れないタイプのポートに接続している場合、有線接続のままOSを終了すると無線接続等に切り替えることができなくなる。
この状態からの復旧方法としてはケーブルを1回抜くことである。そうするとキーボード側も一旦シャットダウンし、再度電源ボタンを長押しすれば電源を入れられる。

おそらくこのバグはファームウェアがこういった接続先を考慮できていなく、また終了時のステート遷移の処理が甘いのが原因だろう。

他にもワークアラウンドはあって、有線接続のままPCをシャットダウンしなければいい。
OSがシャットダウンしきる前に無線接続に切り替えたらこの問題は起きない。

### なぜKinesisに戻したのか

そんな理由はなく、半分は気分、もう半分は自作キーボードの画像を何枚か見たのでまたこの形のキーボードを使いたくなったから。つまり100%気分である。

自作キーボードは極一部を除いてフラットなキー配置だが、Kinesisのこのお碗型のキー配置はとても良い。
最小の移動でどのキーにも届くし、一般的なキーボードに比べて一行多く押せるキーが多いのも特徴。（Zの行より下にキーがあり、ここも人差し指から小指で押すことができる。矢印キーなどが配置されている）

自作キーボードはDactyl Keyboardなら作ってみたいと思っているが3Dプリンターは持ってないのでケースを購入する必要があり、そうすると$350とかなのでKinesisでいいか、となる。（Kinesis Advantage2は$349、国内代理店から買うと43,000円程）

### PCの切り替え

KinesisにはUSBケーブルが1本しか生えてないので工夫する必要がある。

そこで [サンワサプライのPC切替器](https://www.elecom.co.jp/products/KM-A22BBK.html) を買った。ボタンを押せばUSBの接続先が切り替わるという代物でPCの切り替えはこれで問題なさそうである。

### 日本語配列から英語配列

実は日本語配列と英語配列を混ぜて使うのは問題ない。

元来日本語配列派（一番下の行のキーが日本語配列の方が多いので。英語配列のキーボードは基本的に買う気がない）なのでMacBook Proの内蔵キーボードもJIS配列だし、HHKBも日本語配列を買っている。
Kinesisは英語配列しかないので英語配列で使っているが、両方の配列を交互に使っても特に問題ない。
どちらかしか使えないという人が多いらしいが、別に自分はどちらでも良いし両方を同時に使うことができる。

### ショートカット問題

Kinesisは親指で押せるキーが多いのでそれと組み合わせたショートカットを設定したい。
元々HHKBでもLinuxではxkbで少しキーマップをいじっていたのでKinesis用の設定を作った。

他にIDEのショートカットも便利だと思う配置がHHKBとは異なる部分がある。
これはKinesis用のキーマップを作って切り替えることで対応した。
