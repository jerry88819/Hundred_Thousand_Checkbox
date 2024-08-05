
# Hundred Thousand Checkbox

此專案是由 https://onemillioncheckboxes.com/ 發想的，使用 WebSocket 來管理一個復選框網格，每個復選框的狀態儲存在 Redis 中，應用程式還即時顯示連線的使用者數量。

## 專案結構

```
myproject/
├── docker-compose.yml
├── Dockerfile
├── main.go
├── index.html
├── go.mod
├── go.sum
└── redis/
```

## 先決條件

確保您的機器上已安裝 Docker 和 Docker Compose。

## 快速入門

### 1. Clone 下專案後，使用 Docker Compose 構建並運行專案

```sh
docker-compose up --build
```

此命令將構建 Docker 映像並啟動 Go 應用程式和 Redis 的容器。

### 2. 訪問應用程式

打開瀏覽器並導航到 `http://localhost:8080`。您應該會看到復選框網格和顯示在頂部的連線使用者數量。

## 專案詳情

### 後端

後端使用 Go 編寫，執行以下任務：
- 建立與客戶端的 WebSocket 連線。
- 使用bitmap操作將每個復選框的狀態儲存在 Redis 中。
- 週期性地向所有客戶端廣播連線的使用者數量。

### 前端

前端是一個簡單的 HTML 頁面，嵌入了 JavaScript 來：
- 連接到後端的 WebSocket 伺服器。
- 發送和接收消息來更新復選框狀態並顯示連線使用者數量。

### Docker Compose

`docker-compose.yml` 文件定義了兩個服務：
- `app`: Go 應用程式服務。
- `redis`: Redis 服務。

### 環境變數

您可以通過在 `docker-compose.yml` 文件中定義的環境變數來配置 Redis 連線和其他設置。

### API 端點

WebSocket 伺服器提供以下消息類型：
- `full_state`: 從伺服器發送給客戶端，包含所有復選框的完整狀態。
- `toggle`: 從客戶端發送給伺服器，用於切換特定復選框的狀態。
- `user_count`: 從伺服器發送給所有客戶端，包含當前連線的使用者總數。
