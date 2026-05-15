const themeOrder = ["system", "light", "dark"];
const darkMediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

function getStoredTheme() {
  return localStorage.getItem("site-theme") || "system";
}

function getEffectiveTheme(theme) {
  if (theme === "dark") {
    return "dark";
  }

  if (theme === "light") {
    return "light";
  }

  return darkMediaQuery.matches ? "dark" : "light";
}

function applyTheme(theme) {
  document.documentElement.setAttribute("data-theme", theme);
  const button = document.querySelector("[data-theme-toggle]");
  if (button) {
    const effectiveTheme = getEffectiveTheme(theme);
    const nextTheme = effectiveTheme === "dark" ? "light" : "dark";
    const actionLabel = nextTheme === "dark" ? "切换到深色模式" : "切换到浅色模式";

    button.dataset.theme = effectiveTheme;
    button.setAttribute("aria-label", actionLabel);
    button.setAttribute("title", actionLabel);
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
    const currentTheme = document.documentElement.getAttribute("data-theme") || theme;
    const nextTheme = getEffectiveTheme(currentTheme) === "dark" ? "light" : "dark";
    localStorage.setItem("site-theme", nextTheme);
    applyTheme(nextTheme);
  });

  darkMediaQuery.addEventListener("change", () => {
    const currentTheme = document.documentElement.getAttribute("data-theme") || "system";
    if (currentTheme === "system") {
      applyTheme("system");
    }
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

// 终端交互
class Terminal {
  constructor(element) {
    this.element = element;
    this.output = element.querySelector('[data-terminal-output]');
    this.input = element.querySelector('[data-terminal-input]');
    this.dataElement = element.querySelector('.terminal-data');
    this.history = [];
    this.historyIndex = -1;
    this.commands = new Map();
    this.init();
  }

  init() {
    this.registerCommands();
    this.input.addEventListener('keydown', (e) => this.handleKeydown(e));
    this.element.addEventListener('click', () => this.input.focus());
  }

  getData(key) {
    return this.dataElement?.dataset[key] || '';
  }

  registerCommands() {
    this.commands.set('help', {
      description: '显示所有可用命令',
      handler: () => this.showHelp()
    });

    this.commands.set('stats', {
      description: '显示站点统计',
      handler: () => this.getStats()
    });

    this.commands.set('articles', {
      description: '显示最新文章',
      handler: () => this.getArticles()
    });

    this.commands.set('categories', {
      description: '显示分类列表',
      handler: () => this.getCategories()
    });

    this.commands.set('tags', {
      description: '显示标签云',
      handler: () => this.getTags()
    });

    this.commands.set('about', {
      description: '关于博主',
      handler: () => this.getAbout()
    });

    this.commands.set('skills', {
      description: '技术栈展示',
      handler: () => this.getSkills()
    });

    this.commands.set('contact', {
      description: '联系方式',
      handler: () => this.getContact()
    });

    this.commands.set('theme', {
      description: '切换主题',
      handler: () => this.toggleTheme()
    });

    this.commands.set('clear', {
      description: '清屏',
      handler: () => this.clearScreen()
    });

    this.commands.set('date', {
      description: '显示当前日期时间',
      handler: () => this.getDate()
    });

    this.commands.set('whoami', {
      description: '显示当前用户',
      handler: () => 'guest@md-blog'
    });

    this.commands.set('hello', {
      description: '打个招呼',
      handler: () => 'Hello! 欢迎来到我的博客! 👋'
    });

    this.commands.set('ping', {
      description: '测试连通性',
      handler: () => 'PONG! 🏓'
    });

    this.commands.set('coffee', {
      description: '来杯咖啡',
      handler: () => '☕ 给你一杯咖啡，继续探索吧!'
    });

    this.commands.set('sudo', {
      description: '超级用户',
      handler: (args) => {
        if (args[0] === 'rm' && args[1] === '-rf' && args[2] === '/') {
          return '<span class="terminal-error">Permission denied: 你没有权限执行此操作 :)</span>';
        }
        return '<span class="terminal-error">Permission denied</span>';
      }
    });

    this.commands.set('matrix', {
      description: 'Matrix效果',
      handler: () => this.showMatrix()
    });
  }

  handleKeydown(e) {
    if (e.key === 'Enter') {
      e.preventDefault();
      this.execute(this.input.value);
      this.input.value = '';
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      this.navigateHistory(-1);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      this.navigateHistory(1);
    } else if (e.key === 'Tab') {
      e.preventDefault();
      this.autocomplete();
    }
  }

  navigateHistory(direction) {
    if (this.history.length === 0) return;

    this.historyIndex += direction;
    if (this.historyIndex < 0) this.historyIndex = 0;
    if (this.historyIndex >= this.history.length) {
      this.historyIndex = this.history.length;
      this.input.value = '';
      return;
    }

    this.input.value = this.history[this.historyIndex] || '';
  }

  autocomplete() {
    const input = this.input.value.trim().toLowerCase();
    if (!input) return;

    const matches = Array.from(this.commands.keys()).filter(cmd => cmd.startsWith(input));
    if (matches.length === 1) {
      this.input.value = matches[0] + ' ';
    } else if (matches.length > 1) {
      this.printLine(input, matches.join('  '), 'info');
    }
  }

  execute(input) {
    if (!input.trim()) return;

    this.history.push(input);
    this.historyIndex = this.history.length;

    const parts = input.trim().split(/\s+/);
    const cmd = parts[0].toLowerCase();
    const args = parts.slice(1);

    const command = this.commands.get(cmd);
    if (command) {
      const output = command.handler(args);
      this.printLine(input, output);
    } else {
      this.printLine(input, `command not found: ${cmd}`, 'error');
    }
  }

  printLine(input, output, type = 'output') {
    const line = document.createElement('div');
    line.className = 'terminal-line';

    const prompt = document.createElement('span');
    prompt.className = 'terminal-prompt';
    prompt.textContent = `~/ ${this.getData('siteName')} $`;

    const command = document.createElement('span');
    command.className = 'terminal-command';
    command.textContent = input;

    line.appendChild(prompt);
    line.appendChild(command);

    if (output) {
      const outputDiv = document.createElement('div');
      outputDiv.className = `terminal-${type}`;
      outputDiv.innerHTML = output;
      line.appendChild(outputDiv);
    }

    this.output.appendChild(line);
    this.scrollToBottom();
  }

  scrollToBottom() {
    this.output.scrollTop = this.output.scrollHeight;
  }

  showHelp() {
    const commands = Array.from(this.commands.entries())
      .map(([name, cmd]) => `<strong>${name.padEnd(12)}</strong> ${cmd.description}`)
      .join('<br>');
    return `<div class="terminal-output-text">${commands}</div>`;
  }

  getStats() {
    return `<div class="terminal-output-text">
📊 站点统计<br>
├── 文章数: ${this.getData('articlesCount')}<br>
├── 分类数: ${this.getData('categoriesCount')}<br>
├── 标签数: ${this.getData('tagsCount')}<br>
└── 归档数: ${this.getData('archivesCount')}
</div>`;
  }

  async getArticles() {
    this.printLine('articles', '<span class="terminal-info">加载中...</span>');
    try {
      const response = await fetch('/api/terminal/articles');
      const data = await response.json();
      if (data.code !== 0) {
        this.printLine('', data.message, 'error');
        return;
      }
      const articles = data.data.articles || [];
      if (articles.length === 0) {
        this.printLine('', '暂无文章', 'info');
        return;
      }
      const list = articles.map(a => `├── ${a.date}  ${a.title}`).join('<br>');
      this.printLine('', `<div class="terminal-output-text">📝 最新文章<br>${list}</div>`);
    } catch {
      this.printLine('', '获取文章列表失败', 'error');
    }
  }

  async getCategories() {
    this.printLine('categories', '<span class="terminal-info">加载中...</span>');
    try {
      const response = await fetch('/api/terminal/categories');
      const data = await response.json();
      if (data.code !== 0) {
        this.printLine('', data.message, 'error');
        return;
      }
      const categories = data.data.categories || [];
      if (categories.length === 0) {
        this.printLine('', '暂无分类', 'info');
        return;
      }
      const list = categories.map(c => `├── ${c.name}${c.description ? ' - ' + c.description : ''}`).join('<br>');
      this.printLine('', `<div class="terminal-output-text">📁 分类列表<br>${list}</div>`);
    } catch {
      this.printLine('', '获取分类列表失败', 'error');
    }
  }

  async getTags() {
    this.printLine('tags', '<span class="terminal-info">加载中...</span>');
    try {
      const response = await fetch('/api/terminal/tags');
      const data = await response.json();
      if (data.code !== 0) {
        this.printLine('', data.message, 'error');
        return;
      }
      const tags = data.data.tags || [];
      if (tags.length === 0) {
        this.printLine('', '暂无标签', 'info');
        return;
      }
      const list = tags.map(t => `#${t.name}`).join('  ');
      this.printLine('', `<div class="terminal-output-text">🏷️ 标签云<br>${list}</div>`);
    } catch {
      this.printLine('', '获取标签列表失败', 'error');
    }
  }

  getAbout() {
    return `<div class="terminal-output-text">
👨‍💻 关于博主<br>
<br>
这是一个使用 Go + SSR 构建的技术博客<br>
专注于分享编程技术和开发经验<br>
<br>
输入 skills 查看技术栈
</div>`;
  }

  getSkills() {
    return `<div class="terminal-output-text">
🛠️ 技术栈<br>
<br>
├── 后端: Go, Chi, GORM<br>
├── 前端: Vue 3, Vanilla JS, CSS<br>
├── 数据库: SQLite, MySQL<br>
└── 部署: Docker, Linux
</div>`;
  }

  getContact() {
    return `<div class="terminal-output-text">
📬 联系方式<br>
<br>
├── GitHub: 查看页面底部链接<br>
└── RSS: /rss.xml
</div>`;
  }

  toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    localStorage.setItem('site-theme', next);
    applyTheme(next);
    return `<span class="terminal-success">主题已切换为 ${next} 模式</span>`;
  }

  clearScreen() {
    this.output.innerHTML = '';
    return '';
  }

  getDate() {
    const now = new Date();
    return `<div class="terminal-output-text">${now.toLocaleString('zh-CN')}</div>`;
  }

  showMatrix() {
    const chars = '01アイウエオカキクケコ';
    let count = 0;
    const maxCount = 20;

    const interval = setInterval(() => {
      const line = document.createElement('div');
      line.className = 'terminal-line';
      let text = '';
      for (let i = 0; i < 40; i++) {
        text += chars[Math.floor(Math.random() * chars.length)];
      }
      line.innerHTML = `<span class="terminal-success">${text}</span>`;
      this.output.appendChild(line);
      this.scrollToBottom();
      count++;
      if (count >= maxCount) {
        clearInterval(interval);
        this.printLine('', '<span class="terminal-info">Matrix 效果结束</span>');
      }
    }, 100);

    return '<span class="terminal-info">启动 Matrix 效果...</span>';
  }
}

document.addEventListener("DOMContentLoaded", () => {
  initThemeToggle();
  initCodeBlocks();
  initToc();
  const terminalElement = document.querySelector('[data-terminal]');
  if (terminalElement) {
    new Terminal(terminalElement);
  }
});
