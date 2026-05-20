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
        button.style.transform = "scale(0.9)";
        window.setTimeout(() => { button.style.transform = ""; }, 150);
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

function initCategoryModal() {
  const modal = document.querySelector("[data-category-modal]");
  if (!modal) {
    return;
  }

  const openButton = document.querySelector("[data-category-modal-open]");
  const closeButtons = modal.querySelectorAll("[data-category-modal-close]");
  const panel = modal.querySelector(".overlay-modal__panel");

  const close = () => {
    modal.setAttribute("hidden", "");
    document.body.classList.remove("is-modal-open");
  };

  const open = () => {
    modal.removeAttribute("hidden");
    document.body.classList.add("is-modal-open");
    panel?.focus();
  };

  openButton?.addEventListener("click", open);
  closeButtons.forEach((button) => button.addEventListener("click", close));

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && !modal.hasAttribute("hidden")) {
      close();
    }
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
      const list = categories.map((c) => {
        const count = c.articleCount ? ` (${c.articleCount})` : '';
        return `├── ${c.name}${count}${c.description ? ' - ' + c.description : ''}`;
      }).join('<br>');
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
  initCategoryModal();
  initToc();
  const terminalElement = document.querySelector('[data-terminal]');
  if (terminalElement) {
    new Terminal(terminalElement);
  }
  const catStage = document.querySelector('[data-pixel-cat]');
  if (catStage) {
    new PixelCat(catStage);
  }
});

/* =====================
   Pixel Cat — 像素风小猫
   ===================== */
class PixelCat {
  constructor(stage) {
    this.stage = stage;
    this.canvas = stage.querySelector('.pixel-cat-canvas');
    this.bubble = stage.querySelector('[data-cat-bubble]');
    this.ctx = this.canvas.getContext('2d');

    // 逻辑像素尺寸（每个"像素"= SCALE 个 canvas 像素）
    this.SCALE = 4;
    this.CAT_W = 16; // 猫身逻辑宽
    this.CAT_H = 16; // 猫身逻辑高

    // 世界坐标（逻辑像素）
    this.x = 30;
    this.y = 0; // 将在 resize 后设置
    this.vx = 0;
    this.facing = 1; // 1=右 -1=左

    // 鼠标（相对 stage，逻辑像素）
    this.mouseX = -999;
    this.mouseY = -999;
    this.mouseInStage = false;

    // 状态机
    // states: idle | walk | sit | sleep | roll | run | scared
    this.state = 'idle';
    this.frame = 0;
    this.frameTick = 0;
    this.FRAME_DURATION = 8; // 每帧持续多少 tick

    // 气泡
    this.bubbleTimer = null;

    // idle/sleep 计时
    this.idleTick = 0;
    this.IDLE_TO_SIT = 180;   // ~3s → sit
    this.SIT_TO_SLEEP = 360;  // ~6s → sleep

    // roll 参数
    this.rollTick = 0;
    this.isRolling = false;

    // scared 计时
    this.scaredTick = 0;

    this.resize();
    this.bindEvents();
    this.loop();
  }

  /* ── 画布尺寸同步 ── */
  resize() {
    const rect = this.stage.getBoundingClientRect();
    this.canvas.width = rect.width;
    this.canvas.height = rect.height;
    this.stageW = rect.width / this.SCALE;
    this.stageH = rect.height / this.SCALE;
    this.groundY = this.stageH - this.CAT_H - 2;
    if (this.y <= 0) this.y = this.groundY;
    this.x = Math.min(this.x, this.stageW - this.CAT_W);
  }

