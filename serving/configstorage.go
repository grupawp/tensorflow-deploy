package serving

import "context"

type ConfigStorage interface {
	ReadConfig(ctx context.Context, team, project string) ([]byte, error)
	SaveConfig(ctx context.Context, team, project string, config []byte) error
}
