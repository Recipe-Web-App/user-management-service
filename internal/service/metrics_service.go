package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto_model "github.com/prometheus/client_model/go"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

const (
	percentMultiplier = 100.0
	msMultiplier      = 1000.0
	quantileP50       = 0.50
	quantileP95       = 0.95
	quantileP99       = 0.99
	gbDivisor         = 1024 * 1024 * 1024
	mbDivisor         = 1024 * 1024
)

// MetricsService handles metrics gathering.
type MetricsService interface {
	GetPerformanceMetrics(ctx context.Context) (*dto.PerformanceMetricsResponse, error)
	GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error)
	GetSystemMetrics(ctx context.Context) (*dto.SystemMetricsResponse, error)
}

// SystemCollector defines the interface for gathering system metrics.
type SystemCollector interface {
	GetCPUPercent() (float64, error)
	GetMemoryInfo() (*mem.VirtualMemoryStat, error)
	GetDiskUsage() (*disk.UsageStat, error)
	GetProcessInfo() (*dto.ProcessInfo, error)
}

// RedisClient defines the interface for interacting with Redis cache metrics.
type RedisClient interface {
	GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error)
}

type metricsService struct {
	db        *database.Service
	redis     RedisClient
	sys       SystemCollector
	gatherer  prometheus.Gatherer
	startTime time.Time
}

// NewMetricsService creates a new metrics service.
func NewMetricsService(db *database.Service, redis RedisClient, sys SystemCollector) MetricsService {
	return &metricsService{
		db:        db,
		redis:     redis,
		sys:       sys,
		gatherer:  prometheus.DefaultGatherer,
		startTime: time.Now(),
	}
}

// gopsutilCollector implements SystemCollector using gopsutil.
type gopsutilCollector struct {
	proc *process.Process
	mu   sync.Mutex
}

// NewSystemCollector creates a new system collector.
func NewSystemCollector() SystemCollector {
	return &gopsutilCollector{}
}

func (c *gopsutilCollector) GetCPUPercent() (float64, error) {
	percentages, err := cpu.Percent(0, false)
	if err != nil || len(percentages) == 0 {
		return 0, fmt.Errorf("failed to get cpu percent: %w", err)
	}

	return percentages[0], nil
}

func (c *gopsutilCollector) GetMemoryInfo() (*mem.VirtualMemoryStat, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	return stat, nil
}

func (c *gopsutilCollector) GetDiskUsage() (*disk.UsageStat, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	return stat, nil
}

func (c *gopsutilCollector) GetProcessInfo() (*dto.ProcessInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.proc == nil {
		proc, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		c.proc = proc
	}

	memInfo, _ := c.proc.MemoryInfo()
	cpuPercent, _ := c.proc.Percent(0)
	numThreads, _ := c.proc.NumThreads()
	openFiles, _ := c.proc.OpenFiles()

	info := &dto.ProcessInfo{
		CPUPercent: cpuPercent,
		NumThreads: int(numThreads),
		OpenFiles:  len(openFiles),
	}

	if memInfo != nil {
		info.MemoryRSSMB = float64(memInfo.RSS) / mbDivisor
		info.MemoryVMSMB = float64(memInfo.VMS) / mbDivisor
	}

	return info, nil
}

// GetPerformanceMetrics aggregates metrics from Prometheus and DB.
func (s *metricsService) GetPerformanceMetrics(ctx context.Context) (*dto.PerformanceMetricsResponse, error) {
	mfs, err := s.gatherer.Gather()
	if err != nil {
		return nil, fmt.Errorf("failed to gather metrics: %w", err)
	}

	metrics := &dto.PerformanceMetricsResponse{}

	stats := processPrometheusMetrics(mfs)

	metrics.RequestCounts = calculateRequestCounts(stats)
	metrics.ErrorRates = calculateErrorRates(stats)
	metrics.ResponseTimes = calculateResponseTimes(stats)
	metrics.Database = s.getDatabaseMetrics()

	return metrics, nil
}

// ErrRedisServiceUnavailable is returned when the Redis service is nil or not configured.
var ErrRedisServiceUnavailable = errors.New("redis service not available")

// ErrSystemCollectorUnavailable is returned when the system collector is nil or not configured.
var ErrSystemCollectorUnavailable = errors.New("system collector not available")

// GetCacheMetrics retrieves cache metrics from the Redis service.
func (s *metricsService) GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error) {
	if s.redis == nil {
		return nil, ErrRedisServiceUnavailable
	}

	metrics, err := s.redis.GetCacheMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache metrics: %w", err)
	}

	return metrics, nil
}

type metricStats struct {
	totalRequests float64
	totalErrors   float64
	errors4xx     float64
	errors5xx     float64
	durationSum   float64
	durationCount float64
	buckets       map[float64]uint64
}

func processPrometheusMetrics(mfs []*dto_model.MetricFamily) metricStats {
	var stats metricStats

	stats.buckets = make(map[float64]uint64)

	for _, mf := range mfs {
		switch mf.GetName() {
		case "user_management_http_requests_total":
			processRequestTotals(mf, &stats)
		case "user_management_http_request_duration_seconds":
			processRequestDurations(mf, &stats)
		}
	}

	return stats
}

