package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/benelog/uploader/internal/auth"
	"github.com/benelog/uploader/internal/uploader"
)

func main() {
	httpPort := flag.Int("httpPort", 8080, "http port to listen on")
	username := flag.String("user", "", "HTTP Basic auth username (default: the OS login account)")
	password := flag.String("password", "", "HTTP Basic auth password (default: randomly generated at startup)")
	noAuth := flag.Bool("no-auth", false, "serve without any authentication (anyone reaching the port can write files)")
	flag.Parse()

	if *noAuth && (*username != "" || *password != "") {
		log.Fatal("--no-auth cannot be combined with --user or --password")
	}

	handler, err := uploader.NewHandler()
	if err != nil {
		log.Fatalf("failed to initialize uploader: %v", err)
	}

	var creds auth.Credentials
	if !*noAuth {
		if creds, err = auth.Resolve(*username, *password); err != nil {
			log.Fatalf("cannot generate a password: %v", err)
		}
	}

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(*httpPort),
		Handler: buildHandler(handler.Routes(), creds, *noAuth),
		// Only the headers are given a deadline: a large upload may legitimately
		// take as long as the network needs.
		ReadHeaderTimeout: 20 * time.Second,
	}

	printAuthPolicy(os.Stdout, creds, *noAuth, *httpPort)

	log.Printf("uploader is listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// buildHandler puts the whole mux behind the Basic auth gate, so a route added
// later is protected without any per-handler code. Only --no-auth removes it.
func buildHandler(routes http.Handler, creds auth.Credentials, authDisabled bool) http.Handler {
	if authDisabled {
		return routes
	}
	return auth.Require(creds, routes)
}

// printAuthPolicy explains the credentials policy at boot. A generated
// password is printed because this is the only chance to see it; an
// operator-supplied one is never echoed.
func printAuthPolicy(out io.Writer, creds auth.Credentials, authDisabled bool, port int) {
	const rule = "-----------------------------"

	lines := []string{rule}
	if authDisabled {
		lines = append(lines,
			"Authentication: DISABLED by --no-auth.",
			"   Anyone who can reach this port can write files anywhere this",
			"   process has permission to. Use it on a trusted network only,",
			"   and stop the server when you are done.",
			rule, "")
	} else {
		lines = append(lines,
			"Authentication: HTTP Basic auth is required on every URL",
			"   (/, /form.html, /upload.html).",
			"   Username : "+creds.Username+usernameOrigin(creds))
		if creds.GeneratedPassword {
			lines = append(lines,
				"   Password : "+creds.Password+"  (randomly generated for this run)",
				"   The password changes on every restart. Use --user / --password to fix it.")
		} else {
			lines = append(lines, "   Password : (hidden; the one given with --password)")
		}
		lines = append(lines,
			"",
			"   Browsers prompt for these credentials on the first request.",
			"   From the command line:",
			fmt.Sprintf("      curl -u '%s:%s' -F 'path=/tmp/' -F 'file=@big.iso' \\",
				creds.Username, passwordForExample(creds)),
			fmt.Sprintf("        'http://localhost:%d/upload.html'", port),
			"",
			"   Pass --no-auth to serve without any authentication.",
			rule, "")
	}

	_, _ = io.WriteString(out, strings.Join(lines, "\n")+"\n")
}

func usernameOrigin(creds auth.Credentials) string {
	if creds.DefaultedUsername {
		return "  (OS login account)"
	}
	return "  (from --user)"
}

func passwordForExample(creds auth.Credentials) string {
	if creds.GeneratedPassword {
		return creds.Password
	}
	return "<password>"
}
