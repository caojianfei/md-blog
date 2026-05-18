package web

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	appcontainer "github.com/cybernote/md-blog/internal/container"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	markdownSvc "github.com/cybernote/md-blog/internal/service/markdown"
	seoSvc "github.com/cybernote/md-blog/internal/service/seo"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	c *appcontainer.Container
}

type PageData struct {
	Site            *model.SiteSetting
	Meta            seoSvc.Meta
	Articles        []model.Article
	Article         *model.Article
	Headings        []markdownSvc.Heading
	Categories      []model.Category
	Tags            []model.Tag
	Archives        []repository.ArchiveItem
	CurrentPage     int
	TotalPages      int
	Path            string
	Query           string
	CurrentTag      *model.Tag
	CurrentCategory *model.Category
	PrevArticle     *model.Article
	NextArticle     *model.Article
	AboutHTML       string
	PublishedCount  int64
	ArchiveCount    int
	IsPreview       bool
}

func New(c *appcontainer.Container) *Handler { return &Handler{c: c} }

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	items, total, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, Page: page, PageSize: 10})
	data := h.baseData("首页", "/", "", "")
	data.Articles = items
	data.CurrentPage = page
	data.TotalPages = totalPages(total, 10)
	h.c.Renderer.Render(w, "home", data, 0)
}

func (h *Handler) Article(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	article, err := h.c.Article.FindBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	isPreview := false
	if previewKey := r.URL.Query().Get("preview_key"); previewKey != "" {
		if previewKey != h.c.Config.App.PreviewSecret {
			http.NotFound(w, r)
			return
		}
		ok, _, authErr := h.c.Auth.CurrentUser(r)
		if authErr != nil || !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		isPreview = true
	} else {
		if article.Status != model.ArticleStatusPublished {
			http.NotFound(w, r)
			return
		}
	}

	prev, next, _ := h.c.Article.PrevNext(article)
	data := h.baseData(article.Title, r.URL.Path, article.SEODescription, article.SEOKeywords)
	data.Article = article
	data.PrevArticle = prev
	data.NextArticle = next
	data.IsPreview = isPreview
	if rendered, renderErr := h.c.Markdown.Render(article.Content); renderErr == nil {
		data.Article.HTMLContent = rendered.HTML
		data.Headings = rendered.Headings
	}
	h.c.Renderer.Render(w, "article", data, 0)
}

func (h *Handler) Category(w http.ResponseWriter, r *http.Request) {
	category, err := h.c.CategoryRepo.FindBySlug(chi.URLParam(r, "slug"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	page := parsePage(r)
	items, total, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, CategoryID: category.ID, Page: page, PageSize: 10})
	data := h.baseData(category.Name, r.URL.Path, category.Description, category.Name)
	data.Articles = items
	data.CurrentCategory = category
	data.CurrentPage = page
	data.TotalPages = totalPages(total, 10)
	h.c.Renderer.Render(w, "category", data, 0)
}

func (h *Handler) Tag(w http.ResponseWriter, r *http.Request) {
	tag, err := h.c.TagRepo.FindBySlug(chi.URLParam(r, "slug"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	page := parsePage(r)
	items, total, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, TagID: tag.ID, Page: page, PageSize: 10})
	data := h.baseData(tag.Name, r.URL.Path, tag.Description, tag.Name)
	data.Articles = items
	data.CurrentTag = tag
	data.CurrentPage = page
	data.TotalPages = totalPages(total, 10)
	h.c.Renderer.Render(w, "tag", data, 0)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	q := r.URL.Query().Get("q")
	items, total, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, Query: q, Page: page, PageSize: 10})
	data := h.baseData("搜索", r.URL.Path, "站内搜索结果", q)
	data.Articles = items
	data.Query = q
	data.CurrentPage = page
	data.TotalPages = totalPages(total, 10)
	h.c.Renderer.Render(w, "search", data, 0)
}

