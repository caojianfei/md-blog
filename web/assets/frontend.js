const themeOrder = ["system", "light", "dark"];
const themeLabels = {
  system: "theme: auto",
  light: "theme: light",
  dark: "theme: dark",
};

function getStoredTheme() {
  return localStorage.getItem("site-theme") || "system";
}

function applyTheme(theme) {
  document.documentElement.setAttribute("data-theme", theme);
  const label = document.querySelector("[data-theme-toggle-label]");
  if (label) {
    label.textContent = themeLabels[theme] || themeLabels.system;
  }
}

function initThemeToggle() {
  const button = document.querySelector("[data-theme-toggle]");
  if (!button) {
    return;
  }

  let theme = getStoredTheme();
  if (!themeOrder.includes(theme)) {
    theme = button.dataset.themeDefault || "system";
  }

  applyTheme(theme);

  button.addEventListener("click", () => {
    const currentIndex = themeOrder.indexOf(document.documentElement.getAttribute("data-theme"));
    const nextTheme = themeOrder[(currentIndex + 1 + themeOrder.length) % themeOrder.length];
    localStorage.setItem("site-theme", nextTheme);
    applyTheme(nextTheme);
  });
}

function initCodeBlocks() {
  const buttons = document.querySelectorAll("[data-copy-code]");
  buttons.forEach((button) => {
    button.addEventListener("click", async () => {
      const code = button.closest(".code-block")?.querySelector(".code-block__body code");
      if (!code) {
        return;
      }

      try {
        await navigator.clipboard.writeText(code.innerText);
        const original = button.textContent;
        button.textContent = "已复制";
        window.setTimeout(() => {
          button.textContent = original;
        }, 1600);
      } catch (error) {
        button.textContent = "复制失败";
        window.setTimeout(() => {
          button.textContent = "复制";
        }, 1600);
      }
    });
  });
}

function initToc() {
  const toc = document.querySelector("[data-toc]");
  if (!toc) {
    return;
  }

  const toggle = document.querySelector("[data-toc-toggle]");
  if (toggle) {
    toggle.addEventListener("click", () => {
      toc.toggleAttribute("data-open");
    });
  }

  const links = Array.from(toc.querySelectorAll('a[href^="#"]'));
  const headings = links
    .map((link) => {
      const target = document.querySelector(link.getAttribute("href"));
      if (!target) {
        return null;
      }
      return { link, target };
    })
    .filter(Boolean);

  if (!headings.length) {
    return;
  }

  const activate = (id) => {
    links.forEach((link) => {
      const isActive = link.getAttribute("href") === `#${id}`;
      link.classList.toggle("is-active", isActive);
      if (isActive) {
        toc.querySelectorAll("[data-toc-item]").forEach((item) => {
          item.classList.toggle("is-active", item.contains(link));
        });
      }
    });
  };

  const observer = new IntersectionObserver(
    (entries) => {
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((left, right) => left.boundingClientRect.top - right.boundingClientRect.top);
      if (visible[0]?.target?.id) {
        activate(visible[0].target.id);
      }
    },
    {
      rootMargin: "-20% 0px -65% 0px",
      threshold: [0, 1],
    }
  );

  headings.forEach(({ target }) => observer.observe(target));
  activate(headings[0].target.id);
}

document.addEventListener("DOMContentLoaded", () => {
  initThemeToggle();
  initCodeBlocks();
  initToc();
});
