let minimizedModalsByType = {};
let modalCounter = 0;

export default {
  show({
    id = "custom-modal",
    title = "",
    content = "",
    buttons = [],
    type = "default",
  }) {
    this.setNoSelect();
    if (document.getElementById(id)) {
      id = id + "-" + ++modalCounter;
    }
    let modal = document.createElement("div");
    modal.id = id;
    modal.dataset.type = type;
    modal.className =
      "fixed bottom-0 left-0 w-full bg-white shadow-xl rounded-t-xl transform translate-y-full transition-transform duration-300 z-50";
    modal.innerHTML = `
          <div class="cursor-pointer w-full bg-gray-300 h-1.5 rounded-full mt-2 mx-auto max-w-16"></div>
          <div class="mx-auto p-4 border-b">
            <div class="max-w-7xl mx-auto flex items-center justify-between">
              <h2 class="text-lg font-semibold" id="${id}-title">${title}</h2>
              <div id="${id}-buttons" class="space-x-2"></div>
            </div>
          </div>
          <div class="max-w-7xl mx-auto p-4 overflow-auto max-h-[60vh]" id="${id}-content">
            ${content}
          </div>
        `;
    document.body.appendChild(modal);
    this.updateButtons(id, buttons);
    this.enableDrag(id);
    setTimeout(() => {
      modal.style.transform = "translateY(0px)";
    }, 10);
  },

  hide(id = "custom-modal") {
    this.removeNoSelect();
    const modal = document.getElementById(id);
    if (modal) {
      modal.style.transform = "translateY(100%)";
      setTimeout(() => modal.remove(), 300);
      this._removeMinimized(modal);
    }
  },

  setTitle(id, newTitle) {
    const titleElem = document.getElementById(`${id}-title`);
    if (titleElem) titleElem.textContent = newTitle;
  },

  updateButtons(id, buttons) {
    const buttonsContainer = document.getElementById(`${id}-buttons`);
    if (!buttonsContainer) return;
    buttonsContainer.innerHTML = "";
    buttons.forEach(({ id, text, icon, onClick, hidden = false }) => {
      const btn = document.createElement("button");
      btn.id = `${id}`;
      if (icon) {
        btn.innerHTML = `<span class="material-symbols-outlined" aria-label="${text}">${icon}</span>`;
        btn.className = `px-4 py-1 rounded ${
          hidden ? "hidden" : ""
        } text-gray-700`;
      } else {
        btn.textContent = text;
        btn.className = `px-4 py-1 rounded ${
          hidden ? "hidden" : ""
        } bg-blue-500 text-white hover:bg-blue-600`;
      }
      btn.addEventListener("click", onClick);
      buttonsContainer.appendChild(btn);
    });
  },

  toggleButton(id, buttonId, show) {
    const button = document.getElementById(buttonId);
    if (button) button.classList.toggle("hidden", !show);
  },

  enableDrag(id) {
    let startY,
      isDragging = false;
    let modal = document.getElementById(id);
    let pill = modal.querySelector(".cursor-pointer");
    let initialTransform = 0;
    let currentTransform = 0;
    const anchorThreshold = 150;
    pill.addEventListener("mousedown", (e) => {
      isDragging = true;
      startY = e.clientY;
      const transform = getComputedStyle(modal).transform;
      if (transform !== "none") {
        const matrix = transform.match(/matrix.*\((.+)\)/)[1].split(", ");
        initialTransform = parseFloat(matrix[5]);
      } else {
        initialTransform = 0;
      }
      modal.style.transition = "none";
    });
    document.addEventListener("mousemove", (e) => {
      if (!isDragging) return;
      let diff = e.clientY - startY;
      let newTransform = initialTransform + diff;
      newTransform = Math.max(newTransform, 0);
      modal.style.transform = `translateY(${newTransform}px)`;
      currentTransform = newTransform;
    });
    document.addEventListener("mouseup", () => {
      if (!isDragging) return;
      isDragging = false;
      modal.style.transition = "all 0.3s ease";
      if (currentTransform >= anchorThreshold) {
        this.minimizeModal(modal);
      } else {
        modal.style.transform = "translateY(0px)";
      }
    });
  },

  minimizeModal(modal) {
    modal.style.transform = "translateY(0px)";
    modal.style.height = "74px";
    modal.style.width = "200px";
    modal.style.bottom = "10px";
    modal.style.zIndex = "0";
    let type = modal.dataset.type || "default";
    if (!minimizedModalsByType[type]) {
      minimizedModalsByType[type] = [];
    }
    const arr = minimizedModalsByType[type];
    const index = arr.length;
    arr.push(modal);
    const spacing = 10;
    const modalWidth = 200;
    let left = 10 + index * (modalWidth + spacing);
    const maxLeft = window.innerWidth - modalWidth - 10;
    if (left > maxLeft) left = maxLeft;
    modal.style.left = `${left}px`;
    modal.classList.add("rounded-xl");
    const title = modal.querySelector(".font-semibold");
    if (title) {
      title.classList.add("truncate", "text-sm");
    }
    const restoreHandler = (e) => {
      modal.style.transition = "all 0.3s ease";
      modal.style.height = "";
      modal.style.width = "100%";
      modal.style.left = "0";
      modal.style.zIndex = "50";
      modal.classList.remove("rounded-xl");
      if (title) {
        title.classList.remove("truncate", "text-sm");
      }
      modal.removeEventListener("click", restoreHandler);
      this._removeMinimized(modal);
    };
    modal.addEventListener("click", restoreHandler);
    modal._restoreHandler = restoreHandler;
  },

  _removeMinimized(modal) {
    let type = modal.dataset.type || "default";
    if (minimizedModalsByType[type]) {
      minimizedModalsByType[type] = minimizedModalsByType[type].filter(
        (m) => m !== modal
      );
      minimizedModalsByType[type].forEach((m, i) => {
        const spacing = 10;
        const modalWidth = 200;
        let left = 10 + i * (modalWidth + spacing);
        const maxLeft = window.innerWidth - modalWidth - 10;
        if (left > maxLeft) left = maxLeft;
        m.style.left = `${left}px`;
      });
    }
  },

  setNoSelect() {
    document.body.style.userSelect = "none";
  },

  removeNoSelect() {
    document.body.style.userSelect = "auto";
  },
};
