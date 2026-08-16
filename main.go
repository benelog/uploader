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

	// A nil creds means --no-auth: the pairing of "is auth on?" and "with
	// which credentials?" travels as one value.
	var creds *auth.Credentials
	if !*noAuth {
		c, err := auth.Resolve(*username, *password)
		if err != nil {
			log.Fatalf("cannot generate a password: %v", err)
		}
		creds = &c
	}

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(*httpPort),
		Handler: buildHandler(handler.Routes(), creds),
		// Only the headers are given a deadline: a large upload may legitimately
		// take as long as the network needs.
		ReadHeaderTimeout: 20 * time.Second,
	}

	printAuthPolicy(os.Stdout, creds, *httpPort)

	log.Printf("uploader is listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// buildHandler puts the whole mux behind the Basic auth gate, so a route added
// later is protected without any per-handler code. Only a nil creds
// (--no-auth) serves the routes bare.
func buildHandler(routes http.Handler, creds *auth.Credentials) http.Handler {
	if creds == nil {
		return routes
	}
	return auth.Require(*creds, routes)
}

// printAuthPolicy explains the credentials policy at boot. A generated
// password is printed because this is the only chance to see it; an
// operator-supplied one is never echoed.
func printAuthPolicy(out io.Writer, creds *auth.Credentials, port int) {
	const rule = "-----------------------------"

	lines := []string{rule}
	if creds == nil {
		lines = append(lines,
			"Authentication: DISABLED by --no-auth.",
			"   Anyone who can reach this port can write files anywhere this",
			"   process has permission to. Use it on a trusted network only,",
			"   and stop the server when you are done.")
	} else {
		origin := "  (from --user)"
		if creds.DefaultedUsername {
			origin = "  (OS login account)"
		}
		lines = append(lines,
			"Authentication: HTTP Basic auth is required on every URL.",
			"   Username : "+creds.Username+origin)
		example := "<password>"
		if creds.GeneratedPassword {
			example = creds.Password
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
				creds.Username, example),
			fmt.Sprintf("        'http://localhost:%d/upload.html'", port),
			"",
			"   Pass --no-auth to serve without any authentication.")
	}
	lines = append(lines, rule, "")

	_, _ = io.WriteString(out, strings.Join(lines, "\n")+"\n")
}
