# document-version-compare-service

文档版本比对服务，提供文本版本保存、段落差异、批注、回滚、审计和简单 Web 界面。服务默认将运行数据持久化到 `data/document-store.json`，也提供可迁移的 SQL 表结构。

## 本地运行

```bash
go build ./...
go test ./...
go vet ./...
go run ./cmd/server
```

打开 <http://localhost:8080>。生产写接口需要 `X-User-ID` 和 `X-Role: editor` 或 `admin`；回滚需要 `admin`。

## Docker

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh document-version-compare-service linux/amd64
./build_benzhi_docker.sh document-version-compare-service linux/arm64
docker run --rm -p 8080:8080 document-version-compare-service:latest
```

## 验证

```bash
go test -race ./...
go vet ./...
go build ./...
```

项目生产 Go 代码保持 21-24 个文件、2001-2199 行的规模约束。
