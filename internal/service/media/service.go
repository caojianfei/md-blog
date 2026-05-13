package media

import (
    "context"
    "fmt"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/cybernote/md-blog/internal/config"
    "github.com/cybernote/md-blog/internal/model"
    "github.com/cybernote/md-blog/internal/repository"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type Service struct {
    cfg    config.Config
    repo   *repository.MediaRepository
    client *minio.Client
}

func New(cfg config.Config, repo *repository.MediaRepository) (*Service, error) {
    service := &Service{cfg: cfg, repo: repo}
    if cfg.Storage.Driver == "s3" {
        client, err := minio.New(cfg.Storage.S3Endpoint, &minio.Options{
            Creds:  credentials.NewStaticV4(cfg.Storage.S3AccessKey, cfg.Storage.S3SecretKey, ""),
            Secure: cfg.Storage.S3UseSSL,
            Region: cfg.Storage.S3Region,
        })
        if err != nil {
            return nil, err
        }
        service.client = client
    }
    return service, nil
}

func (s *Service) Upload(file multipart.File, header *multipart.FileHeader) (*model.Media, error) {
    defer file.Close()
    objectKey := fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeFilename(header.Filename))
    media := &model.Media{Filename: objectKey, Original: header.Filename, MIMEType: header.Header.Get("Content-Type"), Size: header.Size, StorageType: s.cfg.Storage.Driver, ObjectKey: objectKey}

    if s.cfg.Storage.Driver == "s3" && s.client != nil {
        _, err := s.client.PutObject(context.Background(), s.cfg.Storage.S3Bucket, objectKey, file, header.Size, minio.PutObjectOptions{ContentType: media.MIMEType})
        if err != nil {
            return nil, err
        }
        if s.cfg.Storage.S3PublicURL != "" {
            media.URL = s.cfg.Storage.S3PublicURL + "/" + objectKey
        } else {
            media.URL = fmt.Sprintf("https://%s/%s/%s", s.cfg.Storage.S3Endpoint, s.cfg.Storage.S3Bucket, objectKey)
        }
    } else {
        if err := os.MkdirAll(s.cfg.Storage.LocalDir, 0o755); err != nil {
            return nil, err
        }
        targetPath := filepath.Join(s.cfg.Storage.LocalDir, objectKey)
        target, err := os.Create(targetPath)
        if err != nil {
            return nil, err
        }
        defer target.Close()
        if _, err := target.ReadFrom(file); err != nil {
            return nil, err
        }
        media.URL = s.cfg.Storage.LocalBaseURL + "/" + objectKey
    }

    if err := s.repo.Save(media); err != nil {
        return nil, err
    }
    return media, nil
}

func (s *Service) List() ([]model.Media, error) { return s.repo.List() }
func (s *Service) Delete(id uint) error         { return s.repo.Delete(id) }

func sanitizeFilename(name string) string {
    lowered := strings.ToLower(strings.TrimSpace(name))
    replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "..", "", "_", "-")
    return replacer.Replace(lowered)
}
