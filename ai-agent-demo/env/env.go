package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// LoadEnv 使用 Viper 从当前工作目录读取 .env 文件，并同步到 os.Environ，
// 这样现有代码里的 os.Getenv 无需修改。若 .env 不存在则忽略。
func LoadEnv(opts ...LoadEnvOption) error {
	o := &loadEnvOpts{override: true}
	for _, fn := range opts {
		fn(o)
	}

	startDir, err := os.Getwd()
	if err != nil {
		return err
	}

	envPath, ok := findEnvPath(startDir)
	if !ok {
		return nil
	}

	v := viper.New()
	v.SetConfigFile(envPath)
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取 .env: %w", err)
	}

	for _, key := range v.AllKeys() {
		envKey := strings.ToUpper(key)
		if !o.override && os.Getenv(envKey) != "" {
			continue
		}
		os.Setenv(envKey, v.GetString(key))
	}
	return nil
}

func findEnvPath(startDir string) (string, bool) {
	dir := startDir
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

type loadEnvOpts struct {
	override bool
}

// LoadEnvOption 用于配置 LoadEnv 行为。
type LoadEnvOption func(*loadEnvOpts)

// LoadEnvNoOverride 表示不覆盖已有环境变量（.env 仅填充未设置的变量）。
func LoadEnvNoOverride() LoadEnvOption {
	return func(o *loadEnvOpts) { o.override = false }
}
