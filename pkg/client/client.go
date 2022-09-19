package client

import (
	"context"

	"github.com/MaciejTe/accountapiclient/pkg/models"
)

type Client interface {
	Create(ctx context.Context, accountData models.AccountData) (*models.AccountData, error)
	Fetch(ctx context.Context, id string) (*models.AccountData, error)
	Delete(ctx context.Context, id, version string) error
}
