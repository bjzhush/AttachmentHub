# AttachmentHub

`AttachmentHub` 由两个独立模块组成：

- `AttHub`：附件管理系统本体（Go API + SQLite + 本地附件存储）
- `ObsidianImport`：导入与同步工具（通过 HTTP API 调用 `AttHub`）

文档统一维护在本文件。

## 目录结构

```text
AttachmentHub/
├── AttHub/          # API 服务
└── ObsidianImport/  # 导入/同步脚本
```

## AttHub（附件系统）

### 当前能力

- 导入附件：支持 `PDF/HTML`
- 自动生成 12 位公开 ID：`public_id`
- 内部自增 ID：`id`
- 搜索、更新 URL/备注、删除
- `GET /f/{public_id}` 直接访问附件
- Web 管理页：`/web/attachments`

去重策略：

- 按 `sha256` 去重
- 重复上传时返回 `200` + 已有记录（包含 `public_id`）
- 首次上传返回 `201`

### API 概览

- `POST /api/v1/attachments/import`
- `GET /api/v1/attachments`
- `GET /api/v1/attachments/{id}`
- `GET /api/v1/attachments/public/{public_id}`
- `PATCH /api/v1/attachments/{id}`
- `DELETE /api/v1/attachments/{id}`
- `GET /f/{public_id}`
- `GET /healthz`

导入示例：

```bash
curl -X POST http://127.0.0.1:10001/api/v1/attachments/import \
  -F "file=@/path/to/file.pdf" \
  -F "url=https://example.com/article" \
  -F "note=离线备份"
```

### 运行与管理

开发运行：

```bash
cd AttHub
go mod tidy
go run ./cmd/server
```

服务脚本（单实例）：

```bash
cd AttHub
./scripts/atthub-service.sh start|stop|restart|status|logs|reset
```

等价 Make 命令：

```bash
make service-start
make service-status
make service-restart
make service-stop
make service-reset
```

开发期重置（清空 SQLite + 附件目录）：

```bash
cd AttHub
make reset-dev
```

默认配置（可通过环境变量覆盖）：

- `ATTHUB_PORT=10001`
- `ATTHUB_DB_PATH=./data/attachmenthub.db`
- `ATTHUB_STORAGE_DIR=./attachments`
- `ATTHUB_MAX_UPLOAD_MB=100`

## ObsidianImport（导入/同步）

### 工具 1：目录扫描上传脚本（Shell）

脚本：`ObsidianImport/upload_from_dir.sh`

功能：

- 递归扫描目录里的 `PDF/HTML`
- 调用 `AttHub` 导入接口上传
- 失败文件移动到 `${SCAN_DIR}/failed`
- 支持 `SingleFile` 导出 HTML 的头部 `url: ...` 自动提取

配置：

```bash
cd ObsidianImport
cp config.env.example config.env
```

`config.env` 关键项：

- `API_URL`：如 `http://127.0.0.1:10001`
- `SCAN_DIR`：待扫描目录

执行：

```bash
./upload_from_dir.sh
./upload_from_dir.sh --once
```

### 工具 2：Obsidian Vault 本地链接同步（Go）

程序：`ObsidianImport/cmd/vault_local_file_sync`

功能：

- 扫描指定 vault 下所有 `.md`
- 匹配 `[本地文件](...)`
- 仅处理 `.html/.htm/.pdf` 链接
- 上传成功后替换为附件系统链接
- Markdown 写回成功后删除原附件文件（可关闭）
- 末尾输出失败明细（文件 + 原因）

替换格式：

```md
[附件系统   12位ID](http://你的AttHub地址/f/12位ID)
```

说明：`附件系统` 和 `12位ID` 之间固定 3 个半角空格。

执行示例：

```bash
cd ObsidianImport
go run ./cmd/vault_local_file_sync /path/to/your-vault --api-url http://127.0.0.1:10001
```

也支持：

```bash
go run ./cmd/vault_local_file_sync --vault /path/to/your-vault --api-url http://127.0.0.1:10001
```

常用参数：

- `--dry-run`：仅预览，不写入、不删除
- `--backup`：回写前生成 `.bak`
- `--keep-files`：上传成功后不删除原附件
- `--public-base-url`：生成链接用的外部地址（默认与 `--api-url` 相同）

## 模块边界

- `ObsidianImport` 不直接访问 `AttHub` 的 SQLite 或附件目录
- 只通过 HTTP API 交互
- 便于两部分独立演进与发布
