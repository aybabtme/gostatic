# gostatic

Command `gostatic` takes a list of directories and create a Go file
containing all of those directories that you can include in your project.

# Example

Say we have a `static` folder containing typical assets:

```bash
$ tree static
static
├── css
│   ├── bootstrap-theme.css
│   ├── bootstrap-theme.css.map
│   ├── bootstrap-theme.min.css
│   ├── bootstrap.css
│   ├── bootstrap.css.map
│   └── bootstrap.min.css
├── fonts
│   ├── glyphicons-halflings-regular.eot
│   ├── glyphicons-halflings-regular.svg
│   ├── glyphicons-halflings-regular.ttf
│   └── glyphicons-halflings-regular.woff
├── index.go.html
└── js
    ├── bootstrap.min.js
    ├── cubism.v1.js
    ├── d3.js
    ├── jquery-2.1.0.min.js
    └── rickshaw.js

3 directories, 16 files
```

To bundle them in a distributable Go binary, you would do:

```bash
$ gostatic static
[info] Created directory for package "staticfs"
[info] 15KB ->  2.4KB   "static/css/bootstrap-theme.css"
[info] 38KB ->  11KB    "static/css/bootstrap-theme.css.map"
[info] 13KB ->  2.3KB   "static/css/bootstrap-theme.min.css"
[info] 121KB    ->  24KB    "static/css/bootstrap.css"
[info] 246KB    ->  71KB    "static/css/bootstrap.css.map"
[info] 100KB    ->  23KB    "static/css/bootstrap.min.css"
[info] 20KB ->  27KB    "static/fonts/glyphicons-halflings-regular.eot"
[info] 63KB ->  23KB    "static/fonts/glyphicons-halflings-regular.svg"
[info] 41KB ->  31KB    "static/fonts/glyphicons-halflings-regular.ttf"
[info] 23KB ->  31KB    "static/fonts/glyphicons-halflings-regular.woff"
[info] 4.0KB    ->  1.6KB   "static/index.go.html"
[info] 29KB ->  10KB    "static/js/bootstrap.min.js"
[info] 41KB ->  14KB    "static/js/cubism.v1.js"
[info] 326KB    ->  94KB    "static/js/d3.js"
[info] 84KB ->  39KB    "static/js/jquery-2.1.0.min.js"
[info] 98KB ->  31KB    "static/js/rickshaw.js"
[info] saving to "staticfs/static.go", usable with function GetStatic
```

The file `staticfs/static.go` now contains a function
`GetStatic(filename) (*bytes.Reader, bool)`, you can use it like this:

```go
// req.URL.Path == "static/{js/css/fonts}/*"
content, found := staticfs.GetStatic(req.URL.Path)
if !found {
    http.NotFound(rw, req)
    return
}
http.ServeContent(rw, req, req.URL.Path, time.Now(), content)
```

The file it generates is in a package. The file is typically __smaller__ than
your original content since the strings it stores are gzipped.

# Sample file:

The file we generated in the example above looks like this:

```go
package staticfs

import (
    "bytes"
    "compress/gzip"
    "encoding/base64"
    "io/ioutil"
    "log"
)

// GetStatic will lookup the static assets. It returns a *bytes.Reader
// and true if found, false otherwise. The static assets contain exactly the
// following entries:
//
//   static/css/bootstrap-theme.css
//   static/css/bootstrap-theme.css.map
//   static/css/bootstrap-theme.min.css
//   static/css/bootstrap.css
//   static/css/bootstrap.css.map
//   static/css/bootstrap.min.css
//   static/fonts/glyphicons-halflings-regular.eot
//   static/fonts/glyphicons-halflings-regular.svg
//   static/fonts/glyphicons-halflings-regular.ttf
//   static/fonts/glyphicons-halflings-regular.woff
//   static/index.go.html
//   static/js/bootstrap.min.js
//   static/js/cubism.v1.js
//   static/js/d3.js
//   static/js/jquery-2.1.0.min.js
//   static/js/rickshaw.js
//
func GetStatic(filename string) (*bytes.Reader, bool) {
    data, ok := decompressedStatic[filename]
    return bytes.NewReader(data), ok
}

var decompressedStatic = make(map[string][]byte)

func init() {

    var compressed = [...]struct {
        name   string
        gzip64 string
    }{
        {"static/css/bootstrap-theme.css", "H4sIAAAJbogA/+xbX ... truncated"},
        {"static/css/bootstrap-theme.css.map", "H4sIAAAJbogA/+x9C ... truncated"},
        {"static/css/bootstrap-theme.min.css", "H4sIAAAJbogA/+RaT ... truncated"},
        {"static/css/bootstrap.css", "H4sIAAAJbogA/+y9b ... truncated"},
        {"static/css/bootstrap.css.map", "H4sIAAAJbogA/+y9C ... truncated"},
        {"static/css/bootstrap.min.css", "H4sIAAAJbogA/+y9X ... truncated"},
        {"static/fonts/glyphicons-halflings-regular.eot", "H4sIAAAJbogA/4z7V ... truncated"},
        {"static/fonts/glyphicons-halflings-regular.svg", "H4sIAAAJbogA/+x9b ... truncated"},
        {"static/fonts/glyphicons-halflings-regular.ttf", "H4sIAAAJbogA/8z9C ... truncated"},
        {"static/fonts/glyphicons-halflings-regular.woff", "H4sIAAAJbogA/2R3c ... truncated"},
        {"static/index.go.html", "H4sIAAAJbogA/9xXX ... truncated"},
        {"static/js/bootstrap.min.js", "H4sIAAAJbogA/9R96 ... truncated"},
        {"static/js/cubism.v1.js", "H4sIAAAJbogA/+R9a ... truncated"},
        {"static/js/d3.js", "H4sIAAAJbogA/+S96 ... truncated"},
        {"static/js/jquery-2.1.0.min.js", "H4sIAAAJbogA/8y9e ... truncated"},
        {"static/js/rickshaw.js", "H4sIAAAJbogA/+y9/ ... truncated"},
    }

    for _, file := range compressed {
        gzipdata, err := base64.StdEncoding.DecodeString(file.gzip64)
        if err != nil {
            log.Panicf("Couldn't decode base64 data for %q: %v", file.name, err)
        }
        gr, err := gzip.NewReader(bytes.NewBuffer(gzipdata))
        if err != nil {
            log.Panicf("Couldn't open gzip stream for data for %q: %v", file.name, err)
        }
        data, err := ioutil.ReadAll(gr)
        if err != nil {
            log.Panicf("Couldn't decompress gzip data in %q: %v", file.name, err)
        }
        decompressedStatic[file.name] = data
    }
}

```
