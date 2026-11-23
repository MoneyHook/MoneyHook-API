# 開発ガイド

## ローカル環境とDB初期化

起動と疎通確認は[README](../README.md#ローカル開発)を参照する。環境変数・ポート・起動順序の設定は[Compose](../compose.yaml)で確認できる。

- 通常のComposeはFirebase Auth Emulatorがhealthyになってから、migration、master data、sample dataの順にDBを初期化し、APIを起動する。APIはAuth Emulatorへ `firebase:9099` で接続する。
- Dev Containerは[Compose上書き](../.devcontainer/compose.devcontainer.yaml)でGoサービスを待機させる。[Run and Debug](../.vscode/launch.json)の `Seed Database`、`Launch API (Air)` の順に実行する。
- [E2E用Compose](../compose.e2e.yaml)は通常の起動手順を引き継ぎ、E2E用のCORS originを上書きする。

DB初期化コマンドは `app/` から `go run ./cmd/migrate` を実行する。最初に `postgres` メンテナンスDBへ接続して、`POSTGRES_DATABASE` が存在しなければ作成する。初回作成時はPostgreSQLユーザーにDB作成権限が必要。

通常のComposeは次の開発用フラグを有効にする。本番環境では両方を無効にする。

- `ENABLE_SEED_DATA` はmigration時のサンプルデータ投入を制御する。`strconv.ParseBool` が受け付ける真値（`true`、`1` など）で投入し、未設定・空文字・偽値で省略する。不正な値ではmigrationが失敗する。
- `ENABLE_DEVELOPMENT_USER` はAPI起動時に固定UIDの開発用Googleユーザー（表示名「開発ユーザー」、メール `developer@example.com`）をAuth Emulatorに用意する。[UIDなどの定義](../app/common)

seedを有効にすると、固定開発ユーザーの取引、固定費、予算、支払方法、カスタム／非表示サブカテゴリ、表示設定を実行のたびに再生成する。再生成する個人取引に紐づく家族の共有記録・控えも削除する。家族の所属、他の利用者の取引、原本を参照しない代理入力は保持する。開発ユーザーの変更を保持してmigrationするには `ENABLE_SEED_DATA=false go run ./cmd/migrate` を実行する。

## 検証

Goのコマンドはmodule rootの `app/` で実行する。[Goのバージョン](../app/go.mod)

| 変更 | 選ぶ検証 |
| --- | --- |
| 文書 | `git diff --check`、変更したリンクと記述の確認 |
| Goコード | 変更したGoファイルを `gofmt -w` で整形し、変更に関係するpackage・ケースを `go test` で実行 |
| route・HTTP method・認証境界 | `go test ./handler` と、変更した機能・認証middlewareのテスト |
| migration・DB制約・SQL・トランザクション | 対象packageのテストと、変更したDB挙動に該当する[PostgreSQL統合テスト](#postgresql統合テスト) |
| 起動時の依存組み立て・build設定 | 影響する起動テストと、対象の `go build` |
| HTTP契約 | Goの関連テストと[React開発ガイドの「API契約の変更」](https://github.com/MoneyHook/MoneyHook-FRONT/blob/develop/docs/DEVELOPMENT.md) |

routeのみのcaseを確認する例は `go test ./handler -run '^TestRegisterIncludesCompleteBusinessRouteSet$' -v`。対象を絞った場合は、出力でテストの実行を確認する。

全体の `go test ./...`、`go vet ./...`、`go build ./...` は、全体に影響する変更、必要なCIチェック、または明示的な依頼に応じて選ぶ。本番のデプロイでは [workflow](../.github/workflows/deploy.yml) が `go test ./...` を実行する。通常テストではスキーマ定義やマスターデータ定義をDB接続なしで検証する。

### PostgreSQL統合テスト

専用のテスト用PostgreSQLと、スキーマを作成・削除できる認証情報を使う。各テストは専用スキーマを作成し、終了時に削除する。`MIGRATION_TEST_POSTGRES_DSN` が空の場合はスキップされる。出力で対象のcaseの実行を確認し、スキップは未実行として報告する。

```sh
cd app
export MIGRATION_TEST_POSTGRES_DSN='postgres://moneyhook:password@localhost:5432/moneyhook?sslmode=disable'
go test -tags=integration ./db/migration -v
```

migrationは `./db/migration`、Storeの権限・共有・ロールバック・並行処理などは `./store_postgres` を指定する。両方に影響する場合は同じコマンドに両packageを指定できる。

## 家族機能の移行と公開

本番ではAPIの配備前にCloud Run Jobでmigrationを実行する。手順2の旧API停止は家族機能の初回移行時に別途実施する。[デプロイ workflow](../.github/workflows/deploy.yml)

1. 移行前に利用者DBのバックアップと復元手順を確認する。本番相当のCockroachDBでDDLの再実行、partial index、ロック、制約、`40001` の再試行を検証する。PostgreSQL統合テストとは別に実施する。
2. 家族機能の初回移行では、旧APIを停止してからmigrationを実行する。家族の所属・招待・記録、個人取引のversionと論理削除、入力先の既定設定を導入する。既存の原本は非共有、家族参照は空のままにし、個人設定の所有者を維持する。
3. 移行前後で既存原本の件数、ユーザー・月ごとの符号付き金額合計、参照整合性を照合する。初回公開時は既存原本の共有が0件であることと、migrationの再実行で重複しないことを確認する。
4. 招待の発行・照合には `HOUSEHOLD_INVITATION_SECRET` を秘密管理から渡す。32byte以上のランダム値を使い、未設定時は招待処理だけが503を返す。鍵を変更すると既発行の未使用招待は再発行が必要。招待code/tokenの平文はDBに保存しない。[Compose](../compose.yaml)の固定値は開発専用。
5. [React開発ガイドの「API契約の変更」](https://github.com/MoneyHook/MoneyHook-FRONT/blob/develop/docs/DEVELOPMENT.md)に従って契約と生成物を同期し、旧API・v1ともに共有と論理削除に対応したAPIを全インスタンスに配備する。CORSで `Idempotency-Key` を許可する。招待の接続元制限はsocket peerを使う。プロキシ配下でIP制限が共有される影響をアカウント制限と併せて確認する。退出、控え、旧APIの整合性を確認してからReactを公開する。
6. 切り戻し時はReactでの公開を止め、必要に応じて家族データと論理削除に対応したAPIへ戻す。家族データや論理削除の利用開始後は、控えを保持できる互換APIを使用する。

旧 `household_event` と旧取引履歴テーブルがある場合はmigrationで削除される。控えの最新訂正は先に `household_entry_data.corrected_payload` へ移す。旧 `household_entry_data.source_version` 列も削除し、控えの内容・保存日時・最新訂正は保持する。削除される過去の操作・変更履歴の復元にはバックアップが必要。冪等キーは期限なしで保存され、保存期間と掃除の運用は未決定。

## 日次ジョブ

APIは起動時に `JOB_NAME`、`SCHEDULER_AUDIENCE`、`SCHEDULER_SERVICE_ACCOUNT_EMAIL` を読み取る。ローカルの起動用設定は[Compose](../compose.yaml)を参照する。実際のジョブ呼び出しには、設定と一致するGoogle署名付きOIDC ID tokenが必要。[認証の流れ](ARCHITECTURE.md#httpリクエストの流れ)

本番の初回構成では、[デプロイ workflow](../.github/workflows/deploy.yml)に指定したScheduler専用サービスアカウントに、対象Cloud RunサービスのInvokerだけを付与する。デプロイ用サービスアカウントには、Scheduler専用アカウントに対するService Account Userと、Scheduler jobの参照・更新（`cloudscheduler.jobs.get`、`cloudscheduler.jobs.update`）を許可するカスタムロール `moneyHookSchedulerJobUpdater` を付与する。以降のOIDC設定はworkflowが既存Scheduler jobに反映する。

## ドキュメント方針

利用者向け仕様、HTTP契約、設定項目、起動・検証・運用手順、作業ルールが変わるときは、変更と同時に該当する文書の正本を更新する。関連する文書を検索してコード・設定と照合し、古い説明や重複を同じ変更で整理する。文書だけの変更でも記述の根拠と変更したリンクを確認する。

- HTTPの入出力と認証は[API契約](ARCHITECTURE.md#apiバージョンと契約)、DBの制約と移行処理は[Go migration](../app/db/migration)を起点に確認する。
- 起動順、ポート、環境変数の実値は [Compose](../compose.yaml)、配備ブランチやジョブの実行順は[デプロイ workflow](../.github/workflows/deploy.yml)を確認する。本文には操作や判断に必要な情報を置く。
- 運用手順はこのガイド、設計上の制約は[アーキテクチャ](ARCHITECTURE.md)、Gitの手順は[Git Workflow](GIT_WORKFLOW.md)にまとめる。他の文書からは関連する節へリンクする。検証・配備の履歴には確認時点を記す。
