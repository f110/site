---
title: "Wireguard and Unifi Controller"
date: 2020-04-21
isCJKLanguage: true
tags: ["WireGuard", "Home Network", "UniFi"]
---

UniFi Controllerは内部に持っているMongoDBに設定が保存されており、プロビジョニング（設定の反映）を行うとCLIから行ったものは消えてしまいます。

なので [先のポスト](../vpn-with-wireguard) のように設定をするとプロビジョニングのたびに手動でconfigureしないといけなくなります。

UniFi Controllerにはコントローラの設定にオーバーレイする形で独自の設定をインジェクトできる仕組みがあるのでそれを使ってWireGuardの設定を行います。

# オーバーレイ

適切なパスに `config.gateway.json` というファイルを作るとUniFi Controllerはこのファイルの内容をマージしてから各デバイスに設定を反映します。

これを使いWireGuardの設定を行います。
まず `config.gateway.json` は以下のような内容にします。

```json
{
  "firewall": {
  "group": {
    "network-group": {
    "wireguard_network": {
      "description": "Remote WireGuard (VPN) subnets",
      "network": [
      "192.168.12.0/24"
      ]
    }
    }
  },
  "name": {
    "WAN_LOCAL": {
    "rule": {
      "20": {
      "action": "accept",
      "description": "WireGuard",
      "protocol": "udp",
      "destination": {
        "port": "51820"
      }
      }
    }
    },
    "GUEST_IN": {
    "rule": {
      "20": {
      "action": "drop",
      "description": "drop packet to wireguard network",
      "destination": {
        "group": {
        "network-group": "wireguard_network"
        }
      }
      }
    }
    }
  }
  },
  "interfaces": {
  "wireguard": {
    "wg0": {
    "listen-port": "51820",
    "private-key": "/config/auth/wg.key",
    "address": [
      "192.168.12.1/24"
    ],
    "route-allowed-ips": "true",
    "peer": [
      {
      "bwqE81/MNgb/D6klMd+AFGGB3FXBVRv1RC+p8JTk6wE=": {
        "endpoint": "[your vpn host or ip]:51820",
        "allowed-ips": [
        "192.168.12.2/32"
        ]
      }
      }
    ]
    }
  }
  },
  "service": {
  "nat": {
    "rule": {
    "5001": {
      "type": "masquerade",
      "description": "MASQ wireguard_network to WAN",
      "log": "disable",
      "outbound-interface": "eth0",
      "protocol": "all",
      "source": {
      "group": {
        "network-group": "wireguard_network"
      }
      }
    }
    }
  }
  }
}
```

CLIで行っていた設定を適切なデータ構造に置き換えただけです。

現在の設定をJSON形式でみたい場合は

```console
usg$ mca-ctrl -t dump-cfg
```

で手に入れることができます。

現在の設定をJSON形式を参考にしながらWireGuard用のJSONを作りました。

# 配置

ファイルを作るよりファイルを適切なパスに置く方が少し難しいです。

UniFi Controllerをどこかのホストで動かしているのであれば `<unifi_base>/data/sites/<site_id>` に置くだけですが、うちの場合はUniFi Controllerがk8sのPodとして動作しています。

