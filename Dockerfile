# 第一阶段：编译环境 (起名叫 builder)
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app

# 配置国内 GOPROXY，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# 先拷贝 go.mod 和 go.sum，下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝项目所有代码并编译
COPY . .
# CGO_ENABLED=0 确保编译出静态链接的纯净二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp main.go

# 第二阶段：运行环境 (极其轻量的 alpine 镜像)
FROM alpine:latest

WORKDIR /app

# 从 builder 阶段把编译好的可执行文件捞过来
COPY --from=builder /app/myapp .

# 暴露 8080 端口
EXPOSE 8080

# 启动程序
CMD ["./myapp"]