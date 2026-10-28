package main

import (
	"log"
	"log/slog"
	"os"
)

// logger: 日志记录器
// NewJSONHandler：日志处理器，决定日志的格式和处理方式

// 标准库 log 不支持Handler参数
func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
	logger.Println("hello world")
	log.Println("hello world")

	slog.Info("hello, world")
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // 最低日志级别

		AddSource: true, // 显示源码位置
	})
	slogger := slog.New(handler)
	slogger.Debug("服务启动", "port", 8080)
}
