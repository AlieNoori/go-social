package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQeury struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (fq *PaginatedFeedQeury) Parse(r *http.Request) error {
	qv := r.URL.Query()
	limit := qv.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return err
		}

		fq.Limit = l
	}

	offset := qv.Get("offset")
	if offset != "" {
		off, err := strconv.Atoi(offset)
		if err != nil {
			return err
		}

		fq.Offset = off
	}

	sort := qv.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := qv.Get("tags")
	if len(tags) > 0 {
		fq.Tags = strings.Split(tags, ",")
	}

	search := qv.Get("search")
	if search != "" {
		fq.Search = search
	}

	since := qv.Get("since")
	if since != "" {
		fq.Since = parseTime(since)
	}

	until := qv.Get("until")
	if until != "" {
		fq.Until = parseTime(until)
	}

	return nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateOnly)
}
