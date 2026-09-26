# MoneyHooks API アーキテクチャ

## ディレクトリ構造

```text
moneyHook_api/
├── app/
│   ├── main.go                  # プロセス起動、middleware、依存の組み立て
│   ├── handler/
│   │   ├── handler.go           # Dependenciesから機能別Handlerを構築
│   │   ├── routes.go            # 全業務APIのルート定義
│   │   ├── internal/httpx/      # 認証contextとv1共通HTTP処理
│   │   ├── transaction/         # 旧取引APIとv1取引API
│   │   ├── analytics/           # v1分析APIと集計処理
│   │   ├── fixed/               # 固定費API
│   │   ├── category/            # カテゴリAPI
│   │   ├── subcategory/         # サブカテゴリAPI
│   │   ├── payment/             # 支払方法API
│   │   └── job/                 # 日次ジョブAPI
│   ├── category/                # Category Store interface
│   ├── fixed/                   # Fixed Store interface
│   ├── job/                     # Job Store interface
│   ├── paymentresource/         # Payment resource Store interface
│   ├── subcategory/             # Subcategory Store interface
│   ├── transaction/             # Transaction Store interfaceとdomain error
│   ├── user/                    # User Store interfaceとdomain error
│   ├── store_postgres/          # PostgreSQL向けStore実装
│   ├── db/                      # DB接続、Store生成、migration
│   ├── model/                   # DB・domain間で共有するデータ構造
│   ├── router/                  # Firebase初期化、開発ユーザーprovision、認証middleware
│   └── message/                 # 旧APIのメッセージ取得
├── firebase/                    # Firebase Emulator構成
├── psql/                        # PostgreSQLローカル構成
└── compose.yaml
```

## 起動時の初期化

`cmd/migrate`はDB接続、schema migration、master data、`ENABLE_SEED_DATA`に応じたsample data投入を担当します。`main.go`はFirebase Auth clientとDB Storeを初期化し、`ENABLE_DEVELOPMENT_USER=true`の場合に固定UIDの開発ユーザーを冪等にprovisionしてからHTTP APIを起動します。API起動処理はDB schemaやsample dataを変更しません。

通常のComposeではFirebase Auth Emulatorのhealthy確認後、`cmd/migrate`とAPIを順に起動します。Dev Containerは専用Compose上書きでGoサービスを待機させ、Run and Debugの `Seed Database (Migration + Sample Data)` と `Launch Echo Server via Air (Hot Reload + Debug)` がDB初期化とAPI起動をそれぞれ担当します。E2E用Composeは通常の一括フローを引き継ぎます。

開発ユーザーのUID・表示名・emailは`app/common`で定義し、Auth provisionとsample seedで共有します。Auth userが既に存在する場合は必要なプロフィールとGoogle provider情報を補正し、provider UIDの競合や予期しないAuthエラーは起動失敗として扱います。

## 依存方向

```text
main
  ├─ db ──> store_postgres ──> model
  └─ handler root
       ├─> feature handlers ──> feature Store interfaces ──> model
       └─> internal/httpx ──> router authentication
```

- `main`がPostgreSQLの具体的なStore実装を`handler.Dependencies`へ渡します。
- `handler/routes.go`がURLとHTTPメソッドの唯一の一覧です。機能別Handlerはルートを登録しません。
- 機能別Handlerは必要なStore interfaceだけに依存し、`store_postgres`を直接importしません。
- Store実装はHTTP packageをimportしません。DB固有の処理は各Store実装に閉じ込めます。
- 複数機能で共有するHTTP処理は`handler/internal/httpx`に限定します。機能固有のDTO、validation、response変換は各機能packageに置きます。

## HTTPリクエストの流れ

1. `main.go`が公開ヘルスチェック`GET /`を登録します。
2. `handler.Register`が一般業務APIにFirebase認証middlewareを適用し、`/api/job/daily`にはScheduler専用OIDC認証middlewareを適用します。
3. 一般業務APIではFirebase ID tokenとGoogle provider、verified emailを検証し、解決した`user_no`をrequest contextへ保存します。
4. 日次ジョブではGoogle署名、固定audience、許可したScheduler service accountのemailを検証し、Schedulerヘッダーを追加確認します。
5. 機能別Handlerが入力を読み取り、Store interfaceを呼び出し、HTTP responseへ変換します。
6. PostgreSQLのStoreが永続化処理を行います。

## APIバージョン

既存クライアント向けの旧APIとReact向けv1 APIは同時に提供します。

- v1: `/api/v1/...`。未知のJSON fieldを拒否し、`status`、`code`、`message`を持つエラー形式を使用します。
- 旧API: `/api/transaction`、`/api/fixed`など。後方互換性のため既存のrequest・response・status codeを維持します。
- 新しいReact機能は原則としてv1へ追加します。旧APIの置換や削除は、利用クライアントとOpenAPI契約を確認した別変更として扱います。

HTTP契約の機械可読な正本は隣接する`moneyhooks-react/contracts/openapi.yaml`です。パス、method、query、JSON、status code、認証要件を変更する場合はOpenAPIと生成クライアントを同時に更新します。

## Handler追加時の規則

- 既存機能のendpointは該当する`handler/<feature>`へ追加します。
- 新しい機能は専用packageと、必要なStore interfaceを作成します。
- routeは必ず`handler/routes.go`へ追加し、`handler/routes_test.go`の期待値も更新します。
- request/response DTOは機能packageに置き、共通`request`・`response` packageを再作成しません。
- `handler/internal/httpx`へ追加するのは、複数機能で同じHTTP意味を持つ小さな処理だけです。

## 既知の課題

今回の構造整理では、次の既存挙動を変更していません。

- 旧取引APIの登録・一括登録・編集は、日付、金額、符号、名称、ID、必須項目を保存前に検証します。固定費・支払方法など他の旧APIには、request tagによるvalidationが未実装の箇所が残っています。
- 旧APIではendpointによってエラーstatus codeとresponse形式が異なります。
- `model`はDB record、Store入出力、HTTP変換元の構造を広く共有しています。

これらを改善する場合は、構造変更と混ぜず、HTTP・DB互換性とOpenAPI更新を含む独立した変更として扱います。

## 旧取引APIのエラーと保存境界

旧取引APIの取得処理はStoreのDBエラーをHTTP 500として返し、取引詳細のデータ不存在だけを404として扱います。取引の登録・一括登録・編集と固定費の登録・編集は、サブカテゴリの検索・作成から本体の保存までを同じDBトランザクションで実行します。失敗時は新規サブカテゴリもロールバックし、同名サブカテゴリの同時作成は既存の一意制約で重複を防ぎます。対象のない編集は422を返します。
