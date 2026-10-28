# 使用指南

## 快速开始

### 1. 准备环境

确保以下服务已启动并运行在 `192.168.40.154`:

- **Ollama**: 端口 11434
  ```bash
  # 确保已加载模型
  ollama pull qwen2.5:7b
  ollama pull bge-m3
  ```

- **Qdrant**: 端口 6333
  ```bash
  # 使用 Docker 运行 Qdrant
  docker run -p 6333:6333 qdrant/qdrant
  ```

### 2. 安装依赖

```bash
# 安装 PDF 解析工具 (macOS)
brew install poppler

# 或 Ubuntu/Debian
sudo apt-get install poppler-utils
```

### 3. 运行服务

```bash
# 编译
go build -o rag-server

# 运行
./rag-server -port 8080
```

### 4. 使用 API

#### 上传文档

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@your_document.pdf"
```

#### 查询问答

```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"question": "文档中提到了什么重要内容？"}'
```

## Python 客户端示例

```python
import requests

# 上传文档
with open('document.pdf', 'rb') as f:
    files = {'file': f}
    response = requests.post('http://localhost:8080/api/upload', files=files)
    print(response.json())

# 查询
response = requests.post(
    'http://localhost:8080/api/query',
    json={'question': '文档的主要内容是什么？'}
)
result = response.json()
print(f"答案: {result['answer']}")
print(f"来源: {result['sources']}")
```

## JavaScript 客户端示例

```javascript
// 上传文档
const formData = new FormData();
formData.append('file', fileInput.files[0]);

fetch('http://localhost:8080/api/upload', {
  method: 'POST',
  body: formData
})
.then(res => res.json())
.then(data => console.log(data));

// 查询
fetch('http://localhost:8080/api/query', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({question: '文档的主要内容是什么？'})
})
.then(res => res.json())
.then(data => {
  console.log('答案:', data.answer);
  console.log('来源:', data.sources);
});
```

## 常见问题

### Q: PDF 解析失败？

A: 确保已安装 `pdftotext` 工具：
```bash
which pdftotext  # 检查是否安装
```

### Q: 连接 Ollama 失败？

A: 检查 Ollama 服务是否运行：
```bash
curl http://192.168.40.154:11434/api/tags
```

### Q: Qdrant 连接失败？

A: 检查 Qdrant 服务是否运行：
```bash
curl http://192.168.40.154:6333/collections
```

### Q: 如何修改服务地址？

A: 使用环境变量：
```bash
export OLLAMA_URL=http://your-ollama-host:11434
export QDRANT_URL=http://your-qdrant-host:6333
./rag-server
```
