import Toast from "./toasts.js";
import Search from "./search.js";

async function render(containerId, searchTerm = "") {
  const resp = await fetch("/templates/config.html");
  const templateText = await resp.text();
  const template = Handlebars.compile(templateText);

  const configResp = await fetch("/api/config");
  const configData = await configResp.json();

  const html = template({
    ...configData,
    searchTerm,
  });
  document.querySelector(containerId).innerHTML = html;

  document
    .getElementById("saveConfigButton")
    .addEventListener("click", async () => {
      const apiPort = document.getElementById("apiPort").value;
      const dashboardPort = document.getElementById("dashboardPort").value;
      const updatedConfig = {
        api_port: parseInt(apiPort),
        dashboard_port: parseInt(dashboardPort),
        enabled_plugins: configData.enabled_plugins,
      };
      const res = await fetch("/api/config", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updatedConfig),
      });
      const newConfig = await res.json();
      Toast.show("Config saved.", "info", 3000);
      render(containerId);
    });

  window.currentView.applySearch = (searchTerm) => {
    Search.highlightText(".block.text-sm.font-medium", searchTerm);
  };
}

export default { render };
