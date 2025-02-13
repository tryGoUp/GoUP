import Search from "./search.js";

export default {
  render: async (containerId, searchTerm = "") => {
    const resp = await fetch("/templates/tools.html");
    const templateText = await resp.text();
    const template = Handlebars.compile(templateText);

    const html = template({ searchTerm });
    document.querySelector(containerId).innerHTML = html;

    document
      .getElementById("cleanupLogsBtn")
      .addEventListener("click", async () => {
        if (confirm("Sei sicuro di voler fare il backup e pulire i log?")) {
          const res = await fetch("/api/tools/cleanuplogs", { method: "POST" });
          const result = await res.json();
          alert(result.message);
        }
      });

    document
      .getElementById("physicalRestartBtn")
      .addEventListener("click", async () => {
        if (confirm("Riavvia fisicamente il server?")) {
          const res = await fetch("/api/restart", { method: "POST" });
          alert("Server in riavvio...");
          setTimeout(() => window.location.reload(), 6000);
        }
      });
  },
};
