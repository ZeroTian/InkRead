# InkRead - 电纸书阅读服务

## 项目概述

InkRead 是一个轻量级的电纸书阅读服务后端，提供书架管理、电子书阅读器 API 以及 AI 总结功能。

## 技术栈

- **语言**: Go 1.21+
- **数据库**: SQLite（书籍元数据存储）
- **文件系统**: 本地存储（电子书文件）
- **Web 框架**: Gin
- **AI 集成**: OpenAI GPT API

## 项目结构

```
inkread/
├── main.go              # 程序入口
├── go.mod               # Go 模块文件
├── README.md            # 项目文档
├── api/
│   └── handlers.go      # API 处理器
├── models/
│   └── book.go          # 数据模型
├── services/
│   ├── book_service.go  # 书籍业务逻辑
│   └── ai_service.go    # AI 服务
├── storage/
│   └── sqlite.go        # SQLite 数据库操作
└── uploads/             # 电子书存储目录
```

## API 接口

### 书架管理

| 方法   | 路径            | 描述         | 请求体                  |
|--------|-----------------|--------------|------------------------|
| GET    | /api/books      | 获取书架列表   | -                      |
| POST   | /api/books      | 上传电子书     | multipart/form-data    |
| GET    | /api/books/:id  | 获取书籍内容   | -                      |
| DELETE | /api/books/:id  | 删除书籍      | -                      |

### AI 总结

| 方法   | 路径               | 描述      | 请求体                      |
|--------|--------------------|-----------|----------------------------|
| POST   | /api/ai/summarize  | AI 总结书籍 | `{"book_id": "xxx"}`       |

## 数据模型

### Book (书籍)

```json
{
  "id": "uuid",
  "title": "书名",
  "author": "作者",
  "file_path": "/uploads/xxx.epub",
  "file_size": 1024000,
  "file_type": "epub",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## 响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "error message",
  "data": null
}
```

## 环境变量

| 变量              | 描述                  | 默认值        |
|-------------------|-----------------------|---------------|
| PORT              | 服务端口              | 8080          |
| UPLOAD_DIR        | 电子书上传目录        | ./uploads     |
| DATABASE_PATH     | SQLite 数据库路径     | ./inkread.db  |
| OPENAI_API_KEY    | OpenAI API 密钥       | -             |
| OPENAI_MODEL      | OpenAI 模型          | gpt-3.5-turbo |

## 使用示例

### 1. 启动服务

```bash
go run main.go
```

### 2. 上传电子书

```bash
curl -X POST http://localhost:8080/api/books \
  -F "file=@/path/to/book.epub"
```

### 3. 获取书架列表

```bash
curl http://localhost:8080/api/books
```

### 4. 获取书籍内容

```bash
curl http://localhost:8080/api/books/{id}
```

### 5. 删除书籍

```bash
curl -X DELETE http://localhost:8080/api/books/{id}
```

### 6. AI 总结

```bash
curl -X POST http://localhost:8080/api/ai/summarize \
  -H "Content-Type: application/json" \
  -d '{"book_id": "xxx"}'
```

## 开发计划

- [x] 项目结构设计
- [x] SQLite 数据库集成
- [x] 书架 CRUD API
- [x] 电子书文件存储
- [x] AI 总结接口
- [x] EPUB 解析
- [x] 阅读进度管理
- [x] 单元测试