func (h *Handler) Archives(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	archives, _ := h.c.Article.Archives()
	data := h.baseData("归档", r.URL.Path, "文章归档", "归档")
	data.Archives = archives
	data.ArchiveCount = len(archives)

	if yearStr != "" && monthStr != "" {
		year, _ := strconv.Atoi(yearStr)
		month, _ := strconv.Atoi(monthStr)

		page := parsePage(r)
		items, total, _ := h.c.Article.List(repository.ArticleFilter{
			OnlyPublic: true,
			Year:       year,
			Month:      month,
			Page:       page,
			PageSize:   10,
		})

		data.Articles = items
		data.CurrentPage = page
		data.TotalPages = totalPages(total, 10)
		data.Query = fmt.Sprintf("%d-%02d", year, month)
	}

	h.c.Renderer.Render(w, "archives", data, 0)
}

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	data := h.baseData("关于", r.URL.Path, "关于本站", "关于")
	rendered, _ := h.c.Markdown.Render(data.Site.AboutContent)
	if rendered != nil {
		data.AboutHTML = rendered.HTML
	}
	archives, _ := h.c.Article.Archives()
	_, total, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, Page: 1, PageSize: 1})
	data.Archives = archives
	data.ArchiveCount = len(archives)
	data.PublishedCount = total
	h.c.Renderer.Render(w, "about", data, 0)
}

func (h *Handler) RSS(w http.ResponseWriter, _ *http.Request) {
	items, _, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, Page: 1, PageSize: 20})

	type Item struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
	}
	type Channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Items       []Item `xml:"item"`
	}
	type RSS struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Channel Channel  `xml:"channel"`
	}

	site, _ := h.c.SettingRepo.Get()
	feed := RSS{Version: "2.0", Channel: Channel{Title: h.c.Config.App.Name, Link: h.c.Config.App.BaseURL, Description: site.SiteDescription}}
	for _, article := range items {
		feed.Channel.Items = append(feed.Channel.Items, Item{Title: article.Title, Link: h.c.Config.App.BaseURL + "/posts/" + url.PathEscape(article.Slug), Description: article.Excerpt})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_ = xml.NewEncoder(w).Encode(feed)
}

func (h *Handler) Sitemap(w http.ResponseWriter, _ *http.Request) {
	items, _, _ := h.c.Article.List(repository.ArticleFilter{OnlyPublic: true, Page: 1, PageSize: 500})

	type URL struct {
		Loc string `xml:"loc"`
	}
	type URLSet struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		URLs    []URL    `xml:"url"`
	}

	urls := []URL{{Loc: h.c.Config.App.BaseURL}, {Loc: h.c.Config.App.BaseURL + "/archives"}, {Loc: h.c.Config.App.BaseURL + "/about"}}
	for _, article := range items {
		urls = append(urls, URL{Loc: h.c.Config.App.BaseURL + "/posts/" + url.PathEscape(article.Slug)})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_ = xml.NewEncoder(w).Encode(URLSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls})
}

func (h *Handler) Robots(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\nAllow: /\nSitemap: " + h.c.Config.App.BaseURL + "/sitemap.xml\n"))
}

func (h *Handler) baseData(title, path, description, keywords string) PageData {
	site, _ := h.c.SettingRepo.Get()
	categories, _ := h.c.CategoryRepo.List()
	tags, _ := h.c.TagRepo.List()
	archives, _ := h.c.Article.Archives()
	return PageData{
		Site:         site,
		Categories:   categories,
		Tags:         tags,
		Archives:     archives,
		ArchiveCount: len(archives),
		Path:         path,
		Meta:         h.c.SEO.Build(title, description, keywords, path),
	}
}

func parsePage(r *http.Request) int {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		return 1
	}
	return page
}

func totalPages(total int64, size int) int {
	if total == 0 {
		return 1
	}
	return int((total + int64(size) - 1) / int64(size))
}
