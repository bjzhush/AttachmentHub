package api

const webAppHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>AttachmentHub</title>
  <style>
    :root {
      --bg-1: #f3f8ff;
      --bg-2: #eef5ff;
      --surface: rgba(255, 255, 255, 0.86);
      --surface-strong: #ffffff;
      --line: #d8e2ef;
      --text: #152033;
      --muted: #5d6f87;
      --primary: #0a66d1;
      --primary-strong: #064c9f;
      --danger: #cf2e45;
      --shadow: 0 14px 30px rgba(20, 45, 90, 0.12);
    }
    body {
      margin: 0;
      font-family: "Avenir Next", "SF Pro Text", "Segoe UI", sans-serif;
      background:
        radial-gradient(circle at 12% 20%, #dceeff 0%, rgba(220, 238, 255, 0) 42%),
        radial-gradient(circle at 88% 12%, #e4fbf4 0%, rgba(228, 251, 244, 0) 44%),
        linear-gradient(180deg, var(--bg-1) 0%, var(--bg-2) 100%);
      color: var(--text);
    }
    .bg-orb {
      position: fixed;
      z-index: -1;
      width: 340px;
      height: 340px;
      border-radius: 999px;
      filter: blur(12px);
      opacity: 0.42;
      pointer-events: none;
    }
    .orb-a {
      top: -120px;
      left: -120px;
      background: #cde2ff;
    }
    .orb-b {
      bottom: -120px;
      right: -90px;
      background: #c7f2ea;
    }
    .wrap {
      max-width: 1680px;
      margin: 18px auto 52px;
      padding: 0 16px;
    }
    .panel {
      background: var(--surface);
      backdrop-filter: blur(6px);
      border: 1px solid var(--line);
      border-radius: 16px;
      padding: 20px 22px;
      margin-bottom: 16px;
      box-shadow: var(--shadow);
    }
    .hero {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 16px;
      padding: 18px;
      background: linear-gradient(120deg, rgba(10, 102, 209, 0.14), rgba(8, 184, 132, 0.1));
      border: 1px solid rgba(10, 102, 209, 0.2);
      border-radius: 18px;
      margin-bottom: 16px;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 12px;
    }
    .brand-mark {
      width: 44px;
      height: 44px;
      border-radius: 12px;
      display: grid;
      place-items: center;
      font-weight: 800;
      letter-spacing: 0.08em;
      color: #fff;
      background: linear-gradient(135deg, var(--primary), #0f8fcd);
      box-shadow: 0 8px 18px rgba(10, 102, 209, 0.36);
    }
    .brand-title {
      margin: 0;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.02em;
      color: #0d2e58;
      text-shadow: 0 1px 0 rgba(255, 255, 255, 0.7);
    }
    .brand-subtitle {
      margin-top: 2px;
      font-size: 12px;
      color: #3f5978;
      letter-spacing: 0.06em;
      text-transform: uppercase;
    }
    .quick-hint {
      font-size: 12px;
      color: #33567f;
      background: rgba(255, 255, 255, 0.7);
      border: 1px solid rgba(51, 86, 127, 0.18);
      border-radius: 999px;
      padding: 7px 10px;
    }
    form {
      display: grid;
      grid-template-columns: repeat(12, minmax(0, 1fr));
      gap: 12px;
      align-items: center;
    }
    .section-title {
      margin: 0 0 12px;
      font-size: 15px;
      letter-spacing: 0.02em;
      color: #234160;
    }
    input[type="text"], input[type="file"], textarea {
      width: 100%;
      box-sizing: border-box;
      border: 1px solid var(--line);
      border-radius: 10px;
      padding: 9px 11px;
      font-size: 14px;
      background: var(--surface-strong);
      color: var(--text);
      transition: box-shadow .18s ease, border-color .18s ease;
    }
    input:focus, textarea:focus {
      outline: none;
      border-color: rgba(10, 102, 209, 0.45);
      box-shadow: 0 0 0 4px rgba(10, 102, 209, 0.12);
    }
    textarea {
      min-height: 92px;
      resize: vertical;
      font-family: inherit;
    }
    input::placeholder,
    textarea::placeholder {
      color: #b7c6d8;
      opacity: 1;
    }
    .field-label {
      font-size: 11px;
      color: var(--muted);
      text-transform: uppercase;
      letter-spacing: 0.07em;
      margin-bottom: 6px;
    }
    button {
      border: 0;
      border-radius: 10px;
      padding: 9px 12px;
      font-size: 14px;
      font-weight: 600;
      cursor: pointer;
      background: var(--primary);
      color: #fff;
      transition: transform .12s ease, background .18s ease;
    }
    button:hover {
      transform: translateY(-1px);
      background: var(--primary-strong);
    }
    button:active {
      transform: translateY(0);
    }
    button.secondary { background: #4f6787; }
    button.secondary:hover { background: #415675; }
    button.danger { background: var(--danger); }
    button.danger:hover { background: #a82337; }
    .result-list {
      display: grid;
      gap: 11px;
      margin-top: 8px;
    }
    .item {
      border: 1px solid var(--line);
      border-radius: 14px;
      background: rgba(255, 255, 255, 0.96);
      padding: 14px 16px;
      box-shadow: 0 6px 16px rgba(21, 41, 73, 0.08);
    }
    .item-head {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 10px;
      margin-bottom: 10px;
    }
    .item-main {
      min-width: 0;
      flex: 1 1 auto;
    }
    .item-title-row {
      display: flex;
      align-items: center;
      gap: 10px;
      min-width: 0;
    }
    .item-id {
      font-family: ui-monospace, "SFMono-Regular", Menlo, monospace;
      font-size: 13px;
      font-weight: 700;
      color: #0a4f9c;
      background: rgba(10, 102, 209, 0.1);
      border: 1px solid rgba(10, 102, 209, 0.2);
      border-radius: 8px;
      padding: 2px 8px;
      width: fit-content;
      flex: 0 0 auto;
    }
    .item-seq {
      font-family: ui-monospace, "SFMono-Regular", Menlo, monospace;
      font-size: 12px;
      font-weight: 700;
      color: #56708f;
      background: #eef3f9;
      border: 1px solid #d9e4f2;
      border-radius: 8px;
      padding: 2px 8px;
      flex: 0 0 auto;
    }
    .item-name {
      font-size: 15px;
      font-weight: 700;
      margin: 0;
      min-width: 0;
      width: 40ch;
      max-width: 40ch;
      flex: 0 1 40ch;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .item-meta {
      font-size: 12px;
      color: var(--muted);
      margin-top: 3px;
    }
    .item-side {
      display: grid;
      grid-template-columns: minmax(0, 1fr) 230px;
      align-items: start;
      gap: 10px;
      min-width: 520px;
      max-width: 760px;
      width: 52%;
      margin-left: auto;
    }
    .item-snippet {
      flex: 1;
      min-width: 0;
      border: 1px solid #dce7f4;
      border-radius: 10px;
      background: rgba(247, 251, 255, 0.86);
      padding: 7px 9px;
      display: grid;
      gap: 4px;
      transition: border-color .18s ease;
    }
    .snippet-line {
      display: flex;
      gap: 6px;
      align-items: baseline;
      min-width: 0;
    }
    .snippet-label {
      flex: 0 0 auto;
      font-size: 11px;
      color: #4a627f;
      letter-spacing: 0.05em;
      text-transform: uppercase;
      font-weight: 700;
    }
    .snippet-value {
      min-width: 0;
      flex: 1;
      display: block;
      font-size: 12px;
      color: #3a516e;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .snippet-link {
      color: #0a5aa8;
      text-decoration: underline;
      text-decoration-color: rgba(10, 90, 168, 0.36);
      text-underline-offset: 2px;
    }
    .snippet-link:hover {
      color: #084b8a;
      text-decoration-color: rgba(8, 75, 138, 0.52);
    }
    .has-overflow-tip {
      position: relative;
      cursor: default;
    }
    .has-overflow-tip::before {
      content: "";
      position: absolute;
      left: 12px;
      top: calc(100% + 2px);
      width: 8px;
      height: 8px;
      background: rgba(15, 31, 54, 0.96);
      transform: rotate(45deg);
      opacity: 0;
      visibility: hidden;
      transition: opacity .14s ease, visibility .14s ease;
      z-index: 39;
      pointer-events: none;
    }
    .has-overflow-tip::after {
      content: attr(data-tip);
      position: absolute;
      left: 0;
      top: calc(100% + 8px);
      max-width: min(820px, calc(100vw - 48px));
      background: rgba(15, 31, 54, 0.96);
      color: #fff;
      font-size: 12px;
      line-height: 1.45;
      border-radius: 10px;
      padding: 8px 10px;
      box-shadow: 0 12px 24px rgba(8, 18, 35, 0.35);
      white-space: normal;
      word-break: break-word;
      opacity: 0;
      visibility: hidden;
      transform: translateY(5px);
      transition: opacity .14s ease, transform .14s ease, visibility .14s ease;
      z-index: 40;
      pointer-events: none;
    }
    .has-overflow-tip:hover::before,
    .has-overflow-tip:hover::after,
    .has-overflow-tip:focus-visible::before,
    .has-overflow-tip:focus-visible::after {
      opacity: 1;
      visibility: visible;
      transform: translateY(0);
    }
    .item-actions {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      width: 230px;
      justify-content: flex-end;
      align-content: flex-start;
      opacity: 0;
      visibility: hidden;
      transform: translateY(-2px);
      pointer-events: none;
      transition: opacity .16s ease, transform .16s ease, visibility .16s ease;
    }
    .item:hover .item-actions,
    .item:focus-within .item-actions {
      opacity: 1;
      visibility: visible;
      transform: translateY(0);
      pointer-events: auto;
    }
    .item:hover .item-snippet,
    .item:focus-within .item-snippet {
      border-color: #c7d9ef;
    }
    .open-link {
      font-size: 13px;
      text-decoration: none;
      color: #0a539f;
      border: 1px solid rgba(10, 83, 159, 0.23);
      background: rgba(10, 83, 159, 0.08);
      border-radius: 9px;
      padding: 8px 10px;
      font-weight: 600;
    }
    .pagination {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      margin-top: 14px;
      border-top: 1px dashed #d9e3ef;
      padding-top: 12px;
    }
    .pagination-left {
      font-size: 13px;
      color: #4e6480;
    }
    .pagination-actions {
      display: flex;
      gap: 8px;
    }
    .pagination button[disabled] {
      opacity: 0.45;
      cursor: not-allowed;
      transform: none;
    }
    .muted { color: var(--muted); }
    .status {
      margin-top: 12px;
      min-height: 20px;
      color: var(--muted);
      font-size: 13px;
      padding-left: 2px;
    }
    .status.error { color: #a92c3e; }
    .status.ok { color: #126a52; }
    .center-notice {
      position: fixed;
      left: 50%;
      top: 50%;
      transform: translate(-50%, -50%) scale(0.96);
      z-index: 1500;
      min-width: min(640px, calc(100vw - 36px));
      max-width: min(860px, calc(100vw - 24px));
      padding: 14px 18px;
      border-radius: 12px;
      border: 1px solid rgba(145, 28, 47, 0.52);
      background: rgba(176, 35, 57, 0.97);
      color: #fff;
      text-align: center;
      font-size: 15px;
      font-weight: 700;
      letter-spacing: 0.01em;
      box-shadow: 0 18px 42px rgba(66, 12, 23, 0.42);
      opacity: 0;
      visibility: hidden;
      pointer-events: none;
      transition: opacity .2s ease, transform .2s ease, visibility .2s ease;
    }
    .center-notice.show {
      opacity: 1;
      visibility: visible;
      transform: translate(-50%, -50%) scale(1);
    }
    .modal {
      position: fixed;
      inset: 0;
      display: none;
      align-items: center;
      justify-content: center;
      padding: 16px;
      background: rgba(6, 21, 38, 0.48);
      backdrop-filter: blur(2px);
    }
    .modal.show {
      display: flex;
    }
    .modal-card {
      width: min(640px, 100%);
      background: #fff;
      border-radius: 14px;
      border: 1px solid #d9e4f1;
      box-shadow: 0 18px 36px rgba(5, 21, 45, 0.26);
      padding: 16px;
    }
    .modal-title {
      margin: 0;
      font-size: 17px;
      color: #183355;
    }
    .modal-sub {
      font-size: 12px;
      color: #5c6f88;
      margin: 4px 0 12px;
    }
    .modal-file {
      font-size: 12px;
      color: #4f647f;
      margin: -6px 0 12px;
      line-height: 1.45;
      white-space: normal;
      word-break: break-word;
    }
    .modal-actions {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
      margin-top: 12px;
    }
    @media (max-width: 900px) {
      .quick-hint {
        display: none;
      }
      form {
        gap: 10px;
      }
      #import-form input[type="file"] { grid-column: span 12 !important; }
      #import-form input[name="url"] { grid-column: span 12 !important; }
      #import-form input[name="note"] { grid-column: span 12 !important; }
      #import-form button { grid-column: span 12 !important; }
      #search-form input { grid-column: span 12 !important; }
      #search-form button { grid-column: span 4 !important; }
      .pagination {
        flex-direction: column;
        align-items: flex-start;
      }
      .item-side {
        min-width: 0;
        max-width: none;
        width: 100%;
        grid-template-columns: minmax(0, 1fr) 230px;
      }
      .item-actions {
        width: 230px;
        opacity: 1;
        visibility: visible;
        transform: none;
        pointer-events: auto;
      }
    }
    @media (max-width: 600px) {
      .hero {
        flex-direction: column;
        align-items: flex-start;
      }
      .item-head {
        flex-direction: column;
      }
      .item-actions button,
      .item-actions .open-link {
        width: 100%;
        text-align: center;
        box-sizing: border-box;
      }
      .item-name {
        width: auto;
        max-width: 100%;
        flex: 1 1 auto;
        white-space: normal;
      }
      .item-side {
        grid-template-columns: 1fr;
      }
      .item-actions {
        width: 100%;
        justify-content: flex-start;
      }
    }
    @media (hover: none) {
      .item-actions {
        width: 230px;
        opacity: 1;
        visibility: visible;
        transform: none;
        pointer-events: auto;
      }
    }
  </style>
</head>
<body>
  <div class="bg-orb orb-a"></div>
  <div class="bg-orb orb-b"></div>
  <div id="center-notice" class="center-notice" role="alert" aria-live="assertive"></div>
  <div class="wrap">
    <header class="hero">
      <div class="brand">
        <div class="brand-mark">AH</div>
        <div>
          <h1 class="brand-title">附件管理系统</h1>
          <div class="brand-subtitle">AttachmentHub</div>
        </div>
      </div>
      <div class="quick-hint">Web + API unified management</div>
    </header>

    <section class="panel">
      <h2 class="section-title">导入附件</h2>
      <form id="import-form">
        <input style="grid-column: span 4;" name="file" type="file" required />
        <input style="grid-column: span 3;" name="url" type="text" placeholder="Optional URL" />
        <input style="grid-column: span 3;" name="note" type="text" placeholder="Optional note" />
        <button style="grid-column: span 2;" type="submit">Import</button>
      </form>
      <div id="import-status" class="status"></div>
    </section>

    <section class="panel">
      <h2 class="section-title">附件列表</h2>
      <form id="search-form">
        <input style="grid-column: span 5;" id="keyword" type="text" placeholder="Search URL / note" />
        <input style="grid-column: span 5;" id="filename-keyword" type="text" placeholder="Search stored_name" />
        <button style="grid-column: span 1;" type="submit">Search</button>
        <button style="grid-column: span 1;" type="button" class="secondary" id="refresh-btn">Refresh</button>
      </form>
      <div id="search-status" class="status"></div>
      <div id="result-body" class="result-list"></div>
      <div id="pagination" class="pagination">
        <div id="pagination-info" class="pagination-left"></div>
        <div class="pagination-actions">
          <button id="prev-page-btn" type="button" class="secondary">Prev</button>
          <button id="next-page-btn" type="button" class="secondary">Next</button>
        </div>
      </div>
    </section>
  </div>

  <div id="editor-modal" class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <div class="modal-card">
      <h3 class="modal-title" id="modal-title">编辑附件元数据</h3>
      <div id="modal-sub" class="modal-sub"></div>
      <div id="modal-file" class="modal-file"></div>
      <input id="edit-id" type="hidden" />
      <div style="display:grid; gap:10px;">
        <label>
          <div class="field-label">URL</div>
          <input id="edit-url" type="text" placeholder="https://example.com" />
        </label>
        <label>
          <div class="field-label">Note</div>
          <textarea id="edit-note" placeholder="Optional note"></textarea>
        </label>
      </div>
      <div class="modal-actions">
        <button type="button" class="secondary" id="modal-cancel">Cancel</button>
        <button type="button" id="modal-save">Save</button>
      </div>
    </div>
  </div>

  <script>
    const importForm = document.getElementById("import-form");
    const importStatus = document.getElementById("import-status");
    const searchForm = document.getElementById("search-form");
    const searchStatus = document.getElementById("search-status");
    const resultBody = document.getElementById("result-body");
    const keywordInput = document.getElementById("keyword");
    const filenameKeywordInput = document.getElementById("filename-keyword");
    const refreshBtn = document.getElementById("refresh-btn");
    const pagination = document.getElementById("pagination");
    const paginationInfo = document.getElementById("pagination-info");
    const prevPageBtn = document.getElementById("prev-page-btn");
    const nextPageBtn = document.getElementById("next-page-btn");
    const editorModal = document.getElementById("editor-modal");
    const editIDInput = document.getElementById("edit-id");
    const editURLInput = document.getElementById("edit-url");
    const editNoteInput = document.getElementById("edit-note");
    const modalCancelBtn = document.getElementById("modal-cancel");
    const modalSaveBtn = document.getElementById("modal-save");
    const modalSub = document.getElementById("modal-sub");
    const modalFile = document.getElementById("modal-file");
    const centerNotice = document.getElementById("center-notice");
    const itemStore = new Map();
    let currentPage = 1;
    let currentPageSize = 50;
    let currentTotal = 0;
    let searchMode = false;
    let centerNoticeTimer = null;

    async function loadList(requestedPage) {
      const keyword = keywordInput.value.trim();
      const filename = filenameKeywordInput.value.trim();
      const isSearch = keyword.length > 0 || filename.length > 0;
      const page = isSearch ? 1 : Math.max(1, requestedPage || currentPage || 1);

      const params = new URLSearchParams();
      if (keyword) {
        params.set("keyword", keyword);
      }
      if (filename) {
        params.set("filename", filename);
      }
      if (!isSearch) {
        params.set("page", String(page));
      }

      const response = await fetch("/api/v1/attachments?" + params.toString());
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "Search failed");
      }

      searchMode = isSearch;
      currentPage = data.page || page;
      currentPageSize = data.page_size || 50;
      currentTotal = data.total || 0;

      renderRows(data.items || []);
      renderPagination();

      if (isSearch) {
        let filterLabel = "filters";
        if (keyword && filename) {
          filterLabel = "URL/note + stored_name";
        } else if (keyword) {
          filterLabel = "URL/note";
        } else if (filename) {
          filterLabel = "stored_name";
        }
        setStatus(searchStatus, "Found " + currentTotal + " result(s) by " + filterLabel + ", showing first " + Math.min(currentTotal, currentPageSize) + ".", "ok");
      } else {
        setStatus(searchStatus, "Loaded page " + currentPage + " (" + (data.items || []).length + " item(s)).", "ok");
      }
    }

    function renderRows(items) {
      itemStore.clear();
      resultBody.innerHTML = "";
      if (!items.length) {
        const empty = document.createElement("div");
        empty.className = "item muted";
        empty.textContent = "No attachments found.";
        resultBody.appendChild(empty);
        return;
      }

      for (const item of items) {
        itemStore.set(String(item.id), item);

        const card = document.createElement("article");
        card.className = "item";
        card.dataset.id = item.id;
        card.tabIndex = 0;
        card.innerHTML =
          "<div class='item-head'>" +
          "  <div class='item-main'>" +
          "    <div class='item-title-row'>" +
          "      <div class='item-seq'>#" + item.id + "</div>" +
          "      <div class='item-id'>" + escapeHtml(item.public_id) + "</div>" +
          "      <div class='item-name js-overflow-tip' data-full='" + escapeAttr(tooltipText(item.original_name)) + "'>" + escapeHtml(item.original_name) + "</div>" +
          "    </div>" +
          "    <div class='item-meta'>" + escapeHtml(item.content_type) + " · " + formatSizeMB(item.file_size) + " · 上传时间：" + escapeHtml(formatUploadTime(item.created_at)) + "</div>" +
          "  </div>" +
          "  <div class='item-side'>" +
          "    <div class='item-snippet'>" +
          "      <div class='snippet-line'><span class='snippet-label'>URL</span>" + renderURLValue(item.url) + "</div>" +
          "      <div class='snippet-line'><span class='snippet-label'>Note</span><span class='snippet-value js-overflow-tip' data-full='" + escapeAttr(tooltipText(item.note)) + "'>" + previewText(item.note) + "</span></div>" +
          "    </div>" +
          "    <div class='item-actions'>" +
          "      <a class='open-link' href='/f/" + encodeURIComponent(item.public_id) + "' target='_blank'>Open</a>" +
          "      <button class='secondary edit-btn' type='button'>Edit</button>" +
          "      <button class='danger del-btn' type='button'>Delete</button>" +
          "    </div>" +
          "  </div>" +
          "</div>";
        resultBody.appendChild(card);
      }
      updateOverflowTips();
    }

    function renderPagination() {
      if (searchMode) {
        pagination.style.display = "none";
        return;
      }

      pagination.style.display = "flex";
      const totalPages = Math.max(1, Math.ceil(currentTotal / currentPageSize));
      paginationInfo.textContent = "Page " + currentPage + " / " + totalPages + " · Total " + currentTotal;
      prevPageBtn.disabled = currentPage <= 1;
      nextPageBtn.disabled = currentPage >= totalPages;
    }

    async function saveEditorChanges() {
      const id = editIDInput.value;
      const url = editURLInput.value;
      const note = editNoteInput.value;
      if (!id) {
        throw new Error("Missing attachment id");
      }
      const response = await fetch("/api/v1/attachments/" + id, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url, note })
      });
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "Update failed");
      }
      setStatus(searchStatus, "Updated " + data.public_id + ".", "ok");
    }

    async function deleteRowByID(id) {
      const response = await fetch("/api/v1/attachments/" + id, { method: "DELETE" });
      if (!response.ok) {
        let message = "Delete failed";
        try {
          const data = await response.json();
          message = data.error || message;
        } catch (_) {}
        throw new Error(message);
      }
      setStatus(searchStatus, "Deleted attachment #" + id + ".", "ok");
    }

    function openEditorByID(id) {
      const item = itemStore.get(String(id));
      if (!item) {
        setStatus(searchStatus, "Attachment not found in current list.", "error");
        return;
      }
      editIDInput.value = String(item.id);
      editURLInput.value = item.url || "";
      editNoteInput.value = item.note || "";
      modalSub.textContent = "#" + item.id + " · " + item.public_id;
      modalFile.textContent = "文件名： " + item.original_name;
      editorModal.classList.add("show");
    }

    function closeEditor() {
      editorModal.classList.remove("show");
      editIDInput.value = "";
      editURLInput.value = "";
      editNoteInput.value = "";
      modalSub.textContent = "";
      modalFile.textContent = "";
    }

    function setStatus(node, message, type) {
      node.className = "status";
      if (type) {
        node.classList.add(type);
      }
      node.textContent = message;
      if (type === "error" && message) {
        showCenterNotice(message);
      }
    }

    function showCenterNotice(message) {
      if (!centerNotice || !message) {
        return;
      }

      centerNotice.textContent = String(message);
      centerNotice.classList.add("show");

      if (centerNoticeTimer) {
        window.clearTimeout(centerNoticeTimer);
      }

      centerNoticeTimer = window.setTimeout(() => {
        centerNotice.classList.remove("show");
      }, 4200);
    }

    importForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      setStatus(importStatus, "Importing...");
      try {
        const formData = new FormData(importForm);
        const response = await fetch("/api/v1/attachments/import", {
          method: "POST",
          body: formData
        });
        const data = await response.json();
        if (!response.ok) {
          throw new Error(data.error || "Import failed");
        }
        importForm.reset();
        setStatus(importStatus, "Imported " + data.public_id + " (" + data.original_name + ").", "ok");
        await loadList(searchMode ? 1 : currentPage);
      } catch (error) {
        setStatus(importStatus, error.message, "error");
      }
    });

    searchForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      setStatus(searchStatus, "Searching...");
      try {
        await loadList(1);
      } catch (error) {
        setStatus(searchStatus, error.message, "error");
      }
    });

    refreshBtn.addEventListener("click", async () => {
      setStatus(searchStatus, "Refreshing...");
      try {
        await loadList(searchMode ? 1 : currentPage);
      } catch (error) {
        setStatus(searchStatus, error.message, "error");
      }
    });

    resultBody.addEventListener("click", async (event) => {
      const target = event.target;
      const card = target.closest(".item");
      if (!card) {
        return;
      }
      const id = card.dataset.id;

      if (target.classList.contains("edit-btn")) {
        openEditorByID(id);
        return;
      }

      if (target.classList.contains("del-btn")) {
        if (!window.confirm("Delete this attachment and metadata?")) {
          return;
        }
        setStatus(searchStatus, "Deleting...");
        try {
          await deleteRowByID(id);
          await loadList(searchMode ? 1 : currentPage);
        } catch (error) {
          setStatus(searchStatus, error.message, "error");
        }
      }
    });

    modalCancelBtn.addEventListener("click", () => {
      closeEditor();
    });

    modalSaveBtn.addEventListener("click", async () => {
      setStatus(searchStatus, "Saving...");
      try {
        await saveEditorChanges();
        closeEditor();
        await loadList(searchMode ? 1 : currentPage);
      } catch (error) {
        setStatus(searchStatus, error.message, "error");
      }
    });

    prevPageBtn.addEventListener("click", async () => {
      if (searchMode || currentPage <= 1) {
        return;
      }
      setStatus(searchStatus, "Loading previous page...");
      try {
        await loadList(currentPage - 1);
      } catch (error) {
        setStatus(searchStatus, error.message, "error");
      }
    });

    nextPageBtn.addEventListener("click", async () => {
      if (searchMode) {
        return;
      }
      const totalPages = Math.max(1, Math.ceil(currentTotal / currentPageSize));
      if (currentPage >= totalPages) {
        return;
      }
      setStatus(searchStatus, "Loading next page...");
      try {
        await loadList(currentPage + 1);
      } catch (error) {
        setStatus(searchStatus, error.message, "error");
      }
    });

    editorModal.addEventListener("click", (event) => {
      if (event.target === editorModal) {
        closeEditor();
      }
    });

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape" && editorModal.classList.contains("show")) {
        closeEditor();
      }
    });

    function escapeHtml(value) {
      if (value === null || value === undefined) {
        return "";
      }
      value = String(value);
      return value
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;");
    }

    function escapeAttr(value) {
      return escapeHtml(value).replaceAll("'", "&#39;").replaceAll("\"", "&quot;");
    }

    function previewText(value) {
      if (!value || !value.trim()) {
        return "-";
      }
      return escapeHtml(String(value).replaceAll("\n", " ").trim());
    }

    function tooltipText(value) {
      if (!value || !String(value).trim()) {
        return "-";
      }
      return String(value).replaceAll("\n", " ").trim();
    }

    function rawText(value) {
      if (value === null || value === undefined) {
        return "";
      }
      return String(value).replaceAll("\n", " ").trim();
    }

    function sanitizeExternalURL(value) {
      const text = rawText(value);
      if (!text) {
        return "";
      }
      if (/^https?:\/\//i.test(text)) {
        return text;
      }
      return "";
    }

    function renderURLValue(value) {
      const text = rawText(value);
      if (!text) {
        return "<span class='snippet-value'>-</span>";
      }

      const safeText = escapeHtml(text);
      const fullText = escapeAttr(text);
      const href = sanitizeExternalURL(text);
      if (!href) {
        return "<span class='snippet-value js-overflow-tip' data-full='" + fullText + "'>" + safeText + "</span>";
      }

      return "<a class='snippet-value snippet-link js-overflow-tip' data-full='" + fullText + "' href='" + escapeAttr(href) + "' target='_blank' rel='noopener noreferrer'>" + safeText + "</a>";
    }

    function formatSizeMB(value) {
      const bytes = Number(value);
      if (!Number.isFinite(bytes) || bytes <= 0) {
        return "0.00 MB";
      }
      return (bytes / (1024 * 1024)).toFixed(2) + " MB";
    }

    function pad2(value) {
      return String(value).padStart(2, "0");
    }

    function formatUploadTime(value) {
      const text = rawText(value);
      if (!text) {
        return "-";
      }

      const date = new Date(text);
      if (Number.isNaN(date.getTime())) {
        return text;
      }

      return date.getFullYear() + "-" +
        pad2(date.getMonth() + 1) + "-" +
        pad2(date.getDate()) + " " +
        pad2(date.getHours()) + ":" +
        pad2(date.getMinutes()) + ":" +
        pad2(date.getSeconds());
    }

    function updateOverflowTips() {
      const nodes = resultBody.querySelectorAll(".js-overflow-tip");
      for (const node of nodes) {
        const full = node.getAttribute("data-full") || "";
        node.classList.remove("has-overflow-tip");
        node.removeAttribute("data-tip");
        if (node.scrollWidth > node.clientWidth + 1) {
          node.classList.add("has-overflow-tip");
          node.setAttribute("data-tip", full);
        }
      }
    }

    window.addEventListener("resize", () => {
      updateOverflowTips();
    });

    loadList().catch((error) => {
      setStatus(searchStatus, error.message, "error");
    });
  </script>
</body>
</html>
`
