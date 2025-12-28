# GAMERS Server

GAMERSプラットフォームのバックエンドAPIサーバー

## 技術スタック
- Go 1.25
- Gin Framework
- GORM
- Wire (DI)
- Swagger
- Docker

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

## Author
| Sunwoo An                                                             |
|-----------------------------------------------------------------------|
| <img src="https://www.github.com/Sunja-An.png" width=240 height=240 > |
| Nationality - 🇰🇷 Republic of Korea                                  |
| role - FullStack Developer                                            |
