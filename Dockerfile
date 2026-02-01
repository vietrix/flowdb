FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY backend ./backend
COPY migrations ./migrations
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 go build -ldflags "-X flowdb/backend/util.Version=${VERSION} -X flowdb/backend/util.CommitSHA=${COMMIT} -X flowdb/backend/util.BuildTime=${BUILD_TIME}" -o /out/flowdb ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/flowdb /app/flowdb
COPY migrations /app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["./flowdb"]
