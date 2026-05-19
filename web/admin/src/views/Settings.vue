<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { request } from "../utils/request";
import { Save, Settings, Info, FileText, Monitor, Globe, Image, Upload, Shield, HardDrive } from "lucide-vue-next";
import CoverUploadField from "../components/CoverUploadField.vue";
import { uploadMedia } from "../utils/media";

const settings = reactive({
  siteName: "",
  siteSubtitle: "",
  siteKeywords: "",
  searchPlaceholder: "",
  siteDescription: "",
  aboutContent: "",
  heroTitle: "",
  heroDescription: "",
  logo: "",
  footerText: "",
  themeDefault: "",
  githubUrl: "",
  icp: "",
  defaultOgImage: "",
  baseUrl: "",
  previewSecret: "",
  maxUploadSize: 8 * 1024 * 1024,
  storageDriver: "",
  storageLocalPath: "",
  storageLocalBaseUrl: "",
  storageS3Endpoint: "",
  storageS3Region: "",
  storageS3Bucket: "",
  storageS3AccessKey: "",
  storageS3SecretKey: "",
  storageS3UseSsl: false,
  storageS3PublicUrl: "",
  storagePublicUrl: "",
  aiEnabled: false,
  aiProvider: "openai_compatible",
  aiModel: "",
  aiApiKey: "",
  aiBaseUrl: "https://api.openai.com/v1",
  aiTimeoutSeconds: 15,
});

const tabs = [
  { key: "site", label: "站点", icon: Settings },
  { key: "content", label: "内容", icon: FileText },
  { key: "seo", label: "SEO", icon: Info },
  { key: "storage", label: "存储", icon: HardDrive },
  { key: "ai", label: "AI", icon: Monitor },
  { key: "security", label: "安全", icon: Shield },
];

const isSaving = ref(false);
const isUploadingLogo = ref(false);
const isSavingAccount = ref(false);
const activeTab = ref("site");
const account = reactive({
  username: "",
  newPassword: "",
  confirmPassword: "",
});
const showSettingsSave = computed(() => activeTab.value !== "security");
const showAIBaseUrl = computed(() => settings.aiProvider === "openai_compatible");

const isValidTab = (tab) => tabs.some((item) => item.key === tab);

