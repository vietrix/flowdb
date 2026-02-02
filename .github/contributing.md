# Contributing

Cảm ơn bạn đã quan tâm đến FlowDB. Tài liệu này mô tả cách đóng góp hiệu quả và đúng chuẩn cho repo.

## Quy trình chung

1. Fork repo và tạo nhánh mới từ `main`.
2. Thực hiện thay đổi, cập nhật tài liệu liên quan nếu cần.
3. Chạy kiểm tra tối thiểu trước khi gửi PR:
   - Backend: `go test ./...`
   - Frontend:
     - `pnpm install --frozen-lockfile`
     - `pnpm lint`
4. Tạo pull request, mô tả rõ mục tiêu và phạm vi thay đổi.

## Yêu cầu môi trường

- Go 1.22 (theo CI).
- Node.js 20 và pnpm 10.x (theo CI).
- Docker (khuyến nghị để chạy nhanh toàn bộ stack).

## Quy ước

- Giữ thay đổi nhỏ, dễ review.
- Tránh thay đổi không liên quan trong cùng PR.
- Nếu thêm tính năng mới, hãy cập nhật `docs/` và/hoặc README.

## Báo lỗi / đề xuất tính năng

Hãy dùng các issue templates trong `.github/ISSUE_TEMPLATE/`.
