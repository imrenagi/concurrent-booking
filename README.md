# Simple Concert Ticketing

This application is used for demonstrating the application instrumentation, especially trace and metrics.

## Running it locally

### Enable new relic (optional)

To enable new relic instrumentation:

1. Create free newrelic account and grab the license key.

1. Set these following env var:

    * `OTEL_NEW_RELIC_EXPORTER_OTLP_ENDPOINT` to `https://otlp.nr-data.net:4317`
    * `OTEL_NEW_RELIC_EXPORTER_API_KEY` to ingest license key from newrelic account
         
### Enable Google Cloud Tracing and Monitoring

1. Create GCP service account and ensure that service account has (at minimum) `roles/monitoring.metricWriter` and `roles/cloudtrace.agent` role.

1. Update value of `exporters.googlecloud.project` in [otel-config file](./hack/observability/otel-config.yaml) to your google cloud project id/ 

### Run Docker Compose

    ```shell
    docker-compose up
    ```

### Open Dashboard

1. Jaeger Tracing Dashboard: [http://127.0.0.1:16686/](http://127.0.0.1:16686/)
2. Prometheus Dashboard: [http://127.0.0.1:9090/](http://127.0.0.1:9090/)
3. Grafana Dashboard: [http://127.0.0.1:3000/](http://127.0.0.1:3000/)
4. Asynqmon Dashboard: [http://127.0.0.1:8011/](http://127.0.0.1:8011/)
5. Newrelic dashboard (optional)
6. Google Cloud Monitoring
7. Google Cloud Tracing

### Run load generator

1. Install hey

1. Run load generator

    ```shell
    hey -z 1m -c 2 -q 2  -m POST -d '{"show_id": "b9b0d5da-98a4-4b09-b5f5-83dc0c3b9964"}' http://localhost:9999/api/v1/booking
    hey -z 1m -c 2 -q 2  -m POST -d '{"show_id": "b9b0d5da-98a4-4b09-b5f5-83dc0c3b9964"}' http://localhost:9999/api/v2/booking
    ```

## Running it on kubernetes

WARNING: This setup is not properly documented.

1. Setup CloudSQL Postgres Instance. Update `secret.yaml.example`, `app.yaml`, and `worker.yaml` with correct value for database config.
    ```shell
    docker run -d \
        -v ${PWD}/hack/observability/secrets/psql-local.json:/config/key.json \
        -p 127.0.0.1:5432:5432 \
        gcr.io/cloudsql-docker/gce-proxy:1.31.0 /cloud_sql_proxy \
        -instances=io-extended-2022:asia-southeast1:booking=tcp:0.0.0.0:5432 -credential_file=/config/key.json
    ```

1. Install bitnami/helm chart

    ```shell
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm install redis --namespace redis --values redis-values.yml bitnami/redis
    ```

1. Apply all manifest in `k8s` directory.

1. Port forward prometheus dashboard

    ```shell
    kubectl -n gmp-test port-forward svc/frontend 9090
    ```

1. Run load test

    ```shell
    k apply -f k8s/hey.yaml
    ```