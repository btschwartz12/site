package repo

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/btschwartz12/site/internal/repo/db"
	"github.com/google/uuid"
)

const (
	pictureUploadDir = "pictures"
	maxUploadMb      = 5
	maxUploadSize    = maxUploadMb << 20
)

var (
	allowedExtensionsRe = regexp.MustCompile(`\.(jpe?g|png|gif)$`)
)

type Picture struct {
	ID          int64
	Author      string
	Url         string
	Description string
	Extension   string
	NumLikes    int64
	NumDislikes int64
	Pit         time.Time
}

func (p *Picture) fromDb(row *db.Picture) {
	p.ID = row.ID
	p.Author = row.Author
	p.Url = row.Url
	p.Description = row.Description
	p.Extension = row.Extension
	p.NumLikes = row.NumLikes
	p.NumDislikes = row.NumDislikes
	p.Pit = row.Pit
}

func (r *Repo) InsertPicture(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	author string,
	description string,
) (*Picture, error) {
	if r.storageFull() {
		return nil, fmt.Errorf("storage full")
	}

	if header.Size > maxUploadSize {
		return nil, fmt.Errorf("file too large (max %d MB)", maxUploadMb)
	}
	r.logger.Infow("uploading picture", "size", header.Size, "max", maxUploadSize)

	ext := filepath.Ext(header.Filename)
	if !allowedExtensionsRe.MatchString(ext) {
		return nil, fmt.Errorf("invalid file extension")
	}

	newPath := filepath.Join(r.varDir, pictureUploadDir, uuid.New().String()+ext)

	newFile, err := os.Create(newPath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		return nil, fmt.Errorf("error copying file content: %w", err)
	}

	params := db.InsertPictureParams{
		Author:      author,
		Url:         newPath,
		Description: description,
		Extension:   ext,
	}
	q := db.New(r.db)
	row, err := q.InsertPicture(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error inserting picture: %w", err)
	}

	p := Picture{}
	p.fromDb(&row)
	return &p, nil
}

func (r *Repo) GetAllPictures(ctx context.Context) ([]Picture, error) {
	q := db.New(r.db)
	rows, err := q.GetAllPictures(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting pictures: %w", err)
	}

	pictures := make([]Picture, 0, len(rows))
	for _, row := range rows {
		p := Picture{}
		p.fromDb(&row)
		pictures = append(pictures, p)
	}
	return pictures, nil
}

func (r *Repo) GetPicture(ctx context.Context, basename string) (*Picture, error) {
	id, ext, err := parsePictureBasename(basename)
	if err != nil {
		return nil, fmt.Errorf("error parsing basename: %w", err)
	}

	q := db.New(r.db)
	row, err := q.GetPicture(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting picture: %w", err)
	}

	if row.Extension != ext {
		return nil, fmt.Errorf("invalid extension")
	}

	p := Picture{}
	p.fromDb(&row)
	return &p, nil
}

func (r *Repo) DeletePicture(ctx context.Context, idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting id to int64: %w", err)
	}
	q := db.New(r.db)
	url, err := q.DeletePicture(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting picture: %w", err)
	}

	err = os.Remove(url)
	if err != nil {
		return fmt.Errorf("error removing file: %w", err)
	}

	return nil
}

func (r *Repo) LikePicture(ctx context.Context, idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting id to int64: %w", err)
	}
	q := db.New(r.db)
	err = q.AddLikeToPicture(ctx, id)
	if err != nil {
		return fmt.Errorf("error liking picture: %w", err)
	}
	return nil
}

func (r *Repo) DislikePicture(ctx context.Context, idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting id to int64: %w", err)
	}
	q := db.New(r.db)
	err = q.AddDislikeToPicture(ctx, id)
	if err != nil {
		return fmt.Errorf("error disliking picture: %w", err)
	}
	return nil
}

func (r *Repo) UpdateLikesOfPicture(ctx context.Context, idStr string, likes int64, dislikes int64) (*Picture, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting id to int64: %w", err)
	}
	q := db.New(r.db)
	picture, err := q.UpdateLikesDislikesOfPicture(ctx, db.UpdateLikesDislikesOfPictureParams{
		ID:          id,
		NumLikes:    likes,
		NumDislikes: dislikes,
	})
	if err != nil {
		return nil, fmt.Errorf("error updating likes/dislikes: %w", err)
	}
	p := Picture{}
	p.fromDb(&picture)
	return &p, nil
}

func parsePictureBasename(basename string) (int64, string, error) {
	re := regexp.MustCompile(`^(\d+)(\.\w+)$`)
	matches := re.FindStringSubmatch(basename)
	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid basename format")
	}

	i, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("error converting id to int64: %w", err)
	}

	return i, matches[2], nil
}
