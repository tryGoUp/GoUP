export default {
  show(message, type = "info", duration = 3000) {
    let container = document.getElementById("toast-container");
    if (!container) {
      container = document.createElement("div");
      container.id = "toast-container";
      container.className = "fixed top-4 right-4 space-y-2 z-50";
      document.body.appendChild(container);
    }

    const toast = document.createElement("div");
    let bgColor = "bg-cyan-600";
    if (type === "error") {
      bgColor = "bg-red-600";
    } else if (type === "success") {
      bgColor = "bg-cyan-600";
    } else if (type === "warning") {
      bgColor = "bg-yellow-600";
    }

    toast.className = `p-4 rounded shadow text-white ${bgColor} opacity-100 transition-opacity duration-500`;
    toast.innerText = message;
    container.appendChild(toast);

    setTimeout(() => {
      toast.classList.add("opacity-0");
      setTimeout(() => {
        if (container.contains(toast)) {
          container.removeChild(toast);
        }
      }, 500);
    }, duration);
  },
};
