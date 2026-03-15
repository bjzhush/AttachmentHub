# ObsidianImport

该目录预留给“历史数据批量导入脚本”模块。

边界约束：

- 与 `AttHub` 系统本体独立开发、独立运行
- 通过 `AttHub` 的 HTTP API 导入附件与元数据
- 不直接访问 `AttHub` 的 SQLite 数据库和文件存储目录

这样可以保证系统内核稳定，同时让导入逻辑可按来源数据单独演进。

## 目录扫描上传脚本（最简版）

该脚本会递归扫描指定目录中的 `PDF/HTML` 文件，并调用 `AttHub` 上传接口导入。

当前行为：

- `API_URL` 只配置域名/IP+端口，脚本内部固定拼接上传路径 `/api/v1/attachments/import`
- 上传失败的文件会移动到 `SCAN_DIR/failed`（保留相对目录结构）
- 兼容 `SingleFile` 导出的 HTML：如果前 4 行出现 `url: https://...`，会提取该链接并作为 `url` 字段提交
- 轮巡间隔会按规则变化：
  - 本轮有上传成功：下轮 1 分钟后
  - 本轮没有成功上传：间隔翻倍到 2/4/8/16/32 分钟（上限 32 分钟）

### 1) 准备配置

```bash
cd ObsidianImport
cp config.env.example config.env
```

编辑 `config.env` 两项（都必填）：

- `API_URL`：本地 `AttHub` 地址（仅域名/IP+端口，不带接口路径）
- `SCAN_DIR`：待扫描目录

### 2) 执行上传

```bash
./upload_from_dir.sh
```

如果你使用了自定义配置文件路径：

```bash
./upload_from_dir.sh /path/to/your-config.env
```

开发测试时如果只想跑一轮就退出：

```bash
./upload_from_dir.sh --once
```
