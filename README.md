# MoneyHooks API

MoneyHooksのバックエンドAPIです。GoとEchoでHTTP APIを提供します。

## ローカル開発

Docker ComposeでPostgreSQL、Firebase Auth Emulator、APIを起動します。

```bash
docker compose up --build
```

公開ヘルスチェックでAPIの起動を確認します。

```bash
curl http://localhost:8080/
```

Firebase Emulator UIは `http://localhost:4000/auth`、ホストからのAuth Emulator接続先は `http://localhost:9099` です。

一般業務APIはFirebaseのGoogleログインを使います。日次ジョブはCloud Scheduler用のOIDC認証を使います。[HTTPリクエストの流れ](docs/ARCHITECTURE.md#httpリクエストの流れ)

## ドキュメント

| 確認する内容 | 文書 |
| --- | --- |
| Compose・Dev Container・E2E・seed | [ローカル環境とDB初期化](docs/DEVELOPMENT.md#ローカル環境とdb初期化) |
| 変更範囲に応じたGo検証・PostgreSQL統合テスト | [検証](docs/DEVELOPMENT.md#検証) |
| 家族機能の本番migration・公開・切り戻し | [家族機能の移行と公開](docs/DEVELOPMENT.md#家族機能の移行と公開) |
| 日次ジョブの認証・Scheduler設定 | [日次ジョブ](docs/DEVELOPMENT.md#日次ジョブ) |
| package責務・依存方向・API契約 | [アーキテクチャ](docs/ARCHITECTURE.md) |
| branch・commit・PR・release | [Git Workflow](docs/GIT_WORKFLOW.md) |
| エージェント向け作業方針 | [AGENTS.md](AGENTS.md) |
