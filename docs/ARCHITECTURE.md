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

`main.go`はFirebase Auth clientを初期化し、`ENABLE_SEED_DATA=true`かつ`ENABLE_DEVELOPMENT_USER=true`の場合に固定UIDの開発ユーザーを冪等にprovisionします。その後、DB接続、migration、master data、sample dataを実行してからHTTP APIを起動します。サンプルデータが有効な場合、固定開発ユーザーのユーザー固有データは毎回シナリオへ再生成されます。ComposeではFirebase Auth EmulatorがhealthyになってからAPIコンテナを起動します。

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
2. `handler.Register`が`/api`以下にFirebase認証middlewareを適用します。
3. middlewareがFirebase ID tokenとGoogle provider、verified emailを検証し、解決した`user_no`をrequest contextへ保存します。
4. 機能別Handlerが入力を読み取り、Store interfaceを呼び出し、HTTP responseへ変換します。
5. PostgreSQLのStoreが永続化処理を行います。

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

- 旧APIはrequest structにvalidation tagがあってもEcho Validatorを実行しておらず、一部の不正入力はDB処理まで拒否されません。
- 旧APIではendpointによってエラーstatus codeとresponse形式が異なります。
- `model`はDB record、Store入出力、HTTP変換元の構造を広く共有しています。

これらを改善する場合は、構造変更と混ぜず、HTTP・DB互換性とOpenAPI更新を含む独立した変更として扱います。
