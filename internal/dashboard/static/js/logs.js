import Search from "./search.js";
import Modal from "./modal.js";

export default {
  render: async (containerId, searchTerm = "") => {
    const resp = await fetch("/templates/logs.html");
    const templateText = await resp.text();
    const template = Handlebars.compile(templateText);
    document.querySelector(containerId).innerHTML = template();

    document
      .getElementById("fetchLogsBtn")
      .addEventListener("click", async () => {
        const start = document.getElementById("startDate").value;
        const end = document.getElementById("endDate").value;
        const plugin = document.getElementById("pluginFilter").value;

        let url = "/api/logfiles";
        const params = [];
        if (start) params.push(`start=${start}`);
        if (end) params.push(`end=${end}`);
        if (plugin) params.push(`plugin=${plugin}`);
        if (params.length > 0) {
          url += "?" + params.join("&");
        }

        const res = await fetch(url);
        const files = await res.json();

        const logsList = document.getElementById("logsList");
        logsList.innerHTML = "";

        files.forEach((f) => {
          const li = document.createElement("li");
          li.className =
            "px-4 py-3 hover:bg-gray-50 flex justify-between items-center";

          li.innerHTML = `
            <div>
              <p class="font-medium">${f.domain}</p>
              <p class="text-sm text-gray-400">
                ${f.file_name} | ${f.size_bytes} bytes
              </p>
            </div>
            ${
              f.size_bytes > 0
                ? `<button
                    class="text-blue-600 hover:text-blue-500"
                    data-file="${encodeURIComponent(f.file_name)}">
                    View
                  </button>`
                : `<span class="text-gray-400">Empty</span>`
            }
          `;

          const btn = li.querySelector("button");
          if (btn) {
            btn.addEventListener("click", async (e) => {
              const file = e.target.getAttribute("data-file");
              const contentResp = await fetch(`/api/logfiles/${file}`);
              const rawText = await contentResp.text();
              showLogContent(f.file_name, rawText);
            });
          }

          logsList.appendChild(li);
        });
      });

    function showLogContent(fileName, raw) {
      try {
        let fix = raw.trim();
        fix = fix.replace(/}\s*{/g, "},{");
        fix = `[${fix}]`;
        const parsed = JSON.parse(fix);

        let html = `<div class="flex flex-col space-y-2 text-sm">`;
        for (const entry of parsed) {
          const { level, time, ...rest } = entry;
          let mainMessage = "";
          if (typeof rest.message === "string") {
            mainMessage = rest.message;
            delete rest.message;
          }
          let body = "";
          const keys = Object.keys(rest);
          if (keys.length > 0) {
            body += `<div class="mt-1 text-xs text-gray-600">`;
            for (const k of keys) {
              body += `<div><strong>${k}:</strong> ${JSON.stringify(
                rest[k]
              )}</div>`;
            }
            body += `</div>`;
          }
          html += `
            <div class="bg-gray-50 border border-gray-200 rounded p-2 shadow-sm">
              <div class="text-xs text-gray-500 mb-1">
                <span class="font-semibold uppercase">${level || "info"}</span>
                @
                <span>${time || ""}</span>
              </div>
              <div class="text-gray-800 whitespace-pre-line">${mainMessage}</div>
              ${body}
            </div>
          `;
        }
        html += `</div>`;

        Modal.show({
          id: "logModal",
          title: fileName,
          content: html,
          buttons: [
            {
              id: "closeBtn",
              text: "Close",
              icon: "close",
              onClick: () => Modal.hide("logModal"),
            },
          ],
        });
      } catch {
        showRawModal(fileName, raw);
      }
    }

    function showRawModal(fileName, raw) {
      const safe = raw.replace(/</g, "&lt;").replace(/>/g, "&gt;");
      Modal.show({
        id: "logModal",
        title: fileName,
        content: `<pre class="whitespace-pre-wrap">${safe}</pre>`,
        buttons: [
          {
            id: "closeBtn",
            text: "Close",
            onClick: () => Modal.hide("logModal"),
          },
        ],
      });
    }

    window.currentView.applySearch = (term) => {
      Search.filterList("#logsList li", term);
    };
  },
};
