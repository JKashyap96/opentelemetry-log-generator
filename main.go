package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Config represents the configuration for the exporter
type Config struct {
	Endpoint string
	Insecure bool
	Headers  map[string]string
}

// cryptoRandIntn generates a cryptographically secure random integer in the range [0, max)
func cryptoRandIntn(max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(fmt.Sprintf("failed to generate random number: %v", err))
	}
	return int(nBig.Int64())
}

// randomSeverity returns a random log severity and its corresponding text
func randomSeverity() (log.Severity, string) {
	severities := []struct {
		level log.Severity
		text  string
	}{
		{log.SeverityTrace1, "Trace"},
		{log.SeverityDebug, "Debug"},
		{log.SeverityInfo, "Info"},
		{log.SeverityWarn, "Warn"},
		{log.SeverityError, "Error"},
		{log.SeverityFatal, "Fatal"},
	}
	randomIdx := cryptoRandIntn(len(severities))
	return severities[randomIdx].level, severities[randomIdx].text
}

// randomSentence generates a random meaningful sentence with operational and observability terms
func randomSentence() string {
	// Define possible sentence structures
	structures := []string{
		"Service %s encountered a %s error while processing %s request in %s.",
		"The %s for service %s exceeded threshold with %s %s in %s.",
		"Failed to process request from %s due to %s in %s.",
		"System detected %s failure in %s service affecting %s component. Status: %s.",
		"Health check failed for %s. Status: %s, Error: %s.",
		"Service %s reported %s latency. Monitoring alert triggered in %s component. Metric: %s.",
	}

	// Define possible terms to fill the sentence
	services := []string{"service_A", "service_B", "service_C", "service_D", "service_E", "service_F", "service_G", "service_H"}
	components := []string{"container", "pod", "node", "service", "database", "network"}
	errors := []string{"timeout", "failure", "connection", "unavailable", "high_load", "out_of_memory"}
	statuses := []string{"critical", "warning", "degraded", "normal"}
	metrics := []string{"response_time", "error_rate", "request_count", "latency", "throughput"}
	actors := []string{"client", "user", "admin", "system", "service"}

	// Pick a random sentence structure
	sentenceTemplate := structures[cryptoRandIntn(len(structures))]

	// Randomly fill the structure with terms
	// Ensure the order of arguments matches the placeholders in the template
	switch sentenceTemplate {
	case "Service %s encountered a %s error while processing %s request in %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			errors[cryptoRandIntn(len(errors))],
			actors[cryptoRandIntn(len(actors))],
			components[cryptoRandIntn(len(components))],
		)
	case "The %s for service %s exceeded threshold with %s %s in %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			components[cryptoRandIntn(len(components))],
			services[cryptoRandIntn(len(services))],
			statuses[cryptoRandIntn(len(statuses))],
			metrics[cryptoRandIntn(len(metrics))],
			components[cryptoRandIntn(len(components))],
		)
	case "Failed to process request from %s due to %s in %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			actors[cryptoRandIntn(len(actors))],
			errors[cryptoRandIntn(len(errors))],
			components[cryptoRandIntn(len(components))],
		)
	case "System detected %s failure in %s service affecting %s component. Status: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			errors[cryptoRandIntn(len(errors))],
			services[cryptoRandIntn(len(services))],
			components[cryptoRandIntn(len(components))],
			statuses[cryptoRandIntn(len(statuses))],
		)
	case "Health check failed for %s. Status: %s, Error: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			statuses[cryptoRandIntn(len(statuses))],
			errors[cryptoRandIntn(len(errors))],
		)
	case "Service %s reported %s latency. Monitoring alert triggered in %s component. Metric: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			errors[cryptoRandIntn(len(errors))],
			components[cryptoRandIntn(len(components))],
			metrics[cryptoRandIntn(len(metrics))],
		)
	default:
		// Fallback if none of the templates match
		return "Default message"
	}
}

// generateLog generates a log with random details and sends it to the logger
func generateLog(i int, phase string, logger log.Logger) {
	// Randomize severity and text
	severity, severityText := randomSeverity()

	// Create the log body with multiple variations
	logBody := fmt.Sprintf(
		"Log %d: %s phase: %s | Message : %s | UUID: %s | Timestamp: %s",
		i, severityText, phase, randomSentence(), uuid.NewString(), time.Now().Format(time.RFC3339),
	)

	// Create the log record
	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetObservedTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(severityText)
	record.SetBody(log.StringValue(logBody))

	// Add attributes directly to the record
	attrs := []log.KeyValue{
		log.String("worker_id", fmt.Sprintf("%d", i)),
		log.String("phase", phase),
		log.String("http.target", fmt.Sprintf("/api/v1/resource/%d", i)),
		log.String("k8s.namespace.name", "default"),
		log.String("k8s.container.name", opentelemetry-logs-generator"),
	}
	record.AddAttributes(attrs...)
	logger.Emit(context.Background(), record)

	// Optionally, print the log record to the console for debugging or verification
	fmt.Printf("Body: %s\n", record.Body())
}

// createExporter creates a new OTLP gRPC exporter based on the provided config
func createExporter(c *Config) (logsdk.Exporter, error) {
	ctx := context.Background()
	var exp logsdk.Exporter
	var err error

	// Create gRPC exporter
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(c.Endpoint),
	}
	if c.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(c.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(c.Headers))
	}
	exp, err = otlploggrpc.New(ctx, opts...)

	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP gRPC exporter: %w", err)
	}

	return exp, nil
}

func main() {
	// Define a flag for the exporter endpoint
	endpointFlag := flag.String("exporter-endpoint", "0.0.0.0:4317", "gRPC exporter endpoint")

	// Parse the flags
	flag.Parse()

	// Configure exporter (using gRPC)
	config := &Config{
		Endpoint: *endpointFlag,       // Use the endpoint passed as a flag
		Insecure: true,                // Use insecure connection
		Headers:  map[string]string{}, // Optional headers
	}

	// Create the exporter
	exporter, err := createExporter(config)
	if err != nil {
		fmt.Printf("Failed to create exporter: %v\n", err)
		os.Exit(1)
	}
	defer exporter.Shutdown(context.Background())

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.HostNameKey.String("node-1"),
	)

	// Create the batch processor
	batchProcessor := logsdk.NewBatchProcessor(exporter,
		logsdk.WithMaxQueueSize(2048),
		logsdk.WithExportMaxBatchSize(512),
		logsdk.WithExportInterval(1*time.Second),
	)

	// Create the logger provider with the batch processor
	loggerProvider := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(batchProcessor),
		logsdk.WithResource(res),
	)
	logger := loggerProvider.Logger("opentelemetry-logger")

	// Create a channel to listen for termination signals (SIGINT, SIGTERM)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Create a context that will be canceled on signal
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start generating logs in a loop in a separate goroutine
	go func() {
		for i := 0; ; i++ {
			// Use different phases for demonstration
			var phase string
			switch i % 5 {
			case 0:
				phase = "start"
			case 1:
				phase = "processing"
			case 2:
				phase = "queued"
			case 3:
				phase = "completed"
			case 4:
				phase = "error"
			}

			generateLog(i, phase, logger)

			// Sleep for a smaller duration to speed up log generation
			time.Sleep(100 * time.Nanosecond) // Adjust the log generation frequency
		}
	}()

	// Wait for termination signal (SIGINT or SIGTERM)
	<-signalChan
	fmt.Println("\nReceived termination signal, shutting down gracefully...")

	// Cleanup or flush any final logs (if necessary)
	cancel()

	// Ensure the logger is properly flushed if needed
	fmt.Println("Shutdown complete.")
}
