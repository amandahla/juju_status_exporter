# Juju status exporter

Prometheus exporter for juju status output in Go.

Built with go1.22.6.

## Requirement

juju status is set.

## Build

The following command in create the juju_status_exporter binary.

```
make build
```

## Usage

The following flags are available:

- **`-port`**: The port on which the exporter will start listening.
- **`-model_name`**: Label for the `instance` grouping in the Pushgateway. Default is `"instance_1"`.

Also, the following environment variables can be used:

- **`JUJU_MODEL`**
- **`JUJU_STATUS_EXPORTER_PORT`**

### Examples

1. **Expose Metrics via HTTP**:
   ```sh
   ./juju_exporter -model_name=testing
   ```
   ```
