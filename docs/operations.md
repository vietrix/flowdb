# Operations

## Auto-update qua GitHub Release

### API

- `GET /api/v1/system/version`: trả phiên bản đang chạy.
- `GET /api/v1/system/update`: kiểm tra bản mới (admin).
- `POST /api/v1/system/update/apply`: tải và cài bản mới (admin).

### Biến môi trường

- `UPDATE_REPO`: repo GitHub (mặc định `vietrix/flowdb`).
- `UPDATE_CHECK_INTERVAL`: chu kỳ cache kiểm tra (mặc định `5m`).
- `UPDATE_GITHUB_TOKEN`: token để tránh rate limit (khuyến nghị).
- `UPDATE_AUTO_RESTART`: `true` để tự restart sau khi cập nhật.

### Ghi chú

- Update dùng asset từ GitHub Release theo tên: `flowdb_<os>_<arch>.(tar.gz|zip)` và `.sha256`.
- Khi `UPDATE_AUTO_RESTART=true`, server sẽ tự thoát sau khi cập nhật để tiến trình/compose khởi động lại.
