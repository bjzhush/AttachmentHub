package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var localFileLinkPattern = regexp.MustCompile(`\[本地文件\]\(([^)\r\n]+)\)`)

type config struct {
	VaultDir      string
	APIBaseURL    string
	PublicBaseURL string
	ImportURL     string
	DryRun        bool
	KeepFiles     bool
	Backup        bool
	Timeout       time.Duration
}

type stats struct {
	MDFilesScanned         int
	MDFilesWithMatches     int
	LinksMatched           int
	LinksEligible          int
	LinksUploaded          int
	LinksReplaced          int
	MDFilesUpdated         int
	AttachmentFilesDeleted int
	Failures               int
}

type uploadResult struct {
	PublicID string
}

type importResponse struct {
	PublicID string `json:"public_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type runner struct {
	cfg        config
	httpClient *http.Client
	cache      map[string]uploadResult
	deleted    map[string]bool
	stats      stats
	failures   []failureItem
}

type failureItem struct {
	Scope  string
	Reason string
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置错误: %v\n", err)
		os.Exit(2)
	}

	r := runner{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		cache:   make(map[string]uploadResult),
		deleted: make(map[string]bool),
	}

	if err := r.run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n执行失败: %v\n", err)
		os.Exit(1)
	}
}

func parseConfig() (config, error) {
	var cfg config

	defaultAPI := strings.TrimSpace(os.Getenv("API_URL"))
	defaultPublicBase := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL"))
	vaultFlag := ""

	timeoutSec := 120

	flag.StringVar(&vaultFlag, "vault", "", "Obsidian vault 目录（也可作为第一个位置参数传入）")
	flag.StringVar(&cfg.APIBaseURL, "api-url", defaultAPI, "AttHub 服务地址（也可用环境变量 API_URL）")
	flag.StringVar(&cfg.PublicBaseURL, "public-base-url", defaultPublicBase, "生成链接使用的公开地址（默认同 api-url，可用 PUBLIC_BASE_URL）")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "仅打印将要变更的内容，不写文件、不删附件")
	flag.BoolVar(&cfg.KeepFiles, "keep-files", false, "上传成功后保留原始 HTML/PDF 文件")
	flag.BoolVar(&cfg.Backup, "backup", false, "回写前先生成 .bak 备份")
	flag.IntVar(&timeoutSec, "timeout-sec", timeoutSec, "HTTP 请求超时时间（秒）")
	flagArgs, posArgs := splitMixedArgs(os.Args[1:])
	if err := flag.CommandLine.Parse(flagArgs); err != nil {
		return config{}, err
	}

	switch {
	case len(posArgs) > 1:
		return config{}, errors.New("位置参数最多只能传 1 个 vault 目录")
	case len(posArgs) == 1:
		cfg.VaultDir = posArgs[0]
	default:
		cfg.VaultDir = vaultFlag
	}

	if strings.TrimSpace(cfg.VaultDir) == "" {
		return config{}, errors.New("缺少 vault 目录，请使用 --vault 或第一个位置参数")
	}

	absVault, err := filepath.Abs(cfg.VaultDir)
	if err != nil {
		return config{}, fmt.Errorf("解析 vault 路径失败: %w", err)
	}
	info, err := os.Stat(absVault)
	if err != nil {
		return config{}, fmt.Errorf("vault 路径不可用: %w", err)
	}
	if !info.IsDir() {
		return config{}, fmt.Errorf("vault 路径不是目录: %s", absVault)
	}
	cfg.VaultDir = absVault

	apiBase, err := normalizeBaseURL(cfg.APIBaseURL)
	if err != nil {
		return config{}, fmt.Errorf("api-url 非法: %w", err)
	}
	cfg.APIBaseURL = apiBase

	if strings.TrimSpace(cfg.PublicBaseURL) == "" {
		cfg.PublicBaseURL = cfg.APIBaseURL
	}
	publicBase, err := normalizeBaseURL(cfg.PublicBaseURL)
	if err != nil {
		return config{}, fmt.Errorf("public-base-url 非法: %w", err)
	}
	cfg.PublicBaseURL = publicBase
	cfg.ImportURL = strings.TrimRight(cfg.APIBaseURL, "/") + "/api/v1/attachments/import"

	if timeoutSec <= 0 {
		return config{}, errors.New("timeout-sec 必须大于 0")
	}
	cfg.Timeout = time.Duration(timeoutSec) * time.Second
	return cfg, nil
}

func splitMixedArgs(args []string) ([]string, []string) {
	flagArgs := make([]string, 0, len(args))
	positional := make([]string, 0, 1)

	for i := 0; i < len(args); i++ {
		current := args[i]
		if current == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}

		if strings.HasPrefix(current, "-") {
			flagArgs = append(flagArgs, current)
			if flagNeedsValue(current) && i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
			continue
		}

		positional = append(positional, current)
	}

	return flagArgs, positional
}

func flagNeedsValue(flagToken string) bool {
	switch {
	case strings.HasPrefix(flagToken, "--vault="):
		return false
	case strings.HasPrefix(flagToken, "--api-url="):
		return false
	case strings.HasPrefix(flagToken, "--public-base-url="):
		return false
	case strings.HasPrefix(flagToken, "--timeout-sec="):
		return false
	}

	switch flagToken {
	case "--vault", "--api-url", "--public-base-url", "--timeout-sec":
		return true
	default:
		return false
	}
}

func normalizeBaseURL(raw string) (string, error) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return "", errors.New("不能为空")
	}
	u, err := neturl.Parse(cleaned)
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("必须是 http/https 且包含主机名")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func (r *runner) run() error {
	mdFiles, err := collectMarkdownFiles(r.cfg.VaultDir)
	if err != nil {
		return err
	}

	fmt.Printf("Vault: %s\n", r.cfg.VaultDir)
	fmt.Printf("Import API: %s\n", r.cfg.ImportURL)
	fmt.Printf("Public Link Base: %s\n", r.cfg.PublicBaseURL)
	fmt.Printf("Markdown files: %d\n\n", len(mdFiles))

	for _, mdPath := range mdFiles {
		r.stats.MDFilesScanned++
		if err := r.processMarkdown(mdPath); err != nil {
			r.addFailure(mdPath, "处理 Markdown 失败: %v", err)
		}
	}

	fmt.Println("\n===== Summary =====")
	fmt.Printf("MD scanned: %d\n", r.stats.MDFilesScanned)
	fmt.Printf("MD with [本地文件]: %d\n", r.stats.MDFilesWithMatches)
	fmt.Printf("Matched links: %d\n", r.stats.LinksMatched)
	fmt.Printf("Eligible links(html/pdf): %d\n", r.stats.LinksEligible)
	fmt.Printf("Uploaded links: %d\n", r.stats.LinksUploaded)
	fmt.Printf("Replaced links: %d\n", r.stats.LinksReplaced)
	fmt.Printf("Updated MD files: %d\n", r.stats.MDFilesUpdated)
	fmt.Printf("Deleted attachments: %d\n", r.stats.AttachmentFilesDeleted)
	fmt.Printf("Failures: %d\n", r.stats.Failures)
	if len(r.failures) > 0 {
		fmt.Println("\n失败明细:")
		for i, item := range r.failures {
			fmt.Printf("%d. %s\n", i+1, item.Scope)
			fmt.Printf("   原因: %s\n", item.Reason)
		}
	}

	if r.stats.Failures > 0 {
		return fmt.Errorf("存在 %d 个失败项", r.stats.Failures)
	}
	return nil
}

func collectMarkdownFiles(root string) ([]string, error) {
	files := make([]string, 0, 256)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func (r *runner) processMarkdown(mdPath string) error {
	original, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("读取 markdown 失败: %w", err)
	}

	content := string(original)
	matches := localFileLinkPattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return nil
	}

	r.stats.MDFilesWithMatches++
	r.stats.LinksMatched += len(matches)

	var builder strings.Builder
	builder.Grow(len(content) + len(matches)*24)

	last := 0
	changed := false
	deleteCandidates := make(map[string]struct{})

	for _, m := range matches {
		fullStart, fullEnd := m[0], m[1]
		targetStart, targetEnd := m[2], m[3]

		builder.WriteString(content[last:fullStart])
		last = fullEnd

		rawTarget := strings.TrimSpace(content[targetStart:targetEnd])
		pathTarget := normalizeLinkTarget(rawTarget)
		resolvedPath, eligible, resolveErr := resolveAttachmentPath(mdPath, pathTarget, r.cfg.VaultDir)
		if resolveErr != nil {
			r.addFailure(mdPath, "解析链接失败 %q: %v", rawTarget, resolveErr)
			builder.WriteString(content[fullStart:fullEnd])
			continue
		}
		if !eligible {
			builder.WriteString(content[fullStart:fullEnd])
			continue
		}
		r.stats.LinksEligible++

		uploaded, ok := r.cache[resolvedPath]
		if !ok {
			result, uploadErr := r.uploadFile(resolvedPath)
			if uploadErr != nil {
				r.addFailure(resolvedPath, "上传失败（引用于 %s）: %v", mdPath, uploadErr)
				builder.WriteString(content[fullStart:fullEnd])
				continue
			}
			uploaded = result
			r.cache[resolvedPath] = result
			r.stats.LinksUploaded++
		}

		replacement := fmt.Sprintf("[附件系统   %s](%s/f/%s)", uploaded.PublicID, strings.TrimRight(r.cfg.PublicBaseURL, "/"), uploaded.PublicID)
		builder.WriteString(replacement)
		deleteCandidates[resolvedPath] = struct{}{}
		changed = true
		r.stats.LinksReplaced++
		fmt.Printf("[OK] %s -> %s\n", resolvedPath, uploaded.PublicID)
	}

	builder.WriteString(content[last:])
	if !changed {
		return nil
	}

	newContent := []byte(builder.String())
	if r.cfg.DryRun {
		fmt.Printf("[DRY-RUN] %s 将更新并替换本地链接\n", mdPath)
		return nil
	}

	if r.cfg.Backup {
		backupPath := mdPath + ".bak"
		if err := copyFile(mdPath, backupPath); err != nil {
			return fmt.Errorf("写入备份失败: %w", err)
		}
	}

	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(mdPath); statErr == nil {
		mode = info.Mode()
	}
	if err := writeFileAtomic(mdPath, newContent, mode); err != nil {
		return fmt.Errorf("写回 markdown 失败: %w", err)
	}
	r.stats.MDFilesUpdated++
	fmt.Printf("[WRITE] %s\n", mdPath)

	if r.cfg.KeepFiles {
		return nil
	}

	for path := range deleteCandidates {
		if r.deleted[path] {
			continue
		}
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				r.deleted[path] = true
				continue
			}
			r.addFailure(path, "删除附件失败（引用于 %s）: %v", mdPath, err)
			continue
		}
		r.deleted[path] = true
		r.stats.AttachmentFilesDeleted++
		fmt.Printf("[DELETE] %s\n", path)
	}

	return nil
}

func normalizeLinkTarget(raw string) string {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "<") && strings.HasSuffix(clean, ">") {
		clean = strings.TrimSpace(clean[1 : len(clean)-1])
	}
	return clean
}

func resolveAttachmentPath(mdPath string, linkTarget string, vaultRoot string) (string, bool, error) {
	if linkTarget == "" {
		return "", false, errors.New("空链接")
	}
	lower := strings.ToLower(linkTarget)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return "", false, nil
	}

	trimmed := linkTarget
	if idx := strings.IndexAny(trimmed, "?#"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "", false, errors.New("链接路径为空")
	}

	if strings.Contains(trimmed, "%") {
		if decoded, err := neturl.PathUnescape(trimmed); err == nil {
			trimmed = decoded
		}
	}

	var candidate string
	if filepath.IsAbs(trimmed) {
		candidate = trimmed
	} else if strings.HasPrefix(trimmed, "/") {
		candidate = filepath.Join(vaultRoot, strings.TrimPrefix(trimmed, "/"))
	} else {
		candidate = filepath.Join(filepath.Dir(mdPath), trimmed)
	}

	absPath, err := filepath.Abs(candidate)
	if err != nil {
		return "", false, err
	}
	ext := strings.ToLower(filepath.Ext(absPath))
	if ext != ".pdf" && ext != ".html" && ext != ".htm" {
		return "", false, nil
	}

	if _, found := strings.CutPrefix(absPath, vaultRoot+string(os.PathSeparator)); !found && absPath != vaultRoot {
		return "", false, fmt.Errorf("附件路径超出 vault: %s", absPath)
	}
	return absPath, true, nil
}

func (r *runner) uploadFile(path string) (uploadResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return uploadResult{}, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return uploadResult{}, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return uploadResult{}, err
	}
	if err := writer.Close(); err != nil {
		return uploadResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, r.cfg.ImportURL, &body)
	if err != nil {
		return uploadResult{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return uploadResult{}, err
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if readErr != nil {
		return uploadResult{}, fmt.Errorf("读取上传响应失败: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var apiErr errorResponse
		if err := json.Unmarshal(respBody, &apiErr); err == nil && strings.TrimSpace(apiErr.Error) != "" {
			return uploadResult{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error)
		}
		return uploadResult{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var payload importResponse
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return uploadResult{}, fmt.Errorf("解析上传响应失败: %w", err)
	}
	if len(strings.TrimSpace(payload.PublicID)) != 12 {
		return uploadResult{}, fmt.Errorf("上传响应缺少合法 public_id: %s", strings.TrimSpace(string(respBody)))
	}
	return uploadResult{PublicID: strings.TrimSpace(payload.PublicID)}, nil
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".vault-sync-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Chmod(mode.Perm()); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func copyFile(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	info, err := input.Stat()
	if err != nil {
		return err
	}

	output, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return output.Close()
}

func (r *runner) addFailure(scope string, format string, args ...any) {
	reason := fmt.Sprintf(format, args...)
	r.stats.Failures++
	r.failures = append(r.failures, failureItem{
		Scope:  scope,
		Reason: reason,
	})
	fmt.Printf("[FAIL] %s: %s\n", scope, reason)
}
