# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway .

# 最终镜像
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/api-gateway .
EXPOSE 3378
CMD ["./api-gateway", "3378"]