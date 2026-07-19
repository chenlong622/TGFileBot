# 运行阶段
FROM alpine:3.20

ARG TARGETARCH

WORKDIR /app

# su-exec 用于在 entrypoint 中从 root 降权到非特权用户后再启动主进程
RUN apk --no-cache add ca-certificates tzdata su-exec

# 复制之前在 Github Actions 中已经编译好的对应架构的可执行文件
COPY TGFileBot-linux-${TARGETARCH} ./TGBot

# 确保可执行权限
RUN chmod +x ./TGBot

# 确保配置文件和目录存在
RUN mkdir -p files

# 创建非特权用户，避免主进程以 root 身份运行
RUN addgroup -S tgbot && adduser -S tgbot -G tgbot

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENV TZ=Asia/Shanghai
ENV LOG=""

# entrypoint.sh 会先修正挂载卷 ./files 的属主，再以非 root 用户 tgbot 启动主进程
ENTRYPOINT ["/entrypoint.sh"]

