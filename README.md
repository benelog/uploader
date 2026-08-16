Uploader
=========
Uploader puts a file on a server through a web browser: run the binary on the receiving machine, open the page it serves, type the destination directory and pick a file.

- Handy when SSH is not at hand
- A single binary with nothing to install and no configuration
- Uploads are streamed to disk, so file size is not a concern
- Password protected from the first run, with nothing to set up

Every URL is behind HTTP Basic auth (see [Authentication](#authentication)), but the traffic is plain HTTP with no TLS, so the credentials travel unencrypted, and whoever holds them can write files anywhere the process has permission to.
Run it on a trusted network, and stop it when you are done.


Using Uploader
---------

Download the binary for your platform from the [latest release](https://github.com/benelog/uploader/releases/latest):

	curl -LO https://github.com/benelog/uploader/releases/latest/download/uploader-linux-amd64
	chmod +x uploader-linux-amd64

Available assets: `uploader-linux-amd64`, `uploader-linux-arm64`, `uploader-darwin-amd64`, `uploader-darwin-arm64`, `uploader-windows-amd64.exe`.
To pin a version, replace `latest/download` with `download/v2.0.0`.

Run it, then open http://localhost:8080/ , fill in the server path and pick a file:

    ./uploader-linux-amd64

The default port is 8080; change it with `--httpPort`:

    ./uploader-linux-amd64 --httpPort=10023


Authentication
---------

Every URL requires **HTTP Basic auth**. By default:

- the **username** is the OS account the server runs as;
- the **password** is randomly generated (128 bits) each time the server starts, and printed to the console once, at startup.

The policy is spelled out at boot:

    -----------------------------
    Authentication: HTTP Basic auth is required on every URL.
       Username : alice  (OS login account)
       Password : 4jddSK1CdMHklMzAifdkiQ  (randomly generated for this run)
       The password changes on every restart. Use --user / --password to fix it.

       Browsers prompt for these credentials on the first request.
       From the command line:
          curl -u 'alice:4jddSK1CdMHklMzAifdkiQ' -F 'path=/tmp/' -F 'file=@big.iso' \
            'http://localhost:8080/upload.html'

       Pass --no-auth to serve without any authentication.
    -----------------------------

A browser shows its usual login prompt on the first request.
Since a generated password changes on every restart, set both explicitly when you want stable credentials for a script or a service unit:

    ./uploader-linux-amd64 --user=ops --password='s3cr3t'

A password given with `--password` is never echoed to the console.

| Option              | Default                       |
| ------------------- | ----------------------------- |
| `--user <name>`     | the OS login account          |
| `--password <pass>` | randomly generated at startup |
| `--no-auth`         | off, that is, auth is required |

`--no-auth` serves the old, unauthenticated uploader: anyone who reaches the port can then write files.
It cannot be combined with `--user` or `--password`, and the startup banner says loudly that authentication is off.


Large files
---------

The request body is streamed straight to the destination file, so an upload is bounded by the free space of the target disk, not by memory: a 1 GiB upload keeps the process under 10 MiB of RSS.

The `path` field must precede the `file` field, since the destination must be known before the content arrives.
The HTML form already does this; keep the same order when posting with a tool such as curl (`-u` carries the credentials printed at startup):

    curl -u 'alice:4jddSK1CdMHklMzAifdkiQ' \
      -F "path=/tmp/uploads/" -F "file=@big.iso" http://localhost:8080/upload.html


---------

Changes of each version are listed in [CHANGELOG.md](CHANGELOG.md).
To build or modify the uploader, see [DEVELOPMENT.md](DEVELOPMENT.md).
Uploader is released under the MIT License; see [LICENSE](LICENSE).
