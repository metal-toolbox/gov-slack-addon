# gov-slack-addon

`gov-slack-addon` is an addon to integrate Slack with Governor.

## Usage

Search this repo for `TODO` to find the places where code would need to be updated which should be limited do:

```bash
.buildkite/pipeline.yml

cmd/serve.go
cmd/root.go
cmd/errors.go

srv/doc.go
srv/errors.go
```

## Development

This repo includes a `docker-compose.yml` and a `Makefile` to make getting started easy.

`make docker-up` will start a basic https server with an audit container
