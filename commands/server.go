package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"script_validation/api"
	service "script_validation/services"
	web "script_validation/web/handlers"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fsnotify/fsnotify"
	"github.com/go-fuego/fuego"
)

func addURLPathToContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the URL path from the request
		path := r.URL.Path
		// Create a new context with the URL path added
		ctx := context.WithValue(r.Context(), "path", path)

		// Create a new request with the updated context
		reqWithCtx := r.WithContext(ctx)

		// Call the next handler with the new request
		next.ServeHTTP(w, reqWithCtx)
	})
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows us to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// logRequest is a middleware that logs the HTTP method, URI, status code, and the time it took to process the request.
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap the original http.ResponseWriter with our custom writer that captures the status code.
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK} // Default to 200 OK

		// Use defer to ensure logging occurs even if there's a panic.
		defer func() {
			if err := recover(); err != nil {
				// Log the panic as an internal server error (500).
				log.Printf("%s %s %d %s [ERROR: %v]", r.Method, r.RequestURI, http.StatusInternalServerError, time.Since(start), err)
				// Write the internal server error status code if it hasn't been written yet.
				if wrapper.statusCode == http.StatusOK {
					wrapper.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				// Log the request details as normal.
				log.Printf("%s %s %d %s", r.Method, r.RequestURI, wrapper.statusCode, time.Since(start))
			}
		}()

		next.ServeHTTP(wrapper, r) // Pass the request to the actual handler
	})
}

func removeURLTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, strings.TrimRight(r.URL.Path, "/"), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func StartServer() {
	server := fuego.NewServer(
		fuego.WithPort(":3000"),
		//fuego.WithTemplateFS(templates.FS),
		fuego.WithTemplateGlobs("./**/*.html"),
	)
	server.DevMode()
	service := service.New("test.db", "./.env")

	fuego.Use(server, logRequest)

	webResources := web.Resources{Service: service}
	webGroup := fuego.Group(server, "/")
	fuego.Use(webGroup, addURLPathToContextMiddleware)
	fuego.Use(webGroup, removeURLTrailingSlash)

	// Serve the static files
	staticFiles := http.FileServer(http.Dir("./static"))
	fuego.Handle(webGroup, "/static/", http.StripPrefix("/static/", staticFiles))

	fuego.GetStd(webGroup, "/sse", sseHandler)
	webResources.RegisterAgentRoutes(webGroup)
	webResources.RegisterConversationRoutes(webGroup)
	webResources.RegisterLLMRoutes(webGroup)
	webResources.RegisterMessageRoutes(webGroup)
	webResources.RegisterPromptRoutes(webGroup)
	webResources.RegisterProviderRoutes(webGroup)
	webResources.RegisterMessageMetadataRoutes(webGroup)

	apiResources := api.Resources{Service: service}

	fuego.Post(server, "/v1/chat/completions", apiResources.ProxyOpenai)

	apiGroup := fuego.Group(server, "/v1/api")
	apiResources.RegisterAgentRoutes(apiGroup)
	apiResources.RegisterConversationRoutes(apiGroup)
	apiResources.RegisterLLMRoutes(apiGroup)
	apiResources.RegisterMessageRoutes(apiGroup)
	apiResources.RegisterPromptRoutes(apiGroup)
	apiResources.RegisterProviderRoutes(apiGroup)
	apiResources.RegisterMessageMetadataRoutes(apiGroup)

	server.Run()
}

func fetchHTMLFromPath(url string) (string, error) {
	// Make a GET request to the path
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the node by ID and get all its contents
	contentHtml, err := doc.Find("#app").Html()
	if err != nil {
		log.Fatal(err)
	}

	// Remove all line breaks
	contentHtml = strings.ReplaceAll(contentHtml, "\n", "")

	return contentHtml, nil
}

func sseHandler(writer http.ResponseWriter, r *http.Request) {
	// Set necessary headers for SSE
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	referer := r.Header.Get("Referer")

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.Walk("/home/epentland/ai/2yfv/script_validation/templates/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Use a loop to handle events
	for {
		select {
		case event := <-watcher.Events:
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)

				// Fetch the HTML from the path
				html, err := fetchHTMLFromPath(referer)
				if err != nil {
					log.Println("Error fetching HTML:", err)
					return
				}

				fmt.Fprintf(writer, "data: %s\n\n", html)

				// Flush the data immediately
				if f, ok := writer.(http.Flusher); ok {
					f.Flush()
				} else {
					log.Println("Unable to flush")
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	} else {
		fmt.Println("problem flushhing")
	}
}
