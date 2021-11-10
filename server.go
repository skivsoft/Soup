package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/subosito/gotenv"
)

// Initialization
func init() {
	gotenv.Load()
}

// Main program.
func main() {
	fmt.Println("Soup  v1.0")
	fmt.Println("SoupUI <==> Soup <==> Target SOAP service")
	fmt.Println()

	port := os.Getenv("PORT")
	log.Println("Starting SOAP proxy at http://localhost:" + port + " ...")

	http.HandleFunc("/", handleRequestAndRedirect)
	http.ListenAndServe(":"+port, nil)
}

// Handle requests and send it to the TARGET_URL environment variable.
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(os.Getenv("TARGET_URL"))
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
