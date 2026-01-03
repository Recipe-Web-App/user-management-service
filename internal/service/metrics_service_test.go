package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dto_model "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// mockGatherer implements prometheus.Gatherer interface.
type mockGatherer struct {
	mfs []*dto_model.MetricFamily
}

func (m *mockGatherer) Gather() ([]*dto_model.MetricFamily, error) {
	return m.mfs, nil
}

type mockSystemCollector struct {
	cpuPercent  float64
	cpuErr      error
	memInfo     *mem.VirtualMemoryStat
	memErr      error
	diskUsage   *disk.UsageStat
	diskErr     error
	processInfo *dto.ProcessInfo
	processErr  error
}

func (m *mockSystemCollector) GetCPUPercent() (float64, error) {
	return m.cpuPercent, m.cpuErr
}

func (m *mockSystemCollector) GetMemoryInfo() (*mem.VirtualMemoryStat, error) {
	return m.memInfo, m.memErr
}

func (m *mockSystemCollector) GetDiskUsage() (*disk.UsageStat, error) {
	return m.diskUsage, m.diskErr
}

func (m *mockSystemCollector) GetProcessInfo() (*dto.ProcessInfo, error) {
	return m.processInfo, m.processErr
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
		redis:    &mockRedisClient{},
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
		redis:    &mockRedisClient{},
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

type mockRedisClient struct {
	metrics     *dto.CacheMetricsResponse
	err         error
	healthStats map[string]string
}

func (m *mockRedisClient) GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error) {
	return m.metrics, m.err
}

func (m *mockRedisClient) Health(ctx context.Context) map[string]string {
	return m.healthStats
}

func TestMetricsServiceGetCacheMetrics(t *testing.T) {
	t.Parallel()

	mockRedis := &mockRedisClient{
		metrics: &dto.CacheMetricsResponse{
			KeysCount:        100,
			MemoryUsage:      "1024",
			MemoryUsageHuman: "1KB",
		},
	}

	db, _, _ := sqlmock.New()

	defer func() {
		_ = db.Close()
	}()

	dbService := database.NewWithDB(db)

	svc := NewMetricsService(dbService, mockRedis, &mockSystemCollector{}, nil)

	metrics, err := svc.GetCacheMetrics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 100, metrics.KeysCount)
	assert.Equal(t, "1KB", metrics.MemoryUsageHuman)
}

func TestMetricsServiceGetCacheMetricsError(t *testing.T) {
	t.Parallel()

	mockRedis := &mockRedisClient{
		err: assert.AnError,
	}

	db, _, _ := sqlmock.New()

	defer func() {
		_ = db.Close()
	}()

	dbService := database.NewWithDB(db)

	svc := NewMetricsService(dbService, mockRedis, &mockSystemCollector{}, nil)

	metrics, err := svc.GetCacheMetrics(context.Background())
	require.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, metrics)
}

func TestMetricsServiceGetCacheMetricsNoRedis(t *testing.T) {
	t.Parallel()

	db, _, _ := sqlmock.New()

	defer func() {
		_ = db.Close()
	}()

	dbService := database.NewWithDB(db)

	// Pass nil as redis client
	svc := NewMetricsService(dbService, nil, &mockSystemCollector{}, nil)

	metrics, err := svc.GetCacheMetrics(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis service not available")
	assert.Nil(t, metrics)
}

func TestMetricsServiceGetSystemMetrics(t *testing.T) {
	t.Parallel()

	mockSys := &mockSystemCollector{
		cpuPercent: 25.5,
		memInfo: &mem.VirtualMemoryStat{
			Total:       16 * 1024 * 1024 * 1024,
			Used:        8 * 1024 * 1024 * 1024,
			UsedPercent: 50.0,
		},
		diskUsage: &disk.UsageStat{
			Total:       500 * 1024 * 1024 * 1024,
			Used:        100 * 1024 * 1024 * 1024,
			UsedPercent: 20.0,
		},
		// gopsutil process.Process has private fields, so hard to mock "filled" process without finding a real PID.
		// For unit test safety, we'll return nil for process to avoid flaky PID lookups,
		// or rely on limited test coverage for that specific helper.
		processInfo: nil,
	}

	db, _, _ := sqlmock.New()

	defer func() { _ = db.Close() }()

	dbService := database.NewWithDB(db)

	svc := NewMetricsService(dbService, &mockRedisClient{}, mockSys, nil)

	metrics, err := svc.GetSystemMetrics(context.Background())
	require.NoError(t, err)

	assert.InDelta(t, 25.5, metrics.System.CPUUsagePercent, 0.01)
	assert.InDelta(t, 16.0, metrics.System.MemoryTotalGB, 0.01)
	assert.InDelta(t, 50.0, metrics.System.MemoryUsagePercent, 0.01)
	assert.InDelta(t, 500.0, metrics.System.DiskTotalGB, 0.01)
}

func BenchmarkMetricsServiceGetSystemMetrics(b *testing.B) {
	mockSys := &mockSystemCollector{
		cpuPercent: 25.5,
		memInfo: &mem.VirtualMemoryStat{
			Total:       16 * 1024 * 1024 * 1024,
			Used:        8 * 1024 * 1024 * 1024,
			UsedPercent: 50.0,
		},
		diskUsage: &disk.UsageStat{
			Total:       500 * 1024 * 1024 * 1024,
			Used:        100 * 1024 * 1024 * 1024,
			UsedPercent: 20.0,
		},
	}

	db, _, _ := sqlmock.New()

	defer func() { _ = db.Close() }()

	dbService := database.NewWithDB(db)

	svc := NewMetricsService(dbService, &mockRedisClient{}, mockSys, nil)
	ctx := context.Background()

	for b.Loop() {
		_, _ = svc.GetSystemMetrics(ctx)
	}
}

func TestMetricsServiceGetDetailedHealthMetrics(t *testing.T) {
	t.Parallel()

	mockRedis := &mockRedisClient{
		healthStats: map[string]string{
			"status": "up",
		},
		metrics: &dto.CacheMetricsResponse{
			ConnectedClients: 10,
			HitRate:          0.8,
			MemoryUsageHuman: "128MB",
		},
	}

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	mock.ExpectPing()

	defer func() { _ = db.Close() }()

	dbService := database.NewWithDB(db)
	cfg := &config.Config{
		Environment: "test-env",
	}

	svc := NewMetricsService(dbService, mockRedis, &mockSystemCollector{}, cfg)

	metrics, err := svc.GetDetailedHealthMetrics(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Check Application Info
	assert.Equal(t, "1.0.0", metrics.Application.Version)
	assert.Equal(t, "test-env", metrics.Application.Environment)

	// Check Redis Health
	assert.Equal(t, "healthy", metrics.Services.Redis.Status)
	assert.Equal(t, 10, metrics.Services.Redis.ConnectedClients)
	assert.InDelta(t, 80.0, metrics.Services.Redis.HitRatePercent, 0.00)
	assert.Equal(t, "128MB", metrics.Services.Redis.MemoryUsage)

	// Check DB Health
	assert.Equal(t, "healthy", metrics.Services.Database.Status)
	assert.GreaterOrEqual(t, metrics.Services.Database.ResponseTimeMs, 0.0)

	// Check Overall Status
	assert.Equal(t, "healthy", metrics.OverallStatus)
}
