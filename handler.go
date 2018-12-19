// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// csr-meta serves Go vanity URLs.
package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type handler struct {
	cacheControl string
}

type pathConfig struct {
	path string
	repo string
}

func newHandler(cacheAge time.Duration) (*handler, error) {
	h := &handler{}
	if cacheAge < 0 {
		return nil, errors.New("cache_max_age is negative")
	}
	h.cacheControl = fmt.Sprintf("public, max-age=%d", cacheAge/time.Second)
	return h, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	current := r.URL.Path
	if strings.HasSuffix(current, "/") {
		current = strings.TrimSuffix(current, "/")
	}

	parts := strings.SplitN(current, "/", 4)

	switch len(parts) {
	case 3:
		parts = append(parts, "")
	case 4:
		// NOP
	default:
		http.NotFound(w, r)
		return
	}

	proj, repo, subpath := parts[1], parts[2], parts[3]
	if proj == "" || repo == "" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", h.cacheControl)
	if err := vanityTmpl.Execute(w, struct {
		Import  string
		Subpath string
		Repo    string
	}{
		Import:  fmt.Sprintf("%s/%s/%s", defaultHost(r), proj, repo),
		Subpath: subpath,
		Repo:    fmt.Sprintf("https://source.developers.google.com/p/%s/r/%s", proj, repo),
	}); err != nil {
		http.Error(w, "cannot render the page", http.StatusInternalServerError)
	}
}

var vanityTmpl = template.Must(template.New("vanity").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Import}} git {{.Repo}}">
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.Import}}/{{.Subpath}}">
</head>
<body>
Nothing to see here; <a href="https://godoc.org/{{.Import}}/{{.Subpath}}">see the package on godoc</a>.
</body>
</html>`))