  /* ── 事件绑定 ── */
  bindEvents() {
    this.stage.addEventListener('mousemove', (e) => {
      const rect = this.stage.getBoundingClientRect();
      this.mouseX = (e.clientX - rect.left) / this.SCALE;
      this.mouseY = (e.clientY - rect.top) / this.SCALE;
      this.mouseInStage = true;

      // 如果猫在睡觉，被惊醒
      if (this.state === 'sleep' || this.state === 'sit') {
        this.setState('scared');
        this.showBubble('！！');
      }
    });

    this.stage.addEventListener('mouseleave', () => {
      this.mouseInStage = false;
    });

    this.stage.addEventListener('mouseenter', () => {
      this.mouseInStage = true;
      if (this.state === 'sleep') {
        this.setState('scared');
        this.showBubble('！！');
      }
    });

    this.stage.addEventListener('click', () => {
      if (!this.isRolling) {
        this.setState('roll');
        const msgs = ['喵～', '(=｀ω´=)', '(=^･ω･^=)', '喵呜！', '(=^▽^=)'];
        this.showBubble(msgs[Math.floor(Math.random() * msgs.length)]);
      }
    });

    window.addEventListener('resize', () => this.resize());
  }

  /* ── 状态切换 ── */
  setState(s) {
    if (this.state === s) return;
    this.state = s;
    this.frame = 0;
    this.frameTick = 0;
    this.idleTick = 0;
    if (s === 'roll') {
      this.rollTick = 0;
      this.isRolling = true;
    }
    if (s === 'scared') {
      this.scaredTick = 0;
      // 往反方向跑
      this.facing = this.mouseX > this.x ? -1 : 1;
    }
  }

  /* ── 气泡 ── */
  showBubble(text) {
    clearTimeout(this.bubbleTimer);
    this.bubble.textContent = text;

    // 定位气泡到猫咪中心上方
    const catCenterX = (this.x + this.CAT_W / 2) * this.SCALE;
    const catTopY = this.y * this.SCALE;
    this.bubble.style.left = `${catCenterX}px`;
    this.bubble.style.bottom = `${this.canvas.height - catTopY + 4}px`;
    this.bubble.style.transform = 'translateX(-50%)';

    this.bubble.classList.add('is-visible');
    this.bubbleTimer = setTimeout(() => {
      this.bubble.classList.remove('is-visible');
    }, 1800);
  }

  /* ── 主循环 ── */
  loop() {
    this.update();
    this.draw();
    requestAnimationFrame(() => this.loop());
  }

  /* ── 逻辑更新 ── */
  update() {
    this.frameTick++;
    if (this.frameTick >= this.FRAME_DURATION) {
      this.frameTick = 0;
      this.frame++;
    }

    switch (this.state) {
      case 'idle': this.updateIdle(); break;
      case 'walk': this.updateWalk(); break;
      case 'sit':  this.updateSit();  break;
      case 'sleep':this.updateSleep();break;
      case 'roll': this.updateRoll(); break;
      case 'run':  this.updateRun();  break;
      case 'scared':this.updateScared();break;
    }

    // 边界钳制
    this.x = Math.max(0, Math.min(this.stageW - this.CAT_W, this.x));
  }

  updateIdle() {
    this.idleTick++;
    if (this.mouseInStage) {
      const dx = this.mouseX - (this.x + this.CAT_W / 2);
      if (Math.abs(dx) > this.CAT_W * 2) {
        this.facing = dx > 0 ? 1 : -1;
        this.setState('walk');
        return;
      }
    }
    if (this.idleTick > this.IDLE_TO_SIT) this.setState('sit');
  }

  updateWalk() {
    if (!this.mouseInStage) {
      this.setState('idle');
      return;
    }
    const dx = this.mouseX - (this.x + this.CAT_W / 2);
    if (Math.abs(dx) < this.CAT_W * 1.5) {
      this.setState('idle');
      this.showBubble('喵～');
      return;
    }
    this.facing = dx > 0 ? 1 : -1;
    this.x += this.facing * 1.2;
  }

  updateSit() {
    this.idleTick++;
    if (this.idleTick > this.SIT_TO_SLEEP) this.setState('sleep');
    if (this.mouseInStage) {
      const dx = this.mouseX - (this.x + this.CAT_W / 2);
      if (Math.abs(dx) > this.CAT_W * 3) {
        this.setState('walk');
      }
    }
  }

