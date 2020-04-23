---
title: MetalLBからBGPで広報する
date: 2020-04-23
isCJKLanguage: true
tags: ["Kubernetes", "homecluster", "UniFi"]
---

自宅のKubernetesクラスタはLoadBalancerとしてMetalLBを使っていてExternalIPはARPを使って広報しています。

自宅で使う上では特にこれで不満もないけども上位のルータがBGPを喋ることができるのでせっかくなのでBGPで広報するレンジも追加してみました。

その一連の作業ですがBGPで広報する必要がなければARPで十分だと思います。
今は2つのレンジを使っていますがBGPだけにしようとは考えてないです。

UniFi Security Gateway側
===========================

まずはUSG側でBGPを話せるようにする必要があるので以下の設定をします。

.. code:: console

    $ configure
    $ set protocols bgp 64512 parameters router-id 192.168.1.1
    $ commit
    $ save

ついでにneighborも設定しておきます。

.. code:: console

    $ set protocols bgp 64512 neighbor 192.168.100.2 remote-as 64513
    $ set protocols bgp 64512 neighbor 192.168.100.3 remote-as 64513

今回はルータ側を64512、MetalLB側を64513のeBGPにしています。

もちろんこのままだとコントローラのプロビジョニングやアップデートで戻ってしまうのでJSONに書いておきます。

.. code:: console

	{
		"protocols": {
			"bgp": {
				"64512": {
					"neighbor": {
						"192.168.100.2": {
						"remote-as": "64513"
						},
						"192.168.100.3": {
							"remote-as": "64513"
						}
					},
					"parameters": {
						"router-id": "192.168.1.1"
					}
				}
			}
		}
	}

これらの設定を反映してUSG側でneighborの情報が見れれば大丈夫です。
この時点で実際のneighborがいないのでBGPのstateはEstablishedにはなりません。

.. code:: console

    $ show ip bgp neighbor

MetalLB側
============

ConfigMapにMetalLBが使用するアドレスのレンジが書かれているのでこれを変更します。

.. code:: yaml

    apiVersion: v1
    kind: ConfigMap
    metadata:
      namespace: metallb-system
      name: config
    data:
      config: |
        address-pools:
          - name: default
            protocol: layer2
            addresses:
              - 192.168.100.128/25
          - name: bgp
            protocol: bgp
            addresses:
              - 192.168.101.0/24
            auto-assign: false
            avoid-buggy-ips: true
        peers:
          - peer-address: 192.168.1.1
            peer-asn: 64512
            my-asn: 64513

``address-pools`` に ``protocol: bgp`` な要素を追加するのと ``peers`` に自分とピアの設定をしておきます。

後はこれをapplyすればMetalLBのコントローラが自動的に読み込み直します。

BGPのレンジは ``auto-assign: false`` なので明示しない限り使われません。

動作テスト
===========

まずはServiceを作ります。

既存の動作しているServiceの定義を持ってきて修正を加えるのがいいと思います。

.. code:: yaml

    apiVersion: v1
    kind: Service
    metadata:
      name: test
    spec:
      type: LoadBalancer
      loadBalancerIP: 192.168.101.32
      ports:
      - name: http
        port: 80
        protocol: TCP
        targetPort: 4002
      selector:
        app: proxy

例えばこのような定義を書きます。 ``loadBalancerIP`` で新しいアドレスのレンジ内を指定するのがポイントです。

このyamlをapplyすればUSG側から広報された経路を見ることができるはずです。

.. code::

    admin@SecurityGateway:~$ show ip bgp
    BGP table version is 0, local router ID is 192.168.1.1
    Status codes: s suppressed, d damped, h history, * valid, > best, i - internal,
                  r RIB-failure, S Stale, R Removed
    Origin codes: i - IGP, e - EGP, ? - incomplete

       Network          Next Hop            Metric LocPrf Weight Path
    *  192.168.101.32/32
                        192.168.100.3                          0 64513 ?
    *                   192.168.100.2                          0 64513 ?


もし経路が広報されてこない場合はServiceの裏にいるPodがReadyかどうか確認してみてください。
Podが全てUnhealthyでトラフィックをバランシング出来ない場合、MetalLBは経路を広報しません。

なぜ全て切り替えないのか
===========================

自宅ではARPで十分。

参考
=====

* http://blog.cowger.us/2019/02/10/using-metallb-with-the-unifi-usg-for-in-home-kubernetes-loadbalancer-services.html
* https://medium.com/@ipuustin/using-metallb-as-kubernetes-load-balancer-with-ubiquiti-edgerouter-7ff680e9dca3
