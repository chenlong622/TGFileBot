#!/bin/sh
set -e

# 兼容历史以 root 身份运行时在挂载卷中创建的文件：
# 启动前先以 root 权限修正属主，再降权到非特权用户 tgbot 运行主进程。
chown -R tgbot:tgbot ./files 2>/dev/null || true

# ${LOG:+-log "$LOG"} 会在 LOG 变量非空时自动展开为 -log 参数
exec su-exec tgbot ./TGBot -files ./files ${LOG:+-log "$LOG"} "$@"
