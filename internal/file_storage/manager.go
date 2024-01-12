package filestorage

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3 struct {
	Session   *session.Session
	Bucket    string
	CdnHost   string
	EnableCdn bool
}

type AttachmentInfo struct {
	Filename    string
	ContentType string
	ByteSize    int64
	ETag        string
	Location    string
}

func NewS3Manager(s3 *session.Session, bucket string, enable_cdn bool, cdn_host string) S3 {
	return S3{Session: s3, Bucket: bucket, CdnHost: cdn_host, EnableCdn: enable_cdn}
}

func (s *S3) UploaderFile(filePath string) (*AttachmentInfo, error) {
	session := s.Session

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	fileSize := fileInfo.Size()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fileSize)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}

	fileType := http.DetectContentType(buffer)

	uploader := s3manager.NewUploader(session)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:        bytes.NewReader(buffer),
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(filepath.Base(filePath)),
		ContentType: aws.String(fileType),
	})
	if err != nil {
		return nil, err
	}

	attachment := AttachmentInfo{
		Filename:    filepath.Base(filePath),
		ContentType: fileType,
		ByteSize:    fileSize,
		ETag:        *result.ETag,
		Location:    result.Location,
	}

	log.Printf("RESULT: %v:", attachment)

	return &attachment, nil
}

func (s *S3) GetFileUrl(key string) string {
	if key == "" {
		return ""
	}

	svc := s3.New(s.Session)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(15 * time.Minute)
	if err != nil {
		log.Println(err)
	}

	return url
}

func (s *S3) ProxyImageUrl(key string) string {
	if s.EnableCdn {
		return fmt.Sprintf("https://%s/images/%s", s.CdnHost, key)
	} else {
		return fmt.Sprintf("/images/%s", key)
	}
}
