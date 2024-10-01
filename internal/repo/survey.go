package repo

import (
	"context"
	"fmt"

	"github.com/btschwartz12/site/internal/repo/db"
)

func (r *Repo) GetSurveyState(ctx context.Context) ([]byte, error) {
	q := db.New(r.db)
	data, err := q.GetSurveyState(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting survey state: %w", err)
	}
	return data, nil
}

func (r *Repo) UpdateSurveyState(ctx context.Context, data []byte) error {
	q := db.New(r.db)
	err := q.UpdateSurveyState(ctx, data)
	if err != nil {
		return fmt.Errorf("error updating survey state: %w", err)
	}
	return nil
}
