package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

type application struct {
	auth struct {
		username string
		password string
	}
	gcp_conf struct {
		gcloud_creds_path string
		gcs_bucket_name   string
	}
}

func (app *application) create_signed_url(w http.ResponseWriter, r *http.Request) {
	log.Printf("- %s - %s - %s", r.RemoteAddr, r.RequestURI, r.Header.Get("User-Agent"))
	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	decoded_uri, err := url.QueryUnescape(r.RequestURI[1:])

	if err != nil {
		log.Fatalf("Failed to decode request: %v", err)
	}

	expires := time.Now().Add(time.Minute * 15)
	url, err := client.Bucket(app.gcp_conf.gcs_bucket_name).SignedURL(decoded_uri, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: expires,
		Scheme:  storage.SigningSchemeV4,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error " + err.Error())
	}
	// we use StatusFound here because StatusMovedPermanently can mean the redirect gets cached
	// Chrome does this, we don't want expired tokens to get cached client-side
	http.Redirect(w, r, url, http.StatusFound)

}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func main() {
	app := new(application)

	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	app.gcp_conf.gcs_bucket_name = os.Getenv("GCS_BUCKET_NAME")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	if app.gcp_conf.gcs_bucket_name == "" {
		log.Fatal("A bucket name must be specified")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.basicAuth(app.create_signed_url))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("starting server on %s", srv.Addr)
	err := srv.ListenAndServe()
	log.Fatal(err)
}
