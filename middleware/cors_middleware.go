package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type corsConfig struct {
	allowCredentials bool
	allowedOrigins   map[string]bool
	allowAll         bool
	exposeHeaders    string
	allowedMethods   string
	allowedHeaders   string
	maxAge           string
}

func CorsMiddleware(_ context.Context, next http.HandlerFunc, conf Config) http.HandlerFunc {
	c, err := parseCorsConfig(conf)
	if err != nil {
		zap.S().Errorf("CorsMiddleware: failed to parse configuration: %v. Failing closed.", err)
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "CorsMiddleware misconfigured", http.StatusServiceUnavailable)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "" {
			next(w, r)
			return
		}

		if !c.allowAll && !c.allowedOrigins[origin] {
			next(w, r)
			return
		}

		if c.allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}

		if c.allowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Preflight request
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			if c.allowedMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", c.allowedMethods)
			}
			if c.allowedHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", c.allowedHeaders)
			}
			if c.maxAge != "" {
				w.Header().Set("Access-Control-Max-Age", c.maxAge)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if c.exposeHeaders != "" {
			w.Header().Set("Access-Control-Expose-Headers", c.exposeHeaders)
		}

		next(w, r)
	}
}

func parseCorsConfig(conf Config) (*corsConfig, error) {
	if conf.Options == nil {
		// No options = allow everything (permissive default)
		return &corsConfig{allowAll: true}, nil
	}

	origins, err := ParseStringSliceOpt(conf.Options, "allowed-origins", []string{"*"})
	if err != nil {
		return nil, err
	}

	methods, err := ParseStringSliceOpt(conf.Options, "allowed-methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	if err != nil {
		return nil, err
	}

	headers, err := ParseStringSliceOpt(conf.Options, "allowed-headers", []string{"Content-Type", "Authorization"})
	if err != nil {
		return nil, err
	}

	expose, err := ParseStringSliceOpt(conf.Options, "expose-headers", nil)
	if err != nil {
		return nil, err
	}

	allowCredentials := ParseBoolOpt(conf.Options, "allow-credentials", false)
	maxAge := ParseIntOpt(conf.Options, "max-age", 0)

	c := &corsConfig{
		allowedOrigins:   make(map[string]bool, len(origins)),
		allowedMethods:   strings.Join(methods, ", "),
		allowedHeaders:   strings.Join(headers, ", "),
		allowCredentials: allowCredentials,
	}

	if expose != nil {
		c.exposeHeaders = strings.Join(expose, ", ")
	}

	if maxAge > 0 {
		c.maxAge = strconv.Itoa(maxAge)
	}

	for _, o := range origins {
		if o == "*" {
			c.allowAll = true
			break
		}
		c.allowedOrigins[o] = true
	}

	return c, nil
}
