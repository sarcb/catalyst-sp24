package service

import (
	"context"
	"net/url"

	"github.com/sarcb/catalyst-sp24/generated/model"
)

func (s *Service) GetLogs(ctx context.Context, reference string) ([]*model.LogEntry, error) {
	id, _ := url.QueryUnescape(reference)

	return s.database.LogList(ctx, id)
}
