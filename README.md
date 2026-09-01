# MoneyHooks API

MoneyHooksのバックエンドAPIです。GoとEchoでHTTP APIを提供し、Firebase Authenticationで利用者を認証します。データベースはPostgreSQLを使用します。

## ローカル開発

Docker Composeを使うと、PostgreSQL、Firebase Auth Emulator、APIをまとめて起動できます。

```bash
docker compose up --build
```

APIは`http://localhost:8080`で起動します。次の公開エンドポイントで起動状態を確認できます。

```bash
curl http://localhost:8080/
```

業務APIは`/api`以下にあり、GoogleプロバイダーのFirebase ID tokenをBearer tokenとして要求します。ローカル構成ではFirebase Auth Emulatorを使用します。

Compose起動時は、Firebase Auth Emulatorの初期化、PostgreSQLのmigration・master data・sample data、API起動、開発ユーザーのprovisionの順で処理されます。`compose.yaml`では`ENABLE_SEED_DATA=true`と`ENABLE_DEVELOPMENT_USER=true`を設定しているため、固定UID `a77a6e94-6aa2-47ea-87dd-129f580fb669`のGoogleユーザー（表示名「開発ユーザー」、`developer@example.com`）と、そのユーザーに紐づくサンプルデータが利用できます。

`ENABLE_SEED_DATA=true`で起動するたびに、この開発ユーザーの取引・固定費・予算履歴・支払い方法・カスタム設定は最新のサンプルシナリオへ再生成されます。手動で変更した内容を残したい場合は、このフラグを無効にしてください。

開発ユーザー作成は`ENABLE_DEVELOPMENT_USER`がtrue、かつサンプルデータ投入が有効な場合だけ実行されます。未設定またはfalseがデフォルトで、本番環境では開発用フラグを有効にしないでください。

## 開発コマンド

Goコマンドはmodule rootの`app/`で実行します。

```bash
cd app
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

DB migrationと開発環境の詳細は[開発ガイド](docs/DEVELOPMENT.md)を参照してください。

## ドキュメント

- [アーキテクチャ](docs/ARCHITECTURE.md): ディレクトリ構造、package責務、依存方向、HTTP処理の流れ
- [開発エージェント向けガイド](AGENTS.md): 変更時の作業方針と検証ルール
- OpenAPI契約: 隣接する`moneyhooks-react/contracts/openapi.yaml`

API実装、DB migration、認証・認可はこのリポジトリを正本とします。HTTP契約を変更する場合は、ReactリポジトリのOpenAPI契約と生成クライアントも同じ変更単位で更新してください。
