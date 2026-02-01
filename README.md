# FlowDB Backend

## Cách chạy

```bash
docker compose up --build
```

Mặc định API lắng nghe tại `http://127.0.0.1:8080`.

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

## Ví dụ SSH port-forward

```bash
ssh -L 8080:127.0.0.1:8080 user@server
```

## Reverse proxy

- Đặt proxy (Nginx/Traefik) phía trước để terminate TLS.
- Nếu bật mTLS, cấu hình proxy thêm header `X-Client-Cert-Verified: true` khi client cert hợp lệ.
- Đảm bảo forward `X-Request-ID` nếu bạn muốn trace thống nhất.
