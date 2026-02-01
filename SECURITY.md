# SECURITY

## Chính sách disclosure

Vui lòng báo cáo lỗ hổng bảo mật qua kênh nội bộ/issue private của dự án. Không công khai chi tiết trước khi có bản vá.

## Hardening checklist

- Đổi `MASTER_KEY` và lưu trữ an toàn (KMS/secret manager).
- Đổi `ADMIN_PASS` khỏi mặc định và bật MFA.
- Bật HTTPS ở reverse proxy, bật HSTS nếu phù hợp.
- Bật `enable_ip_allowlist` và cấu hình CIDR cần thiết.
- Bật `enable_mtls` nếu có proxy hỗ trợ xác thực client cert.
- Bật `enable_signed_audit_log` để chống sửa log.
- Sao lưu PostgreSQL metadata định kỳ.
- Giới hạn network access cho PostgreSQL/MongoDB.
- Thiết lập giám sát và alert cho audit log.

## Bật chế độ enterprise

1. Gọi API `PUT /api/v1/settings/security-mode` với `mode=enterprise`.
2. Bật các cờ cần thiết qua `PUT /api/v1/settings/flags`:
   - `enable_sso_oidc`
   - `enable_mfa`
   - `enable_step_up_auth`
   - `enable_ip_allowlist`
   - `enable_signed_audit_log`
   - `enable_query_approval`
   - `enable_pii_masking`
   - `enable_scim`
3. Cập nhật cấu hình OIDC và network policy tương ứng.
