# Airについて

Golangのホットリロードツールで、ソースコードの変更を自動的に検出して際ビルドしてくれる

[AirのGitリポジトリ](https://github.com/cosmtrek/air)


## 前提
Golangのインストールが済み、開発環境が整っていること

`go version`などのコマンドでバージョンが確認できる状態


## インストール

`go install github.com/cosmtrek/air@latest`でインストールを実行する

zshを使っている場合、`alias air="$(go env GOPATH)/bin/air"`を.zshrcに追加してパスを通す

### インストールの確認
`air -v`で以下のような表示ならOK
```
  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.49.0, built with Go go1.21.6
```


## 初期化

Airは、.air.tomlの情報を元に実行できるため、`air init`を実行してデフォルトの設定ファイルを作成する（これで十分）


## 実行

`air -c .air.toml`でアプリケーションが実行できる

`tmp/main`が作成され、それが実行ファイルになっている模様


## `air`で実行することについて
`air init`せず(.air.tomlを作成せず)、`air`コマンドだけでもプロジェクトを実行可能<br>
おそらく、デフォルトの`.air.toml`の状態で実行しているので、カスタムをしたいならtomlファイルをいじって実行する必要がある
