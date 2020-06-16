package reject

import (
	"context"

	"github.com/puppetlabs/relay-core/pkg/model"
)

type connectionManager struct{}

func (*connectionManager) Get(ctx context.Context, typ, name string) (*model.Connection, error) {
	return nil, model.ErrRejected
}

var ConnectionManager model.ConnectionManager = &connectionManager{}
