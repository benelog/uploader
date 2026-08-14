Using Uploader
=========
Download the binary for your platform from the
[latest release](https://github.com/benelog/uploader/releases/latest):

	curl -LO https://github.com/benelog/uploader/releases/latest/download/uploader-linux-amd64
	chmod +x uploader-linux-amd64

Available assets: `uploader-linux-amd64`, `uploader-linux-arm64`,
`uploader-darwin-amd64`, `uploader-darwin-arm64`, `uploader-windows-amd64.exe`.
To pin a version, replace `latest/download` with `download/v2.0.0`.

Execute

    ./uploader-linux-amd64

Port 8080 is the default http port. You can use "--httpPort" option to change it.

    ./uploader-linux-amd64 --httpPort=10023

Then open http://localhost:8080/ , fill in the server path and pick a file.

There is nothing to install: the binary has no dependency and the templates
are embedded in it.


Large files
---------

The request body is streamed part by part straight to the destination file, so
an upload is bounded by the free space of the target disk, not by memory. A 1
GiB upload keeps the process under 10 MiB of RSS.

The `path` field must be sent before the `file` field, because the destination
must be known before the content is written. The HTML form already does this;
keep the same order when posting with a tool such as curl:

    curl -F "path=/tmp/uploads/" -F "file=@big.iso" http://localhost:8080/upload.html


Build from source
---------

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


Layout
---------

    main.go                                 entry point, --httpPort flag
    internal/uploader/handler.go            /form.html, /upload.html handlers
    internal/uploader/templates/            HTML templates, embedded with go:embed
