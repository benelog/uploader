Developing Uploader
=========
Go 1.22 or later is the only requirement to build and test. The uploader has no
third party dependency.

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

The enabled linters are declared in `.golangci.yml`. The same checks run on
every push and pull request through `.github/workflows/ci.yml`.

The test of a 256 MiB upload is the slow one; skip it with `go test -short ./...`.


Layout
---------

    main.go                                 entry point, --httpPort flag
    internal/uploader/handler.go            /form.html, /upload.html handlers
    internal/uploader/handler_test.go       tests
    internal/uploader/templates/            HTML templates, embedded with go:embed

`Handler.storeRequest` walks the multipart body part by part and hands the file
part to `Store`, which streams it to its destination. Nothing reads the body as
a whole, which is what keeps the memory flat on a large upload; keep it that way
when changing the upload path.


Releasing
---------

Build one binary per platform, tag, then publish the assets:

    rm -rf dist && mkdir -p dist
    for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
      GOOS=${target%/*}; GOARCH=${target#*/}
      out=dist/uploader-$GOOS-$GOARCH
      [ "$GOOS" = windows ] && out=$out.exe
      GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $out .
    done

    git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z
    gh release create vX.Y.Z dist/* --title "vX.Y.Z" --notes-file notes.md

Add the section of the new version to `CHANGELOG.md` first and use it as the
release notes. The download instructions in `README.md` point at
`releases/latest`, so they need no update; only the pinned example does.
