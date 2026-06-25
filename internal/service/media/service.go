package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	boot            config.Config
	settings        *settingSvc.Service
	repo            *repository.MediaRepository
	client          *minio.Client
	clientSignature string
}

func New(cfg config.Config, settings *settingSvc.Service, repo *repository.MediaRepository) (*Service, error) {
	return &Service{boot: cfg, settings: settings, repo: repo}, nil
}

var compressibleMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/avif": true,
}

func (s *Service) Upload(file multipart.File, header *multipart.FileHeader) (*model.Media, error) {
	defer file.Close()
	resolved, err := s.settings.Resolve()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	mimeType := header.Header.Get("Content-Type")
	if resolved.Compression.Enabled && compressibleMIMETypes[mimeType] {
		compressed, err := s.compressWithTinyPNG(resolved.Compression.APIKey, data, mimeType, resolved.Compression.TimeoutSeconds)
		if err != nil {
			return nil, fmt.Errorf("图片压缩失败: %w", err)
		}
		data = compressed
	}

	objectKey := fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeFilename(header.Filename))
	media := &model.Media{
		Filename:    objectKey,
		Original:    header.Filename,
		MIMEType:    mimeType,
		Size:        int64(len(data)),
		StorageType: resolved.Storage.Driver,
		ObjectKey:   objectKey,
	}

	if resolved.Storage.Driver == "s3" {
		client, err := s.ensureS3Client(resolved)
		if err != nil {
			return nil, err
		}
		_, err = client.PutObject(context.Background(), resolved.Storage.S3Bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: media.MIMEType})
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
		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
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

func (s *Service) compressWithTinyPNG(apiKey string, data []byte, mimeType string, timeoutSeconds int) ([]byte, error) {
	httpClient := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}

	req, err := http.NewRequest(http.MethodPost, "https://api.tinify.com/shrink", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", mimeType)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("%s: %s", errResp.Error, errResp.Message)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("tinypng 未返回压缩图地址")
	}

	getReq, err := http.NewRequest(http.MethodGet, location, nil)
	if err != nil {
		return nil, err
	}
	getReq.SetBasicAuth("api", apiKey)

	getResp, err := httpClient.Do(getReq)
	if err != nil {
		return nil, err
	}
	defer getResp.Body.Close()

	return io.ReadAll(getResp.Body)
}

func sanitizeFilename(name string) string {
	lowered := strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "..", "", "_", "-")
	return replacer.Replace(lowered)
}