UniFi Controllerのデプロイはhelmを使っていて [Chart はstable](https://github.com/helm/charts/tree/master/stable/unifi) を使っています。
このchartには `config.gateway.json` を差し込む方法が用意されていないのでchartを修正する必要があります。

## chartの修正

1. ConfigMapで `config.gateway.json` を保持する
1. Deploymentで `/unifi/data/sites/<site_id>/config.gateway.json` にマウントする
1. DeploymentのPodTemplateにjsonファイルのハッシュ値を書き込んでおいて変更されたらPodを再作成するようにする

という変更を加えます。

helmでファイルを差し込むには現状chartをまるごと持ってくるしかないので自分のリポジトリにすべてコピーしてきましょう。

実際に加えた変更はプライベートリポジトリなのでここではcommitをお見せすることは出来ません。
代わりにdiffを載せておくのでこれを参考に修正してください。

```diff
diff --git a/chart/files/config.gateway.json b/chart/files/config.gateway.json
new file mode 100644
index 0000000..64020c6
--- /dev/null
+++ b/chart/files/config.gateway.json
@@ -0,0 +1,2 @@
+{
+}
\ No newline at end of file
diff --git a/chart/templates/configmap.yaml b/chart/templates/configmap.yaml
index 463abb1..94723b2 100644
--- a/chart/templates/configmap.yaml
+++ b/chart/templates/configmap.yaml
@@ -10,4 +10,18 @@ metadata:
     app.kubernetes.io/managed-by: {{ .Release.Service }}
 data:
 {{ toYaml .Values.extraConfigFiles | indent 2 }}
+---
 {{- end }}
+{{- if .Values.customSiteConfig }}
+apiVersion: v1
+kind: ConfigMap
+metadata:
+  name: {{ template "unifi.fullname" . }}-site
+  labels:
+    app.kubernetes.io/name: {{ include "unifi.name" . }}
+    helm.sh/chart: {{ include "unifi.chart" . }}
+    app.kubernetes.io/instance: {{ .Release.Name }}
+    app.kubernetes.io/managed-by: {{ .Release.Service }}
+data:
+{{ (.Files.Glob "files/config.gateway.json").AsConfig | indent 2 }}
+{{- end }}
\ No newline at end of file
diff --git a/chart/templates/deployment.yaml b/chart/templates/deployment.yaml
index c93d444..60082c8 100644
--- a/chart/templates/deployment.yaml
+++ b/chart/templates/deployment.yaml
@@ -26,11 +26,9 @@ spec:
       labels:
         app.kubernetes.io/name: {{ include "unifi.name" . }}
         app.kubernetes.io/instance: {{ .Release.Name }}
-      {{- if .Values.podAnnotations }}
+      {{- if .Values.customSiteConfig }}
       annotations:
-        {{- range $key, $value := .Values.podAnnotations }}
-        {{ $key }}: {{ $value | quote }}
-        {{- end }}
+        checksum/site-config: {{ .Files.Get "files/config.gateway.json" | sha256sum }}
       {{- end }}
     spec:
       containers:
@@ -118,6 +116,10 @@ spec:
             - name: extra-config
               mountPath: /configmap
             {{- end }}
+            {{- if .Values.customSiteConfig }}
+            - name: custom-site-config
+              mountPath: /unifi/data/sites/default
+            {{- end }}
           resources:
 {{ toYaml .Values.resources | indent 12 }}
       volumes:
@@ -133,6 +135,11 @@ spec:
           configMap:
             name: {{ template "unifi.fullname" . }}
         {{- end }}
+        {{- if .Values.customSiteConfig }}
+        - name: custom-site-config
+          configMap:
+            name: {{ template "unifi.fullname" . }}-site
+        {{- end }}
     {{- with .Values.nodeSelector }}
       nodeSelector:
 {{ toYaml . | indent 8 }}
diff --git a/chart/values.yaml b/chart/values.yaml
index 4609dbd..f0bccad 100644
--- a/chart/values.yaml
+++ b/chart/values.yaml
@@ -244,6 +244,8 @@ extraConfigFiles: {}
   #     </Loggers>
   #   </Configuration>
 
+customSiteConfig: false
+
 resources: {}
   # We usually recommend not to specify default resources and to leave this as a conscious
   # choice for the user. This also increases chances charts run on environments with little
```

# 反映

jsonを更新してコントローラを再起動しただけだと反映しないはずです。

強制的に反映させるには `Devices -> Security Gateway -> Config -> Manage Device -> Force Provision` を行う必要があります。

# 参考

* https://help.ui.com/hc/en-us/articles/215458888-UniFi-USG-Advanced-Configuration-Using-config-gateway-json
* https://graham.hayes.ie/posts/wireguard-%2B-unifi/
