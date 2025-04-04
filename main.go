ackage main

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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Config struct {
	Endpoint string
	Insecure bool
	Headers  map[string]string
}

func cryptoRandIntn(max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(fmt.Sprintf("failed to generate random number: %v", err))
	}
	return int(nBig.Int64())
}

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

func randomSentence() string {
	// Define possible sentence structures
	structures := []string{
		"Service %s encountered a %s error while processing %s request in %s. Detail: %s.",
		"The %s for service %s exceeded threshold with %s %s in %s. Action: %s.",
		"Failed to process request from %s due to %s in %s. Immediate action required.",
		"System detected %s failure in %s service affecting %s component. Status: %s, Action: %s.",
		"Health check failed for %s. Status: %s, Error: %s. Issue reported to %s.",
		"Service %s reported %s latency. Monitoring alert triggered in %s component. Metric: %s. Resolution: %s.",
		"Critical error in %s service. %s failure caused by %s. %s status reported. %s.",
	}

	// Define possible terms to fill the sentence
	services := []string{"app-api", "app-manager", "connector", "device-context", "ingestion", "location", "decoder", "reporter", "reporter", "wifi-context", "teleport"}
	components := []string{"container", "pod", "node", "service", "database", "network", "API", "filesystem", "cache", "auth"}
	errors := []string{"timeout", "failure", "connection", "unavailable", "high_load", "out_of_memory", "disk_full", "network_error", "service_crash", "data_corruption"}
	statuses := []string{"critical", "warning", "degraded", "normal", "unknown", "resolved", "in_progress", "paused"}
	metrics := []string{"response_time", "error_rate", "request_count", "latency", "throughput", "availability", "cpu_usage", "memory_usage", "disk_io"}
	actors := []string{"client", "user", "admin", "system", "service", "API", "monitoring_tool", "scheduler"}
	actions := []string{"restart", "escalate", "investigate", "resolve", "ignore", "patch", "alert"}
	details := []string{"timeout exceeded", "connection lost", "high CPU usage", "memory leak detected", "data corruption", "service unresponsive"}
	resolutions := []string{"reboot server", "increase timeout", "scale up service", "reboot container", "clear cache", "restore database"}

	// Pick a random sentence structure
	sentenceTemplate := structures[cryptoRandIntn(len(structures))]

	// Randomly fill the structure with terms
	switch sentenceTemplate {
	case "Service %s encountered a %s error while processing %s request in %s. Detail: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			errors[cryptoRandIntn(len(errors))],
			actors[cryptoRandIntn(len(actors))],
			components[cryptoRandIntn(len(components))],
			details[cryptoRandIntn(len(details))],
		)
	case "The %s for service %s exceeded threshold with %s %s in %s. Action: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			components[cryptoRandIntn(len(components))],
			services[cryptoRandIntn(len(services))],
			statuses[cryptoRandIntn(len(statuses))],
			metrics[cryptoRandIntn(len(metrics))],
			components[cryptoRandIntn(len(components))],
			actions[cryptoRandIntn(len(actions))],
		)
	case "Failed to process request from %s due to %s in %s. Immediate action required.":
		return fmt.Sprintf(
			sentenceTemplate,
			actors[cryptoRandIntn(len(actors))],
			errors[cryptoRandIntn(len(errors))],
			components[cryptoRandIntn(len(components))],
		)
	case "System detected %s failure in %s service affecting %s component. Status: %s, Action: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			errors[cryptoRandIntn(len(errors))],
			services[cryptoRandIntn(len(services))],
			components[cryptoRandIntn(len(components))],
			statuses[cryptoRandIntn(len(statuses))],
			actions[cryptoRandIntn(len(actions))],
		)
	case "Health check failed for %s. Status: %s, Error: %s. Issue reported to %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			statuses[cryptoRandIntn(len(statuses))],
			errors[cryptoRandIntn(len(errors))],
			actors[cryptoRandIntn(len(actors))],
		)
	case "Service %s reported %s latency. Monitoring alert triggered in %s component. Metric: %s. Resolution: %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			errors[cryptoRandIntn(len(errors))],
			components[cryptoRandIntn(len(components))],
			metrics[cryptoRandIntn(len(metrics))],
			resolutions[cryptoRandIntn(len(resolutions))],
		)
	case "Critical error in %s service. %s failure caused by %s. %s status reported. %s status action taken. %s.":
		return fmt.Sprintf(
			sentenceTemplate,
			services[cryptoRandIntn(len(services))],
			errors[cryptoRandIntn(len(errors))],
			errors[cryptoRandIntn(len(errors))],
			statuses[cryptoRandIntn(len(statuses))],
			actions[cryptoRandIntn(len(actions))],
			details[cryptoRandIntn(len(details))],
		)
	default:
		// Fallback if none of the templates match
		return "Default message"
	}
}

