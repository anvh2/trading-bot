package server

import (
	"context"
	"time"
)

func (s *Server) Polling(ctx context.Context, idx int32) error {
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {

	}

	return nil
}
