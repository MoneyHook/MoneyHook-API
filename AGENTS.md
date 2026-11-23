# MoneyHooks API Agent Guide

このファイルはリポジトリ全体に適用します。

- 依頼内容と変更対象を確認し、関連する既存コードと文書の該当節を読んでから変更する。[READMEの資料索引](README.md#ドキュメント)から確認先を選ぶ。
- 依頼範囲外のHTTP契約、DB schema、認証・認可、別リポジトリを変更しない。既存のHTTP契約や保存済みデータに影響する場合は、利用クライアントと互換性を先に確認する。[APIバージョンと契約](docs/ARCHITECTURE.md#apiバージョンと契約)
- route、Handler、Storeの配置と依存方向は[アーキテクチャ](docs/ARCHITECTURE.md#handler追加時の規則)を参照する。
- 変更に必要な検証を[開発ガイド](docs/DEVELOPMENT.md#検証)から選び、結果と必要だが未実行・失敗の項目を報告する。変更した仕様や手順は[ドキュメント方針](docs/DEVELOPMENT.md#ドキュメント方針)に沿って更新する。
- 明示的に依頼されない限り、branch作成、commit、push、PR作成、tag操作を行わない。操作時の詳細は[Git Workflow](docs/GIT_WORKFLOW.md)を参照する。