func processRequestTotals(mf *dto_model.MetricFamily, stats *metricStats) {
	for _, m := range mf.GetMetric() {
		val := m.GetCounter().GetValue()
		stats.totalRequests += val

		for _, label := range m.GetLabel() {
			if label.GetName() == "status" {
				statusCode := label.GetValue()
				if strings.HasPrefix(statusCode, "4") {
					stats.totalErrors += val
					stats.errors4xx += val
				} else if strings.HasPrefix(statusCode, "5") {
					stats.totalErrors += val
					stats.errors5xx += val
				}

				break
			}
		}
	}
}

func processRequestDurations(mf *dto_model.MetricFamily, stats *metricStats) {
	for _, m := range mf.GetMetric() {
		h := m.GetHistogram()
		stats.durationSum += h.GetSampleSum()

		stats.durationCount += float64(h.GetSampleCount())
		for _, b := range h.GetBucket() {
			stats.buckets[b.GetUpperBound()] += b.GetCumulativeCount()
		}
	}
}

func calculateRequestCounts(stats metricStats) dto.RequestCounts {
	return dto.RequestCounts{
		TotalRequests: int(stats.totalRequests),
	}
}

func calculateErrorRates(stats metricStats) dto.ErrorRates {
	rate := 0.0
	if stats.totalRequests > 0 {
		rate = (stats.totalErrors / stats.totalRequests) * percentMultiplier
	}

	return dto.ErrorRates{
		TotalErrors:      int(stats.totalErrors),
		ErrorRatePercent: rate,
		Errors4xx:        int(stats.errors4xx),
		Errors5xx:        int(stats.errors5xx),
	}
}

func calculateResponseTimes(stats metricStats) dto.ResponseTimes {
	avg := 0.0
	if stats.durationCount > 0 {
		avg = (stats.durationSum / stats.durationCount) * msMultiplier
	}

	sortedBuckets := getSortedBuckets(stats.buckets)

	return dto.ResponseTimes{
		AverageMs: avg,
		P50Ms:     calculateQuantile(quantileP50, sortedBuckets, stats.buckets, stats.durationCount) * msMultiplier,
		P95Ms:     calculateQuantile(quantileP95, sortedBuckets, stats.buckets, stats.durationCount) * msMultiplier,
		P99Ms:     calculateQuantile(quantileP99, sortedBuckets, stats.buckets, stats.durationCount) * msMultiplier,
	}
}

func (s *metricsService) getDatabaseMetrics() dto.DatabaseMetrics {
	dbStats := s.db.GetDB().Stats()

	return dto.DatabaseMetrics{
		ActiveConnections: dbStats.OpenConnections,
		MaxConnections:    dbStats.MaxOpenConnections,
	}
}

func getSortedBuckets(buckets map[float64]uint64) []float64 {
	keys := make([]float64, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}

	sort.Float64s(keys)

	return keys
}

// calculateQuantile estimates the quantile from histogram buckets using linear interpolation.
func calculateQuantile(q float64, sortedBuckets []float64, buckets map[float64]uint64, totalCount float64) float64 {
	if totalCount == 0 || len(sortedBuckets) == 0 {
		return 0
	}

	rank := q * totalCount

	var (
		prevBound float64 = 0
		prevCount uint64  = 0
	)

	for _, bound := range sortedBuckets {
		count := buckets[bound]

		if float64(count) >= rank {
			countDiff := float64(count - prevCount)
			if countDiff == 0 {
				return bound
			}

			// Linear interpolation
			rankDiff := rank - float64(prevCount)
			fraction := rankDiff / countDiff

			return prevBound + (bound-prevBound)*fraction
		}

		prevBound = bound
		prevCount = count
	}

	return sortedBuckets[len(sortedBuckets)-1]
}

// GetSystemMetrics retrieves system resource metrics.
func (s *metricsService) GetSystemMetrics(ctx context.Context) (*dto.SystemMetricsResponse, error) {
	if s.sys == nil {
		return nil, ErrSystemCollectorUnavailable
	}

	cpuPercent, _ := s.sys.GetCPUPercent()
	memInfo, _ := s.sys.GetMemoryInfo()
	diskInfo, _ := s.sys.GetDiskUsage()
	procInfo, _ := s.sys.GetProcessInfo()

	response := &dto.SystemMetricsResponse{
		Timestamp: time.Now(),
		System: dto.SystemInfo{
			CPUUsagePercent: cpuPercent,
		},
		UptimeSeconds: int(time.Since(s.startTime).Seconds()),
	}

	if memInfo != nil {
		response.System.MemoryTotalGB = float64(memInfo.Total) / gbDivisor
		response.System.MemoryUsedGB = float64(memInfo.Used) / gbDivisor
		response.System.MemoryUsagePercent = memInfo.UsedPercent
	}

	if diskInfo != nil {
		response.System.DiskTotalGB = float64(diskInfo.Total) / gbDivisor
		response.System.DiskUsedGB = float64(diskInfo.Used) / gbDivisor
		response.System.DiskUsagePercent = diskInfo.UsedPercent
	}

	if procInfo != nil {
		response.Process = *procInfo
	}

	return response, nil
}
