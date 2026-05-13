<script setup>
import { ref, reactive, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { request } from "../utils/request";
import MarkdownIt from "markdown-it";
import hljs from "highlight.js";
import "highlight.js/styles/github-dark.css";
import { 
  ArrowLeft,
  Save,
  Send,
  Image as ImageIcon,
  Link,
  Code,
  Bold,
  Heading1,
  Heading2,
  Columns,
  Eye,
  PenLine,
  ChevronDown,
  ChevronUp
} from "lucide-vue-next";

const route = useRoute();
const router = useRouter();

const md = new MarkdownIt({
  html: false,
  linkify: true,
  highlight(str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return `<pre><code class="hljs">${hljs.highlight(str, { language: lang }).value}</code></pre>`;
    }
    return `<pre><code class="hljs">${md.utils.escapeHtml(str)}</code></pre>`;
  },
});

const editor = reactive({
  id: 0, title: "", slug: "", excerpt: "", content: "", coverImage: "",
  categoryId: null, tagIds: [], status: "draft", seoDescription: "", seoKeywords: "",
});
const categories = ref([]);
const tags = ref([]);
const previewMode = ref("split"); // edit, split, preview
const editorMessage = ref("");
const isSettingsOpen = ref(true); // Toggle meta settings

const isEdit = computed(() => !!route.params.id);
const previewHtml = computed(() => md.render(editor.content || ""));
const editorWords = computed(() => (editor.content || "").trim().split(/\s+/).filter(Boolean).length);

const loadOptions = async () => {
  try {
    const [categoryData, tagData] = await Promise.all([
      request("/api/admin/categories"),
      request("/api/admin/tags"),
    ]);
    categories.value = categoryData || [];
    tags.value = tagData || [];
  } catch (err) {
    console.error(err);
  }
};

const loadArticle = async () => {
  if (!isEdit.value) return;
  try {
    const data = await request(`/api/admin/articles/${route.params.id}`);
    Object.assign(editor, {
      id: data.id, title: data.title, slug: data.slug, excerpt: data.excerpt, content: data.content,
      coverImage: data.coverImage, categoryId: data.categoryId, tagIds: (data.tags || []).map((x) => x.id),
      status: data.status, seoDescription: data.seoDescription, seoKeywords: data.seoKeywords,
    });
  } catch (err) {
    alert(err.message || "加载文章失败");
    router.push("/articles");
  }
};

const saveArticle = async (status = editor.status) => {
  const normalizeID = (value) => {
    if (value === null || value === undefined || value === "") return null;
    const parsed = Number(value);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
  };
  const payload = {
    ...editor,
    id: Number(editor.id) || 0,
    categoryId: normalizeID(editor.categoryId),
    tagIds: (editor.tagIds || []).map((item) => Number(item)).filter((item) => Number.isFinite(item) && item > 0),
    status,
  };
  try {
    const data = await request("/api/admin/articles", { method: "POST", body: JSON.stringify(payload) });
    Object.assign(editor, {
      id: data.id, title: data.title, slug: data.slug, excerpt: data.excerpt, content: data.content,
      coverImage: data.coverImage, categoryId: data.categoryId, tagIds: (data.tags || []).map((x) => x.id),
      status: data.status, seoDescription: data.seoDescription, seoKeywords: data.seoKeywords,
    });
    alert(status === "published" ? "文章已发布" : "草稿已保存");
    if (!isEdit.value) {
      router.replace(`/articles/edit/${data.id}`);
    }
  } catch (err) {
    alert(err.message || "保存失败");
  }
};

const insertText = (text) => { editor.content = `${editor.content || ""}\n${text}`; };
const insertLink = () => {
  const t = window.prompt("链接文本", "点击这里") || "链接";
  const u = window.prompt("链接地址", "https://") || "https://";
  insertText(`[${t}](${u})`);
};
const uploadImage = async (event) => {
  const file = event.target.files?.[0];
  if (!file) return;
  const form = new FormData();
  form.append("file", file);
  editorMessage.value = "图片上传中...";
  try {
    const response = await fetch("/api/admin/media/upload", { method: "POST", credentials: "include", body: form });
    const data = await response.json();
    if (data.code !== 0) throw new Error(data.message || "上传失败");
    insertText(`![${file.name}](${data.data.url})`);
    editorMessage.value = "图片已插入";
    setTimeout(() => { editorMessage.value = ""; }, 3000);
  } catch (err) {
    editorMessage.value = err.message || "上传失败";
  } finally {
    event.target.value = "";
  }
};

watch(() => editor.content, (value) => {
  localStorage.setItem(`draft:${editor.id || "new"}`, value || "");
});

