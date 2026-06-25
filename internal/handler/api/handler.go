package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	appcontainer "github.com/cybernote/md-blog/internal/container"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
	"github.com/cybernote/md-blog/internal/security"
	articleSvc "github.com/cybernote/md-blog/internal/service/article"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Handler struct {
	c *appcontainer.Container
}

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func New(c *appcontainer.Container) *Handler { return &Handler{c: c} }

func (h *Handler) TurnstileConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: map[string]string{
		"siteKey": h.c.Config.Turnstile.SiteKey,
	}})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		TurnstileToken string `json:"turnstileToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	if h.c.Config.Turnstile.SecretKey != "" {
		if err := verifyTurnstile(h.c.Config.Turnstile.SecretKey, payload.TurnstileToken, r.RemoteAddr); err != nil {
			writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "人机验证失败，请重试"})
			return
		}
	}
	admin, err := h.c.Auth.Login(w, r, payload.Username, payload.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, response{Code: 401, Message: "用户名或密码错误"})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: admin})
}

func verifyTurnstile(secretKey, token, remoteIP string) error {
	body := fmt.Sprintf(`{"secret":%q,"response":%q,"remoteip":%q}`, secretKey, token, remoteIP)
	resp, err := http.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("turnstile verification failed")
	}
	return nil
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.c.Auth.Logout(w, r)
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	ok, admin, err := h.c.Auth.CurrentUser(r)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, response{Code: 401, Message: "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: admin})
}

func (h *Handler) Dashboard(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.c.Article.DashboardStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: stats})
}

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	filter := repository.ArticleFilter{
		Query:      r.URL.Query().Get("q"),
		Status:     r.URL.Query().Get("status"),
		CategoryID: uint(parseInt(r.URL.Query().Get("categoryId"))),
		TagID:      uint(parseInt(r.URL.Query().Get("tagId"))),
		Page:       parseIntDefault(r.URL.Query().Get("page"), 1),
		PageSize:   parseIntDefault(r.URL.Query().Get("pageSize"), 10),
	}
	items, total, err := h.c.Article.List(filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: map[string]any{"items": items, "total": total}})
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	article, err := h.c.Article.FindByID(uint(parseInt(chi.URLParam(r, "id"))))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response{Code: status, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: article})
}

func (h *Handler) SaveArticle(w http.ResponseWriter, r *http.Request) {
	var input articleSvc.SaveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	article, err := h.c.Article.Save(r.Context(), input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: article})
}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	if err := h.c.Article.Delete(uint(parseInt(chi.URLParam(r, "id")))); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response{Code: status, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) UpdateArticleStatus(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Status model.ArticleStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	article, err := h.c.Article.UpdateStatus(uint(parseInt(chi.URLParam(r, "id"))), payload.Status)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: article})
}

func (h *Handler) PreviewMarkdown(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	rendered, err := h.c.Article.Preview(payload.Content)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: rendered})
}

func (h *Handler) ListCategories(w http.ResponseWriter, _ *http.Request) {
	items, err := h.c.CategoryRepo.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: items})
}

func (h *Handler) SaveCategory(w http.ResponseWriter, r *http.Request) {
	var item model.Category
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	item.Name = strings.TrimSpace(item.Name)
	item.Slug = normalizeSlug(item.Slug, item.Name)
	if item.Name == "" {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "name is required"})
		return
	}
	if err := h.c.CategoryRepo.Save(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: item})
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := uint(parseInt(chi.URLParam(r, "id")))
	articleCount, err := h.c.ArticleRepo.CountByCategory(categoryID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	if articleCount > 0 && !parseForce(r) {
		writeJSON(w, http.StatusConflict, response{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("该分类下有 %d 篇文章，删除后这些文章将变为未分类", articleCount),
			Data: map[string]any{
				"confirmRequired": true,
				"articleCount":    articleCount,
			},
		})
		return
	}
	err = h.c.DB.Transaction(func(tx *gorm.DB) error {
		articleRepo := repository.NewArticleRepository(tx)
		categoryRepo := repository.NewCategoryRepository(tx)
		if err := articleRepo.ClearCategory(categoryID); err != nil {
			return err
		}
		return categoryRepo.Delete(categoryID)
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) ListTags(w http.ResponseWriter, _ *http.Request) {
	items, err := h.c.TagRepo.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: items})
}

func (h *Handler) SaveTag(w http.ResponseWriter, r *http.Request) {
	var item model.Tag
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	item.Name = strings.TrimSpace(item.Name)
	item.Slug = normalizeSlug(item.Slug, item.Name)
	if item.Name == "" {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "name is required"})
		return
	}
	if err := h.c.TagRepo.Save(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: item})
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID := uint(parseInt(chi.URLParam(r, "id")))
	articleCount, err := h.c.ArticleRepo.CountByTag(tagID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	if articleCount > 0 && !parseForce(r) {
		writeJSON(w, http.StatusConflict, response{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("该标签关联 %d 篇文章，删除后会同步移除这些文章上的该标签", articleCount),
			Data: map[string]any{
				"confirmRequired": true,
				"articleCount":    articleCount,
			},
		})
		return
	}
	err = h.c.DB.Transaction(func(tx *gorm.DB) error {
		articleRepo := repository.NewArticleRepository(tx)
		tagRepo := repository.NewTagRepository(tx)
		if err := articleRepo.DetachTag(tagID); err != nil {
			return err
		}
		return tagRepo.Delete(tagID)
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) GetSettings(w http.ResponseWriter, _ *http.Request) {
	setting, err := h.c.Settings.Get()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: setting})
}

func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var payload model.SiteSetting
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	saved, err := h.c.Settings.Save(&payload)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: saved})
}

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	ok, admin, err := h.c.Auth.CurrentUser(r)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, response{Code: 401, Message: "unauthorized"})
		return
	}
	var payload struct {
		Username        string `json:"username"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	payload.Username = strings.TrimSpace(payload.Username)
	if payload.Username == "" {
		payload.Username = admin.Username
	}
	if payload.NewPassword != "" {
		if len(payload.NewPassword) < 8 {
			writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "新密码至少 8 位"})
			return
		}
		if payload.NewPassword != payload.ConfirmPassword {
			writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "两次输入的密码不一致"})
			return
		}
		hash, err := security.HashPassword(payload.NewPassword)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
			return
		}
		admin.PasswordHash = hash
		admin.PasswordReset = true
	}
	admin.Username = payload.Username
	if err := h.c.AdminRepo.Save(admin); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: admin})
}

