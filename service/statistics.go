package service

import (
	"context"

	"github.com/sarcb/catalyst-sp24/generated/model"
)

func (s *Service) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	return s.database.Statistics(ctx)
}
