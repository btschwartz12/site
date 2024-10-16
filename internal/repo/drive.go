package repo

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/btschwartz12/site/internal/repo/db"
	"github.com/google/uuid"
)

const (
	driveUploadDir    = "drive"
	maxFileUploadMb   = 5
	maxFileUploadSize = maxFileUploadMb << 20
)

type File struct {
	Uuid      uuid.UUID
	Url       string
	Notes     string
	Extension string
	Expires   time.Time
	Pit       time.Time
}

func (p *File) fromDb(row *db.File) {
	p.Uuid = uuid.MustParse(row.Uuid)
	p.Url = row.Url
	p.Notes = row.Notes
	p.Extension = row.Extension
	p.Expires = row.Expires
	p.Pit = row.Pit
}

func (r *Repo) InsertFile(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	expires time.Time,
	notes string,
) (*File, error) {
	if r.storageFull() {
		return nil, fmt.Errorf("storage full")
	}

	if header.Size > maxFileUploadSize {
		return nil, fmt.Errorf("file too large (max %d MB)", maxPictureUploadMb)
	}
	ext := filepath.Ext(header.Filename)

	url, err := r.getFileUrl(ctx, uuid.New().String(), file, header)
	if err != nil {
		return nil, fmt.Errorf("error getting file url: %w", err)
	}

	params := db.InsertFileParams{
		Uuid:      uuid.New().String(),
		Url:       url,
		Notes:     notes,
		Extension: ext,
		Expires:   expires,
	}
	q := db.New(r.db)
	row, err := q.InsertFile(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error inserting picture: %w", err)
	}

	f := &File{}
	f.fromDb(&row)
	return f, nil
}

func (r *Repo) getFileUrl(ctx context.Context, uuid string, file multipart.File,
	header *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(header.Filename)
	path := filepath.Join(r.varDir, driveUploadDir, uuid+ext)
	newFile, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, file)
	if err != nil {
		return "", fmt.Errorf("error copying file content: %w", err)
	}
	return path, nil
}

func (r *Repo) GetFile(ctx context.Context, uuid string) (*File, error) {
	q := db.New(r.db)
	row, err := q.GetFile(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("error getting file: %w", err)
	}
	f := &File{}
	f.fromDb(&row)
	return f, nil
}

func (r *Repo) UpdateFileExpires(ctx context.Context, uuid string, expires time.Time) (*File, error) {
	q := db.New(r.db)
	row, err := q.UpdateFileExpires(ctx, db.UpdateFileExpiresParams{
		Expires: expires,
		Uuid:    uuid,
	})
	if err != nil {
		r.logger.Errorw("error updating file expires", "error", err)
		return nil, fmt.Errorf("error updating file expires: %w", err)
	}
	f := &File{}
	f.fromDb(&row)
	return f, nil
}

func (r *Repo) GetAllFiles(ctx context.Context) ([]*File, error) {
	q := db.New(r.db)
	rows, err := q.GetAllFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all files: %w", err)
	}
	files := make([]*File, 0, len(rows))
	for i := range rows {
		f := &File{}
		f.fromDb(&rows[i])
		files = append(files, f)
	}
	return files, nil
}
