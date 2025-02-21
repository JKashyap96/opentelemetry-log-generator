# opentelemetry-log-generator
This repository contains an OpenTelemetry logs generator.
Logs are emitted to the grpc endpoint: **0.0.0.0:4317** by default.<br/>
You can send logs to any service grpc endpoint by using the flag ```-exporter-endpoint=<YOUR_EXPORTER_ENDPOINT>```.

# Test the logSource in your local environment
## Build the logSource image
Clone the repo and then run ```docker build -t log-simulator -f ./DockerFile .``` to build the **log-simulator** image.
## Install an OpenTelemetry Collector
* Download the latest OpenTelemetry Collector :<br/>
```curl -sSfL https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.114.0/otelcol-contrib_0.114.0_darwin_arm64.tar.gz -o otelcol-contrib.tar.gz```
* Extract the binary :<br/>
```tar -xvf otelcol-contrib.tar.gz```
* Move the binary to /usr/local/bin :<br/>
```mv otelcol-contrib /usr/local/bin/```
* Verify installation :<br/>
```otelcol-contrib --version```
* Create a sample OpenTelemetry collector config file called **otel-collector-config.yaml** to receive the emitted logs :<br/>
```
receivers:
  otlp:
    protocols:
      grpc:
      http:

exporters:
  debug:
    verbosity: basic  # Enables detailed logging of all received data points

service:
  pipelines:
    logs:
      receivers: [otlp]
      exporters: [debug]
```
* Run the OpenTelemetry collector :<br/>
```otelcol-contrib --config otel-collector-config.yaml```
## Run the log-simulator docker container to send logs to your OpenTelemetry Collector in your local macOS environment.
```docker run -d log-simulator -exporter-endpoint="host.docker.internal:4317"``` <br/><br/>
