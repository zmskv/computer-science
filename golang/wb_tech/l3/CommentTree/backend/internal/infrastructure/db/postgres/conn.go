package postgres

import (
	"context"

	"github.com/wb-go/wbf/dbpg"
)

func NewDB(ctx context.Context, masterDSN string, slaveDSNs []string) (*dbpg.DB, error) {
	return dbpg.New(masterDSN, slaveDSNs, nil)
}
