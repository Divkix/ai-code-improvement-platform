# ai-code-improvement-platform

## Development Environment

The frontend expects the variable `VITE_API_URL` to point at the backend REST API.  In local development the backend runs on port **8080** (see `docker-compose.dev.yml`), so create a `.env` file in `frontend/` (or export the variable in your shell) containing:

```bash
VITE_API_URL=http://localhost:8080
```

If you change the backend host or port, update this value accordingly.

> **Why is this important?**  The chat page uses absolute URLs to call the backend.  Without `VITE_API_URL`, those calls would default to the browser origin (e.g. `http://localhost:3000`) and return **404 Not Found**.

Make sure to restart the dev server after modifying environment variables so that Vite picks them up.

## Live API documentation

Once the backend is running you can explore the REST API at:

```
http://localhost:8080/docs/
```

The Swagger-UI page loads the OpenAPI specification directly from `/api/openapi.yaml`, so any edits to `backend/api/openapi.yaml` are reflected on refresh.

To generate a JSON version of the spec (used by some tooling) run:

```bash
make -C backend openapi-json
```
(or execute `backend/scripts/generate-openapi-json.sh` directly). Ensure [`yq`](https://github.com/mikefarah/yq) is installed.
