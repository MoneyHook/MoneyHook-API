# MoneyHooks API Agent Guide

このファイルはリポジトリ全体に適用します。

## 作業方針

- 依頼内容と変更対象を確認し、関連する既存コードを読んでから変更する。
- APIの構造と依存規則は`docs/ARCHITECTURE.md`、DB migrationと開発環境は`docs/DEVELOPMENT.md`を参照する。
- branch、commit、PRの命名と言語ルールは`docs/GIT_WORKFLOW.md`を参照する。
- ユーザーが指定していないHTTP契約、DB schema、認証・認可、別リポジトリを変更しない。
- 既存のURL、JSON、status codeを変更する場合は、利用クライアントとOpenAPIへの影響を先に確認する。
- 明示的に依頼されない限り、branch作成、commit、push、PR作成、tag操作を行わない。

## Packageと依存規則

- 全業務routeは`app/handler/routes.go`へ登録する。
- HTTP handler、DTO、validation、response変換は`app/handler/<feature>/`へ置く。
- 機能別Handlerは必要なStore interfaceだけを受け取り、具体的なDB Storeをimportしない。
- DB StoreはHTTP packageをimportしない。
- 共通HTTP処理は`app/handler/internal/httpx`へ限定し、機能固有ロジックを共通化しない。
- domain packageとGo識別子は正しい英語とGo標準の命名を使う。DB table・column名とJSON field名は契約なので、綴り修正の対象に含めない。

## API契約

OpenAPIは隣接するReactリポジトリの`contracts/openapi.yaml`にあります。API契約変更が依頼に含まれる場合だけ、Reactリポジトリの契約と生成クライアントを同じ変更単位で更新してください。

構造整理では次を維持します。

- HTTP pathとmethod
- query・path parameter名
- request・response JSON
- status codeと認証要件
- DB schemaと保存済みデータ

## 検証

変更範囲に応じ、`app/`で次を実行します。

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

routeを変更した場合は`app/handler/routes_test.go`も更新します。未実行または失敗した確認項目は最終報告に明記してください。
