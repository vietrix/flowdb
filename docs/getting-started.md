# Getting Started

## Yêu cầu

- Docker và Docker Compose.

## Chạy nhanh với 2 lệnh Docker (frontend + backend)

Backend:

```bash
docker run -d --name flowdb-backend \
  -p 8080:8080 \
  -e BIND_ADDR=0.0.0.0:8080 \
  -e DATABASE_URL="postgres://<user>:<pass>@<host>:5432/flowdb?sslmode=disable" \
  -e MASTER_KEY="<BASE64_32_BYTES>" \
  -e ADMIN_USER="<admin_user>" \
  -e ADMIN_PASS="<strong_password>" \
  -e CORS_ALLOW_ORIGINS="http://localhost:3000" \
  ghcr.io/vietrix/flowdb-backend:latest
```

Frontend:

```bash
docker run -d --name flowdb-frontend \
  -p 3000:3000 \
  -e FLOWDB_API_BASE="http://localhost:8080" \
  ghcr.io/vietrix/flowdb-frontend:latest
```

Ghi chú:

- Nếu dùng domain riêng, đổi các URL thành `https://app.example.com` và `https://api.example.com`.
- Nếu reverse proxy dùng cùng domain, bạn có thể bỏ `FLOWDB_API_BASE` để frontend gọi relative path.
- Xem cấu hình reverse proxy trong `docs/deployment.md`.

## Cấu hình `.env`

Tạo file `.env` ở root repo để dùng với `docker compose`:

```bash
DATABASE_URL=postgres://flowdb:flowdb@meta-postgres:5432/flowdb?sslmode=disable
MASTER_KEY=<BASE64_32_BYTES>
ADMIN_USER=admin
ADMIN_PASS=<strong_password>
# Optional
FLOWDB_API_BASE=http://localhost:8080
CORS_ALLOW_ORIGINS=http://localhost:3000
```

Gợi ý tạo `MASTER_KEY`:

```bash
openssl rand -base64 32
```

Nếu cần dùng image khác, có thể override:

```bash
FLOWDB_BACKEND_IMAGE=ghcr.io/vietrix/flowdb-backend:latest
FLOWDB_FRONTEND_IMAGE=ghcr.io/vietrix/flowdb-frontend:latest
```

## Chạy bằng Docker Compose

```bash
docker compose up -d
```

Sau khi chạy:

- API: `http://127.0.0.1:8080`
- UI: `http://127.0.0.1:3000`

## Chạy nhanh backend (không dùng compose)

```bash
docker run -d --name flowdb \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://<user>:<pass>@<host>:5432/flowdb?sslmode=disable" \
  -e MASTER_KEY="<BASE64_32_BYTES>" \
  -e ADMIN_USER="<admin_user>" \
  -e ADMIN_PASS="<strong_password>" \
  ghcr.io/vietrix/flowdb-backend:latest
```

## SSH port-forward (tùy chọn)

```bash
ssh -L 8080:127.0.0.1:8080 user@server
```
