package repo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
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
	Pit       time.Time
}

type Permalink struct {
	Uuid            string
	File            *File
	DurationSeconds int64
	Expires         time.Time
	Pit             time.Time
}

func (p *File) fromDb(row *db.File) {
	p.Uuid = uuid.MustParse(row.Uuid)
	p.Url = row.Url
	p.Notes = row.Notes
	p.Extension = row.Extension
	p.Pit = row.Pit
}

func (r *Repo) permalinkFromDb(row *db.Permalink) (*Permalink, error) {
	f, err := r.GetFile(context.Background(), row.FileUuid)
	if err != nil {
		return nil, fmt.Errorf("error getting file: %w", err)
	}
	return &Permalink{
		Uuid:            row.Uuid,
		File:            f,
		DurationSeconds: row.DurationSeconds,
		Expires:         row.Expires,
		Pit:             row.Pit,
	}, nil
}

func (r *Repo) InsertFile(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
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

func (r *Repo) GetAllFiles(ctx context.Context) ([]File, error) {
	q := db.New(r.db)
	rows, err := q.GetAllFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %w", err)
	}
	files := make([]File, len(rows))
	for i, row := range rows {
		files[i].fromDb(&row)
	}
	return files, nil
}

func (r *Repo) InsertPermalink(ctx context.Context, fileUuid string, durationSeconds int64) (*Permalink, error) {
	_, err := r.GetFile(ctx, fileUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("file not found")
		} else {
			return nil, fmt.Errorf("error getting file: %w", err)
		}
	}

	uuid := ""
	for {
		bytes := make([]byte, 3)
		_, err = rand.Read(bytes)
		if err != nil {
			return nil, fmt.Errorf("error generating uuid: %w", err)
		}
		uuid = hex.EncodeToString(bytes)
		_, err = r.GetPermalink(ctx, uuid)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		r.logger.Warnw("uuid collision", "uuid", uuid)
	}

	params := db.InsertPermalinkParams{
		Uuid:            uuid,
		FileUuid:        fileUuid,
		DurationSeconds: durationSeconds,
		Expires:         time.Now().Add(time.Duration(durationSeconds) * time.Second),
	}
	q := db.New(r.db)
	row, err := q.InsertPermalink(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error inserting permalink: %w", err)
	}
	return r.permalinkFromDb(&row)
}

func (r *Repo) GetPermalink(ctx context.Context, uuid string) (*Permalink, error) {
	q := db.New(r.db)
	row, err := q.GetPermalink(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("error getting permalink: %w", err)
	}
	return r.permalinkFromDb(&row)
}

func (r *Repo) GetAllPermalinks(ctx context.Context) ([]Permalink, error) {
	q := db.New(r.db)
	rows, err := q.GetAllPermalinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting permalinks: %w", err)
	}
	permalinks := make([]Permalink, len(rows))
	for i, row := range rows {
		p, err := r.permalinkFromDb(&row)
		if err != nil {
			return nil, fmt.Errorf("error getting permalink: %w", err)
		}
		permalinks[i] = *p
	}
	return permalinks, nil
}

// TODO use external file server
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
