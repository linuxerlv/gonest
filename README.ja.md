# Gonest

> ASP.NET CoreとNestJSに触発された、シンプルで優雅なGoウェブフレームワーク

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/linuxerlv/gonest/workflows/CI/badge.svg)](https://github.com/linuxerlv/gonest/actions)

[English](README.en.md) | [中文](README.zh.md)

---

## コア機能

| 機能 | 説明 |
|------|------|
| 🏗️ **インターフェース優先アーキテクチャ** | `core/abstract/` でインターフェースを定義、`core/` で実装、拡張はインターフェースに依存 |
| 🔌 **モジュール設計** | ミドルウェア、プロトコル、タスク調度は独立したパッケージ - 必要なものだけ使用 |
| ⚡ **細粒度インターフェース** | ContextAbstract = RequestReader + ResponseWriter + ValueStore - 必要に応じて合成 |
| 📦 **本番対応** | Config (Koanf)、ログ (Zap)、認証、認可 - 完全なソリューション |

---

## 5分で始める

### インストール

```bash
go get github.com/linuxerlv/gonest
```

### クイックスタート

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/middleware/cors"
    "github.com/linuxerlv/gonest/middleware/recovery"
)

func main() {
    // 方法1: クイック作成（シンプルなアプリケーション向け）
    app := core.CreateApplication()
    app.Use(recovery.New(nil))
    app.Use(cors.New(nil))
    app.GET("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })
    app.Run()
}
```

### ビルダーパターン（推奨、依存性注入対応）

```go
package main

import (
    "github.com/linuxerlv/gonest/core"
    "github.com/linuxerlv/gonest/core/abstract"
    "github.com/linuxerlv/gonest/config"
)

func main() {
    // ビルダーを作成
    builder := core.CreateBuilder()
    
    // 設定と環境変数を設定（公開プロパティ）
    builder.Config = config.NewKoanfConfig(".")
    builder.Env.Set("APP_ENV", "production")
    
    // DIコンテナにサービスを登録（Wire注入に対応）
    builder.Services.AddSingleton(&MyService{})
    builder.Services.AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
        return &DbContext{DSN: builder.Env.Get("DATABASE_URL")}
    })
    
    // アプリケーションをビルド
    app := builder.Build().(*core.WebApplication)
    
    // DIコンテナからサービスを取得
    service := core.GetService[*MyService](app.Services)
    
    // ルートを登録
    app.MapGet("/hello", func(ctx abstract.ContextAbstract) error {
        return ctx.JSON(200, map[string]string{"message": "Hello World"})
    })
    
    app.Run()
}
```

---

## プロジェクト構造

```
gonest/
├── core/
│   ├── abstract/              # インターフェース定義（細粒度、合成可能）
│   │   ├── context.go         # ContextAbstract, RequestReaderAbstract, ResponseWriterAbstract...
│   │   ├── router.go          # RouterAbstract, RouteBuilderAbstract...
│   │   ├── middleware.go      # MiddlewareAbstract...
│   │   ├── di.go              # ServiceCollectionAbstract...
│   │   ├── env.go             # EnvAbstract 環境変数インターフェース
│   │   └── ...                # その他のインターフェース
│   │
│   ├── context.go             # HttpContext実装
│   ├── router.go              # HttpRouter実装
│   ├── application.go         # Application実装
│   ├── builder.go             # WebApplicationBuilder実装
│   ├── env.go                 # Env環境変数実装
│   └── ...                    # その他の実装
│
├── config/                    # 設定モジュール
│   ├── config.go              # Configインターフェース（abstract.ConfigAbstractを実装）
│   └── koanf.go               # Koanf実装
│
├── logger/                    # ロガーモジュール
│   ├── logger.go              # Loggerインターフェース（abstract.LoggerAbstractを実装）
│   └── zap.go                 # Zap実装
│
├── middleware/                # ミドルウェア拡張パッケージ
│   ├── auth/                  # JWT認証
│   ├── session/               # セッション管理
│   ├── casbin/                # Casbin権限制御
│   ├── cors/                  # CORS
│   ├── recovery/              # Panicリカバリー
│   ├── ratelimit/             # レート制限
│   ├── timeout/               # タイムアウト制御
│   ├── gzip/                  # Gzip圧縮
│   ├── security/              # セキュリティヘッダー
│   ├── logger/                # ロガーミドルウェア
│   └── requestid/             # リクエストID
│
├── protocol/                  # プロトコル拡張パッケージ
│   ├── websocket/             # WebSocket
│   ├── sse/                   # Server-Sent Events
│   ├── http3/                 # HTTP/3
│   └── grpc/                  # gRPC
│
├── task/                      # タスク調度
│   ├── interface.go           # TaskQueue, CronSchedulerインターフェース
│   ├── asynq.go               # Asynq実装（Redis）
│   ├── cron.go                # Cron実装
│   └── memory.go              # メモリ実装
│
├── ipc/                       # IPCインターフェース（プロセス間通信）
│   └── interface.go           # Endpoint, Publisher, Subscriber...
│
└── gonest.go                  # 後方互換性のある型エイリアス
```

---

## インターフェース設計

### 細粒度インターフェース（core/abstract/）

フレームワークは細粒度インターフェース設計を採用しており、開発者は必要なものだけに依存できます:

```go
// リクエスト読み取りインターフェース
type RequestReaderAbstract interface {
    Method() string
    Path() string
    Header(name string) string
}

