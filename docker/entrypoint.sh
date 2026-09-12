#!/bin/bash
# Canvas 容器守护入口：Go API（:8080）与 Next.js（:3000）均为后台子进程。
# 任一子进程退出时终止另一个并退出容器，交给 Docker 的 restart 策略整体拉起，
# 避免「Go 进程崩溃但容器仍在运行、所有 /api/* 持续 502」的静默故障。
set -u

PORT=8080 /app/server &
API_PID=$!

cd /app/web || exit 1
PORT=3000 node server.js &
WEB_PID=$!

echo "[entrypoint] api pid=$API_PID web pid=$WEB_PID"

shutdown() {
    echo "[entrypoint] 收到停止信号，正在终止子进程..."
    kill -TERM "$API_PID" "$WEB_PID" 2>/dev/null || true
    wait || true
    exit 0
}
trap shutdown TERM INT

# 阻塞直到任一子进程退出
wait -n
STATUS=$?
echo "[entrypoint] 子进程退出（status=$STATUS），容器即将退出并由 Docker 重启"
kill -TERM "$API_PID" "$WEB_PID" 2>/dev/null || true
wait || true
exit "$STATUS"
