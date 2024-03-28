package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

// node is used to build a tree that will traversed for each request.
type node struct {
	// Url is used if there are no more parameters, e.g. this node is exactly the one requested.
	Url string `json:"url"`
	// Children is checked for the next argument that is in list which we are traversing. If a
	// child exists for a given key it is responsible to resolve the rest of the query.
	Children map[string]*node `json:"children"`
	// Template is a last resort, if more keys exist they are formatted into the template string.
	// If there is more than one argument they will be joined by spaces and url-encoded.
	Template string `json:"template"`
}

func (n *node) resolve(p []string) (string, error) {
	if len(p) == 0 && len(n.Url) == 0 {
		return "", ErrNotFound
	}

	if len(p) == 0 {
		return n.Url, nil
	}

	c, ok := n.Children[p[0]]
	if ok {
		return c.resolve(p[1:])
	}

	if n.Template == "" {
		return "", ErrNotFound
	}

	slog.Debug("formatting", "template", n.Template, "parameters", p)

	return fmt.Sprintf(n.Template, url.PathEscape(strings.Join(p, " "))), nil
}

func (n *node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		p   []string
		to  string
		err error
	)

	p = strings.Split(r.URL.Query().Get("to"), " ")

	to, err = n.resolve(p)
	slog.Info("handling request", "to", to, "parts", strings.Join(p, "/"), "error", err)
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Location", to)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
}

func main() {
	err := Main()
	if err != nil {
		slog.Error("execution failed", "error", err)
		os.Exit(1)
	}
}

func Main() error {
	configPath := filepath.Join(os.Getenv("HOME"), ".config/r/config.json")
	if len(os.Args) == 2 {
		configPath = os.Args[1]
	} else if len(os.Args) > 2 {
		slog.Error("usage: r <config.json>")
		return fmt.Errorf("unknown arguments")
	}

	cf, err := os.Open(configPath)
	if err != nil {
		return err
	}

	n := &node{}
	err = json.NewDecoder(cf).Decode(n)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "7091"
	}

	err = http.ListenAndServe(":"+port, n)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}
