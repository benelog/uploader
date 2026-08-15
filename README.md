Uploader
=========
Uploader puts a file on a server through a web browser: run the binary on the
receiving machine, open the page it serves, type the destination directory and
pick a file.

- Handy when SSH is not at hand
- A single binary with nothing to install and no configuration
- Uploads are streamed to disk, so file size is not a concern

Note that Uploader has no authentication: anyone who can reach the port can
write files anywhere the process has permission to. Run it on a trusted network
or behind something that authenticates for it, and stop it when you are done.


Using Uploader
---------

Download the binary for your platform from the
[latest release](https://github.com/benelog/uploader/releases/latest):

	curl -LO https://github.com/benelog/uploader/releases/latest/download/uploader-linux-amd64
	chmod +x uploader-linux-amd64

Available assets: `uploader-linux-amd64`, `uploader-linux-arm64`,
`uploader-darwin-amd64`, `uploader-darwin-arm64`, `uploader-windows-amd64.exe`.
To pin a version, replace `latest/download` with `download/v2.0.0`.

Run it, then open http://localhost:8080/ , fill in the server path and pick a
file:

    ./uploader-linux-amd64

The default port is 8080; change it with `--httpPort`:

    ./uploader-linux-amd64 --httpPort=10023


Large files
---------

The request body is streamed straight to the destination file, so an upload is
bounded by the free space of the target disk, not by memory: a 1 GiB upload
keeps the process under 10 MiB of RSS.

The `path` field must precede the `file` field, since the destination must be
known before the content arrives. The HTML form already does this; keep the
same order when posting with a tool such as curl:

    curl -F "path=/tmp/uploads/" -F "file=@big.iso" http://localhost:8080/upload.html


---------

Changes of each version are listed in [CHANGELOG.md](CHANGELOG.md).
To build or modify the uploader, see [DEVELOPMENT.md](DEVELOPMENT.md).
Uploader is released under the MIT License; see [LICENSE](LICENSE).
