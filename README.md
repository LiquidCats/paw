# paw

## Development commands

The repository has one root `Makefile`. Docker runs each tool against the
selected Go module, so no local Go tooling is required.

Run a command for every compatible module:

```bash
make test
make lint
make mock
make generate-sql
```

Limit a command to one module with its repository-relative path:

```bash
make test MODULE=services/rater
make lint-fix MODULE=lib/workers
make generate-sql MODULE=services/watcher
```

Run `make help` for all targets and `make modules` for the discovered module
lists. Tool image names, versions, and command flags can be overridden through
the variables declared at the top of the `Makefile`.
