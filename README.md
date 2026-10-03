[![Docker Build and Push](https://github.com/Anuj2402/github-actions-practice/actions/workflows/docker-build-push.yml/badge.svg)](https://github.com/Anuj2402/github-actions-practice/actions/workflows/docker-build-push.yml)

# github-actions-practice

A three-tier real-time chat application, Dockerized and wired up with a GitHub Actions CI/CD pipeline that builds, tests, and publishes the image to Docker Hub.

## What it does

- **Frontend** -- static HTML/CSS/JS chat UI (index.html) and a chat history browser (history.html), both served directly by the Go backend.
- **Backend** -- a Go server exposing a WebSocket endpoint (/ws) for real-time messaging and a REST endpoint (/api/chat-history) for querying past conversations.
- **Database** -- MySQL, storing every message sent, queried via GORM.

Two users connect over WebSocket with a from/to username pair, see their prior conversation history replayed on connect, and exchange messages live. The protocol (ws:// vs wss://) is chosen dynamically based on how the page itself was loaded, so it works correctly behind both plain HTTP and TLS-terminating proxies.

## Project layout

```
.
|-- main.go                        # Go backend: WebSocket handler + REST API + MySQL via GORM
|-- go.mod / go.sum
|-- index.html                     # Chat UI
|-- history.html                   # Chat history UI
|-- Dockerfile                     # Multi-stage build, small, non-root runtime image
|-- .dockerignore
|-- scripts/
|   `-- health_check.sh            # Builds the image, runs it against a real MySQL
|                                   # instance, and curls the REST endpoint to confirm
|                                   # the app actually starts and responds correctly
`-- .github/workflows/
    `-- docker-build-push.yml      # CI: build, test, tag, push to Docker Hub
```

## Running locally

```bash
docker build -t chat-app .
```

The app needs MySQL to start (it fails fast if it can't connect). Configuration is entirely via environment variables -- no hardcoded credentials:

| Variable | Default | Description |
|---|---|---|
| APP_PORT | 8080 | Port the server listens on |
| DB_HOST | 127.0.0.1 | MySQL host |
| DB_PORT | 3306 | MySQL port |
| DB_USER | root | MySQL user |
| DB_PASSWORD | (empty) | MySQL password |
| DB_NAME | chatdb | Database name |

See scripts/health_check.sh for a working example of running the app alongside a MySQL container.

## Testing

```bash
./scripts/health_check.sh
```

This builds the image, starts a real MySQL container, starts the app pointed at it, waits for the app to respond, and curls /api/chat-history to confirm it returns a valid response. Everything is cleaned up automatically afterward, whether the check passes or fails.

## CI/CD

On every push to main, GitHub Actions:
1. Checks out the code
2. Builds the Docker image
3. Logs in to Docker Hub using repo secrets (DOCKER_USERNAME, DOCKER_TOKEN)
4. Tags the image as latest and sha-<short-commit-hash>
5. Pushes both tags to Docker Hub -- only when the push is to main (feature branches and PRs build the image but don't publish it)

Pull the published image:

```bash
docker pull anujkumar007/chat-app:latest
```
