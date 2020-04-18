---
title: "VPNをWireGuardにした"
date: 2020-04-18
isCJKLanguage: true
tags: ["WireGuard", "Home Network"]
---

自宅には信用できない経路を使わざるおえない時用にVPNを張れるようにしています。
今まではUniFi Security Gatewayが持っているVPN機能（L2TP over IPSec）を使っていたのですが、これを `WireGuard <https://www.wireguard.com/>`_ も使えるようにしました。

前提
======

使ってるハードウェア
-----------------------

* UniFi Security Gateway
* UniFi nanoHD
* UniFi Switch 8 60W

やりたいこと
--------------

* **全てのトラフィックを** VPN経由にする
* 内部ネットワークと通信できる
* Guestネットワークとは通信できない

仕様
------

* クライアントのレンジは ``192.168.12.0/24``
* 通信するポートは 51820

インストール
===============

EdgeRoute向けのパッケージをビルドしてくれている人がいるのでそれをありがたく使います。 `github.com/Lochnair/vyatta-wireguard <https://github.com/Lochnair/vyatta-wireguard>`_

最新版をUniFi Security Gateay（以下USG）まで持っていってインストールします。

.. code:: console

    usg$ sudo dpkg -i wireguard-ugw3-0.0.20191219-2.deb

設定
======

秘密鍵と公開鍵の生成
------------------------

.. code:: console

    usg$ wg genkey | tee /config/auth/wg.key | wg pubkey > /config/auth/wg.public

インターフェースの設定
-------------------------

WireGuardのPeer（クライアント）は ``192.168.12.0/24`` のレンジを使うことにします。

.. code:: console

    usg$ configure
    usg$ set interfaces wireguard wg0 address 192.168.12.1/24
    usg$ set interfaces wireguard wg0 listen-port 51820
    usg$ set interfaces wireguard wg0 route-allowed-ips true
    usg$ set interfaces wireguard wg0 private-key /config/auth/wg.key

.. _configure-peer:

Peer（クライアント）の設定
-----------------------------

ここではまずクライアント側で秘密鍵と公開鍵を生成してそれを設定します。

.. code:: console

    usg$ set interfaces wireguard wg0 peer bwqE81/MNgb/D6klMd+AFGGB3FXBVRv1RC+p8JTk6wE= endpoint [your vpn host or ip]:51820
    usg$ set interfaces wireguard wg0 peer bwqE81/MNgb/D6klMd+AFGGB3FXBVRv1RC+p8JTk6wE= allowed-ips 1921.168.12.2/32

firewallの設定
---------------

USGの外からUDP/5120に届いたパケットを許可する。

.. code:: console

    usg$ set firewall name WAN_LOCAL rule 20 action accept
    usg$ set firewall name WAN_LOCAL rule 20 protocol udp
    usg$ set firewall name WAN_LOCAL rule 20 description 'WireGuard'
    usg$ set firewall name WAN_LOCAL rule 20 destination port 51820

SNATする。インターネットに出れるようにするために必要です。

.. code:: console

    usg$ set firewall group network-group wireguard_network description "Remote WireGuard (VPN) subnets"
    usg$ set firewall group network-group wireguard_network network 192.168.12.0/24
    usg$ set service nat rule 5001 description "MASQ wireguard_network to WAN"
    usg$ set service nat rule 5001 log disable
    usg$ set service nat rule 5001 outbound-interface eth0
    usg$ set service nat rule 5001 protocol all
    usg$ set service nat rule 5001 source group network-group wireguard_network
    usg$ set service nat rule 5001 type masquerade

Guestネットワークと通信できないようにする。

.. code:: console

    usg$ set firewall name GUEST_IN rule 20 description "drop packet to wireguard network"
    usg$ set firewall name GUEST_IN rule 20 action drop
    usg$ set firewall name GUEST_IN rule 20 destination group network-group wireguard_network

保存
------

.. code:: console

    usg$ commit
    usg$ save
    usg$ exit

クライアント側の設定
======================

macOS
-------

`Mac App Store の WireGuardクライアント <https://apps.apple.com/us/app/wireguard/id1451685025>`_ を入れます。

新しいTunnelを追加して以下を書いてSaveします。

この時、サーバーのPublicKeyが必要になるのでUSGから手に入れておきます。 ``usg$ cat /config/auth/wg.public``

.. _peer-conf:

.. code::

    [Interface]
    PrivateKey = <Generated private key>
    Address = 192.168.12.2/32
    DNS = 192.168.12.1

    [Peer]
    PublicKey = <Server's Public key>
    AllowedIPs = 0.0.0.0/0
    Endpoint = [your vpn host or ip]:51820

``AllowedIPs`` を ``0.0.0.0/0`` とすることで全てのトラフィックがVPNに流れます。

モバイル
---------

iOS、Androidともに公式のストアからWireGuardのアプリをインストールすることが出来ます。

モバイルでも上記のような設定をすることにはなるのですが、これを手書きするのは大変なので代替手段が用意されています。
もちろん手書きでも設定はできるので最初の疎通テストなどは手書きで頑張ってもいいでしょう。実際頑張りました。

手書きをする場合は双方のPublic Keyを交換する必要があるので何らかの手段で交換してください。
自分はDropboxにテキストファイルを置いて交換しました。

他のデバイスではこの方法は面倒なのでQRコードを使って設定します。

.. code:: console

    usg$ wg genkey | tee peer-privatekey | wg pubkey > peer-publickey

``peer-publickey`` を使って Peer の設定をUSG側で行います。 `Peerの設定 <#configure-peer>`_

Private KeyとPublic Keyを適当なマシンに持ってきて `設定ファイル <#peer-conf>`_ を書きます。

.. code:: console

    $ sudo apt install -y qrencode
    $ qrencode -t ansiutf8 < peer.conf

で表示されたQRコードをクライアントアプリで読み取れば設定が完了します。

最後に
=======

これで快適なVPN生活になることを期待しています。

今はFree WiFiを使う機会がないのでVPNも不要なのですが。

参考
======

* https://github.com/Lochnair/vyatta-wireguard
* https://wiki.archlinux.org/index.php/WireGuard
* https://grh.am/2018/wireguard-setup-guide-for-ios/
