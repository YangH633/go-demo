package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/events", handleSSE)
	log.Println("打开 http://localhost:8080 查看 SSE demo")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	// 1. ResponseWriter 本身没有 Flush 方法，先检查实际实现是否支持。
	// 某些中间件包装后的 ResponseWriter 可能不支持 http.Flusher。
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 2. 声明 SSE 事件流，禁止缓存。同源 demo 无需跨域头，
	// 也无需手动设置 Connection: keep-alive。
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	// 若经过 Nginx，此头可关闭代理缓冲；其他代理需配置缓冲和长连接超时。
	w.Header().Set("X-Accel-Buffering", "no")

	// 3. 每条事件以 data: 开头、空行（\n\n）结尾；内容可以直接使用文本。
	fmt.Fprint(w, "data: 连接成功\n\n")
	// 4. 立即发送缓冲数据，不能等处理函数结束才发送。
	flusher.Flush()

	// 定时器只是模拟业务数据，不是 SSE 要求，无需额外 goroutine/channel。
	// 每 3 秒都有数据，无需额外心跳；长期无数据时可写入 ": ping\n\n" 并 Flush。
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// 5. 保持请求运行，持续推送；普通 HTTP 接口一般返回一次就结束。
	for {
		select {
		case now := <-ticker.C:
			if _, err := fmt.Fprintf(w, "data: 当前时间 %s\n\n", now.Format("15:04:05")); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			// 6. 客户端断开后退出，defer 停止定时器并释放资源。
			return
		}
	}
}
