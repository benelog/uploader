Developing Uploader
=========
Go 1.22 or later is the only requirement to build and test.
The uploader has no third party dependency.

Build

    make build      # go build -o uploader .
    make run        # build and start on port 8080


Quality tools
---------

Following https://blog.benelog.net/go-quality-tools :

    make fmt        # goimports -w .
    make lint       # golangci-lint run ./...
    make test       # go test ./...
    make check      # fmt + lint + test, before committing
    make ci         # lint + test, what CI runs

`goimports` and `golangci-lint` (v2) are needed for `fmt` and `lint`:

    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

The enabled linters are declared in `.golangci.yml`.
The same checks run on every push and pull request through `.github/workflows/ci.yml`.

The test of a 256 MiB upload is the slow one; skip it with `go test -short ./...`.


Layout
---------

    main.go                                 entry point, flags, startup banner
    main_test.go                            tests of the handler chain main builds
    internal/auth/auth.go                   Basic auth credentials and the gate
    internal/auth/auth_test.go              tests
    internal/uploader/handler.go            /form.html, /upload.html handlers
    internal/uploader/handler_test.go       tests
    internal/uploader/templates/            HTML templates, embedded with go:embed

`Handler.storeRequest` walks the multipart body part by part and hands the file part to `Store`, which streams it to its destination.
Nothing reads the body as a whole, which is what keeps the memory flat on a large upload; keep it that way when changing the upload path.

`auth.Require` wraps the whole mux in `buildHandler`, so a route added to `Handler.Routes` is protected the moment it is registered: there is no per-handler auth code to forget.
`--no-auth` is the single place that removes the wrapper.
The 401 response must keep its `WWW-Authenticate` header, otherwise browsers never show the login prompt and the form becomes unreachable.
Credentials are compared through `crypto/subtle` on SHA-256 digests, so neither the password nor its length leaks through timing.


Releasing
---------

The procedure lives in the `release` skill (`.claude/skills/release/`).
Write the `## vX.Y.Z` section in `CHANGELOG.md`, update the pinned-version example in `README.md`, commit and push, then run:

    .claude/skills/release/scripts/release.sh vX.Y.Z

The script verifies the repo (clean tree, master in sync with origin, tag free, CHANGELOG section present, README bumped), runs `make ci`, cross-compiles one binary per platform into `dist/` (`build-dist.sh`), tags, and publishes the GitHub release with the CHANGELOG section as its notes.
The download instructions in `README.md` point at `releases/latest`, so they need no update; only the pinned example does.