onMounted(async () => {
  await loadOptions();
  await loadArticle();
  
  if (!isEdit.value) {
    const draft = localStorage.getItem("draft:new");
    if (draft && !editor.content) {
      editor.content = draft;
    }
  }
  
  // Responsive: collapse settings on small screens
  if (window.innerWidth < 1024) {
    isSettingsOpen.value = false;
    previewMode.value = "edit"; // Default to edit only on mobile
  }
});
</script>

<template>
  <div class="h-full flex flex-col gap-4">
    <!-- Header Actions -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white dark:bg-zinc-900 p-4 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm shrink-0">
      <div class="flex items-center gap-3">
        <button 
          @click="router.push('/articles')"
          class="p-2 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition-colors"
          title="返回"
        >
          <ArrowLeft class="w-5 h-5" />
        </button>
        <h1 class="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100 truncate max-w-[200px] sm:max-w-md">
          {{ isEdit ? '编辑文章' : '新建文章' }}
        </h1>
        <span 
          v-if="editor.status"
          class="hidden sm:inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
          :class="editor.status === 'published' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/10 dark:text-emerald-400' : 'bg-amber-100 text-amber-800 dark:bg-amber-500/10 dark:text-amber-400'"
        >
          {{ editor.status === 'published' ? '已发布' : '草稿' }}
        </span>
      </div>
      
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <button 
          @click="saveArticle('draft')"
          class="flex-1 sm:flex-none inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-200 bg-zinc-100 dark:bg-zinc-800 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded-lg transition-colors"
        >
          <Save class="w-4 h-4 mr-2" />
          保存草稿
        </button>
        <button 
          @click="saveArticle('published')"
          class="flex-1 sm:flex-none inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors shadow-sm"
        >
          <Send class="w-4 h-4 mr-2" />
          发布
        </button>
      </div>
    </div>

    <!-- Main Workspace -->
    <div class="flex-1 flex flex-col lg:flex-row gap-4 min-h-0">
      
      <!-- Left: Editor & Preview -->
      <div class="flex-1 flex flex-col bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden min-h-[500px] lg:min-h-0">
        
        <!-- Editor Toolbar -->
        <div class="flex items-center justify-between p-2 border-b border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-800/50 overflow-x-auto shrink-0">
          <div class="flex items-center gap-1 shrink-0">
            <button @click="insertText('# ')" class="p-1.5 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded transition-colors" title="H1"><Heading1 class="w-4 h-4" /></button>
            <button @click="insertText('## ')" class="p-1.5 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded transition-colors" title="H2"><Heading2 class="w-4 h-4" /></button>
            <div class="w-px h-4 bg-zinc-300 dark:bg-zinc-700 mx-1"></div>
            <button @click="insertText('**加粗**')" class="p-1.5 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded transition-colors" title="加粗"><Bold class="w-4 h-4" /></button>
            <button @click="insertText('`代码`')" class="p-1.5 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded transition-colors" title="行内代码"><Code class="w-4 h-4" /></button>
            <button @click="insertLink" class="p-1.5 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded transition-colors" title="链接"><Link class="w-4 h-4" /></button>
            <label class="p-1.5 text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 hover:bg-zinc-200 dark:hover:bg-zinc-700 rounded transition-colors cursor-pointer" title="上传图片">
              <ImageIcon class="w-4 h-4" />
              <input type="file" accept="image/*" class="hidden" @change="uploadImage" />
            </label>
            <span v-if="editorMessage" class="ml-2 text-xs text-blue-500">{{ editorMessage }}</span>
          </div>

          <div class="flex items-center gap-1 bg-zinc-200 dark:bg-zinc-700/50 p-1 rounded-lg shrink-0 ml-4">
            <button 
              @click="previewMode='edit'" 
              :class="['p-1.5 rounded-md transition-colors', previewMode==='edit' ? 'bg-white dark:bg-zinc-600 text-zinc-900 dark:text-zinc-100 shadow-sm' : 'text-zinc-500 dark:text-zinc-400']"
              title="仅编辑"
            ><PenLine class="w-4 h-4" /></button>
            <button 
              @click="previewMode='split'" 
              :class="['p-1.5 rounded-md transition-colors hidden sm:block', previewMode==='split' ? 'bg-white dark:bg-zinc-600 text-zinc-900 dark:text-zinc-100 shadow-sm' : 'text-zinc-500 dark:text-zinc-400']"
              title="分屏"
            ><Columns class="w-4 h-4" /></button>
            <button 
              @click="previewMode='preview'" 
              :class="['p-1.5 rounded-md transition-colors', previewMode==='preview' ? 'bg-white dark:bg-zinc-600 text-zinc-900 dark:text-zinc-100 shadow-sm' : 'text-zinc-500 dark:text-zinc-400']"
              title="仅预览"
            ><Eye class="w-4 h-4" /></button>
          </div>
        </div>

        <!-- Editor Area -->
        <div class="flex-1 flex overflow-hidden min-h-0">
          <textarea 
            v-if="previewMode!=='preview'" 
            v-model="editor.content" 
            class="flex-1 p-4 w-full h-full resize-none outline-none bg-white dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100 font-mono text-sm leading-relaxed"
            :class="{'border-r border-zinc-200 dark:border-zinc-800': previewMode==='split'}"
            placeholder="在此输入 Markdown 内容..."
          ></textarea>
          
          <div 
            v-if="previewMode!=='edit'" 
            class="flex-1 p-6 w-full h-full overflow-y-auto bg-white dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100 prose dark:prose-invert max-w-none" 
            v-html="previewHtml"
          ></div>
        </div>
        
        <!-- Status Bar -->
        <div class="p-2 border-t border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-800/50 text-xs text-zinc-500 dark:text-zinc-400 flex justify-between shrink-0">
          <span>Markdown 支持</span>
          <span>共 {{ editorWords }} 字</span>
        </div>
      </div>

      <!-- Right: Meta Settings -->
      <div class="lg:w-80 flex flex-col gap-4 shrink-0">
        <div class="bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800 shadow-sm overflow-hidden flex flex-col min-h-0">
          
          <!-- Accordion Header for Mobile -->
          <button 
            @click="isSettingsOpen = !isSettingsOpen" 
            class="flex items-center justify-between p-4 bg-zinc-50 dark:bg-zinc-800/50 border-b border-zinc-200 dark:border-zinc-800 text-sm font-semibold text-zinc-900 dark:text-zinc-100 shrink-0"
          >
            <span>文章设置</span>
            <ChevronUp v-if="isSettingsOpen" class="w-4 h-4 lg:hidden" />
            <ChevronDown v-else class="w-4 h-4 lg:hidden" />
          </button>
          
          <!-- Settings Content -->
          <div 
            v-show="isSettingsOpen" 
            class="p-4 space-y-4 overflow-y-auto flex-1"
          >
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">标题 <span class="text-red-500">*</span></label>
              <input 
                v-model="editor.title" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                placeholder="输入文章标题" 
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">自定义 Slug</label>
              <input 
                v-model="editor.slug" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                placeholder="例如: my-first-post" 
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">分类</label>
              <select 
                v-model="editor.categoryId" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors"
              >
                <option :value="null">选择分类</option>
                <option v-for="item in categories" :key="item.id" :value="item.id">{{ item.name }}</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">标签</label>
              <select 
                v-model="editor.tagIds" 
                multiple 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors min-h-[100px]"
              >
                <option v-for="item in tags" :key="item.id" :value="item.id">{{ item.name }}</option>
              </select>
              <p class="text-xs text-zinc-500 mt-1">按住 Ctrl/Cmd 多选</p>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">封面图 URL</label>
              <input 
                v-model="editor.coverImage" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                placeholder="https://" 
              />
              <img v-if="editor.coverImage" :src="editor.coverImage" class="mt-2 w-full h-24 object-cover rounded-lg border border-zinc-200 dark:border-zinc-700" />
            </div>

            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">摘要</label>
              <textarea 
                v-model="editor.excerpt" 
                rows="3" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors resize-none" 
                placeholder="列表展示简介"
              ></textarea>
            </div>

            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">SEO 关键词</label>
              <input 
                v-model="editor.seoKeywords" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors" 
                placeholder="逗号分隔" 
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">SEO 描述</label>
              <textarea 
                v-model="editor.seoDescription" 
                rows="2" 
                class="block w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-colors resize-none" 
              ></textarea>
            </div>
          </div>
        </div>
      </div>
      
    </div>
  </div>
</template>

<style>
/* Prose styles for markdown preview */
.prose pre {
  margin: 1em 0;
  padding: 1em;
  border-radius: 0.5rem;
  overflow-x: auto;
  background-color: #0d1117;
}
.prose code {
  font-family: var(--font-mono);
  font-size: 0.875em;
}
.prose :not(pre) > code {
  padding: 0.2em 0.4em;
  background-color: rgba(128, 128, 128, 0.15);
  border-radius: 0.25rem;
}
.prose img {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
}
.prose a {
  color: #2563eb;
  text-decoration: none;
}
.prose a:hover {
  text-decoration: underline;
}
.prose blockquote {
  border-left: 4px solid #e5e7eb;
  padding-left: 1em;
  color: #6b7280;
  margin: 1em 0;
}
.dark .prose blockquote {
  border-left-color: #374151;
  color: #9ca3af;
}
</style>
