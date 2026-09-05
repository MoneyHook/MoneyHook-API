# MoneyHooks API

MoneyHooksのバックエンドAPIです。GoとEchoでHTTP APIを提供し、Firebase Authenticationで利用者を認証します。データベースはPostgreSQLを使用します。

## ローカル開発

Docker Composeを使うと、PostgreSQL、Firebase Auth Emulator、APIをまとめて起動できます。通常のComposeではFirebase Auth Emulatorのhealthy確認後、PostgreSQLのmigration・master data・sample dataを実行してからAPIを起動します。

```bash
docker compose up --build
```

APIは`http://localhost:8080`で起動します。次の公開エンドポイントで起動状態を確認できます。

```bash
curl http://localhost:8080/
```

業務APIは`/api`以下にあり、GoogleプロバイダーのFirebase ID tokenをBearer tokenとして要求します。ローカル構成ではFirebase Auth Emulatorを使用します。

通常のComposeでは`ENABLE_SEED_DATA=true`により、起動のたびに固定UID `a77a6e94-6aa2-47ea-87dd-129f580fb669`の開発ユーザーに紐づくsample dataを再生成します。API起動時は`ENABLE_DEVELOPMENT_USER=true`により、Firebaseの開発ユーザー（表示名「開発ユーザー」、`developer@example.com`）もprovisionされます。

Dev Containerは専用Compose上書きによりgoコンテナを待機状態で起動します。Run and Debugから次の順番で実行してください。

1. `Seed Database (Migration + Sample Data)`
2. `Launch Echo Server via Air (Hot Reload + Debug)`

sample dataを再生成せずmigrationだけを実行したい場合は、`ENABLE_SEED_DATA=false go run ./cmd/migrate`を使用してください。これらの開発用フラグは本番環境では有効にしないでください。

E2Eの`compose.e2e.yaml`は通常Composeの自動migration・seed・API起動を引き継ぎ、E2E用のCORS originだけを上書きします。

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
