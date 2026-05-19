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
	settingSvc "github.com/cybernote/md-blog/internal/service/setting"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Service struct {
	boot           config.Config
	settings       *settingSvc.Service
	repo           *repository.MediaRepository
	client         *minio.Client
	clientSignature string
}

func New(cfg config.Config, settings *settingSvc.Service, repo *repository.MediaRepository) (*Service, error) {
	return &Service{boot: cfg, settings: settings, repo: repo}, nil
}

func (s *Service) Upload(file multipart.File, header *multipart.FileHeader) (*model.Media, error) {
	defer file.Close()
	resolved, err := s.settings.Resolve()
	if err != nil {
		return nil, err
	}
	objectKey := fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeFilename(header.Filename))
	media := &model.Media{Filename: objectKey, Original: header.Filename, MIMEType: header.Header.Get("Content-Type"), Size: header.Size, StorageType: resolved.Storage.Driver, ObjectKey: objectKey}

	if resolved.Storage.Driver == "s3" {
		client, err := s.ensureS3Client(resolved)
		if err != nil {
			return nil, err
		}
		_, err = client.PutObject(context.Background(), resolved.Storage.S3Bucket, objectKey, file, header.Size, minio.PutObjectOptions{ContentType: media.MIMEType})
		if err != nil {
			return nil, err
		}
		if resolved.Storage.S3PublicURL != "" {
			media.URL = resolved.Storage.S3PublicURL + "/" + objectKey
		} else {
			scheme := "https"
			if !resolved.Storage.S3UseSSL {
				scheme = "http"
			}
			media.URL = fmt.Sprintf("%s://%s/%s/%s", scheme, resolved.Storage.S3Endpoint, resolved.Storage.S3Bucket, objectKey)
		}
	} else {
		if err := os.MkdirAll(resolved.Storage.LocalDirAbs, 0o755); err != nil {
			return nil, err
		}
		targetPath := filepath.Join(resolved.Storage.LocalDirAbs, objectKey)
		target, err := os.Create(targetPath)
		if err != nil {
			return nil, err
		}
		defer target.Close()
		if _, err := target.ReadFrom(file); err != nil {
			return nil, err
		}
		media.URL = resolved.Storage.LocalBaseURL + "/" + objectKey
	}

	if err := s.repo.Save(media); err != nil {
		return nil, err
	}
	return media, nil
}

func (s *Service) List() ([]model.Media, error) { return s.repo.List() }
func (s *Service) Delete(id uint) error         { return s.repo.Delete(id) }

func (s *Service) ensureS3Client(resolved *settingSvc.ResolvedSettings) (*minio.Client, error) {
	signature := strings.Join([]string{
		resolved.Storage.S3Endpoint,
		resolved.Storage.S3Region,
		resolved.Storage.S3Bucket,
		resolved.Storage.S3AccessKey,
		resolved.Storage.S3SecretKey,
		fmt.Sprintf("%t", resolved.Storage.S3UseSSL),
	}, "|")
	if s.client != nil && s.clientSignature == signature {
		return s.client, nil
	}
	client, err := minio.New(resolved.Storage.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(resolved.Storage.S3AccessKey, resolved.Storage.S3SecretKey, ""),
		Secure: resolved.Storage.S3UseSSL,
		Region: resolved.Storage.S3Region,
	})
	if err != nil {
		return nil, err
	}
	s.client = client
	s.clientSignature = signature
	return client, nil
}

func sanitizeFilename(name string) string {
	lowered := strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "..", "", "_", "-")
	return replacer.Replace(lowered)
}
