Uploader
=========
Uploader puts a file on a server through a web browser.

Run the binary on the machine that should receive the files, open the page it
serves, type the directory to write to and pick a file. That is the whole tool.
It is handy when SSH is not at hand: from a phone, from a borrowed machine, or
for someone who should be able to drop a file on a server without a shell
account.

It is a single binary with nothing to install and no configuration, it starts in
a moment, and an upload is streamed to disk, so the size of a file is not a
concern.

Note that Uploader has no authentication: anyone who can reach the port can
write a file anywhere the process has permission to. Run it on a trusted
network, or behind something that does the authentication for it, and stop it
when you are done.


Using Uploader
---------

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


Large files
---------

The request body is streamed part by part straight to the destination file, so
an upload is bounded by the free space of the target disk, not by memory. A 1
GiB upload keeps the process under 10 MiB of RSS.

The `path` field must be sent before the `file` field, because the destination
must be known before the content is written. The HTML form already does this;
keep the same order when posting with a tool such as curl:

    curl -F "path=/tmp/uploads/" -F "file=@big.iso" http://localhost:8080/upload.html


---------

Changes of each version are listed in [CHANGELOG.md](CHANGELOG.md).
To build or modify the uploader, see [DEVELOPMENT.md](DEVELOPMENT.md).
