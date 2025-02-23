package couchbase

import (
	"context"
	"errors"
	"github.com/couchbase/gocb/v2"
	"go.uber.org/zap"
	"golang-fiber-poc/domain"
	"time"
)

type Repository struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
}

func NewRepository() *Repository {

	cluster, err := gocb.Connect("couchbase://localhost", gocb.ClusterOptions{
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: 3 * time.Second,
			KVTimeout:      3 * time.Second,
			QueryTimeout:   3 * time.Second,
		},
		Authenticator: gocb.PasswordAuthenticator{
			Username: "Administrator",
			Password: "123456789",
		},
		Transcoder: gocb.NewJSONTranscoder(),
	})

	if err != nil {
		zap.L().Fatal("Failed to connect to couchbase", zap.Error(err))
	}

	bucket := cluster.Bucket("products")
	err = bucket.WaitUntilReady(5*time.Second, &gocb.WaitUntilReadyOptions{})
	if err != nil {
		zap.L().Fatal("Failed to connect to bucket", zap.Error(err))
	}

	return &Repository{
		cluster: cluster,
		bucket:  bucket,
	}
}

func (r *Repository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	data, err := r.bucket.DefaultCollection().Get(id, &gocb.GetOptions{
		Timeout: 3 * time.Second,
		Context: ctx,
	})
	if err != nil {
		if errors.Is(err, gocb.ErrDocumentNotFound) {
			return nil, errors.New("product not found")
		}
		zap.L().Error("Failed to get product", zap.Error(err))
		return nil, err
	}

	var product domain.Product

	if err := data.Content(&product); err != nil {
		return nil, err
	}

	return &product, nil

}

func (r *Repository) CreateProduct(ctx context.Context, product *domain.Product) error {
	_, err := r.bucket.DefaultCollection().Insert(product.ID, product, &gocb.InsertOptions{
		Timeout: 3 * time.Second,
		Context: ctx,
	})
	return err
}
