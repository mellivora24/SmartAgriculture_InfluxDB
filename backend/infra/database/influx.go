package database

import (
	"backend/internal/shared"
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxDB struct {
	Client   influxdb2.Client
	WriteAPI api.WriteAPI
	QueryAPI api.QueryAPI
	Config   *shared.InfluxDBConfig
}

func NewInfluxDB(config *shared.InfluxDBConfig) (*InfluxDB, error) {
	options := influxdb2.DefaultOptions()
	options.SetBatchSize(uint(config.BatchSize))
	options.SetFlushInterval(uint(config.FlushInterval.Milliseconds()))
	options.SetHTTPRequestTimeout(uint(config.Timeout.Seconds()))

	client := influxdb2.NewClientWithOptions(
		config.URL,
		config.Token,
		options,
	)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	if health.Status != "pass" {
		return nil, fmt.Errorf("InfluxDB health check failed: %s", health.Message)
	}

	writeAPI := client.WriteAPI(config.Org, config.Bucket)
	queryAPI := client.QueryAPI(config.Org)

	log.Printf("InfluxDB connected successfully at %s", config.URL)

	influx := &InfluxDB{
		Client:   client,
		WriteAPI: writeAPI,
		QueryAPI: queryAPI,
		Config:   config,
	}

	go influx.handleWriteErrors()

	return influx, nil
}

func (i *InfluxDB) handleWriteErrors() {
	errorsCh := i.WriteAPI.Errors()
	for err := range errorsCh {
		log.Printf("InfluxDB write error: %v", err)
	}
}

func (i *InfluxDB) WritePoint(point *write.Point) {
	i.WriteAPI.WritePoint(point)
}

func (i *InfluxDB) WriteBatch(points []*write.Point) {
	for _, point := range points {
		i.WriteAPI.WritePoint(point)
	}
}

func (i *InfluxDB) Flush() {
	i.WriteAPI.Flush()
}

func (i *InfluxDB) Query(ctx context.Context, query string) (interface{}, error) {
	result, err := i.QueryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer result.Close()

	var records []map[string]interface{}
	for result.Next() {
		record := make(map[string]interface{})
		for key, value := range result.Record().Values() {
			record[key] = value
		}
		records = append(records, record)
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("query iteration error: %w", result.Err())
	}

	return records, nil
}

func (i *InfluxDB) QueryLatestData(ctx context.Context, measurement string, limit int) (interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "%s")
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: %d)
	`, i.Config.Bucket, measurement, limit)

	return i.Query(ctx, query)
}

func (i *InfluxDB) QueryByTimeRange(ctx context.Context, measurement string, start, stop time.Time) (interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> sort(columns: ["_time"], desc: false)
	`, i.Config.Bucket, start.Format(time.RFC3339), stop.Format(time.RFC3339), measurement)

	return i.Query(ctx, query)
}

func (i *InfluxDB) QueryAggregation(ctx context.Context, measurement, field, aggregation string, window time.Duration) (interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "%s" and r._field == "%s")
			|> aggregateWindow(every: %s, fn: %s, createEmpty: false)
			|> yield(name: "%s")
	`, i.Config.Bucket, measurement, field, window.String(), aggregation, aggregation)

	return i.Query(ctx, query)
}

func (i *InfluxDB) Close() {
	i.WriteAPI.Flush()
	i.Client.Close()

	log.Println("InfluxDB connection closed")
}

func (i *InfluxDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), i.Config.Timeout)
	defer cancel()

	health, err := i.Client.Health(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if health.Status != "pass" {
		return fmt.Errorf("health check failed: %s", health.Message)
	}

	return nil
}

func (i *InfluxDB) CreatePoint(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) *write.Point {
	return write.NewPoint(measurement, tags, fields, ts)
}
