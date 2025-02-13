import Toast from "./toasts.js";
import Search from "./search.js";

export default {
  render: async (containerId, searchTerm = "") => {
    const resp = await fetch("/templates/plugins.html");
    const templateText = await resp.text();
    const template = Handlebars.compile(templateText);

    const pluginsResp = await fetch("/api/plugins");
    let pluginsData = await pluginsResp.json();

    const html = template({ plugins: pluginsData });
    document.querySelector(containerId).innerHTML = html;

    document.querySelectorAll(".toggle-plugin").forEach((checkbox) => {
      checkbox.addEventListener("change", async (e) => {
        const pluginName = e.target.dataset.plugin;
        await fetch(`/api/plugins/${pluginName}/toggle`, { method: "POST" });
        Toast.show(
          "Plugin modified. Server will restart shortly.",
          "info",
          3000
        );
        setTimeout(() => Toast.show("Server restarted."), 6000);
      });
    });

    window.currentView.applySearch = (searchTerm) => {
      Search.filterList(".divide-y li", searchTerm);
    };
  },
};
