# AttHub

一个面向个人使用的离线附件管理 API 服务，目标数据规模 `<= 10,000` 条。

## 技术选型

- 语言与 API：Go（`net/http` + `chi`）
- 数据库：SQLite（单文件、低维护成本、足够支撑 1w 条数据）
- 附件存储：本地文件目录（数据库保存元数据，不保存二进制）
- 搜索：基于 `URL` 与 `note` 的关键词匹配（`instr(lower(...))`）

这个组合在单用户场景下简单、稳定，且开发与部署成本最低。

## 当前实现（v1）

- 导入附件：支持 `PDF/HTML` 文件 + 可选 `url` + 可选 `note`
- 每个附件自动分配 8 位字母/数字 ID（`public_id`）
- 搜索附件：按关键词匹配 `url` 或 `note` 子串
- 更新附件：更新 `url` / `note`（空字符串会被清空为 `NULL`）
- 删除附件：删除附件记录，并清理对应文件
- 通过 `public_id` 直接打开附件：`GET /f/{public_id}`
- 简易管理页面：`GET /web/attachments`
- 健康检查：`GET /healthz`

## API

### 1) 导入附件

`POST /api/v1/attachments/import`

`multipart/form-data` 字段：

- `file`：必填，`pdf/html/htm`
- `url`：选填
- `note`：选填

示例：

```bash
curl -X POST http://localhost:8080/api/v1/attachments/import \
  -F "file=@/path/to/file.pdf" \
  -F "url=https://example.com/article" \
  -F "note=离线备份"
```

### 2) 搜索附件

`GET /api/v1/attachments?keyword=xxx&page=1&page_size=20`

### 3) 获取单条附件

`GET /api/v1/attachments/{id}`

### 4) 按 8 位 ID 获取附件元数据

`GET /api/v1/attachments/public/{public_id}`

### 5) 更新 URL / 备注

`PATCH /api/v1/attachments/{id}`

请求体（JSON）：

```json
{
  "url": "https://new.example.com",
  "note": "新的备注"
}
```

### 6) 删除附件

`DELETE /api/v1/attachments/{id}`

### 7) 直接打开附件文件

`GET /f/{public_id}`

## 数据模型

主表 `attachments` 核心字段：

- 访问 ID：`public_id`（8 位字母/数字）
- 文件信息：`original_name`, `stored_name`, `file_ext`, `content_type`, `file_size`, `sha256`
- 元数据：`source_url`, `note`
- 时间字段：`created_at`, `updated_at`

设计上每条记录对应一个附件文件，不共享 URL/备注实例。

## 运行方式

```bash
cd AttHub
go mod tidy
go run ./cmd/server
```

浏览器页面：

```bash
open http://localhost:8080/web/attachments
```

开发期一键清空测试数据（SQLite + 附件目录）：

```bash
make reset-dev
```

默认配置（可通过环境变量覆盖）：

- `ATTHUB_PORT=8080`
- `ATTHUB_DB_PATH=./data/attachmenthub.db`
- `ATTHUB_STORAGE_DIR=./attachments`
- `ATTHUB_MAX_UPLOAD_MB=100`

## 与 ObsidianImport 的边界

- `ObsidianImport` 不直接写库，不直接操作 `attachments`
- 仅通过 HTTP API 调用 `AttHub` 导入/更新
- 两者独立发布与迭代，避免耦合

## 你需求里容易遗漏但建议提前确认的点

- 去重策略：同一文件重复导入是否允许（可按 `sha256` 控制）
- 文件大小上限：单文件是否可能超过 100MB
- 存储增长：1w 条记录下总磁盘占用预估与告警阈值
- 备份策略：SQLite 文件与附件目录必须一起备份
- 删除策略：是否需要“软删除”与误删恢复
- URL 健康检查：是否需要后台任务定期验证链接可访问性
- 搜索增强：若后续需要更复杂检索，可升级 SQLite FTS5

## 已知技术风险与应对

- SQLite 并发写入能力有限：单用户场景可接受，已启用 WAL
- 文件与数据库一致性：当前删除是“先删记录后清理文件（尽力）”，后续可增加孤儿文件巡检
- 仅依赖文件后缀与 MIME 校验：可继续加强为内容特征校验