  updateSleep() {
    // 呼吸循环，什么都不做，等事件唤醒
  }

  updateRoll() {
    this.rollTick++;
    // 撒娇翻滚持续约60帧
    if (this.rollTick > 60) {
      this.isRolling = false;
      this.setState('idle');
    }
  }

  updateRun() {
    this.x += this.facing * 2.5;
    if (this.x <= 0 || this.x >= this.stageW - this.CAT_W) {
      this.facing *= -1;
      this.setState('idle');
    }
    if (!this.mouseInStage) this.setState('idle');
  }

  updateScared() {
    this.scaredTick++;
    this.x += this.facing * 2.8;
    if (this.scaredTick > 40 || this.x <= 0 || this.x >= this.stageW - this.CAT_W) {
      this.setState('idle');
    }
  }

  /* ── 绘制 ── */
  draw() {
    const ctx = this.ctx;
    const S = this.SCALE;
    ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

    const isDark = document.documentElement.getAttribute('data-theme') === 'dark' ||
      (document.documentElement.getAttribute('data-theme') === 'system' &&
        window.matchMedia('(prefers-color-scheme: dark)').matches);

    const colors = isDark
      ? { O: '#30363d', W: '#e6edf3', A: '#29b6f6', S: '#161b22', E: '#00e5ff', N: '#ff8a80', B: '#ffee58' }
      : { O: '#1c2833', W: '#ffffff', A: '#03a9f4', S: '#263238', E: '#00b0ff', N: '#ff5252', B: '#ffeb3b' };

    const frames = {
      idle: [
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWSEESSEESWO    ",
        "OWSEESSEESWO    ",
        "OWSSSNNSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWWWWWWWWWWO  OO",
        " OOOOOOOOOO  OAO",
        " OAAABBAAAO  OAO",
        " OWWWWWWWWO  OAO",
        " OWOWWOWWOO OAO ",
        "  OO OO OO OOO  "
      ],
      walk1: [
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWSEESSEESWO    ",
        "OWSEESSEESWO    ",
        "OWSSSNNSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWWWWWWWWWWO  OO",
        " OOOOOOOOOO  OAO",
        " OAAABBAAAO  OAO",
        " OWWWWWWWWO  OAO",
        " OWOWWOWWOO OAO ",
        "  OO OO OO OOO  "
      ],
      walk2: [
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWSEESSEESWO    ",
        "OWSEESSEESWO    ",
        "OWSSSNNSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWWWWWWWWWWO  OO",
        " OOOOOOOOOO  OAO",
        " OAAABBAAAO OAO ",
        " OWWWWWWWWO OAO ",
        " OWWOWWWOWOOOO  ",
        " OO  OO  OO     "
      ],
      sit: [
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWSEESSEESWO    ",
        "OWESSEESSEWO    ",
        "OWSSSNNSSSWO  OO",
        "OWSSSSSSSSWO OAO",
        "OWWWWWWWWWWOOAO ",
        " OOOOOOOOOO OAO ",
        " OAAABBAAAOOAO  ",
        " OWWWWWWWWOOO   ",
        " OWWWWWWWWOO    ",
        "  OOOOOOOOO     "
      ],
      sleep: [
        "                ",
        "                ",
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO  OO",
        "OWSEESSEESWO OAO",
        "OWSSSNNSSSWO OAO",
        "OWWWWWWWWWWOOAO ",
        " OOOOOOOOOO OAO ",
        " OAAABBAAAOOAO  ",
        " OWWWWWWWWWOO   ",
        "  OOOOOOOOOO    "
      ],
      scared1: [
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWEESSSSEEWO    ",
        "OWSEESSEESWO    ",
        "OWSSSNNSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWWWWWWWWWWO OO ",
        " OOOOOOOOOO OAO ",
        " OAAABBAAAOOAO  ",
        " OWWWWWWWWWOO   ",
        " OWWOWWWOWOO    ",
        "  OO  OO  O     "
      ],
      scared2: [
        " OO      OO     ",
        "OAAO    OAAO    ",
        "OAAO    OAAO    ",
        "OWWOOOOOOWWO    ",
        "OWSSSSSSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWEESSSSEEWO    ",
        "OWSEESSEESWO    ",
        "OWSSSNNSSSWO    ",
        "OWSSSSSSSSWO    ",
        "OWWWWWWWWWWOOO  ",
        " OOOOOOOOOO OAO ",
        " OAAABBAAAO OAO ",
        " OWWWWWWWWO  OO ",
        " OWOWWOWWOO     ",
        " OO OO OO       "
      ]
    };

    const f = this.frame % 2;
    const isSleep = this.state === 'sleep';
    const isRoll = this.state === 'roll';
    
    let currentFrame = frames.idle;
    if (this.state === 'walk' || this.state === 'run') {
      currentFrame = f === 0 ? frames.walk1 : frames.walk2;
    } else if (this.state === 'sit') {
      currentFrame = frames.sit;
    } else if (isSleep) {
      currentFrame = frames.sleep;
    } else if (this.state === 'scared') {
      currentFrame = f === 0 ? frames.scared1 : frames.scared2;
    } else if (isRoll) {
      currentFrame = frames.sit;
    }

    ctx.save();
    
    const bx = this.x;
    const by = this.y;

    if (isRoll) {
      const cx = (bx + this.CAT_W / 2) * S;
      const cy = (by + this.CAT_H / 2) * S;
      ctx.translate(cx, cy);
      const angle = Math.sin(this.rollTick * 0.2) * Math.PI * 0.5;
      ctx.rotate(angle);
      ctx.translate(-cx, -cy);
    }

    if (this.facing === 1) {
      ctx.translate((this.x + this.CAT_W) * S, 0);
      ctx.scale(-1, 1);
      ctx.translate(-this.x * S, 0);
    }

    // Blink logic
    this.tick = (this.tick || 0) + 1;
    const isBlinking = (this.state === 'idle' || this.state === 'walk' || this.state === 'sit') && (this.tick % 120 < 6);

    for (let y = 0; y < 16; y++) {
      const row = currentFrame[y];
      for (let x = 0; x < 16; x++) {
        const char = row[x];
        if (char !== ' ') {
          let color = colors[char];
          if (isBlinking && char === 'E' && this.state !== 'scared') {
            color = colors.S;
          }
          ctx.fillStyle = color;
          ctx.fillRect(
            Math.round((bx + x) * S),
            Math.round((by + y) * S),
            S, S
          );
        }
      }
    }

    ctx.restore();

    // Sleep zZ
    if (isSleep && this.frame % 4 < 2) {
      ctx.fillStyle = colors.A;
      ctx.font = `bold ${S * 2.5}px monospace`;
      let zx1 = bx + 14, zx2 = bx + 16;
      if (this.facing === 1) {
        zx1 = bx - 1;
        zx2 = bx - 3;
      }
      ctx.fillText('z', zx1 * S, (by + 4) * S);
      if (this.frame % 8 < 4) {
        ctx.fillText('Z', zx2 * S, (by + 1) * S);
      }
    }

    // 地面阴影
    ctx.save();
    ctx.fillStyle = isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.06)';
    ctx.beginPath();
    ctx.ellipse(
      (this.x + this.CAT_W / 2) * S,
      (this.groundY + this.CAT_H + 1) * S,
      this.CAT_W * S * 0.45,
      1.5 * S,
      0, 0, Math.PI * 2
    );
    ctx.fill();
    ctx.restore();

    // 更新气泡位置
    const catCenterX = (this.x + this.CAT_W / 2) * S;
    const catTopY = this.y * S;
    this.bubble.style.left = `${catCenterX}px`;
    this.bubble.style.bottom = `${this.canvas.height - catTopY + 4}px`;
  }
}
