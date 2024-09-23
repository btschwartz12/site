package repo

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/btschwartz12/site/internal/ipdata"
	"github.com/btschwartz12/site/internal/repo/db"
	"github.com/btschwartz12/site/internal/slack"
)

type Visitor struct {
	ID      int64
	Path    string
	Message string
	Ip      string
	City    string
	Region  string
	Country string
	Pit     string
}

func (r *Repo) RecordVisitor(ctx context.Context, req *http.Request, message string, slackBlocks []slack.Block) error {
	ip := ipdata.GetIp(req)
	params := db.InsertVisitorParams{
		Ip:      sql.NullString{String: ip.String(), Valid: true},
		Path:    req.URL.Path,
		Message: message,
	}

	info := ipdata.GetIpinfoRecord(ip)
	if info == nil {
		r.logger.Errorw("error getting IP info", "ip", ip)
	} else {
		params.Country = sql.NullString{String: info.Country, Valid: true}
		params.Region = sql.NullString{String: info.Region, Valid: true}
		params.City = sql.NullString{String: info.City, Valid: true}
	}

	if slackBlocks != nil {
		blocks := ipdata.GetVisitBlocks(req, ip, info)
		blocks = append(blocks, slackBlocks...)
		slack.SendAlert(r.logger, "visit", blocks)
	}

	q := db.New(r.db)
	err := q.InsertVisitor(ctx, params)
	if err != nil {
		r.logger.Errorw("error inserting visitor", "error", err)
		return fmt.Errorf("error inserting visitor: %w", err)
	}
	return nil
}

func (r *Repo) GetAllVisitors(ctx context.Context) ([]Visitor, error) {
	q := db.New(r.db)
	rows, err := q.GetAllVisitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all visitors: %w", err)
	}
	visitors := make([]Visitor, 0, len(rows))
	for _, v := range rows {
		visitors = append(visitors, Visitor{
			ID:      v.ID,
			Path:    v.Path,
			Message: v.Message,
			Ip:      v.Ip.String,
			City:    v.City.String,
			Region:  v.Region.String,
			Country: v.Country.String,
			Pit:     v.Pit.String(),
		})
	}
	return visitors, nil
}
