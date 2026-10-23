# 安装指南

## 快速安装

### 1. 安装 PDF 解析工具

#### macOS
```bash
brew install poppler
```

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install poppler-utils
```

#### 验证安装
```bash
which pdftotext
pdftotext -v
```

### 2. 安装 Go 依赖

```bash
go mod tidy
```

### 3. 确保服务运行

#### Ollama 服务
确保 Ollama 在 `192.168.40.154:11434` 运行，并已加载模型：

```bash
# 检查 Ollama 是否运行
curl http://192.168.40.154:11434/api/tags

# 如果未加载模型，在 Ollama 服务器上运行：
ollama pull qwen2.5:7b
ollama pull bge-m3
```

#### Qdrant 服务
确保 Qdrant 在 `192.168.40.154:6333` 运行：

```bash
# 检查 Qdrant 是否运行
curl http://192.168.40.154:6333/collections
```

### 4. 编译和运行

```bash
# 编译
go build -o rag-server

# 运行
./rag-server -port 8080
```

## 故障排除

### PDF 解析失败

**错误**: `pdftotext: executable file not found in $PATH`

**解决方案**:
1. 安装 poppler (见上方)
2. 确保 `pdftotext` 在 PATH 中：
   ```bash
   echo $PATH
   which pdftotext
   ```

### Ollama 连接失败

**错误**: `Failed to connect to Ollama`

**解决方案**:
1. 检查 Ollama 服务是否运行：
   ```bash
   curl http://192.168.40.154:11434/api/tags
   ```
2. 检查防火墙设置
3. 确认模型已加载：
   ```bash
   ollama list
   ```

### Qdrant 连接失败

**错误**: `Failed to connect to Qdrant`

**解决方案**:
1. 检查 Qdrant 服务是否运行：
   ```bash
   curl http://192.168.40.154:6333/collections
   ```
2. 检查网络连接
3. 如果使用 Docker：
   ```bash
   docker ps | grep qdrant
   ```

### JSON 解析错误

**错误**: `jq: parse error` 或 `failed to parse response`

**解决方案**:
1. 检查 API 响应格式
2. 查看服务器日志
3. 确认 Ollama API 返回格式正确

## 环境变量配置

如果需要修改服务地址，可以设置环境变量：

```bash
export OLLAMA_URL=http://your-ollama-host:11434
export EMBEDDING_URL=http://your-ollama-host:11434/api/embeddings
export QDRANT_URL=http://your-qdrant-host:6333
export QDRANT_API_KEY=your-api-key  # 可选
export COLLECTION_NAME=rag_documents  # 可选

./rag-server
```

## 测试安装

运行测试脚本：

```bash
chmod +x example_client.sh
./example_client.sh
```

如果所有步骤都成功，你应该看到：
1. 健康检查返回 `{"status": "ok"}`
2. 文档上传成功
3. 查询返回答案和来源
