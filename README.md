# gostatic

Command `gostatic` takes a list of directories and create a Go file
containing all of those directories that you can include in your project.

```bash
$ gostatic -pkgname helloworld static/
[info] Created directory for package "helloworld"
[info] 15KB ->  3.4KB   "static/css/bootstrap-theme.css"
[info] 38KB ->  16KB    "static/css/bootstrap-theme.css.map"
[info] 13KB ->  3.3KB   "static/css/bootstrap-theme.min.css"
# ....
[info] saving to "helloworld/static.go", usable with functions GetStatic and ListStatic
```

# Features

* Gives out handy `bytes.Reader`.
* Compresses data with gzip.
* Decompresses on init.

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
[info] 15KB ->  3.4KB   "static/css/bootstrap-theme.css"
[info] 38KB ->  16KB    "static/css/bootstrap-theme.css.map"
[info] 13KB ->  3.3KB   "static/css/bootstrap-theme.min.css"
[info] 121KB    ->  34KB    "static/css/bootstrap.css"
[info] 246KB    ->  100KB   "static/css/bootstrap.css.map"
[info] 100KB    ->  32KB    "static/css/bootstrap.min.css"
[info] 6.0KB    ->  2.7KB   "static/css/rickshaw.min.css"
[info] 20KB ->  38KB    "static/fonts/glyphicons-halflings-regular.eot"
[info] 63KB ->  32KB    "static/fonts/glyphicons-halflings-regular.svg"
[info] 41KB ->  44KB    "static/fonts/glyphicons-halflings-regular.ttf"
[info] 23KB ->  44KB    "static/fonts/glyphicons-halflings-regular.woff"
[info] 4.0KB    ->  2.2KB   "static/index.go.html"
[info] 29KB ->  14KB    "static/js/bootstrap.min.js"
[info] 41KB ->  19KB    "static/js/cubism.v1.js"
[info] 326KB    ->  133KB   "static/js/d3.js"
[info] 84KB ->  55KB    "static/js/jquery-2.1.0.min.js"
[info] 98KB ->  43KB    "static/js/rickshaw.js"
[info] saving to "staticfs/static.go", usable with functions GetStatic and ListStatic
```

By default, it saves to package `staticfs`. You can specify your own package name:

```bash
$ gostatic -pkgname helloworld static/
[info] Created directory for package "helloworld"
# ...
[info] saving to "helloworld/static.go", usable with functions GetStatic and ListStatic
```

The file `staticfs/static.go` now contains two functions:

* `ListStatic() map[string]*bytes.Reader`: return a map of all the assets, keyed by name.
* `GetStatic(filename) (*bytes.Reader, bool)`, fetch an asset by name.

For example, you can use `GetStatic`:

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
//   static/css/rickshaw.min.css
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

// ListStatic will return all the static assets sharing root
// Static.
func ListStatic() (map[string]*bytes.Reader) {
    out := make(map[string]*bytes.Reader, len(decompressedStatic))
    for k, v := range decompressedStatic {
        out[k] = bytes.NewReader(v)
    }
    return out
}

var decompressedStatic = make(map[string][]byte)

func init() {

    var compressed = [...]struct {
        name   string
        gzip256 string
    }{
        {"static/css/bootstrap-theme.css", `H4sIAAAJbogA/+xbXW+bTBa+ .. truncated `},
        {"static/css/bootstrap-theme.css.map", `H4sIAAAJbogA/+x9C3Pb .. truncated `},
        {"static/css/bootstrap-theme.min.css", `H4sIAAAJbogA/+RaTW/b .. truncated `},
        {"static/css/bootstrap.css", `H4sIAAAJbogA/+y9bZMjuY0w+L1/hd .. truncated `},
        {"static/css/bootstrap.css.map", `H4sIAAAJbogA/+y9C3/iOLI4+l .. truncated `},
        {"static/css/bootstrap.min.css", `H4sIAAAJbogA/+y9XZPjuLEo+H .. truncated `},
        {"static/css/rickshaw.min.css", `H4sIAAAJbogA/5xY/W6rNhR/FdS .. truncated `},
        {"static/fonts/glyphicons-halflings-regular.eot", `H4sIAAAJb .. truncated `},
        {"static/fonts/glyphicons-halflings-regular.svg", `H4sIAAAJb .. truncated `},
        {"static/fonts/glyphicons-halflings-regular.ttf", `H4sIAAAJb .. truncated `},
        {"static/fonts/glyphicons-halflings-regular.woff", `H4sIAAAJ .. truncated `},
        {"static/index.go.html", `H4sIAAAJbogA/9xXUW/bNhB+969gvQJxMV .. truncated `},
        {"static/js/bootstrap.min.js", `H4sIAAAJbogA/9R963PbRpL49/0r .. truncated `},
        {"static/js/cubism.v1.js", `H4sIAAAJbogA/+R9a3fbRrLgd/0KmLOx .. truncated `},
        {"static/js/d3.js", `H4sIAAAJbogA/+S963ZbN5Iw+vvkKbb5dXo2LYo .. truncated `},
        {"static/js/jquery-2.1.0.min.js", `H4sIAAAJbogA/8y9eXsbx9Ev+ .. truncated `},
        {"static/js/rickshaw.js", `H4sIAAAJbogA/+y9/X8TOZIw/nP8V4jcP .. truncated `},
    }

    for _, file := range compressed {
        gzipdata, err := base64.StdEncoding.DecodeString(file.gzip256)
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
