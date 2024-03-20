package service

import (
	"context"

	"github.com/sarcb/catalyst-sp24/generated/model"
)

func (s *Service) ListTasks(ctx context.Context) ([]*model.TaskWithContext, error) {
	return s.database.TaskList(ctx)
}
