# RAG 检索增强生成系统

基于 Go 语言实现的 RAG (Retrieval-Augmented Generation) 系统，支持中文文档处理和 PDF 解析。

## 功能特性

- ✅ 支持 PDF 文档解析（使用 pdftotext）
- ✅ 支持中文文档处理
- ✅ 使用 BGE-M3 模型生成向量嵌入
- ✅ 使用 Qdrant 作为向量数据库
- ✅ 使用 Ollama Qwen2.5 7B 模型进行文本生成
- ✅ RESTful API 接口

## 系统架构

### 技术栈

- **语言模型**: Ollama Qwen2.5 7B (192.168.40.154:11434)
- **嵌入模型**: BGE-M3 (192.168.40.154:11434)
- **向量数据库**: Qdrant (192.168.40.154:6333)
- **编程语言**: Go 1.22+

## 安装和运行

### 前置要求

1. 安装 Go 1.22 或更高版本
2. 安装 pdftotext 工具（用于 PDF 解析）:
   ```bash
   # 自动安装（推荐）
   ./install_dependencies.sh
   
   # 或手动安装
   # macOS
   brew install poppler
   
   # Ubuntu/Debian
   sudo apt-get install poppler-utils
   ```
3. 确保 Ollama 和 Qdrant 服务在 192.168.40.154 上运行

### 运行服务

```bash
# 编译
go build -o rag-server

# 运行
./rag-server -port 8080
```

或者直接运行：

```bash
go run main.go -port 8080
```

## API 接口

### 1. 健康检查

```bash
GET /api/health
```

响应：
```json
{
  "status": "ok"
}
```

### 2. 上传文档

```bash
POST /api/upload
Content-Type: multipart/form-data

file: <文件内容>
```

响应：
```json
{
  "success": true,
  "message": "Document uploaded and processed successfully"
}
```

### 3. 查询问答

```bash
POST /api/query
Content-Type: application/json

{
  "question": "你的问题"
}
```

响应：
```json
{
  "answer": "生成的答案",
  "sources": [
    {
      "text": "相关文档片段",
      "filename": "文档名.pdf",
      "score": 0.85
    }
  ]
}
```

## 使用示例

### 使用 curl 上传文档

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@document.pdf"
```

### 使用 curl 查询

```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"question": "文档中提到了什么？"}'
```

## 配置

可以通过环境变量配置系统参数：

- `OLLAMA_URL`: Ollama 服务地址（默认: http://192.168.40.154:11434）
- `EMBEDDING_URL`: 嵌入模型 API 地址（默认: http://192.168.40.154:11434/api/embeddings）
- `QDRANT_URL`: Qdrant 服务地址（默认: http://192.168.40.154:6333）
- `QDRANT_API_KEY`: Qdrant API 密钥（可选）
- `COLLECTION_NAME`: Qdrant 集合名称（默认: rag_documents）

## 项目结构

```
rag-test/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块定义
├── internal/
│   ├── api/               # API 处理器
│   │   └── handler.go
│   ├── config/            # 配置管理
│   │   └── config.go
│   ├── parser/            # 文档解析器
│   │   └── pdf.go
│   ├── rag/               # RAG 核心逻辑
│   │   └── rag.go
│   └── storage/           # 存储服务
│       └── storage.go
└── README.md
```

## 工作流程

详见 [流程图文档](FLOWCHART.md)

## 注意事项

1. PDF 解析依赖系统安装的 `pdftotext` 工具
2. 确保 Ollama 服务已加载 qwen2.5:7b 和 bge-m3 模型
3. Qdrant 集合会在首次上传文档时自动创建
4. 文档会被分割成 500 字符的块，重叠 50 字符

## 许可证

MIT