func (h *Handler) ListMedia(w http.ResponseWriter, _ *http.Request) {
	items, err := h.c.Media.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: items})
}

func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	resolved, err := h.c.Settings.Resolve()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	if resolved.Compression.Enabled {
		// 压缩超时由配置决定，额外加 30 秒余量供响应写入
		deadline := time.Duration(resolved.Compression.TimeoutSeconds+30) * time.Second
		_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(deadline))
	}
	if err := r.ParseMultipartForm(resolved.MaxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "missing file"})
		return
	}
	media, err := h.c.Media.Upload(file, header)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: media})
}

func (h *Handler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	if err := h.c.Media.Delete(uint(parseInt(chi.URLParam(r, "id")))); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) PreviewConfig(w http.ResponseWriter, _ *http.Request) {
	resolved, err := h.c.Settings.Resolve()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: map[string]string{
		"previewKey": resolved.PreviewSecret,
	}})
}

func writeJSON(w http.ResponseWriter, status int, payload response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func parseIntDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func normalizeSlug(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", ",", "", ":", "", "?", "", "!", "")
	value = strings.Trim(replacer.Replace(value), "-")
	if value == "" {
		return "item"
	}
	return value
}

func (h *Handler) TerminalArticles(w http.ResponseWriter, _ *http.Request) {
	filter := repository.ArticleFilter{
		Status:   "published",
		Page:     1,
		PageSize: 5,
	}
	items, _, err := h.c.Article.List(filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}

	articles := make([]map[string]string, 0, len(items))
	for _, item := range items {
		articles = append(articles, map[string]string{
			"title": item.Title,
			"date":  item.PublishedAt.Format("2006-01-02"),
			"slug":  item.Slug,
		})
	}

	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: map[string]any{"articles": articles}})
}

func (h *Handler) TerminalCategories(w http.ResponseWriter, _ *http.Request) {
	categories, err := h.c.CategoryRepo.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}

	categories = visibleCategories(categories)
	result := make([]map[string]any, 0, len(categories))
	for _, cat := range categories {
		result = append(result, map[string]any{
			"name":         cat.Name,
			"slug":         cat.Slug,
			"description":  cat.Description,
			"articleCount": cat.ArticleCount,
		})
	}

	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: map[string]any{"categories": result}})
}

func (h *Handler) TerminalTags(w http.ResponseWriter, _ *http.Request) {
	tags, err := h.c.TagRepo.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}

	tags = visibleTags(tags)
	result := make([]map[string]string, 0, len(tags))
	for _, tag := range tags {
		result = append(result, map[string]string{
			"name": tag.Name,
			"slug": tag.Slug,
		})
	}

	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: map[string]any{"tags": result}})
}

func parseForce(r *http.Request) bool {
	value := strings.TrimSpace(r.URL.Query().Get("force"))
	return value == "1" || strings.EqualFold(value, "true")
}

func visibleCategories(items []model.Category) []model.Category {
	filtered := make([]model.Category, 0, len(items))
	for _, item := range items {
		if item.ArticleCount <= 0 {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].ArticleCount != filtered[j].ArticleCount {
			return filtered[i].ArticleCount > filtered[j].ArticleCount
		}
		if filtered[i].Sort != filtered[j].Sort {
			return filtered[i].Sort < filtered[j].Sort
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})
	return filtered
}

func visibleTags(items []model.Tag) []model.Tag {
	filtered := make([]model.Tag, 0, len(items))
	for _, item := range items {
		if item.ArticleCount <= 0 {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].ArticleCount != filtered[j].ArticleCount {
			return filtered[i].ArticleCount > filtered[j].ArticleCount
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})
	return filtered
}

var _ = fmt.Sprintf
