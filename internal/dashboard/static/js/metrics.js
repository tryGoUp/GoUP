import Search from "./search.js";

export default {
  render: async (containerId, searchTerm = "") => {
    const resp = await fetch("/templates/metrics.html");
    const templateText = await resp.text();
    const template = Handlebars.compile(templateText);

    const metricsResp = await fetch("/api/metrics");
    const metricsData = await metricsResp.json();

    const logWeightResp = await fetch("/api/logweight");
    const logWeightData = await logWeightResp.json();

    const pluginUsageResp = await fetch("/api/pluginusage");
    const pluginUsageData = await pluginUsageResp.json();

    const data = {
      metrics: [
        { name: "Requests Total", value: metricsData.requests_total },
        { name: "Average Latency (ms)", value: metricsData.latency_avg_ms },
        { name: "CPU Usage (%)", value: metricsData.cpu_usage },
        { name: "RAM Usage (MB)", value: metricsData.ram_usage_mb },
      ],
      logWeight: logWeightData.log_weight_bytes,
      pluginUsage: pluginUsageData,
      searchTerm,
    };

    document.querySelector(containerId).innerHTML = template(data);

    window.currentView.applySearch = (searchTerm) => {
      Search.highlightText(".border-b", searchTerm);
    };
  },
};
