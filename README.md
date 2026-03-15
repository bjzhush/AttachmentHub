# AttachmentHub

`AttachmentHub` 按模块拆分为两个独立部分：

- `AttHub`：附件管理系统本体（Go API + SQLite + 本地附件存储）
- `ObsidianImport`：后续可独立开发的批量导入脚本模块（与系统本体仅通过 API 交互）

当前已实现 `AttHub` 第一版，支持导入 `PDF/HTML` 附件、8 位附件 ID、搜索、更新 URL/备注、删除，以及 Web 管理页面。
