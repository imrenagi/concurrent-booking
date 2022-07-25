# Simple Concert Ticketing

* There is a concert record with quota for the ticket.
* Concurrent users are competing to get the ticket




```bigquery
helm install redis --namespace redis --values redis-values.yml bitnami/redis
```

docker run -d \
    -v ${PWD}/hack/observability/secrets/psql-local.json:/config/key.json \
    -p 127.0.0.1:5432:5432 \
    gcr.io/cloudsql-docker/gce-proxy:1.31.0 /cloud_sql_proxy \
    -instances=io-extended-2022:asia-southeast1:booking=tcp:0.0.0.0:5432 -credential_file=/config/key.json


kubectl -n gmp-test port-forward svc/frontend 9090