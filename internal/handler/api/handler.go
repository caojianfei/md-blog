package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	appcontainer "github.com/cybernote/md-blog/internal/container"
	"github.com/cybernote/md-blog/internal/model"
	"github.com/cybernote/md-blog/internal/repository"
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	admin, err := h.c.Auth.Login(w, r, payload.Username, payload.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, response{Code: 401, Message: "用户名或密码错误"})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: admin})
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
	article, err := h.c.Article.Save(input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: article})
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
	if err := h.c.CategoryRepo.Delete(uint(parseInt(chi.URLParam(r, "id")))); err != nil {
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
	if err := h.c.TagRepo.Delete(uint(parseInt(chi.URLParam(r, "id")))); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) GetSettings(w http.ResponseWriter, _ *http.Request) {
	setting, err := h.c.SettingRepo.Get()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: setting})
}

func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	current, err := h.c.SettingRepo.Get()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Code: 500, Message: err.Error()})
		return
	}
	var payload model.SiteSetting
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: "invalid payload"})
		return
	}
	payload.ID = current.ID
	if strings.TrimSpace(payload.SiteName) == "" {
		payload.SiteName = current.SiteName
	}
	if err := h.c.SettingRepo.Save(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok", Data: payload})
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
	if err := r.ParseMultipartForm(h.c.Config.Storage.MaxUploadSize); err != nil {
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

var _ = fmt.Sprintf
