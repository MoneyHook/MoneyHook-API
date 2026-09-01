# 開発ガイド

## DBマイグレーションとシード

通常のテストスイートは、データベースを使わずにスキーマ定義とマスターデータ定義を検証します。

```sh
cd app
go test ./...
```

スキーママイグレーションは `go run ./cmd/migrate` で実行します。Cloud Run本番環境では、APIをリリースする前にCloud Run Jobでこのコマンドを実行します。通常のAPI起動処理はスキーマを変更しません。

`ENABLE_SEED_DATA` がGoの `strconv.ParseBool` でtrueとして解釈できる値（`true`、`TRUE`、`True`、`1` など）に設定されている場合、開発用サンプルデータも投入されます。未設定、空文字、またはfalseの場合はサンプルデータを投入しません。不正な値の場合はAPIの起動に失敗します。

サンプルデータを有効にすると、固定開発ユーザーの取引、固定費、予算、支払い方法、カスタム／非表示サブカテゴリ、表示設定が、起動のたびに現在のサンプルシナリオへ再生成されます。開発ユーザーの変更内容を保持したい場合は、このフラグを無効にしてください。

PostgreSQLでは、最初に `postgres` メンテナンスデータベースへ接続し、`POSTGRES_DATABASE` が存在しなければ作成します。そのため、初回起動時には設定したPostgreSQLユーザーにデータベース作成権限が必要です。

Compose起動時はマイグレーション・シード後にAPIを起動します。`ENABLE_DEVELOPMENT_USER=true` かつ `ENABLE_SEED_DATA=true` の場合に限り、API起動時に固定UID `a77a6e94-6aa2-47ea-87dd-129f580fb669` の開発用Googleユーザーがprovisionされます。これらの開発用フラグは本番環境では有効にしないでください。

## マイグレーション統合テスト

統合テストは、指定されたPostgreSQLサーバー上に専用スキーマを作成して実行し、終了時に削除します。

```sh
cd app
MIGRATION_TEST_POSTGRES_DSN='postgres://moneyhook:password@localhost:5432/moneyhook?sslmode=disable' \
go test -tags=integration ./db/migration -v
```

指定するPostgreSQLユーザーにはスキーマの作成・削除権限が必要です。テスト用のDDL権限を付与しても問題ない認証情報だけを使用してください。
