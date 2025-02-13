export default {
  init() {
    const searchInput = document.getElementById("globalSearch");
    if (!searchInput) return;

    searchInput.addEventListener("input", (event) => {
      const searchTerm = event.target.value.toLowerCase().trim();
      if (
        window.currentView &&
        typeof window.currentView.applySearch === "function"
      ) {
        window.currentView.applySearch(searchTerm);
      }
    });
  },

  filterList(selector, searchTerm) {
    document.querySelectorAll(selector).forEach((item) => {
      const text = item.textContent.toLowerCase();
      item.style.opacity = text.includes(searchTerm) ? "1" : "0.4";
    });
  },

  highlightText(selector, searchTerm) {
    document.querySelectorAll(selector).forEach((element) => {
      const originalText =
        element.getAttribute("data-original-text") || element.textContent;
      element.setAttribute("data-original-text", originalText);
      element.innerHTML = searchTerm
        ? originalText.replace(
            new RegExp(`(${searchTerm})`, "gi"),
            "<mark class='bg-yellow-300'>$1</mark>"
          )
        : originalText;
    });
  },
};
