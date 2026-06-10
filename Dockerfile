FROM golang:1.21-alpine AS builder

WORKDIR /app

# 注入国内镜像劫线代理，阻击 proxy.golang.org 的 TLS EOF 脱网现象
ENV GOPROXY=https://goproxy.cn,direct

# [第一层缓存] 仅引入依赖声明。一旦 go.mod/go.sum 未改变，Docker 将永久缓存这一层的完整容器状态
COPY go.mod go.sum ./
# 使用 BuildKit 挂载加速，即便第一层被破坏，也能瞬间从宿主机的隐藏抽屉中读出模块
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# [第二层缓存] 业务代码植入（此层改动频繁，不再会击碎上方依赖层）
COPY . .

# 全量跨环境解绑静态微缩编译，挂载缓存以提速链接环节
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -ldflags="-w -s" -o mdshare .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
# 只将二进制可调用载体进行迁移，达到最简化防攻击微型系统
COPY --from=builder /app/mdshare .

# 将在 yaml 里被主导使用的访问接通孔径声明出来
EXPOSE 18080

CMD ["./mdshare"]
