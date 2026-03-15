package api

const webAppHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>AttachmentHub</title>
  <style>
    :root {
      --bg: #f5f7fa;
      --card: #ffffff;
      --line: #dde3ea;
      --text: #1f2937;
      --muted: #607089;
      --accent: #0066cc;
      --danger: #c53030;
    }
    body {
      margin: 0;
      font-family: "Avenir Next", "Segoe UI", sans-serif;
      background: linear-gradient(180deg, #f7fbff 0%, #eef2f7 100%);
      color: var(--text);
    }
    .wrap {
      max-width: 1080px;
      margin: 24px auto 48px;
      padding: 0 16px;
    }
    .card {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 12px;
      padding: 16px;
      margin-bottom: 16px;
      box-shadow: 0 8px 20px rgba(17, 24, 39, 0.06);
    }
    h1 {
      margin: 0 0 16px;
      font-size: 24px;
    }
    form, .row {
      display: grid;
      grid-template-columns: repeat(12, minmax(0, 1fr));
      gap: 10px;
      align-items: center;
    }
    input[type="text"], input[type="file"], textarea {
      width: 100%;
      box-sizing: border-box;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 8px 10px;
      font-size: 14px;
      background: #fff;
    }
    textarea { min-height: 58px; resize: vertical; }
    button {
      border: 0;
      border-radius: 8px;
      padding: 8px 12px;
      font-size: 14px;
      cursor: pointer;
      background: var(--accent);
      color: #fff;
    }
    button.secondary { background: #526581; }
    button.danger { background: var(--danger); }
    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 13px;
    }
    th, td {
      border-top: 1px solid var(--line);
      padding: 8px;
      vertical-align: top;
    }
    th {
      text-align: left;
      color: var(--muted);
      font-weight: 600;
    }
    .muted { color: var(--muted); }
    .status {
      margin-top: 10px;
      min-height: 20px;
      color: var(--muted);
      font-size: 13px;
    }
    .mono {
      font-family: ui-monospace, "SFMono-Regular", Menlo, monospace;
      font-size: 12px;
    }
    .nowrap { white-space: nowrap; }
    @media (max-width: 860px) {
      .hide-mobile { display: none; }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="card">
      <h1>AttachmentHub</h1>
      <form id="import-form">
        <input style="grid-column: span 4;" name="file" type="file" required />
        <input style="grid-column: span 3;" name="url" type="text" placeholder="Optional URL" />
        <input style="grid-column: span 3;" name="note" type="text" placeholder="Optional note" />
        <button style="grid-column: span 2;" type="submit">Import</button>
      </form>
      <div id="import-status" class="status"></div>
    </div>

    <div class="card">
      <form id="search-form">
        <input style="grid-column: span 10;" id="keyword" type="text" placeholder="Search by URL or note" />
        <button style="grid-column: span 1;" type="submit">Search</button>
        <button style="grid-column: span 1;" type="button" class="secondary" id="refresh-btn">Refresh</button>
      </form>
      <div id="search-status" class="status"></div>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th class="hide-mobile">File</th>
            <th>URL</th>
            <th>Note</th>
            <th class="nowrap">Actions</th>
          </tr>
        </thead>
        <tbody id="result-body"></tbody>
      </table>
    </div>
  </div>

  <script>
    const importForm = document.getElementById("import-form");
    const importStatus = document.getElementById("import-status");
    const searchForm = document.getElementById("search-form");
    const searchStatus = document.getElementById("search-status");
    const resultBody = document.getElementById("result-body");
    const keywordInput = document.getElementById("keyword");
    const refreshBtn = document.getElementById("refresh-btn");

    async function loadList() {
      const keyword = keywordInput.value.trim();
      const query = keyword ? "?keyword=" + encodeURIComponent(keyword) : "";
      const response = await fetch("/api/v1/attachments" + query);
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "Search failed");
      }
      renderRows(data.items || []);
      searchStatus.textContent = "Loaded " + data.total + " attachment(s).";
    }

    function renderRows(items) {
      resultBody.innerHTML = "";
      if (!items.length) {
        const row = document.createElement("tr");
        row.innerHTML = "<td colspan='5' class='muted'>No attachments found.</td>";
        resultBody.appendChild(row);
        return;
      }

      for (const item of items) {
        const row = document.createElement("tr");
        row.dataset.id = item.id;
        row.innerHTML =
          "<td class='mono'>" + item.public_id + "</td>" +
          "<td class='hide-mobile'><div>" + escapeHtml(item.original_name) + "</div><div class='muted mono'>" + escapeHtml(item.content_type) + " / " + item.file_size + " B</div></td>" +
          "<td><input type='text' class='url-input' value='" + escapeAttr(item.url || "") + "'></td>" +
          "<td><input type='text' class='note-input' value='" + escapeAttr(item.note || "") + "'></td>" +
          "<td class='nowrap'>" +
          "<a href='/f/" + item.public_id + "' target='_blank'>Open</a> " +
          "<button class='secondary save-btn' type='button'>Save</button> " +
          "<button class='danger del-btn' type='button'>Delete</button></td>";
        resultBody.appendChild(row);
      }
    }

    async function updateRow(row) {
      const id = row.dataset.id;
      const url = row.querySelector(".url-input").value;
      const note = row.querySelector(".note-input").value;
      const response = await fetch("/api/v1/attachments/" + id, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url, note })
      });
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "Update failed");
      }
      searchStatus.textContent = "Updated " + data.public_id + ".";
    }

    async function deleteRow(row) {
      const id = row.dataset.id;
      const response = await fetch("/api/v1/attachments/" + id, { method: "DELETE" });
      if (!response.ok) {
        let message = "Delete failed";
        try {
          const data = await response.json();
          message = data.error || message;
        } catch (_) {}
        throw new Error(message);
      }
      row.remove();
      searchStatus.textContent = "Deleted attachment #" + id + ".";
    }

    importForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      importStatus.textContent = "Importing...";
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
        importStatus.textContent = "Imported " + data.public_id + " (" + data.original_name + ").";
        await loadList();
      } catch (error) {
        importStatus.textContent = error.message;
      }
    });

    searchForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      searchStatus.textContent = "Searching...";
      try {
        await loadList();
      } catch (error) {
        searchStatus.textContent = error.message;
      }
    });

    refreshBtn.addEventListener("click", async () => {
      searchStatus.textContent = "Refreshing...";
      try {
        await loadList();
      } catch (error) {
        searchStatus.textContent = error.message;
      }
    });

    resultBody.addEventListener("click", async (event) => {
      const target = event.target;
      const row = target.closest("tr");
      if (!row) {
        return;
      }
      if (target.classList.contains("save-btn")) {
        searchStatus.textContent = "Saving...";
        try {
          await updateRow(row);
        } catch (error) {
          searchStatus.textContent = error.message;
        }
      }
      if (target.classList.contains("del-btn")) {
        if (!window.confirm("Delete this attachment and metadata?")) {
          return;
        }
        searchStatus.textContent = "Deleting...";
        try {
          await deleteRow(row);
        } catch (error) {
          searchStatus.textContent = error.message;
        }
      }
    });

    function escapeHtml(value) {
      return value
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;");
    }

    function escapeAttr(value) {
      return escapeHtml(value).replaceAll("'", "&#39;").replaceAll("\"", "&quot;");
    }

    loadList().catch((error) => {
      searchStatus.textContent = error.message;
    });
  </script>
</body>
</html>
`
