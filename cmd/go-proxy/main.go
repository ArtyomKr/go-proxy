package main

import (
	"crypto/tls"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Couldn't find env variables in .env file")
	}

	port := os.Getenv("PORT")
	urlToProxy := os.Getenv("TARGET_URL")
	certFile := os.Getenv("CERT_FILE") // Path to SSL certificate
	keyFile := os.Getenv("KEY_FILE")   // Path to SSL private key

	log.Printf("Url %v", urlToProxy)

	targetURL, err := url.Parse(urlToProxy)
	if err != nil {
		log.Fatal("Failed to parse url", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.Transport = &http.Transport{
		ProxyConnectHeader: http.Header{},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	proxy.Director = func(req *http.Request) {
		req.URL.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
		req.Host = targetURL.Host
	}

	proxy.ModifyResponse = func(res *http.Response) error {
		log.Printf("Response %v", res.Status)

		// Prevent redirects
		if res.StatusCode >= 300 && res.StatusCode <= 399 {
			res.StatusCode = 200
		}

		return nil
	}

	proxyHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Proxying %v", req.URL)
		proxy.ServeHTTP(w, req)
	}

	http.HandleFunc("/", proxyHandler)

	log.Println("Listing for requests at https://localhost:" + port)
	log.Fatal(http.ListenAndServeTLS(":"+port, certFile, keyFile, nil))
}
