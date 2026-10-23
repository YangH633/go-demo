# SSE Demo

最小 Go + 原生 JavaScript 示例：连接后立即收到消息，之后每 3 秒推送服务器时间。
SSE 的实现要点已直接注释在 `main.go` 和 `static/index.html` 中。

## 运行

```bash
cd sse-demo
go run .
```

打开 http://localhost:8080，点击“连接”查看消息，点击“断开”停止接收。
也可以用 `curl -N http://localhost:8080/events` 查看事件流。

## 关键点

- 后端设置事件流响应头，使用 `data: ...\n\n` 格式，写入后 `Flush()`。
- 保持请求运行，客户端断开后释放定时器。
- 前端使用 `EventSource` 接收，断线自动重连，主动关闭时调用 `close()`。
- SSE 是单向推送，客户端提交数据需另外发送普通 HTTP 请求。
- 经过代理时需关闭响应缓冲，并配置适合长连接的超时。

本 demo 不包含认证、广播和断线消息补发。定时推送已保持连接活跃，因此没有额外心跳。
