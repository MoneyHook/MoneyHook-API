# MoneyHooks API アーキテクチャ

## 主な責務

| 参照先 | 責務 |
| --- | --- |
| [main.go](../app/main.go) | 起動時の依存の組み立て、共通middleware、公開ヘルスチェック |
| [cmd/migrate](../app/cmd/migrate) | DB接続、schema migration、master data、設定に応じたsample data |
| [handler](../app/handler) | route登録、機能別のHTTP処理 |
| [router](../app/router) | Firebase初期化、開発ユーザーのprovision、認証middleware |
| [db](../app/db)、[store_postgres](../app/store_postgres) | DB接続とStoreの作成、永続化 |
| [model](../app/model) | DBとdomain間のデータ構造 |

API起動処理はDB schemaやsample dataを変更しません。開発ユーザーのprovisionは冪等で、既存ユーザーのプロフィールとGoogle provider情報を補正します。provider UIDの競合や予期しないAuthエラーは起動エラーとして扱います。[ローカルの起動とDB初期化](DEVELOPMENT.md#ローカル環境とdb初期化)

## 依存方向

```text
main
  ├─ db ──> store_postgres ──> model
  └─ handler root
       ├─> feature handlers ──> feature Store interfaces ──> model
       └─> internal/httpx ──> router authentication
```

`main` が具体的なStore実装を `handler.Dependencies` へ渡します。機能別Handlerは必要なStore interfaceだけに依存します。StoreはHTTP packageをimportせず、DB固有の処理をStore実装に閉じ込めます。

## HTTPリクエストの流れ

1. `main.go`が公開ヘルスチェック`GET /`を登録します。
2. `handler.Register`が一般業務APIにFirebase認証middlewareを適用し、`/api/job/daily`にはScheduler専用OIDC認証middlewareを適用します。
3. 一般業務APIではFirebase ID tokenとGoogle provider、verified emailを検証し、解決した`user_no`をrequest contextへ保存します。
4. 日次ジョブではGoogle署名、固定audience、許可したScheduler service accountのemailを検証し、Schedulerヘッダーを追加確認します。
5. 機能別Handlerが入力を読み取り、Store interfaceを呼び出し、HTTP responseへ変換します。
6. PostgreSQLのStoreが永続化処理を行います。

## APIバージョンと契約

既存クライアント向けの旧APIとReact向けv1 APIを同時に提供します。

- v1は `/api/v1/...` 配下です。未知のJSON fieldを拒否し、`status`、`code`、`message` を持つエラー形式を使用します。新しいReact機能は原則としてv1へ追加します。
- 旧APIは既存クライアントとの互換性を維持します。置換や削除は、利用クライアントへの影響を確認した別変更として扱います。

HTTP契約の正本はReactリポジトリの [`contracts/openapi.yaml`](https://github.com/MoneyHook/MoneyHook-FRONT/blob/develop/contracts/openapi.yaml) です。route・method、parameter、request/response、status code、認証要件の変更は、[React開発ガイドの「API契約の変更」](https://github.com/MoneyHook/MoneyHook-FRONT/blob/develop/docs/DEVELOPMENT.md)を参照してください。構造整理ではこれらのHTTP挙動、DB schema、保存済みデータを維持します。

## Handler追加時の規則

- 業務routeは [`handler/routes.go`](../app/handler/routes.go) に登録し、[`handler/routes_test.go`](../app/handler/routes_test.go) の期待値も更新します。機能別Handlerでrouteを登録しません。
- Handler、DTO、validation、response変換は対応する `handler/<feature>` に置きます。新機能は専用packageと必要なStore interfaceを作り、[依存方向](#依存方向)を守ります。
- 共通HTTP処理は [`handler/internal/httpx`](../app/handler/internal/httpx) に置きます。複数の機能で同じ意味を持つ小さな処理に限り、機能固有の処理は機能packageに置きます。
- domain packageとGo識別子は正しい英語とGo標準の命名を使います。DB table・column名とJSON field名は互換性のため綴りを維持します。

## 既知の課題

- 旧取引APIの登録・一括登録・編集は、日付、金額、符号、名称、ID、必須項目を保存前に検証します。固定費・支払方法など他の旧APIには、request tagによるvalidationが未実装の箇所が残っています。
- 旧APIではendpointによってエラーstatus codeとresponse形式が異なります。
- `model`はDB record、Store入出力、HTTP変換元の構造を広く共有しています。

これらを改善する場合は、HTTPとDBの互換性を確認し、API契約の更新を含む独立した変更として扱います。

## 旧取引APIのエラーと保存境界

旧取引APIの取得処理はStoreのDBエラーをHTTP 500として返し、取引詳細のデータ不存在だけを404として扱います。取引の登録・一括登録・編集と固定費の登録・編集は、サブカテゴリの検索・作成から本体の保存までを同じDBトランザクションで実行します。失敗時は新規サブカテゴリもロールバックし、同名サブカテゴリの同時作成は既存の一意制約で重複を防ぎます。対象のない編集は422を返します。
