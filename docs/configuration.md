# Configuration

Tài liệu này liệt kê các biến môi trường phổ biến. Tùy theo nhu cầu, bạn có thể cấu hình thêm trong runtime hoặc qua reverse proxy.

## Biến môi trường cốt lõi

- `DATABASE_URL`: chuỗi kết nối PostgreSQL metadata.
- `MASTER_KEY`: khóa AES-GCM dạng base64 (32 bytes).
- `BIND_ADDR`: địa chỉ bind (mặc định `127.0.0.1:8080`).
- `ADMIN_USER`, `ADMIN_PASS`: tài khoản admin khởi tạo.
- `MONGO_URI`: URI MongoDB mặc định.
- `AUTO_MIGRATE`: `true` để tự chạy migration khi khởi động.
- `CORS_ALLOW_ORIGINS`: danh sách origin, phân tách bằng dấu phẩy.

## Frontend

- `FLOWDB_API_BASE`: base URL API dùng ở runtime cho frontend (ví dụ `https://api.example.com` hoặc `http://localhost:8080`). Nếu để trống, frontend sẽ gọi relative path theo domain hiện tại.
- `NEXT_PUBLIC_API_BASE`: tùy chọn cho build-time (phù hợp khi tự build frontend).

## OIDC / SSO

- `OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL`: cấu hình OIDC.
- `OIDC_SCOPES`: scopes OIDC (mặc định `openid,profile,email,groups`).
- `OIDC_GROUP_CLAIM`: claim chứa nhóm (mặc định `groups`).
- `OIDC_ADMIN_GROUP`: tên nhóm sẽ map thành admin.
- `OIDC_ROLE_MAP`: JSON map nhóm -> role (ví dụ `{"db-admin":"admin"}`).

## Session và rate limit

- `SESSION_TTL`: thời hạn session (ví dụ `24h`).
- `LOGIN_RATE_LIMIT_PER_MIN`, `LOGIN_RATE_LIMIT_BURST`: rate limit đăng nhập.
- `GLOBAL_MAX_ROWS`: giới hạn số dòng mặc định.
- `STATEMENT_TIMEOUT`: timeout mặc định cho query.

## Auto-update

- `UPDATE_REPO`: repo GitHub để kiểm tra update.
- `UPDATE_CHECK_INTERVAL`: chu kỳ kiểm tra bản mới.
- `UPDATE_GITHUB_TOKEN`: token GitHub để tránh rate limit.
- `UPDATE_AUTO_RESTART`: tự restart sau khi cập nhật.
