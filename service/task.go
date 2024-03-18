package service

import (
	"context"

	"github.com/sarcb/catalyst/generated/model"
)

func (s *Service) ListTasks(ctx context.Context) ([]*model.TaskWithContext, error) {
	return s.database.TaskList(ctx)
}
