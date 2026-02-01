FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY backend ./backend
COPY migrations ./migrations
RUN CGO_ENABLED=0 go build -o /out/flowdb ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/flowdb /app/flowdb
COPY migrations /app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["./flowdb"]
