# FlowDB

FlowDB là dự án gồm backend Go và frontend web, tập trung vào việc vận hành FlowDB theo mô hình self-hosted. Repo này cung cấp mã nguồn, Docker image và cấu hình CI/CD để chạy nhanh trong môi trường phát triển hoặc triển khai nội bộ.

## Bắt đầu nhanh với Docker

### Chạy nhanh với 2 lệnh Docker (frontend + backend)

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

- Nếu dùng domain riêng, đổi URL thành `https://app.example.com` và `https://api.example.com`.
- Nếu reverse proxy dùng cùng domain, bạn có thể bỏ `FLOWDB_API_BASE` để frontend gọi relative path.
- Xem cấu hình reverse proxy trong `docs/deployment.md`.
- CI/CD tự động build và đẩy image lên GHCR khi push `main` hoặc tag `v*` (xem `.github/workflows/docker.yml`).

### Chạy nhanh backend (1 lệnh)

```bash
docker run -d --name flowdb \
  -p 8080:8080 \
  -e BIND_ADDR=0.0.0.0:8080 \
  -e DATABASE_URL="postgres://<user>:<pass>@<host>:5432/flowdb?sslmode=disable" \
  -e MASTER_KEY="<BASE64_32_BYTES>" \
  -e ADMIN_USER="<admin_user>" \
  -e ADMIN_PASS="<strong_password>" \
  ghcr.io/vietrix/flowdb-backend:latest
```

### Docker Compose (frontend + backend)

```bash
docker compose up -d
```

Mặc định API lắng nghe tại `http://127.0.0.1:8080`, UI tại `http://127.0.0.1:3000`.

Lưu ý: bạn cần thiết lập `DATABASE_URL`, `MASTER_KEY`, `ADMIN_USER`, `ADMIN_PASS` (ví dụ qua file `.env`) trước khi chạy compose. Xem chi tiết trong `docs/getting-started.md`.

## Cấu hình

Danh sách biến môi trường và hướng dẫn cấu hình nằm trong `docs/configuration.md`.

## Tài liệu

- `docs/README.md`: mục lục tài liệu
- `docs/getting-started.md`: chạy local với Docker
- `docs/configuration.md`: biến môi trường và cấu hình
- `docs/deployment.md`: reverse proxy và triển khai
- `docs/operations.md`: auto-update và vận hành

## Đóng góp

Xem hướng dẫn trong `.github/contributing.md`.

## Bảo mật

Chính sách báo cáo lỗ hổng nằm ở `.github/SECURITY.md`.

## Giấy phép

Xem `LICENSE`.
