---
title: UniFi Dream Machine Proのファームウェアをアップデートした
date: 2021-03-03T12:56:00+09:00
isCJKLanguage: true
tags: [UniFi]
---

UniFi Dream Machine Pro の [新しいファームウェア v1.9.0](https://community.ui.com/releases/UniFi-Dream-Machine-Firmware-1-9-0/36607188-4bbb-420a-9749-5af3eb85e522) が出ていたのでアップデートした。

アップデートは GUI もしくはコンソールから行うことができるが、今回は GUI から行った。

### アップデート後の感想

いまのところ特に問題はない。

うちの UDMP はコントローラだけではなくいくつかのコンテナを動かしていて、それを動かすために [udm-utilities](https://github.com/boostchicken/udm-utilities) を使用している。
これも特に問題なくファームウェアアップデート後にリブートしてもプロセスは再起動した。