func generateLog(i int, phase string, logger log.Logger) {
	severity, severityText := randomSeverity()

	jobs := []string{
		"payment-service", "auth-service", "inventory-service", "order-service", "analytics-service",
		"billing-service", "user-service", "recommendation-engine", "fraud-detection", "email-service", "notification-service",
		"search-service", "chat-service", "streaming-service", "identity-service", "geo-location-service",
		"reporting-service", "subscription-service", "workflow-engine", "api-gateway", "rate-limiter",
		"cdn-cache", "image-processor", "task-scheduler", "data-sync", "audit-logger",
		"event-dispatcher", "feature-flag-service", "customer-support", "ads-engine", "session-manager",
		"traffic-shaper", "file-storage", "document-parser", "push-service", "captcha-service",
		"background-worker", "load-balancer", "analytics-pipeline", "search-indexer", "log-aggregator",
		"video-transcoder", "music-recommender", "social-graph-service", "device-registry", "map-renderer",
		"sensor-data-collector", "content-moderation", "news-aggregator", "voice-recognition", "machine-learning-inference",
		"chatbot-engine", "game-matchmaking", "ecommerce-checkout", "cart-service", "external-api-proxy",
		"biometric-authentication", "rewards-program", "partnership-integration", "pdf-generator", "digital-wallet",
		"vehicle-tracking", "compliance-monitoring", "rate-converter", "cloud-metrics", "security-scanner",
		"app-performance-monitoring", "incident-management", "ci-cd-runner", "cloud-orchestrator", "hardware-provisioning",
		"iot-command-center", "remote-access-service", "content-delivery", "data-cleaning", "graphql-service",
		"timeseries-database", "iot-event-stream", "blockchain-node", "crypto-trading-bot", "low-latency-streaming",
		"identity-verification", "synthetic-monitoring", "feedback-collector", "a-b-testing", "access-control",
		"edge-computing", "cyber-threat-detection", "continuous-integration", "hyperledger-fabric", "voice-assistant",
		"natural-language-processing", "deep-learning-service", "image-classification", "cloud-storage", "object-detection",
		"handwriting-recognition", "enterprise-resource-planning", "kubernetes-controller", "financial-reporting", "database-migration",
	}

	containers := []string{
		"nginx-container", "redis-cache", "mysql-db", "postgres-db", "mongodb-primary",
		"elasticsearch-node", "kafka-broker", "rabbitmq-queue", "zookeeper-node", "haproxy",
		"vault-secrets", "consul-agent", "etcd-cluster", "fluentd-logger", "logstash-processor",
		"filebeat-agent", "metricbeat-monitor", "prometheus-server", "grafana-dashboard", "thanos-store",
		"jaeger-tracer", "zipkin-collector", "opentelemetry-collector", "istio-ingress", "linkerd-proxy",
		"coredns-resolver", "calico-networking", "cilium-security", "flannel-overlay", "weave-net",
		"traefik-edge-router", "envoy-sidecar", "istiod-control-plane", "knative-serving", "argo-workflows",
		"tekton-pipeline", "helm-controller", "kustomize-deployer", "keda-scaler", "cert-manager",
		"external-dns", "kyverno-policy", "gatekeeper-admission", "kubeflow-training", "mlflow-tracker",
		"spark-operator", "airflow-scheduler", "dagster-executor", "prefect-agent", "nifi-dataflow",
		"dagster-orchestrator", "drone-ci-runner", "tekton-task", "argocd-repo-server", "kubebuilder-manager",
		"crossplane-runtime", "velero-backup", "rook-ceph", "openebs-control-plane", "longhorn-manager",
		"metallb-speaker", "csi-driver", "openfaas-gateway", "kubeless-function", "knative-eventing",
		"chaos-mesh", "litmuschaos-runner", "falco-security", "sysdig-agent", "twistlock-defender",
		"aquasec-scanner", "snyk-monitor", "trivy-scanner", "anchore-engine", "clair-vulnerability-scanner",
		"stackrox-sensor", "sonarqube-analyzer", "codeql-scanner", "cucumber-test-runner", "gatling-load-tester",
		"jmeter-benchmark", "locust-performance", "k6-load-tester", "selenium-webdriver", "cypress-automation",
		"playwright-test-runner", "puppeteer-tester", "behat-bdd-runner", "postman-api-tester", "soapui-service",
		"newrelic-agent", "datadog-monitor", "instana-tracer", "appdynamics-collector", "splunk-forwarder",
		"sumologic-ingest", "loggly-receiver", "papertrail-logger", "humio-indexer", "graylog-ingest",
	}

	instances := []string{}
	for idx := 1; idx <= 100; idx++ {
		instances = append(instances, fmt.Sprintf("instance-%03d", idx)) // instance-001 to instance-100
	}

	regions := []string{
		"us-east-1", "us-west-2", "eu-central-1", "ap-south-1", "sa-east-1",
		"us-east-2", "eu-west-1", "ap-northeast-1", "ap-southeast-2", "ca-central-1",
		"us-west-1", "eu-west-2", "ap-northeast-2", "ap-southeast-1", "eu-north-1",
		"af-south-1", "me-south-1", "sa-east-2", "ap-east-1", "ap-south-2",
		"eu-central-2", "us-gov-west-1", "us-gov-east-1", "cn-north-1", "cn-northwest-1",
		"ap-northeast-3", "us-west-3", "us-west-4", "eu-west-3", "eu-south-1",
		"ap-southeast-3", "ap-southeast-4", "us-east-3", "us-east-4", "us-west-5",
	}

	namespaces := []string{
		"production", "staging", "development", "testing", "sandbox",
		"qa", "pre-production", "demo", "research", "training",
		"legacy", "internal", "customer-facing", "backup", "dr",
		"cloud", "edge", "ops", "monitoring", "analytics",
		"marketing", "sales", "finance", "hr", "it",
		"management", "compliance", "security", "business", "support",
		"api", "mobile", "web", "devops", "platform",
		"service1", "service2", "service3", "service4", "service5",
		"ci-cd", "prod-backup", "pre-prod", "staging-internal", "data-warehouse",
	}

	instanceTypes := []string{
		"t2.micro", "t2.medium", "m5.large", "m5.xlarge", "r5.large",
		"c5.2xlarge", "m5a.large", "t3.micro", "c5.large", "c5.xlarge",
	}

	labels := []string{
		"web", "mobile", "api", "devops", "analytics",
		"frontend", "backend", "database", "auth", "logging",
		"monitoring", "storage", "networking", "messaging", "streaming",
		"security", "performance", "scaling", "cloud", "on-prem",
		"ci-cd", "testing", "staging", "production", "sandbox",
		"support", "maintenance", "debugging", "release", "migration",
		"load-balancer", "proxy", "gateway", "data", "batch",
		"real-time", "business-intelligence", "big-data", "machine-learning", "ai",
		"edge", "iot", "computing", "scheduling", "infra",
		"deployment", "orchestration", "container", "kubernetes", "docker",
		"serverless", "functions", "api-gateway", "graphql", "webhook",
		"client", "user", "service", "container-runtime", "docker-image",
		"data-pipeline", "event-driven", "realtime-stream", "data-lake", "data-warehouse",
		"admin", "support-ticket", "incident", "alert", "backup",
		"configuration", "network-policy", "firewall", "service-mesh", "distributed-system",
	}

	versions := []string{
		"1.0.0", "1.1.0", "1.2.0", "1.3.0", "2.0.0",
		"2.1.0", "2.2.0", "3.0.0", "3.1.0", "3.2.0",
		"4.0.0", "4.1.0", "5.0.0", "5.1.0", "6.0.0",
		"6.1.0", "7.0.0", "7.1.0", "8.0.0", "8.1.0",
		"9.0.0", "9.1.0", "10.0.0", "10.1.0", "11.0.0",
		"12.0.0", "13.0.0", "14.0.0", "15.0.0", "1.0.1",
		"1.1.1", "1.2.1", "2.0.1", "2.1.1", "2.2.1",
		"3.0.1", "3.1.1", "3.2.1", "4.0.1", "5.0.1",
		"6.0.1", "7.0.1", "8.0.1", "9.0.1", "10.0.1",
		"1.2.0-beta", "2.0.0-alpha", "3.0.0-RC", "4.0.0-stable", "5.0.0-latest",
		"1.0.0-preview", "2.1.0-alpha", "3.0.1-beta", "4.1.0-stable", "5.1.0-release",
	}

	availabilityZones := []string{
		"us-east-1a", "us-east-1b", "us-east-1c", "us-west-1a", "us-west-1b",
		"us-west-1c", "eu-central-1a", "eu-central-1b", "eu-central-1c", "ap-south-1a",
		"ap-south-1b", "ap-south-1c", "sa-east-1a", "sa-east-1b", "sa-east-1c",
		"us-east-2a", "us-east-2b", "us-east-2c", "eu-west-1a", "eu-west-1b",
		"eu-west-1c", "ap-northeast-1a", "ap-northeast-1b", "ap-northeast-1c", "ap-southeast-2a",
		"ap-southeast-2b", "ap-southeast-2c", "ca-central-1a", "ca-central-1b", "ca-central-1c",
		"us-west-2a", "us-west-2b", "us-west-2c", "eu-north-1a", "eu-north-1b",
		"eu-north-1c", "ap-northeast-2a", "ap-northeast-2b", "ap-northeast-2c", "me-south-1a",
		"me-south-1b", "me-south-1c", "af-south-1a", "af-south-1b", "af-south-1c",
		"us-west-3a", "us-west-3b", "us-west-3c", "us-west-4a", "us-west-4b",
		"us-west-4c", "cn-north-1a", "cn-north-1b", "cn-north-1c", "cn-northwest-1a",
		"cn-northwest-1b", "us-gov-west-1a", "us-gov-west-1b", "us-gov-west-1c", "us-gov-east-1a",
		"us-gov-east-1b", "us-gov-east-1c", "ap-southeast-3a", "ap-southeast-3b", "ap-southeast-3c",
	}

	cloudAccountIDs := []string{
		// AWS Account IDs (12-digit numbers)
		"123456789012",
		"987654321001",
		"567890123456",
		"246813579024",
		"135792468013",

		// Azure Account IDs (GUID format)
		"9a5f8d94-a0b3-4336-9f3d-61abf85b5974",
		"8b7c5f60-4320-49e1-bfba-90c5cfa4a9c0",
		"40b3647a-bd0e-4e9d-b3ad-cd567984dba9",
		"a12f0c7b-4295-4b0b-bb7f-58cbd2859d19",
		"001d4d9b-3d9b-4d76-8765-990ce98a295e",

		// GCP Account IDs (project-based format)
		"my-cloud-project-12345",
		"gcp-prod-project-67890",
		"dev-cloud-project-54321",
		"prod-cloud-account-98765",
		"test-cloud-account-11223",
	}

	deviceModelIdentifiers := []string{
		"iPhone-13-Pro-Max-A2341",
		"Galaxy-S21-5G-XYZ123",
		"Pixel-6-Pro-XYZ456",
		"MacBook-Pro-16-M1-2021",
		"Surface-Pro-7-XXY987",
		"iPad-Air-4th-Gen-ABCD1234",
		"XPS-13-Intel-2022-5678",
		"ThinkPad-X1-Carbon-Gen9-9876",
		"PlayStation-5-AB1234XY",
		"Galaxy-Tab-S7-XYZ789",
		"Echo-Dot-4th-Gen-XYZ1122",
		"Fire-HD-10-2021-QR1234",
		"Kindle-Paperwhite-10th-Gen-567890",
		"Apple-Watch-SE-44mm-AW12345",
		"Samsung-Galaxy-Buds-Pro-SBP2021",
		"Fitbit-Charge-5-FBC345",
		"GoPro-Hero-10-Black-GH2021",
		"Raspberry-Pi-4B-8GB-RP4500",
		"DJI-Mavic-Air-2-DJI001",
		"Nest-Hub-2nd-Gen-NH123456",
		"HP-Envy-X360-15-2021-HP987",
		"Logitech-C922-Pro-Webcam-LCW123",
		"OnePlus-9-Pro-ABC456DEF",
		"Sony-WH-1000XM4-Sony2022",
		"Xbox-Series-X-XBX98765",
		"Lenovo-Legion-5-2021-LG5678",
		"Roku-Ultra-2021-RU123",
		"Xbox-One-S-XOS567",
		"Apple-Tv-4K-2021-ATV1234",
		"Google-Nest-Mini-Gen-2-NM567",
		"Mac-Mini-Apple-Silicon-M1-MM2022",
		"HP-Omen-15-2022-HO1234",
		"Garmin-Vivoactive-4-GV567",
		"Microsoft-Surface-Go-2-SG2022",
	}

	osDescriptions := []string{
		"Windows 10 Pro 64-bit, Version 21H1",
		"macOS Big Sur 11.6.3",
		"Ubuntu 20.04.3 LTS (Focal Fossa)",
		"Red Hat Enterprise Linux 8.4",
		"CentOS 7 (Core), Kernel 3.10.0-1160.el7.x86_64",
		"Debian 10.9 (Buster)",
		"Fedora 34 (Workstation Edition)",
		"Android 12 (API 31)",
		"iOS 15.4.1",
		"Windows Server 2019 Datacenter",
		"macOS Monterey 12.3.1",
		"Chrome OS 98.0.4758.102 (Official Build)",
		"Amazon Linux 2 Kernel 5.10.47-69.196.amzn2.x86_64",
		"Arch Linux, Kernel 5.16.10-arch1-1",
		"Linux Mint 20.3 (Una), Cinnamon 5.2.7",
		"SUSE Linux Enterprise Server 15 SP3",
		"FreeBSD 13.0-RELEASE",
		"Raspberry Pi OS (32-bit), Version: April 2021",
		"Windows 11 Pro 64-bit, Version 21H2",
		"macOS Sierra 10.12.6",
		"OpenSUSE Leap 15.3",
		"Pop!_OS 21.10 (Intel/AMD)",
		"Kali Linux 2021.3",
		"Oracle Linux 8.5",
		"Windows 7 Ultimate 64-bit",
		"Ubuntu 18.04.5 LTS (Bionic Beaver)",
		"iOS 14.7.1",
		"Android 11 (API 30)",
		"Linux Mint 19.3 (Tricia), Cinnamon 4.4.8",
		"Slackware 14.2",
		"Mac OS X 10.9.5 (Mavericks)",
		"Zorin OS 16 Core, 64-bit",
		"Android 10 (API 29)",
		"Windows Server 2016 Standard",
		"Fedora 32, GNOME 3.38",
		"macOS Catalina 10.15.7",
		"Ubuntu 21.04 (Hirsute Hippo)",
		"Windows 8.1 Pro 64-bit",
		"openSUSE Tumbleweed, Kernel 5.13.8",
		"Chrome OS 93.0.4577.82",
		"Red Hat Enterprise Linux 7.9 (Maipo)",
		"Arch Linux ARM, Kernel 5.10.60-1-ARCH",
		"Sailfish OS 4.0.1.48",
		"ReactOS 0.4.14",
		"Fedora 33, KDE Plasma 5.20.5",
		"Manjaro Linux 21.0.7 (Pahvo)",
		"macOS High Sierra 10.13.6",
		"Android 9 (Pie)",
		"Windows Server 2012 R2 Standard",
		"Ubuntu 16.04.7 LTS (Xenial Xerus)",
		"Windows XP Professional SP3",
		"CentOS Stream 8",
		"Linux Mint 20 (Ulyana), MATE 1.24.1",
	}

	processOwners := []string{
		"root", "admin", "systemd", "nobody", "apache", "www-data", "mysql", "postgres", "nginx", "docker",
		"oracle", "sshd", "ftp", "mail", "vagrant", "mongodb", "elastic", "consul", "jenkins", "ci-cd",
		"system", "postgresql", "redis", "appuser", "root", "testuser", "adminuser", "ftpuser", "usermyapp",
		"root", "shared", "backup", "service-account", "www", "guest", "lambda", "service", "system_process",
		"dev", "adminstrator", "cloud-user", "sysadmin", "debugger", "installer", "monitor", "exim", "cassandra",
		"root", "build", "vfs", "root", "sys", "unprivileged", "service123", "gobuster", "cloud-admin", "bot",
		"worker", "god", "tftp", "local", "system_root", "sbx", "envuser", "infra", "sysadmin", "prod-user", "network",
	}

	webEngineNames := []string{
		"Chrome", "Firefox", "Safari", "Edge", "Opera", "Internet Explorer", "Brave", "Vivaldi", "Chromium", "UC Browser",
		"Samsung Internet", "Yandex Browser", "Baidu Browser", "Epiphany", "Maxthon", "Pale Moon", "Waterfox", "Tor Browser",
		"Comodo Dragon", "SRWare Iron", "Falkon", "Lynx", "Qutebrowser", "Midori", "Konqueror", "Mosaic", "Links", "Dillo",
	}

	sdkLanguages := []string{
		"Go", "Java", "Python", "JavaScript", "Ruby", "C#", "PHP", "C++", "Swift", "Kotlin",
		"Rust", "TypeScript", "Objective-C", "Scala", "Perl", "Lua", "Elixir", "Haskell", "F#", "Dart",
		"Vala", "R", "C", "PowerShell", "Shell", "Visual Basic", "Tcl", "OCaml", "Golang", "CoffeeScript",
	}

	dbConnectionStrings := []string{
		"Server=myserver.database.windows.net;Database=mydb;User Id=myuser;Password=mypassword;",
		"jdbc:mysql://localhost:3306/mydb?user=myuser&password=mypassword",
		"postgres://myuser:mypassword@localhost:5432/mydb",
		"mongodb://myuser:mypassword@localhost:27017/mydb",
		"sqlite:///path/to/mydb.sqlite",
		"sqlserver://myuser:mypassword@localhost:1433;database=mydb",
		"mysql://myuser:mypassword@localhost:3306/mydb",
		"oracle://myuser:mypassword@localhost:1521/mydb",
		"jdbc:postgresql://localhost:5432/mydb?user=myuser&password=mypassword",
		"jdbc:sqlserver://localhost:1433;databaseName=mydb;user=myuser;password=mypassword",
		"mongodb+srv://myuser:mypassword@cluster0.mongodb.net/mydb",
		"db2://myuser:mypassword@localhost:50000/mydb",
		"redis://localhost:6379/0",
		"cockroachdb://myuser:mypassword@localhost:26257/mydb",
		"jdbc:oracle:thin:@localhost:1521:xe",
	}

	dbNames := []string{
		"mydb",
		"users_db",
		"inventory",
		"sales_data",
		"analytics_db",
		"orders_db",
		"customer_data",
		"employees_db",
		"products_db",
		"transaction_db",
		"logs_db",
		"web_app_db",
		"financial_records",
		"audit_db",
		"marketing_data",
		"test_db",
		"hr_system_db",
		"content_management",
		"data_warehouse",
		"payment_gateway_db",
	}

	dbOperations := []string{
		"insert",
		"update",
		"delete",
		"select",
		"merge",
		"upsert",
		"truncate",
		"alter",
		"create",
		"drop",
		"index",
		"select_for_update",
		"grant",
		"revoke",
		"backup",
		"restore",
		"copy",
		"vacuum",
		"analyze",
		"repair",
	}

	netPeerIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"203.0.113.5",
		"198.51.100.2",
		"240.0.0.0",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"fe80::1",
		"10.1.1.1",
		"192.168.0.100",
		"169.254.169.254",
		"198.51.100.3",
		"192.0.2.1",
		"172.31.255.255",
		"10.255.255.255",
		"127.0.0.1",
		"198.51.100.5",
		"203.0.113.10",
		"10.10.10.10",
		"192.168.2.2",
	}

	codeFunctions := []string{
		"processData",
		"initializeConnection",
		"handleRequest",
		"parseJSON",
		"generateReport",
		"validateInput",
		"sendMessage",
		"fetchData",
		"calculateSum",
		"executeQuery",
		"parseQueryString",
		"loadConfiguration",
		"handleError",
		"transformData",
		"renderPage",
		"updateRecord",
		"deleteRecord",
		"connectToDatabase",
		"closeConnection",
		"validateUser",
		"processTransaction",
	}

	httpMethods := []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"PATCH",
		"HEAD",
		"OPTIONS",
		"TRACE",
		"CONNECT",
		"LINK",
		"UNLINK",
	}

	job := jobs[cryptoRandIntn(len(jobs))]
	instance := instances[cryptoRandIntn(len(instances))]
	container := containers[cryptoRandIntn(len(containers))]
	hostname := fmt.Sprintf("node-%03d", cryptoRandIntn(100)+1)
	region := regions[cryptoRandIntn(len(regions))]
	namespace := namespaces[cryptoRandIntn(len(namespaces))]
	instanceType := instanceTypes[cryptoRandIntn(len(instanceTypes))]
	label := labels[cryptoRandIntn(len(labels))]
	version := versions[cryptoRandIntn(len(versions))]
	availabilityZone := availabilityZones[cryptoRandIntn(len(availabilityZones))]
	cloudAccountID := cloudAccountIDs[cryptoRandIntn(len(cloudAccountIDs))]
	deviceModelIdentifier := deviceModelIdentifiers[cryptoRandIntn(len(deviceModelIdentifiers))]
	osDescription := osDescriptions[cryptoRandIntn(len(osDescriptions))]
	processOwner := processOwners[cryptoRandIntn(len(processOwners))]
	webEngineName := webEngineNames[cryptoRandIntn(len(webEngineNames))]
	sdkLanguage := sdkLanguages[cryptoRandIntn(len(sdkLanguages))]
	dbConnectionString := dbConnectionStrings[cryptoRandIntn(len(dbConnectionStrings))]
	dbName := dbNames[cryptoRandIntn(len(dbNames))]
	dbOperation := dbOperations[cryptoRandIntn(len(dbOperations))]
	netPeerIP := netPeerIPs[cryptoRandIntn(len(netPeerIPs))]
	codeFunction := codeFunctions[cryptoRandIntn(len(codeFunctions))]
	httpMethod := httpMethods[cryptoRandIntn(len(httpMethods))]

	logFormats := []string{
		"Log %d: %s phase: %s | Message : %s | UUID: %s | Timestamp: %s",
		"Event %d: Level=%s, Phase=%s, Detail=%s, UUID=%s, Time=%s",
		"[LogRecord-%d] Severity: %s | Phase: %s | Details: %s | UUID: %s | Time: %s",
		"Transaction %d | Severity: %s | Phase: %s | Info: %s | UUID: %s | Time: %s",
		"[Log-%d] Status: %s | Step: %s | Desc: %s | Ref: %s | Time: %s",
		"Log #%d -- [Phase: %s]- Severity: %s | Msg: %s | TraceID: %s | At: %s",
		"Worker %d reports %s | Current Step: %s | Context: %s | ID: %s | Recorded: %s",
		"ID: %d | Alert: %s detected in Phase: %s | Message: %s | Code: %s | Time: %s",
		"Diagnostic Log %d | Type: %s | Phase: %s | Summary: %s | Session: %s | Timestamp: %s",
		"System Log #%d | Priority: %s | Task: %s | Content: %s | UUID: %s | Logged At: %s",
		"Record %d: Severity=%s | Stage=%s | Description=%s | UUID=%s | Logged Time=%s",
		"Trace %d | Level: %s | Phase: %s | Message: %s | Identifier: %s | Timestamp: %s",
		"[Entry-%d] Status: %s | Segment: %s | Notes: %s | UID: %s | Time Log: %s",
		"Process %d -- Status: %s | Phase: %s | Log: %s | Key: %s | Timestamp: %s",
		"Checkpoint %d | Importance: %s | Stage: %s | Report: %s | UniqueID: %s | Time Mark: %s",
		"[Monitor %d] Level: %s | Step: %s | Message: %s | Ref ID: %s | Logged At: %s",
		"Alert %d: Severity=%s | Workflow=%s | Info=%s | UUID=%s | Time Recorded=%s",
		"Log Entry %d -- Priority: %s | Execution Phase: %s | Context: %s | ID: %s | Recorded Time: %s",
		"Task %d | Alert: %s | Phase: %s | Summary: %s | Reference: %s | Timestamp: %s",
		"Execution Log %d: Level=%s | Step=%s | Message=%s | Key=%s | Logged On=%s",
		"Worker %d Status: %s | Processing Step: %s | Log Details: %s | Tracking ID: %s | Timestamp: %s",
		"[Event-%d] Urgency: %s | Stage: %s | Log: %s | Session: %s | Time Record: %s",
		"Service %d - Status: %s | Phase: %s | Output: %s | Correlation ID: %s | Time: %s",
		"Operation %d -> Level: %s | Phase: %s | Event: %s | Token: %s | Recorded: %s",
		"Routine Log %d | Criticality: %s | Phase: %s | Notes: %s | Trace Key: %s | Time Stamp: %s",
		"Transaction Log %d - Severity: %s | Process: %s | Entry: %s | Reference: %s | Time Logged: %s",
		"Session %d: Alert=%s | Status=%s | Details=%s | Log ID=%s | Timestamp=%s",
		"[Audit-%d] Priority: %s | Phase: %s | Log Entry: %s | Token: %s | Recorded Time: %s",
		"Monitor Event %d: Importance=%s | Stage=%s | Context=%s | UID=%s | Time Stamp=%s",
		"System Log %d -- Level: %s | Task: %s | Description: %s | Unique Key: %s | Logged On: %s",
		"Trace Event %d | Urgency: %s | Step: %s | Log Info: %s | Identifier: %s | Time Logged: %s",
		"Registry %d: Type=%s | Workflow=%s | Summary=%s | ID=%s | Time Recorded=%s",
		"[LogData-%d] Alert: %s | Task: %s | Notes: %s | Session: %s | Timestamp: %s",
		"Checkpoint Log %d - Criticality: %s | Segment: %s | Notes: %s | Tracking ID: %s | Logged At: %s",
		"Run %d -- Status: %s | Phase: %s | Details: %s | UUID: %s | Timestamp: %s",
		"Operation Log %d | Priority: %s | Step: %s | Log Message: %s | Unique Ref: %s | Logged At: %s",
		"[Status-%d] Level: %s | Execution Phase: %s | Report: %s | Session ID: %s | Time Marked: %s",
		"Routine %d | Alert Type: %s | Task: %s | Event: %s | ID Key: %s | Timestamp: %s",
		"Diagnostic Log %d: Severity=%s | Segment=%s | Context=%s | TraceID=%s | Time Recorded=%s",
		"Incident %d -- Priority: %s | Status: %s | Entry: %s | Ref Key: %s | Logged Time: %s",
		"Execution Trace %d: Level=%s | Workflow=%s | Notes=%s | UID=%s | Timestamp=%s",
		"[Process-%d] Alert: %s | Step: %s | Log Info: %s | Session: %s | Logged On: %s",
		"Worker %d | Severity: %s | Step: %s | Details: %s | Identifier: %s | Time Recorded: %s",
		"Service Log %d | Urgency: %s | Phase: %s | Summary: %s | Key: %s | Logged At: %s",
		"Operation Log Entry %d -- Importance: %s | Execution Step: %s | Log Report: %s | Ref: %s | Time: %s",
		"Check %d: Status=%s | Stage=%s | Details=%s | UniqueID=%s | Timestamp=%s",
		"[SessionLog-%d] Level: %s | Task: %s | Description: %s | Token: %s | Time Log: %s",
		"Routine Task %d | Priority: %s | Status: %s | Log Message: %s | Identifier: %s | Logged On: %s",
		"Incident Report %d | Severity: %s | Phase: %s | Details: %s | Trace ID: %s | Timestamp: %s",
		"Execution Monitor %d - Alert: %s | Workflow: %s | Message: %s | Tracking ID: %s | Time Recorded: %s",
		"System Event %d: Priority=%s | Execution Step=%s | Report=%s | UUID=%s | Logged Time=%s",
		"[AuditLog-%d] Severity: %s | Stage: %s | Context: %s | Session ID: %s | Timestamp: %s",
		"Registry Entry %d | Urgency: %s | Task: %s | Summary: %s | Key: %s | Logged At: %s",
		"Trace Log %d -- Criticality: %s | Step: %s | Log Info: %s | Identifier: %s | Logged Time: %s",
		"Check %d: Level=%s | Execution Phase=%s | Notes=%s | Reference ID=%s | Timestamp=%s",
	}

	logBody := fmt.Sprintf(
		logFormats[cryptoRandIntn(len(logFormats))],
		i, severityText, phase, randomSentence(), uuid.New().String(), time.Now().Format(time.RFC3339),
	)

	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(severityText)
	record.SetBody(log.StringValue(logBody))

	attributes := []attribute.KeyValue{
		semconv.K8SJobNameKey.String(job),
		semconv.DeviceIDKey.String(instance),
		semconv.ContainerNameKey.String(container),
		semconv.HostNameKey.String(hostname),
		semconv.CloudRegionKey.String(region),
		semconv.ServiceNamespaceKey.String(namespace),
		semconv.HostArchKey.String(instanceType),
		semconv.ProcessRuntimeNameKey.String(label),
		semconv.ServiceVersionKey.String(version),
		semconv.CloudAvailabilityZoneKey.String(availabilityZone),
		semconv.CloudAccountIDKey.String(cloudAccountID),
		semconv.DeviceModelIdentifierKey.String(deviceModelIdentifier),
		semconv.OSDescriptionKey.String(osDescription),
		semconv.ProcessOwnerKey.String(processOwner),
		semconv.WebEngineNameKey.String(webEngineName),
		semconv.TelemetrySDKLanguageKey.String(sdkLanguage),
		semconv.DBConnectionStringKey.String(dbConnectionString),
		semconv.DBNameKey.String(dbName),
		semconv.DBOperationKey.String(dbOperation),
		semconv.NetPeerIPKey.String(netPeerIP),
		semconv.CodeFunctionKey.String(codeFunction),
		semconv.HTTPMethodKey.String(httpMethod),
	}

	shuffledAttributes := make([]attribute.KeyValue, len(attributes))
	copy(shuffledAttributes, attributes)

	for i := range shuffledAttributes {
		j := cryptoRandIntn(len(shuffledAttributes))
		shuffledAttributes[i], shuffledAttributes[j] = shuffledAttributes[j], shuffledAttributes[i]
	}

	numAttributes := 5
	if len(shuffledAttributes) < numAttributes {
		numAttributes = len(shuffledAttributes)
	}

	for i := 0; i < numAttributes; i++ {
		attr := log.KeyValue{
			Key:   string(shuffledAttributes[i].Key),
			Value: log.StringValue(shuffledAttributes[i].Value.AsString()),
		}
		record.AddAttributes(attr)
	}

	logger.Emit(context.Background(), record)

	fmt.Printf("Body: %s\n", record.Body())
}

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
	endpointFlag := flag.String("exporter-endpoint", "0.0.0.0:4317", "gRPC exporter endpoint")

	flag.Parse()

	config := &Config{
		Endpoint: *endpointFlag,
		Insecure: true,
		Headers:  map[string]string{},
	}

	exporter, err := createExporter(config)
	if err != nil {
		fmt.Printf("Failed to create exporter: %v\n", err)
		os.Exit(1)
	}
	defer exporter.Shutdown(context.Background())

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
	)
	batchProcessor := logsdk.NewBatchProcessor(exporter,
		logsdk.WithMaxQueueSize(4096),
		logsdk.WithExportMaxBatchSize(2096),
		logsdk.WithExportInterval(1*time.Second),
	)
	loggerProvider := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(batchProcessor),
		logsdk.WithResource(res),
	)
	logger := loggerProvider.Logger("otel-logger")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for i := 0; ; i++ {
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

			time.Sleep(1 * time.Nanosecond)
		}
	}()

	<-signalChan
	fmt.Println("\nReceived termination signal, shutting down gracefully...")

	cancel()

	fmt.Println("Shutdown complete.")
}
