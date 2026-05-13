package auth

import (
    "net/http"
    "time"

    "github.com/cybernote/md-blog/internal/model"
    "github.com/cybernote/md-blog/internal/repository"
    "github.com/cybernote/md-blog/internal/security"
    "github.com/gorilla/sessions"
)

const sessionName = "md-blog-admin"
const sessionUserKey = "admin_user_id"

type Service struct {
    repo  *repository.AdminRepository
    store sessions.Store
}

func New(repo *repository.AdminRepository, store sessions.Store) *Service {
    return &Service{repo: repo, store: store}
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request, username, password string) (*model.AdminUser, error) {
    admin, err := s.repo.FindByUsername(username)
    if err != nil {
        return nil, err
    }
    if err := security.VerifyPassword(admin.PasswordHash, password); err != nil {
        return nil, err
    }

    now := time.Now()
    admin.LastLoginAt = &now
    admin.LastLoginIP = r.RemoteAddr
    _ = s.repo.Save(admin)

    session, _ := s.store.Get(r, sessionName)
    session.Values[sessionUserKey] = admin.ID
    if err := session.Save(r, w); err != nil {
        return nil, err
    }
    return admin, nil
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
    session, _ := s.store.Get(r, sessionName)
    session.Options.MaxAge = -1
    return session.Save(r, w)
}

func (s *Service) CurrentUser(r *http.Request) (bool, *model.AdminUser, error) {
    session, _ := s.store.Get(r, sessionName)
    rawID, ok := session.Values[sessionUserKey]
    if !ok {
        return false, nil, nil
    }

    var id uint
    switch value := rawID.(type) {
    case uint:
        id = value
    case int:
        id = uint(value)
    case int64:
        id = uint(value)
    case float64:
        id = uint(value)
    default:
        return false, nil, nil
    }

    admin, err := s.repo.FindByID(id)
    if err != nil {
        return false, nil, err
    }
    return true, admin, nil
}
