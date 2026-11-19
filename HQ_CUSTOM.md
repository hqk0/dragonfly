# 浜急線向けDragonfly差分

このディレクトリは`github.com/df-mc/dragonfly`の
`v0.11.6`を基にしています。接続は上流標準のBedrockプロトコルを使い、
旧バージョン用の互換アダプタは登録しません。

浜急線サーバーに必要な乗車基盤およびリソース保護として、次の最小差分を加えています。

- プレイヤーの`Mount` / `Dismount` API
- Bedrockの乗車リンク送信
- 座席位置メタデータ
- クライアントの降車パケット処理
- リソースパック暗号化キー（`.key` / `key.txt` / `key`）の自動検出と `ContentKey` 適用機能 (`server/conf.go` 内の `loadResources`)
- UUIDごとのHTTP(S) CDN配布と、CDN上のUUID/version整合性検証
- 必須リソースディレクトリの欠落・空ディレクトリ・CDN設定漏れで起動を中止する検証
- 暗号化キー読み込み処理の単体テスト (`server/conf_test.go` 内の `TestLoadResourcesWithKeyFile`)

上流Dragonflyを更新する場合は、この差分と乗車テスト・リソースパック読み込みテストを先に確認してください。
