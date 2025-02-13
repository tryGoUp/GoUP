export default {
  show({ id = "custom-modal", title = "", content = "", buttons = [] }) {
    this.setNoSelect();

    let modal = document.getElementById(id);
    if (!modal) {
      modal = document.createElement("div");
      modal.id = id;
      modal.className =
        "fixed bottom-0 left-0 w-full bg-white shadow-xl rounded-t-xl transform translate-y-full transition-transform duration-300 z-50";
      modal.innerHTML = `
          <div class="cursor-pointer w-full bg-gray-300 h-1.5 rounded-full mt-2 mx-auto max-w-16"></div>
          <div class="p-4 flex items-center justify-between border-b">
            <h2 class="text-lg font-semibold" id="${id}-title">${title}</h2>
            <div id="${id}-buttons" class="space-x-2"></div>
          </div>
          <div class="p-4 overflow-auto max-h-[60vh]" id="${id}-content">
            ${content}
          </div>
        `;
      document.body.appendChild(modal);

      this.updateButtons(id, buttons);

      this.enableDrag(id);

      setTimeout(() => {
        modal.style.transform = "translateY(0)";
      }, 10);
    }
  },

  hide(id = "custom-modal") {
    this.removeNoSelect();

    const modal = document.getElementById(id);
    if (modal) {
      modal.style.transform = "translateY(100%)";
      setTimeout(() => modal.remove(), 300);
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
    buttons.forEach(({ id, text, onClick, hidden = false }) => {
      const btn = document.createElement("button");
      btn.id = `${id}`;
      btn.textContent = text;
      btn.className = `px-4 py-1 rounded ${
        hidden ? "hidden" : ""
      } bg-blue-500 text-white hover:bg-blue-600`;
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

    let initialTransform = 100;
    let currentTransform = initialTransform;

    pill.addEventListener("mousedown", (e) => {
      isDragging = true;
      startY = e.clientY;
      initialTransform =
        parseFloat(getComputedStyle(modal).transform.split(",")[5]) || 0;
      modal.style.transition = "none";
    });

    document.addEventListener("mousemove", (e) => {
      if (!isDragging) return;
      let diff = e.clientY - startY;
      let newTransform = Math.min(Math.max(initialTransform + diff, 0), 100);

      modal.style.transform = `translateY(${newTransform}%)`;
      currentTransform = newTransform;
    });

    document.addEventListener("mouseup", () => {
      if (!isDragging) return;
      isDragging = false;
      modal.style.transition = "transform 0.3s ease";

      if (currentTransform > 70) {
        modal.style.transform = "translateY(100%)";
        setTimeout(() => modal.remove(), 300);
      } else {
        modal.style.transform = "translateY(0)";
      }
    });
  },

  setNoSelect() {
    document.body.style.userSelect = "none";
  },

  removeNoSelect() {
    document.body.style.userSelect = "auto";
  },
};
