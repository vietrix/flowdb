# COMPLIANCE

Lưu ý: Chứng nhận tuân thủ phụ thuộc vào cách triển khai và vận hành. Dự án này cung cấp các kiểm soát kỹ thuật để hỗ trợ tuân thủ.

## SOC 2

| Control | Tính năng/Thiết lập | Trách nhiệm vận hành |
|---|---|---|
| CC6.1 | RBAC/ABAC, session, CSRF | Quản trị quyền, rà soát định kỳ |
| CC6.6 | MFA, step-up auth | Bắt buộc bật MFA cho admin |
| CC7.2 | Audit log, signed audit | Lưu trữ log, giám sát |
| CC7.3 | IP allowlist, mTLS | Cấu hình network policy |
| CC8.1 | Backup metadata DB | Thiết lập backup & restore |

## ISO 27001

| Control | Tính năng/Thiết lập | Trách nhiệm vận hành |
|---|---|---|
| A.5.15 | Chính sách truy cập | Quản trị role/policy |
| A.5.17 | Thông tin xác thực | MFA, OIDC/SAML |
| A.5.23 | Logging & monitoring | Audit log, SIEM export |
| A.8.9 | Bảo vệ dữ liệu | AES-GCM cho secrets |
| A.8.12 | Quản lý cấu hình | Settings + flags |

## NIST CSF

| Function | Tính năng/Thiết lập | Trách nhiệm vận hành |
|---|---|---|
| Identify | Inventory connections | Quản trị kết nối |
| Protect | MFA, RBAC/ABAC, mTLS | Cấu hình và review |
| Detect | Audit log, signed audit | Giám sát & cảnh báo |
| Respond | Query approval | Quy trình phê duyệt |
| Recover | Backup metadata DB | Kế hoạch khôi phục |

## PCI DSS

| Requirement | Tính năng/Thiết lập | Trách nhiệm vận hành |
|---|---|---|
| 7 | RBAC/ABAC, least privilege | Định kỳ review quyền |
| 8 | MFA, SSO | Bắt buộc MFA/SSO |
| 10 | Audit log, SIEM export | Lưu trữ log an toàn |
| 3 | AES-GCM secrets | Quản lý khóa |
