# Deployment

## Reverse proxy và TLS

- Đặt reverse proxy (Nginx/Traefik) phía trước để terminate TLS.
- Nếu bật mTLS, cấu hình proxy thêm header `X-Client-Cert-Verified: true` khi client cert hợp lệ.
- Đảm bảo forward `X-Request-ID` nếu bạn muốn trace thống nhất.

## Ví dụ Nginx (một domain)

```nginx
server {
  server_name app.example.com;

  location / {
    proxy_pass http://127.0.0.1:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  }

  location /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
  }
}
```

Ghi chú:

- Nếu API chạy ở domain riêng (ví dụ `https://api.example.com`), hãy đặt `FLOWDB_API_BASE=https://api.example.com` cho frontend.
- Khi dùng domain riêng, cấu hình `CORS_ALLOW_ORIGINS=https://app.example.com` ở backend.

## Lưu ý môi trường production

- Bảo vệ `MASTER_KEY` và các secret bằng KMS/secret manager.
- Giới hạn network access đến PostgreSQL và MongoDB.
- Thiết lập giám sát và cảnh báo cho audit log.
