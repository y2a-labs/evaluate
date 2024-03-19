package commands

import (
	"fmt"
	"log"
	"net/http"
	"script_validation/api"
	service "script_validation/services"
	"script_validation/static"
	"script_validation/templates"
	web "script_validation/web/handlers"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
)

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
		fmt.Println(r.URL.Path)
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") && r.Header.Get("X-Redirected-From") == "" {
			w.Header().Set("X-Redirected-From", r.URL.Path)
			http.Redirect(w, r, strings.TrimRight(r.URL.Path, "/"), http.StatusMovedPermanently)
			//return
		}
		next.ServeHTTP(w, r)
	})
}

func StartServer(port string, dev bool) {
	options := []func(*fuego.Server){
		fuego.WithPort(":" + port),
		fuego.WithTemplateGlobs("./**/*.html"),
	}

	if !dev {
		options = append([]func(*fuego.Server){fuego.WithTemplateFS(templates.FS)}, options...)
	}

	server := fuego.NewServer(options...)

	// Reparses the templates html/templates on every request
	if dev { server.DevMode() }

	service := service.New("./data/data.db", "./.env")

	// Logs the requests
	fuego.Use(server, logRequest)

	webResources := web.Resources{Service: service}
	webGroup := fuego.Group(server, "/")

	// Removes trailing slashes from URLs
	fuego.Use(webGroup, removeURLTrailingSlash)

	// Serve the static files
	fs := http.FileServerFS(static.FS)
	fuego.Handle(webGroup, "/static/", http.StripPrefix("/static/", fs))

	webResources.RegisterTestRoutes(webGroup)
	webResources.RegisterConversationRoutes(webGroup)
	webResources.RegisterLLMRoutes(webGroup)
	webResources.RegisterMessageRoutes(webGroup)
	webResources.RegisterPromptRoutes(webGroup)
	webResources.RegisterProviderRoutes(webGroup)
	webResources.RegisterMessageMetadataRoutes(webGroup)

	apiResources := api.Resources{Service: service}

	// Create a proxy server
	fuego.Post(server, "/v1/chat/completions", apiResources.ProxyOpenai)

	apiGroup := fuego.Group(server, "/v1/api")
	apiResources.RegisterConversationRoutes(apiGroup)
	apiResources.RegisterLLMRoutes(apiGroup)
	apiResources.RegisterMessageRoutes(apiGroup)
	apiResources.RegisterPromptRoutes(apiGroup)
	apiResources.RegisterProviderRoutes(apiGroup)
	apiResources.RegisterMessageMetadataRoutes(apiGroup)

	err := server.Run()
	if err != nil {
		fmt.Println(err)
	}
}
