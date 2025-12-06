package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
)

func NewServer() *http.Server {
	port := 8080
	idleTimeout := time.Minute
	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second
	if config.Instance != nil {
		port = config.Instance.Server.Port
		idleTimeout = config.Instance.Server.IdleTimeout
		readTimeout = config.Instance.Server.ReadTimeout
		writeTimeout = config.Instance.Server.WriteTimeout
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      RegisterRoutes(),
		IdleTimeout:  idleTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}
