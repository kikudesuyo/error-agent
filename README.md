# error-agent

`error-agent` is an open-source Go service under Apache-2.0 that receives production error events and stores them for later investigation. Buildlog is the first integration target. The initial deployment target is Google Cloud Run.

This initial version provides authenticated event intake, input validation, redaction of common secrets, and PostgreSQL-backed deduplication. Automated code investigation and pull request creation are outside the current scope.

## Requirements

- Go 1.27 or newer
- Docker Compose
- Atlas CLI

## Local development

Start PostgreSQL:

```sh
docker compose up -d postgres
```

Set local-only credentials and apply the migration:

```sh
export DATABASE_URL='postgres://error_agent:local-development-only@localhost:5433/error_agent?sslmode=disable'
export INGEST_TOKEN="$(openssl rand -hex 32)"
export APP_ENV=development
cd api
atlas migrate apply --dir "file://sql/migrations" --url "$DATABASE_URL"
go run .
```

The API listens on port `8080`. `GET /healthz` checks process liveness. `GET /readyz` also checks database connectivity.

## Event API

`POST /v1/error-events` requires `Authorization: Bearer <INGEST_TOKEN>` and `Content-Type: application/json`. The service accepts only 5xx events and returns `202 Accepted` after durable persistence. Repeated events return `202` with `duplicate: true` and update the occurrence count without creating a second record.

```json
{
  "service": "buildlog-frontend",
  "environment": "production",
  "status": 500,
  "method": "GET",
  "url": "https://example.com/contact",
  "path": "/contact",
  "errorType": "Error",
  "message": "request failed",
  "stack": "Error: request failed",
  "commitSha": "0123456789abcdef0123456789abcdef01234567",
  "cfRay": "example-ray-id",
  "userAgent": "example-browser"
}
```

The endpoint rejects query strings, fragments, and credentials in `url`, unknown JSON fields, oversized bodies, and paths that do not match the URL. It stores only the documented fields; it does not accept cookies, authorization headers, or request bodies as event fields. Common bearer credentials, cookie values, key/value secrets, email addresses, and phone-number-like values in text fields are redacted before persistence. Producers must still avoid sending personal information or secrets in error messages, stack traces, paths, and user-agent values because format-based redaction cannot recognize every sensitive value.

The service does not log event bodies, Authorization headers, message, stack, URL, path, or user-agent values. Configure application logs without request-body capture in any proxy in front of the service.

## Cloud Run deployment

Build and deploy the container to Cloud Run. Use a dedicated runtime service account with `roles/secretmanager.secretAccessor` only on the Secret Manager secrets containing `DATABASE_URL` and `INGEST_TOKEN`. Do not pass secret values as command-line arguments or commit them to the repository.

The service must be reachable from the Buildlog Cloudflare Pages server runtime. Cloud Run's HTTPS endpoint is publicly invokable, while the application requires the bearer credential on the ingestion endpoint. Keep the credential server-side in both services and rotate both values together. Liveness/readiness endpoints contain no application data.

Example deployment (replace project, region, service account, and secret names with the deployment's values):

```sh
gcloud run deploy error-agent \
  --source . \
  --region asia-northeast1 \
  --max 5 \
  --service-account error-agent-runtime@PROJECT_ID.iam.gserviceaccount.com \
  --set-env-vars APP_ENV=production \
  --set-secrets DATABASE_URL=error-agent-database-url:latest,INGEST_TOKEN=error-agent-ingest-token:latest \
  --allow-unauthenticated
```

The public Cloud Run invoker setting is required for the Cloudflare server runtime to reach the service; the ingestion route still rejects requests without the application bearer credential. In production, the service refuses a PostgreSQL URL whose `sslmode` does not enable TLS. Apply Atlas migrations through the approved deployment procedure before sending events.

## Validation

```sh
cd api
gofmt -w .
go vet ./...
go test ./...
go build ./...
atlas migrate validate --dir "file://sql/migrations"
```

Never use production credentials or apply a production migration from a local development session.
