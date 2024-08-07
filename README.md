# Juju status exporter

Prometheus exporter for juju status output in Go.

## Requirement

juju status is set.

## Build

The following command in create the juju_status_exporter binary.

```
make build
```

## Usage

The following flags are available:

- **`-push_gateway`**: URL of the Prometheus Pushgateway to push metrics to. If not set, the script will expose metrics via HTTP.
- **`-model_name`**: Label for the `instance` grouping in the Pushgateway. Default is `"instance_1"`.
- **`-period`**: Interval (in seconds) for pushing metrics to the Pushgateway. Default is 30 seconds.

### Examples

1. **Expose Metrics via HTTP**:
   ```sh
   ./juju_exporter -model_name=testing
   ```

2. **Push Metrics to Pushgateway with Default Period**:
   ```sh
   ./juju_exporter -push_gateway=http://my-prometheus-pushgateway -model_name=testing
   ```
