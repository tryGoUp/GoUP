import Config from "./js/config.js";
import Home from "./js/home.js";
import Metrics from "./js/metrics.js";
import Plugins from "./js/plugins.js";
import Sites from "./js/sites.js";
import Tools from "./js/tools.js";
import Search from "./js/search.js";

const router = new Navigo("/", { hash: false });

function render(viewModule) {
  window.currentView = {};
  viewModule.render("#app");
}

router
  .on({
    "/": () => render(Home),
    "/metrics": () => render(Metrics),
    "/plugins": () => render(Plugins),
    "/sites": () => render(Sites),
    "/config": () => render(Config),
    "/tools": () => render(Tools),
  })
  .resolve();

router.updatePageLinks();

Search.init();