const syncTabFromHash = () => {
  const hash = window.location.hash.replace(/^#/, "");
  if (isValidTab(hash)) {
    activeTab.value = hash;
  }
};

const setActiveTab = (tab) => {
  if (isValidTab(tab)) {
    activeTab.value = tab;
  }
};

const handleLogoUpload = async (event) => {
  const file = event.target.files?.[0];
  if (!file) return;
  isUploadingLogo.value = true;
  try {
    const media = await uploadMedia(file);
    settings.logo = media.url || "";
  } catch (err) {
    alert(err.message || "上传失败");
  } finally {
    isUploadingLogo.value = false;
    event.target.value = "";
  }
};

const loadSettings = async () => {
  try {
    const data = await request("/api/admin/settings");
    Object.assign(settings, data);
    account.username = data.username || account.username;
  } catch (err) {
    console.error(err);
  }
};

const loadAccount = async () => {
  try {
    const data = await request("/api/admin/me");
    account.username = data.username || "";
  } catch (err) {
    console.error(err);
  }
};

const saveSettings = async () => {
  isSaving.value = true;
  try {
    const payload = {
      ...settings,
      maxUploadSize: Number(settings.maxUploadSize) || 0,
      aiTimeoutSeconds: Number(settings.aiTimeoutSeconds) || 0,
    };
    const data = await request("/api/admin/settings", { method: "POST", body: JSON.stringify(payload) });
    Object.assign(settings, data);
    alert("设置已保存");
  } catch (err) {
    alert(err.message || "保存失败");
  } finally {
    isSaving.value = false;
  }
};

const saveAccount = async () => {
  if (!account.username.trim()) {
    alert("用户名不能为空");
    return;
  }
  if (account.newPassword && account.newPassword !== account.confirmPassword) {
    alert("两次输入的密码不一致");
    return;
  }
  isSavingAccount.value = true;
  try {
    await request("/api/admin/account", {
      method: "POST",
      body: JSON.stringify(account),
    });
    account.newPassword = "";
    account.confirmPassword = "";
    alert("账号设置已保存");
  } catch (err) {
    alert(err.message || "账号设置保存失败");
  } finally {
    isSavingAccount.value = false;
  }
};

onMounted(() => {
  syncTabFromHash();
  loadSettings();
  loadAccount();
  window.addEventListener("hashchange", syncTabFromHash);
});

onBeforeUnmount(() => {
  window.removeEventListener("hashchange", syncTabFromHash);
});

watch(activeTab, (value) => {
  const nextHash = `#${value}`;
  if (window.location.hash !== nextHash) {
    window.history.replaceState(null, "", nextHash);
  }
});
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-6">
    <div class="flex items-center justify-between gap-4">
      <h1 class="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">站点设置</h1>
      <button
        v-if="showSettingsSave"
        @click="saveSettings"
        :disabled="isSaving"
        class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors shadow-sm"
      >
        <Save class="w-4 h-4 mr-2" :class="{ 'animate-spin': isSaving && false }" />
        {{ isSaving ? "保存中..." : "保存设置" }}
      </button>
    </div>

    <div class="bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden">
      <div class="border-b border-zinc-200 px-4 py-4 dark:border-zinc-800 md:px-6 xl:px-8">
        <div class="flex gap-2 overflow-x-auto pb-1">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="setActiveTab(tab.key)"
            :class="[
              'inline-flex shrink-0 items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors',
              activeTab === tab.key
                ? 'bg-blue-600 text-white shadow-sm'
                : 'bg-zinc-100 text-zinc-600 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-300 dark:hover:bg-zinc-700'
            ]"
          >
            <component :is="tab.icon" class="h-4 w-4" />
            {{ tab.label }}
          </button>
        </div>
      </div>

      <template v-if="activeTab === 'site'">
        <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Settings class="w-5 h-5 text-blue-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">基本信息</h2>
          </div>

          <div class="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">站点名称</label>
              <input
                v-model="settings.siteName"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="例如: MD Blog"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">站点副标题</label>
              <input
                v-model="settings.siteSubtitle"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="例如: 记录生活，分享知识"
              />
            </div>
            <div class="md:col-span-2 xl:col-span-1">
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">搜索占位文案</label>
              <input
                v-model="settings.searchPlaceholder"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="例如: 搜索文章..."
              />
            </div>
          </div>
        </div>

        <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Globe class="w-5 h-5 text-sky-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">站点运行</h2>
          </div>

          <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">站点地址</label>
              <input
                v-model="settings.baseUrl"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="https://example.com"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">预览密钥</label>
              <input
                v-model="settings.previewSecret"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="preview-secret"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">上传大小限制（字节）</label>
              <input
                v-model.number="settings.maxUploadSize"
                type="number"
                min="1"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="8388608"
              />
            </div>
          </div>
        </div>

        <div class="p-6 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Globe class="w-5 h-5 text-cyan-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">站点展示</h2>
          </div>

          <div class="space-y-6">
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">站点图标 (Favicon)</label>
              <p class="mb-3 text-xs text-zinc-500 dark:text-zinc-400">浏览器标签页图标，推荐正方形 PNG/SVG</p>
              <div class="flex w-full items-center gap-3 md:w-1/2">
                <input
                  v-model="settings.logo"
                  class="block flex-1 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="https://example.com/favicon.png"
                />
                <label class="inline-flex shrink-0 cursor-pointer items-center gap-1 rounded-lg bg-zinc-100 px-3 py-2 text-xs font-medium text-zinc-700 transition-colors hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700">
                  <Upload class="h-3.5 w-3.5" />
                  {{ isUploadingLogo ? "上传中..." : "上传" }}
                  <input type="file" accept="image/*" class="hidden" :disabled="isUploadingLogo" @change="handleLogoUpload" />
                </label>
                <img
                  v-if="settings.logo"
                  :src="settings.logo"
                  class="h-10 w-10 shrink-0 rounded border border-zinc-200 bg-zinc-50 object-contain dark:border-zinc-700 dark:bg-zinc-800"
                  alt="icon preview"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">默认主题</label>
                <select
                  v-model="settings.themeDefault"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                >
                  <option value="">跟随系统</option>
                  <option value="light">浅色</option>
                  <option value="dark">深色</option>
                </select>
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">页脚文字</label>
                <input
                  v-model="settings.footerText"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="例如: Built with Go"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">GitHub 地址</label>
                <input
                  v-model="settings.githubUrl"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="https://github.com/..."
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">ICP 备案号</label>
                <input
                  v-model="settings.icp"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="例如: 京ICP备XXXXXXXX号"
                />
              </div>
            </div>
          </div>
        </div>
      </template>

      <template v-else-if="activeTab === 'content'">
        <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Monitor class="w-5 h-5 text-amber-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">首页展示</h2>
          </div>

          <div class="space-y-6">
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">首屏标题</label>
              <input
                v-model="settings.heroTitle"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="例如: 你好，欢迎来到我的博客"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">首屏描述</label>
              <textarea
                v-model="settings.heroDescription"
                rows="2"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="展示在首页的简短描述"
              ></textarea>
            </div>
          </div>
        </div>

        <div class="p-6 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <FileText class="w-5 h-5 text-purple-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">关于页内容</h2>
          </div>

          <div>
            <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Markdown 支持</label>
            <textarea
              v-model="settings.aboutContent"
              rows="12"
              class="block min-h-[280px] w-full resize-y rounded-lg border border-zinc-200 bg-white px-3 py-2 font-mono text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
              placeholder="编写关于页面的内容..."
            ></textarea>
          </div>
        </div>
      </template>

      <template v-else-if="activeTab === 'seo'">
        <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Info class="w-5 h-5 text-emerald-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">SEO 设置</h2>
          </div>

          <div class="space-y-6">
            <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">站点关键词</label>
                <input
                  v-model="settings.siteKeywords"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="例如: 博客, 技术, 编程"
                />
              </div>
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">站点描述</label>
              <textarea
                v-model="settings.siteDescription"
                rows="3"
                class="block w-full resize-none rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="用于搜索引擎的描述信息"
              ></textarea>
            </div>
          </div>
        </div>

        <div class="p-6 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Image class="w-5 h-5 text-rose-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">SEO 高级</h2>
          </div>

          <div>
            <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">默认 OG 图片</label>
            <p class="mb-3 text-xs text-zinc-500 dark:text-zinc-400">社交分享时的默认封面图，文章未设置封面时使用</p>
            <CoverUploadField v-model="settings.defaultOgImage" />
          </div>
        </div>
      </template>

      <template v-else-if="activeTab === 'storage'">
        <div class="p-6 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <HardDrive class="w-5 h-5 text-violet-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">存储配置</h2>
          </div>

          <div class="space-y-6">
            <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">存储驱动</label>
                <select
                  v-model="settings.storageDriver"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                >
                  <option value="local">本地</option>
                  <option value="s3">S3</option>
                </select>
              </div>
            </div>

            <div v-if="settings.storageDriver === 'local'" class="grid grid-cols-1 gap-6 md:grid-cols-2">
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">本地相对路径</label>
                <input
                  v-model="settings.storageLocalPath"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="uploads"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">本地访问前缀</label>
                <input
                  v-model="settings.storageLocalBaseUrl"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="/uploads"
                />
              </div>
            </div>

            <div v-else class="grid grid-cols-1 gap-6 md:grid-cols-2">
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">S3 Endpoint</label>
                <input
                  v-model="settings.storageS3Endpoint"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="127.0.0.1:9000"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Region</label>
                <input
                  v-model="settings.storageS3Region"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="us-east-1"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Bucket</label>
                <input
                  v-model="settings.storageS3Bucket"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="md-blog"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Public URL</label>
                <input
                  v-model="settings.storageS3PublicUrl"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                  placeholder="https://cdn.example.com"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Access Key</label>
                <input
                  v-model="settings.storageS3AccessKey"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                />
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Secret Key</label>
                <input
                  v-model="settings.storageS3SecretKey"
                  class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                />
              </div>
              <label class="inline-flex items-center gap-3 text-sm font-medium text-zinc-700 dark:text-zinc-300">
                <input v-model="settings.storageS3UseSsl" type="checkbox" class="h-4 w-4 rounded border-zinc-300 text-blue-600 focus:ring-blue-500" />
                使用 SSL
              </label>
            </div>
          </div>
        </div>
      </template>

      <template v-else-if="activeTab === 'ai'">
        <div class="border-b border-zinc-200 p-6 dark:border-zinc-800 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Monitor class="w-5 h-5 text-sky-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">AI 自动生成</h2>
          </div>

          <div class="space-y-6">
            <label class="inline-flex items-center gap-3 text-sm font-medium text-zinc-700 dark:text-zinc-300">
              <input v-model="settings.aiEnabled" type="checkbox" class="h-4 w-4 rounded border-zinc-300 text-blue-600 focus:ring-blue-500" />
              启用文章摘要与 SEO 自动生成
            </label>

            <div class="rounded-xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-sm text-zinc-600 dark:border-zinc-800 dark:bg-zinc-950 dark:text-zinc-300">
              <p>OpenAI 兼容可用于 OpenAI、DeepSeek、Qwen 兼容网关、Kimi 兼容网关、SiliconFlow、OpenRouter、Groq、Ollama 等。</p>
              <p class="mt-2">Anthropic 对应 Claude，Gemini 对应 Google Gemini 官方接口。</p>
            </div>
          </div>
        </div>

        <div class="p-6 xl:p-8">
          <div class="mb-6 flex items-center gap-2">
            <Settings class="w-5 h-5 text-violet-500" />
            <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Provider 配置</h2>
          </div>

          <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Provider</label>
              <select
                v-model="settings.aiProvider"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
              >
                <option value="openai_compatible">OpenAI Compatible</option>
                <option value="anthropic">Anthropic</option>
                <option value="gemini">Gemini</option>
              </select>
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">模型名称</label>
              <input
                v-model="settings.aiModel"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="例如: gpt-4.1-mini / claude-sonnet-4-20250514 / gemini-2.5-flash"
              />
            </div>
            <div class="xl:col-span-2">
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">API Key</label>
              <input
                v-model="settings.aiApiKey"
                type="password"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="输入模型服务的访问密钥"
              />
            </div>
            <div v-if="showAIBaseUrl" class="xl:col-span-2">
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">Base URL</label>
              <input
                v-model="settings.aiBaseUrl"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="https://api.openai.com/v1"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">超时（秒）</label>
              <input
                v-model.number="settings.aiTimeoutSeconds"
                type="number"
                min="1"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="15"
              />
            </div>
          </div>
        </div>
      </template>

      <template v-else-if="activeTab === 'security'">
        <div class="p-6 xl:p-8">
          <div class="mb-6 flex items-center justify-between gap-4">
            <div class="flex items-center gap-2">
              <Shield class="w-5 h-5 text-emerald-500" />
              <h2 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">安全设置</h2>
            </div>
            <button
              @click="saveAccount"
              :disabled="isSavingAccount"
              class="inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {{ isSavingAccount ? "保存中..." : "保存账号" }}
            </button>
          </div>

          <div class="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">用户名</label>
              <input
                v-model="account.username"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">新密码</label>
              <input
                v-model="account.newPassword"
                type="password"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
                placeholder="不少于 8 位"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300">确认新密码</label>
              <input
                v-model="account.confirmPassword"
                type="password"
                class="block w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-900 outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
              />
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
