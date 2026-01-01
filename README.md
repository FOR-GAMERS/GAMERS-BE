# GAMERS Server

GAMERSプラットフォームのバックエンドAPIサーバー

## 技術スタック
- Go **1.25**
- Gin Framework
- GORM
- Swagger
- Docker
- GoLang-migrate 

## 実行方法

### 1. 環境設定
```bash
# .envファイルを作成
cp env/.env.example env/.env
```

`env/.env`ファイルを開いて、データベース設定を入力してください:
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=gamers_db
PORT=8080
```

### 2-A. Dockerで実行（推奨）
```bash
cd docker
docker-compose up -d
```

### 2-B. ローカルで実行
```bash
# 依存関係のインストール
go mod download

# Wire生成（初回のみ）
go install github.com/google/wire/cmd/wire@latest
wire ./cmd

# Swaggerドキュメント生成
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server.go --output docs

# サーバー実行
go run ./cmd
```

### 3. 接続確認
サーバーが正常に実行されると、次のURLで確認できます:
- サーバー: http://localhost:8080
- Health Check: http://localhost:8080/health
- Swaggerドキュメント: http://localhost:8080/swagger/index.html

### テスト実行
```bash
# 全テスト
go test ./...

# 特定パッケージのテスト
go test ./test/user/...
```

## デプロイ

### CI/CD パイプライン
このプロジェクトは GitHub Actions を使用した自動デプロイメントをサポートします。

#### 🐳 Docker イメージビルド
- `main` / `develop` ブランチへの push 時に自動実行
- GitHub Container Registry (GHCR) にイメージを公開
- タグ戦略: `latest`, ブランチ名, セマンティックバージョン

#### 🚀 GCP Compute Engine デプロイ
- Docker イメージビルド完了後に自動実行
- SSH 経由で VM に安全にデプロイ
- 自動ヘルスチェックとロールバック機能

詳細な設定手順は [DEPLOYMENT.md](./docs/DEPLOYMENT.md) を参照してください。

### クイックデプロイメント
```bash
# Production 環境
git checkout main
git merge develop
git push origin main

# Staging 環境
git checkout develop
git push origin develop

# バージョンタグ
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## ドキュメント
- [API リファレンス](./docs/API_REFERENCE.md) - RESTful API の詳細仕様
- [アーキテクチャ](./docs/ARCHITECTURE.md) - システムアーキテクチャとデザインパターン
- [デプロイメント](./docs/DEPLOYMENT.md) - GCP への自動デプロイ設定ガイド
- [モニタリング](./docs/MONITORING.md) - 監視とロギング設定
- [テスト](./docs/TESTING.md) - テスト戦略とカバレッジ
- [ERD ダイアグラム](./docs/ERD_DIAGRAM.md) - データベーススキーマ設計

## Author
| Sunwoo An                                                             |
|-----------------------------------------------------------------------|
| <img src="https://www.github.com/Sunja-An.png" width=240 height=240 > |
| Nationality - 🇰🇷 Republic of Korea                                  |
| role - FullStack Developer                                            |
