# FlowDB Backend

## Cách chạy (chỉ cần Docker)

### 1) Chạy nhanh (một lệnh)

```bash
docker run -d --name flowdb \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://<user>:<pass>@<host>:5432/flowdb?sslmode=disable" \
  -e MASTER_KEY="<BASE64_32_BYTES>" \
  -e ADMIN_USER="<admin_user>" \
  -e ADMIN_PASS="<strong_password>" \
  ghcr.io/vietrix/flowdb-backend:latest
```

### 2) Dùng Docker Compose (frontend + backend)

```bash
docker compose up -d
```

Mặc định API lắng nghe tại `http://127.0.0.1:8080`, UI tại `http://127.0.0.1:3000`.

Lưu ý: bạn bắt buộc phải set `DATABASE_URL`, `MASTER_KEY`, `ADMIN_USER`, `ADMIN_PASS` bằng biến môi trường (ví dụ file `.env`) trước khi chạy compose.

## Biến môi trường chính

- `DATABASE_URL`: kết nối PostgreSQL metadata.
- `MASTER_KEY`: khóa AES-GCM dạng base64 (32 bytes).
- `BIND_ADDR`: địa chỉ bind (mặc định `127.0.0.1:8080`).
- `ADMIN_USER`, `ADMIN_PASS`: tài khoản admin khởi tạo.
- `MONGO_URI`: URI MongoDB mặc định.
- `AUTO_MIGRATE`: `true` để tự chạy migration khi khởi động.
- `CORS_ALLOW_ORIGINS`: danh sách origin, phân tách bằng dấu phẩy.
- `OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL`: cấu hình OIDC.
- `OIDC_SCOPES`: scopes OIDC (mặc định `openid,profile,email,groups`).
- `OIDC_GROUP_CLAIM`: claim chứa nhóm (mặc định `groups`).
- `OIDC_ADMIN_GROUP`: tên nhóm sẽ map thành admin.
- `OIDC_ROLE_MAP`: JSON map nhóm->role (ví dụ `{"db-admin":"admin"}`).
- `SESSION_TTL`: thời hạn session (ví dụ `24h`).
- `LOGIN_RATE_LIMIT_PER_MIN`, `LOGIN_RATE_LIMIT_BURST`: rate limit đăng nhập.
- `GLOBAL_MAX_ROWS`: giới hạn số dòng mặc định.
- `STATEMENT_TIMEOUT`: timeout mặc định cho query.
- `UPDATE_REPO`: repo GitHub để kiểm tra update.
- `UPDATE_CHECK_INTERVAL`: chu kỳ kiểm tra bản mới.
- `UPDATE_GITHUB_TOKEN`: token GitHub để tránh rate limit.
- `UPDATE_AUTO_RESTART`: tự restart sau khi cập nhật.

## Ví dụ SSH port-forward

```bash
ssh -L 8080:127.0.0.1:8080 user@server
```

## Reverse proxy

- Đặt proxy (Nginx/Traefik) phía trước để terminate TLS.
- Nếu bật mTLS, cấu hình proxy thêm header `X-Client-Cert-Verified: true` khi client cert hợp lệ.
- Đảm bảo forward `X-Request-ID` nếu bạn muốn trace thống nhất.

## CI/CD và phát hành

Workflow:
- `.github/workflows/ci.yml`: chạy `go test ./...` và `pnpm lint` cho frontend.
- `.github/workflows/release.yml`: khi push tag `v*` sẽ build binary đa nền tảng và tạo GitHub Release kèm file `.sha256`.
- `.github/workflows/docker.yml`: build & push image `ghcr.io/vietrix/flowdb-backend` và `ghcr.io/vietrix/flowdb-frontend` khi push `main` hoặc tag `v*`.

## Tạo repo bằng GitHub CLI và push

```bash
gh auth login
gh repo create vietrix/flowdb --public --source . --remote origin --push
git branch -M main
git push -u origin main
```

Tạo release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Hệ thống auto-update qua GitHub Release

API:
- `GET /api/v1/system/version`: trả phiên bản đang chạy.
- `GET /api/v1/system/update`: kiểm tra bản mới (admin).
- `POST /api/v1/system/update/apply`: tải và cài bản mới (admin).

Biến môi trường:
- `UPDATE_REPO`: repo GitHub (mặc định `vietrix/flowdb`).
- `UPDATE_CHECK_INTERVAL`: chu kỳ cache kiểm tra (mặc định `5m`).
- `UPDATE_GITHUB_TOKEN`: token để tránh rate limit (khuyến nghị).
- `UPDATE_AUTO_RESTART`: `true` để tự restart sau khi cập nhật.

Ghi chú:
- Update sử dụng asset từ GitHub Release theo tên: `flowdb_<os>_<arch>.(tar.gz|zip)` và `.sha256`.
- Khi `UPDATE_AUTO_RESTART=true`, server sẽ tự thoát sau khi cập nhật để tiến trình/compose khởi động lại.
