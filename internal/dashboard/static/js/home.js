import Search from "./search.js";

export default {
  render: async (containerId, searchTerm = "") => {
    const resp = await fetch("/templates/home.html");
    const templateText = await resp.text();
    const template = Handlebars.compile(templateText);

    const sitesResp = await fetch("/api/sites");
    const sitesData = await sitesResp.json();

    let vhostCount = 0;
    const portCount = {};
    sitesData.forEach((site) => {
      portCount[site.port] = (portCount[site.port] || 0) + 1;
    });
    for (let port in portCount) {
      if (portCount[port] > 1) vhostCount++;
    }

    const logWeightResp = await fetch("/api/logweight");
    const logWeightData = await logWeightResp.json();

    const metricsResp = await fetch("/api/metrics");
    const metricsData = await metricsResp.json();

    const data = {
      sitesCount: sitesData.length,
      vhostsCount: vhostCount,
      totalLogWeight: logWeightData.log_weight_bytes,
      metrics: metricsData,
      searchTerm,
    };

    document.querySelector(containerId).innerHTML = template(data);

    window.currentView.applySearch = (searchTerm) => {
      Search.highlightText(
        ".text-lg.font-semibold.mb-2, .space-y-2.text-sm.text-gray-600 li, .bg-white.rounded-xl.shadow span",
        searchTerm
      );
    };
  },
};
