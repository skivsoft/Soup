package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
)

// Default values of the settings
var settings = Settings{
	ServerHost: "http://localhost",
	ServerPort: 8080,
}

// Initialization
func init() {
	settings.Load()
}

//go:embed public
var staticFS embed.FS

// Main program
func main() {
	fmt.Printf("SoapUI <==> SOUP <==> %s\n", settings.TargetUrl)
	fmt.Println()

	publicFS, err := fs.Sub(staticFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	// Routes
	http.Handle("/", http.FileServer(http.FS(publicFS)))
	http.HandleFunc("/proxy", handleRequestAndRedirect)

	// Starting server
	log.Printf("Starting SOAP proxy at %s:%d ...\n", settings.ServerHost, settings.ServerPort)
	log.Printf("Provide SoapUI with following Initial WSDL: %s:%d/proxy?wsdl", settings.ServerHost, settings.ServerPort)
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", settings.ServerPort),
	}
	server.ListenAndServe()
}

// Handle requests and send it to the TARGET_URL environment variable.
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)

	if req.Method == http.MethodPost {
		xml, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return
		}
		err = req.Body.Close()
		if err != nil {
			return
		}

		xml = updateXmlBeforeSending(xml)
		log.Println("Body to send:")
		log.Printf("%s", xml)

		body := ioutil.NopCloser(bytes.NewReader(xml))
		req.Body = body
		req.ContentLength = int64(len(xml))
		req.Header.Set("Content-Length", strconv.Itoa(len(xml)))
	}

	url := settings.TargetUrl
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = rewriteBody

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non-blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

// Update SOAP service response
func rewriteBody(res *http.Response) error {
	xml, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = res.Body.Close()
	if err != nil {
		return err
	}

	log.Println("Received body:")
	log.Printf("%s", xml)
	xml = replaceReceivedXml(xml)

	body := ioutil.NopCloser(bytes.NewReader(xml))
	res.Body = body
	res.ContentLength = int64(len(xml))
	res.Header.Set("Content-Length", strconv.Itoa(len(xml)))
	return nil
}

// Replace links in received xml from target SOAP service
func replaceReceivedXml(xml []byte) []byte {
	log.Printf("Updating xml received from %s\n", settings.TargetUrl)
	replaceText := fmt.Sprintf("%s:%d/proxy", settings.ServerHost, settings.ServerPort)
	xml = bytes.Replace(xml, []byte(settings.ReplaceUrl.String()), []byte(replaceText), -1)
	xml = bytes.Replace(xml, []byte(settings.TargetUrl.String()), []byte(replaceText), -1)
	return xml
}

// Replace links in source xml before sending to target SOAP service
func updateXmlBeforeSending(xml []byte) []byte {
	log.Printf("Updating xml before sending to %s\n", settings.TargetUrl)
	searchText := fmt.Sprintf("%s:%d/proxy", settings.ServerHost, settings.ServerPort)
	xml = bytes.Replace(xml, []byte(searchText), []byte(settings.ReplaceUrl.String()), -1)
	return xml
}
