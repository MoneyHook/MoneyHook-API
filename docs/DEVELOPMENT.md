# 開発ガイド

## DBマイグレーションとシード

通常のテストスイートは、データベースを使わずにスキーマ定義とマスターデータ定義を検証します。

```sh
cd app
go test ./...
```

スキーママイグレーションは `go run ./cmd/migrate` で実行します。Cloud Run本番環境では、APIをリリースする前にCloud Run Jobでこのコマンドを実行します。通常のAPI起動処理はスキーマを変更しません。

`ENABLE_SEED_DATA` がGoの `strconv.ParseBool` でtrueとして解釈できる値（`true`、`TRUE`、`True`、`1` など）に設定されている場合、`go run ./cmd/migrate`で開発用サンプルデータも投入されます。未設定、空文字、またはfalseの場合はサンプルデータを投入しません。不正な値の場合はmigrationの実行に失敗します。

サンプルデータを有効にすると、固定開発ユーザーの取引、固定費、予算、支払い方法、カスタム／非表示サブカテゴリ、表示設定が、起動のたびに現在のサンプルシナリオへ再生成されます。開発ユーザーの変更内容を保持したい場合は、このフラグを無効にしてください。

PostgreSQLでは、最初に `postgres` メンテナンスデータベースへ接続し、`POSTGRES_DATABASE` が存在しなければ作成します。そのため、初回起動時には設定したPostgreSQLユーザーにデータベース作成権限が必要です。

通常の `docker compose up` はFirebase Auth Emulatorのhealthy確認後、`ENABLE_SEED_DATA=true`でmigration・master data・sample dataを実行してからAPIを起動します。API起動時は `ENABLE_DEVELOPMENT_USER=true` の場合に固定UID `a77a6e94-6aa2-47ea-87dd-129f580fb669` の開発用Googleユーザーがprovisionされます。

Dev Containerでは専用Compose上書きによりgoコンテナだけを待機させます。Run and Debugから `Seed Database (Migration + Sample Data)` を実行した後、`Launch Echo Server via Air (Hot Reload + Debug)` を実行してください。seedの有効化はAPI起動とは独立して `ENABLE_SEED_DATA` で制御します。

E2E用の `compose.e2e.yaml` は通常Composeの一括フローを引き継ぎ、E2E用のCORS originだけを上書きします。これらの開発用フラグは本番環境では有効にしないでください。

## 家族の家計 v1

`go run ./cmd/migrate` は家族の所属・招待・記録用テーブルと、個人取引のversion/論理削除、入力先デフォルト設定を追加します。既存取引の共有は自動で開始しません。家族の操作履歴は保存せず、旧 `household_event` がある場合はテーブルを削除します。旧取引履歴テーブルがある場合は、控えの最新訂正を `household_entry_data.corrected_payload` へ移してから旧テーブルを削除します。移行前にバックアップを取得し、旧APIを停止してmigration後に新APIを配備してください。過去の操作・変更履歴は削除され、復元にはバックアップが必要です。家族への所属記録があるサンプルユーザーは、個人原本と家族の参照を保つためサンプル再生成をスキップします。

控え作成時の原本の版は保存せず、旧 `household_entry_data.source_version` 列がある場合はmigrationで削除します。控えの内容・保存日時・最新訂正は保持します。

招待の発行・照合には `HOUSEHOLD_INVITATION_SECRET` が必要です。32byte以上のランダムな値を秘密管理から渡してください。未設定時は招待処理だけが503を返します。Composeの固定値は開発専用です。鍵を変更すると既発行の未使用招待は照合できなくなるため、再発行が必要です。招待code/tokenの平文はDBへ保存しません。

Reactを公開する前にmigration、API、Reactの順で配備します。CORSでは `Idempotency-Key` を許可します。接続元の試行制限はsocket peerを使い、未検証のforwarded headerは信頼しません。プロキシ配下ではIP制限が共有されるので、アカウント制限と併せて実際の構成で確認してください。

PostgreSQLの専用テストDBではmigrationの再実行、共有の権限、退出時の控え、招待の同時承諾を検証します。本番のCockroachDBでのDDL/制約と40001再試行は、公開前に別途検証してください。冪等キーは現在期限なしで保存するため、保存期間と掃除の運用は別途決めます。

```sh
cd app
MIGRATION_TEST_POSTGRES_DSN='postgres://moneyhook:password@localhost:5432/moneyhook?sslmode=disable' \
go test -tags=integration ./db/migration ./store_postgres
```

日次ジョブAPIの起動には `JOB_NAME`、`SCHEDULER_AUDIENCE`、`SCHEDULER_SERVICE_ACCOUNT_EMAIL` が必要です。本番ではCloud SchedulerのOIDC audienceとservice accountを同じ値に設定します。通常のComposeにはローカル起動用の値が設定されていますが、実際のジョブ呼び出しにはGoogle署名付きOIDC ID tokenが必要です。

本番の初回構成では `moneyhook-scheduler@moneyhooks.iam.gserviceaccount.com` をScheduler専用identityとして作成します。このidentityには対象Cloud Run serviceのInvokerだけを付与します。GitHub Actionsのデプロイidentityには、専用identityに対するService Account Userと、カスタムロール `projects/moneyhooks/roles/moneyHookSchedulerJobUpdater`（`cloudscheduler.jobs.get`、`cloudscheduler.jobs.update`）を付与します。以降のOIDC設定はデプロイworkflowが既存Scheduler jobへ反映します。

## マイグレーション統合テスト

統合テストは、指定されたPostgreSQLサーバー上に専用スキーマを作成して実行し、終了時に削除します。

```sh
cd app
MIGRATION_TEST_POSTGRES_DSN='postgres://moneyhook:password@localhost:5432/moneyhook?sslmode=disable' \
go test -tags=integration ./db/migration -v
```

指定するPostgreSQLユーザーにはスキーマの作成・削除権限が必要です。テスト用のDDL権限を付与しても問題ない認証情報だけを使用してください。
