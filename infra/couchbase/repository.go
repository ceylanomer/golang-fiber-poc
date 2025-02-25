package couchbase

import (
	"context"
	"errors"
	"golang-fiber-poc/domain"
	"golang-fiber-poc/pkg/config"
	"time"

	gocbopentelemetry "github.com/couchbase/gocb-opentelemetry"
	"github.com/couchbase/gocb/v2"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

type Repository struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	tracer  *gocbopentelemetry.OpenTelemetryRequestTracer
}

func NewRepository(tp *sdktrace.TracerProvider, couchbaseConfig config.CouchbaseConfig) *Repository {
	tracer := gocbopentelemetry.NewOpenTelemetryRequestTracer(tp)
	cluster, err := gocb.Connect(couchbaseConfig.URL, gocb.ClusterOptions{
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: 3 * time.Second,
			KVTimeout:      3 * time.Second,
			QueryTimeout:   3 * time.Second,
		},
		Authenticator: gocb.PasswordAuthenticator{
			Username: couchbaseConfig.Username,
			Password: couchbaseConfig.Password,
		},
		Transcoder: gocb.NewJSONTranscoder(),
		Tracer:     tracer,
	})

	if err != nil {
		zap.L().Fatal("Failed to connect to couchbase", zap.Error(err))
	}

	bucket := cluster.Bucket(couchbaseConfig.Bucket)
	err = bucket.WaitUntilReady(20*time.Second, &gocb.WaitUntilReadyOptions{})
	if err != nil {
		zap.L().Fatal("Failed to connect to bucket", zap.Error(err))
	}

	return &Repository{
		cluster: cluster,
		bucket:  bucket,
		tracer:  tracer,
	}
}

func (r *Repository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	ctx, span := r.tracer.Wrapped().Start(ctx, "GetProduct")
	defer span.End()
	data, err := r.bucket.DefaultCollection().Get(id, &gocb.GetOptions{
		Timeout:    3 * time.Second,
		Context:    ctx,
		ParentSpan: gocbopentelemetry.NewOpenTelemetryRequestSpan(ctx, span),
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
	ctx, span := r.tracer.Wrapped().Start(ctx, "CreateProduct")
	defer span.End()
	_, err := r.bucket.DefaultCollection().Insert(product.ID, product, &gocb.InsertOptions{
		Timeout:    3 * time.Second,
		Context:    ctx,
		ParentSpan: gocbopentelemetry.NewOpenTelemetryRequestSpan(ctx, span),
	})
	return err
}

func (r *Repository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	ctx, span := r.tracer.Wrapped().Start(ctx, "UpdateProduct")
	defer span.End()
	_, err := r.bucket.DefaultCollection().Replace(product.ID, product, &gocb.ReplaceOptions{
		Timeout:    3 * time.Second,
		Context:    ctx,
		ParentSpan: gocbopentelemetry.NewOpenTelemetryRequestSpan(ctx, span),
	})
	return err
}
