# Changelog

## v2.1.0

Uploader now asks for a password.

### Highlights

- **Authenticated by default, with zero setup.** `/`, `/form.html` and
  `/upload.html` are behind HTTP Basic auth from the first run: no config file
  and no flag to remember, because the username defaults to the OS login
  account and the password is generated at startup and printed once.
- **Fixed credentials when you need them.** `--user` and `--password` override
  either half; a password given on the command line is never echoed to the
  console.
- **The policy is explained at boot.** The startup banner names the protected
  URLs, where each credential came from, and a ready-to-paste `curl -u` line.
- **`--no-auth` keeps the old behaviour** for a trusted network, and says
  loudly at startup that anyone reaching the port can write files. It is
  rejected together with `--user` or `--password`.

### Notes

- A generated password changes on every restart, so scripted clients should
  either read it from the startup output or pin it with `--password`.
- The traffic is still plain HTTP: the credentials travel unencrypted, so this
  remains a trusted-network tool. Put it behind a reverse proxy for TLS.

## v2.0.0

Uploader is rewritten in Go. It is now a single binary with no dependency:
no JVM, no servlet container, no WAR to deploy.

### Highlights

- **Single binary.** Download it, run it. The HTML templates are compiled into
  the binary with `go:embed`, so there is nothing to unpack next to it.
- **Uploads of any size.** The multipart body is streamed part by part directly
  to the destination file. Nothing is buffered in memory and nothing is spooled
  to a temporary file first, so an upload is bounded only by the free space of
  the target disk. A 1 GiB upload keeps the process under 10 MiB of RSS.
- **Same URLs.** `/form.html`, `/upload.html` and the redirect from `/` behave
  as before, so existing links and bookmarks keep working.
- **Same option.** `--httpPort` still selects the port, 8080 by default.

### Fixes

- An uploaded file name can no longer escape the requested directory: only the
  last element of the submitted name is used, so a name such as
  `../../etc/passwd` is written inside the chosen path.
- A server path without a trailing slash now works. The previous version
  concatenated the path and the file name as plain strings, which silently
  wrote the file next to the directory instead of inside it.
- Submitting an empty file no longer truncates a file of the same name at the
  destination; the upload is reported as `empty file` and nothing is written.
- A failed write no longer leaves a half-written file behind.

### Breaking changes

- **Deployment.** `java -jar uploader.jar` is replaced by a native binary per
  platform, published as a release asset. The WAR packaging, the Maven build
  and the Spring configuration files are gone.
- **Field order.** The `path` field must be sent before the `file` field, since
  the destination has to be known before the content is streamed. The HTML form
  already sends them in this order; scripted clients posting `file` first now
  receive `400 Bad Request` with an explanatory message instead of writing to an
  unexpected directory.
- **Method.** `/upload.html` accepts `POST` only and answers `405` otherwise.

### Development

Quality tooling follows https://blog.benelog.net/go-quality-tools :
`make check` runs goimports, golangci-lint and the tests; `make ci` is what the
GitHub Actions workflow runs on every push and pull request.