// レスポンス書き込みインターフェース
type ResponseWriterAbstract interface {
    Status(code int)
    JSON(code int, v any) error
    String(code int, s string) error
}

// 完全なコンテキストインターフェース（合成）
type ContextAbstract interface {
    ContextRunnerAbstract
    FullRequestReaderAbstract
    FullResponseWriterAbstract
    ValueStoreAbstract
}
```

### 使用パターン

```go
// 1. コア全体を使用
import "github.com/linuxerlv/gonest/core"
app := core.CreateApplication()

// 2. インターフェース定義のみを使用（拡張機能を作成）
import "github.com/linuxerlv/gonest/core/abstract"
func MyMiddleware(ctx abstract.ContextAbstract) error { ... }

// 3. 拡張ミドルウェアを使用
import "github.com/linuxerlv/gonest/middleware/auth"
app.Use(auth.New(provider, nil))

// 4. 後方互換性（gonestパッケージの型エイリアス）
import "github.com/linuxerlv/gonest"
app := gonest.NewApplication()
```

---

## APIクイックリファレンス

### アプリケーション作成

```go
import "github.com/linuxerlv/gonest/core"

// 方法1: クイック作成（シンプルなアプリケーション向け）
app := core.CreateApplication()

// 方法2: ビルダーパターン（推奨、依存性注入対応）
builder := core.CreateBuilder()
builder.Config = cfg           // 設定を設定
builder.Env.Set("KEY", "val")  // 環境変数を設定
builder.Services.AddSingleton(&MyService{})  // サービスを登録
app := builder.Build().(*core.WebApplication)

// 公開プロパティにアクセス
cfg := app.Config
env := app.Env
services := app.Services
```

### ルート登録

```go
import "github.com/linuxerlv/gonest/core/abstract"

// Application方法
app.GET("/users", func(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, users)
})

// WebApplication方法（MapXXXメソッド）
app.MapGet("/users", func(ctx abstract.ContextAbstract) error {
    return ctx.JSON(200, users)
})

app.MapPost("/users", func(ctx abstract.ContextAbstract) error {
    var user User
    ctx.Bind(&user)
    return ctx.JSON(201, user)
})

// ルートグループ
api := app.Group("/api/v1")
api.GET("/users", listUsers)
```

### ミドルウェア

```go
import (
    "github.com/linuxerlv/gonest/middleware/cors"
    "github.com/linuxerlv/gonest/middleware/recovery"
    "github.com/linuxerlv/gonest/middleware/ratelimit"
)

app.Use(recovery.New(nil))
app.Use(cors.New(&cors.Config{
    AllowOrigins: []string{"https://example.com"},
}))
app.Use(ratelimit.New(&ratelimit.Config{
    Limit:  100,
    Window: time.Minute,
}))
```

### 依存性注入

```go
// 登録
builder.Services.AddSingleton(&MyService{})
builder.Services.AddScoped(func(s abstract.ServiceCollectionAbstract) *DbContext {
    return &DbContext{}
})

// 取得
service := core.GetService[*MyService](app.Services)
```

### 環境変数

```go
// 環境変数を読み取る
dbUrl := builder.Env.Get("DATABASE_URL")
port := builder.Env.GetOrDefault("PORT", "8080")

// 存在を確認
if builder.Env.Has("DEBUG") {
    // ...
}

// すべてを取得
allEnv := builder.Env.All()
```

### 設定ファイル

```go
import "github.com/linuxerlv/gonest/config"

// JSON設定ファイルをロード
cfg := config.NewKoanfConfig(".")
cfg.Load(
    config.NewFileProvider("config.json", config.NewJSONParser()),
    config.NewJSONParser(),
)

// YAML設定ファイルをロード
cfg.Load(
    config.NewFileProvider("config.yaml", config.NewYAMLParser()),
    config.NewYAMLParser(),
)

// 設定を読み取る
port := cfg.GetString("server.port")
debug := cfg.GetBool("debug")

// 構造体にバインド
type ServerConfig struct {
    Port    string `koanf:"port"`
    Timeout int    `koanf:"timeout"`
}
var serverCfg ServerConfig
cfg.Unmarshal("server", &serverCfg)

// ビルダーに設定
builder := core.CreateBuilder()
builder.Config = cfg
```

---

## ドキュメントナビゲーション

| ドキュメント | 対象 | 内容 |
|---------|------|------|
| **[チュートリアル](TUTORIAL.md)** | 🎓 Go初心者 | 段階的な学習ガイド |
| **[APIリファレンス](API_REFERENCE.md)** | 👨‍💻 アプリケーション開発者 | 完全なAPI文書 |
| **[貢献ガイド](DEVELOPER.md)** | 🛠️ フレームワーク貢献者 | アーキテクチャ設計、コーディング標準、テスト戦略、拡張メカニズム |

---

## ライセンス

MIT License - [LICENSE](LICENSE) ファイルを参照
