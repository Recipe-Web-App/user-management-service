package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dto_model "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
)

// mockGatherer implements prometheus.Gatherer interface.
type mockGatherer struct {
	mfs []*dto_model.MetricFamily
}

func (m *mockGatherer) Gather() ([]*dto_model.MetricFamily, error) {
	return m.mfs, nil
}

func TestMetricsServiceGetPerformanceMetrics(t *testing.T) {
	t.Parallel()

	// Setup DB mock
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	dbService := database.NewWithDB(db)

	mfTotal, mfDuration := createSampleMetricFamilies()
	mockG := &mockGatherer{
		mfs: []*dto_model.MetricFamily{mfTotal, mfDuration},
	}

	// Inject dependencies directly since we are in the same package
	svc := &metricsService{
		db:       dbService,
		gatherer: mockG,
	}

	// Execute
	metrics, err := svc.GetPerformanceMetrics(context.Background())

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Total Requests: 10 + 2 = 12
	assert.Equal(t, 12, metrics.RequestCounts.TotalRequests)

	// Errors: 2
	assert.Equal(t, 2, metrics.ErrorRates.TotalErrors)
	assert.Equal(t, 2, metrics.ErrorRates.Errors5xx)
	assert.Equal(t, 0, metrics.ErrorRates.Errors4xx)

	// Error Rate: 2/12 = 16.66%
	assert.InDelta(t, 16.666, metrics.ErrorRates.ErrorRatePercent, 0.01)

	// Average Latency: 5.0 / 10 = 0.5s = 500ms
	assert.InDelta(t, 500.0, metrics.ResponseTimes.AverageMs, 0.1)

	// Quantiles
	assert.True(t, metrics.ResponseTimes.P50Ms > 100 && metrics.ResponseTimes.P50Ms <= 500)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsServiceGetPerformanceMetricsEmpty(t *testing.T) {
	t.Parallel()

	db, _, _ := sqlmock.New()

	defer func() {
		_ = db.Close()
	}()

	dbService := database.NewWithDB(db)

	mockG := &mockGatherer{
		mfs: []*dto_model.MetricFamily{},
	}

	svc := &metricsService{
		db:       dbService,
		gatherer: mockG,
	}

	metrics, err := svc.GetPerformanceMetrics(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 0, metrics.RequestCounts.TotalRequests)
	assert.Equal(t, 0, metrics.ErrorRates.TotalErrors)
	assert.InDelta(t, 0.0, metrics.ResponseTimes.AverageMs, 0.001)
}

func createSampleMetricFamilies() (*dto_model.MetricFamily, *dto_model.MetricFamily) {
	mfTotal := &dto_model.MetricFamily{
		Name: proto.String("user_management_http_requests_total"),
		Type: dto_model.MetricType_COUNTER.Enum(),
		Metric: []*dto_model.Metric{
			{
				Label: []*dto_model.LabelPair{
					{Name: proto.String("status"), Value: proto.String("200")},
				},
				Counter: &dto_model.Counter{Value: proto.Float64(10)},
			},
			{
				Label: []*dto_model.LabelPair{
					{Name: proto.String("status"), Value: proto.String("500")},
				},
				Counter: &dto_model.Counter{Value: proto.Float64(2)},
			},
		},
	}

	mfDuration := &dto_model.MetricFamily{
		Name: proto.String("user_management_http_request_duration_seconds"),
		Type: dto_model.MetricType_HISTOGRAM.Enum(),
		Metric: []*dto_model.Metric{
			{
				Histogram: &dto_model.Histogram{
					SampleCount: proto.Uint64(10),
					SampleSum:   proto.Float64(5.0), // 0.5s avg
					Bucket: []*dto_model.Bucket{
						{UpperBound: proto.Float64(0.1), CumulativeCount: proto.Uint64(2)},
						{UpperBound: proto.Float64(0.5), CumulativeCount: proto.Uint64(6)},
						{UpperBound: proto.Float64(1.0), CumulativeCount: proto.Uint64(9)},
						{UpperBound: proto.Float64(10.0), CumulativeCount: proto.Uint64(10)},
					},
				},
			},
		},
	}

	return mfTotal, mfDuration
}
