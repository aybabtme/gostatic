package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"github.com/aybabtme/color/brush"
	"github.com/dustin/go-humanize"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"
	"text/template"
	"unicode"
)

var elog = log.New(newLogtab(os.Stderr), brush.Red("[error] ").String(), log.Lshortfile)

func main() {

	log.SetOutput(newLogtab(os.Stdout))
	log.SetPrefix(brush.Blue("[info] ").String())
	log.SetFlags(0)

	if len(os.Args) < 1 {
		log.Fatal("Need to specify a directory")
		return
	}
	for _, arg := range os.Args[1:] {

		err := writeDirectory(arg)
		if err != nil {
			log.Printf("Failed to snapshot %q, %v", arg, err)
		}

	}
}

func writeDirectory(dirname string) error {

	compressSize := 0
	totalSize := 0
	fakefs := make(map[string]string)

	err := filepath.Walk(dirname, func(name string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return err
		}

		data, err := ioutil.ReadFile(name)
		if err != nil {
			log.Printf("couldn't read %q: %v", name, err)
			return err
		}

		totalSize += len(data)
		buf := bytes.NewBuffer(nil)
		gw := gzip.NewWriter(buf)

		if _, err = gw.Write(data); err != nil {
			log.Printf("couldn't compress %q: %v", name, err)
		}
		if gw.Close(); err != nil {
			log.Printf("couldn't close compressed %q: %v", name, err)
		}
		compressSize += buf.Len()

		gzip64data := base64.StdEncoding.EncodeToString(buf.Bytes())

		fakefs[name] = gzip64data

		log.Printf("%s\t->\t%s\t%q",
			humanize.Bytes(uint64(len(data))),
			humanize.Bytes(uint64(len(gzip64data))),
			name)

		return nil
	})
	if err != nil {
		return err
	}

	destfilename := snakify(dirname) + ".go"
	destfunction := camelize(dirname)

	log.Printf("saving to %q, usable with function Get%s", destfilename, destfunction)

	file, err := os.Create(destfilename)
	if err != nil {
		return err
	}

	err = filetempl.Execute(file, struct {
		RootMap  map[string]string
		RootName string
	}{
		RootMap:  fakefs,
		RootName: destfunction,
	})
	if err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

type logtabwriter struct {
	tab *tabwriter.Writer
}

func newLogtab(w io.Writer) io.Writer {
	return &logtabwriter{tabwriter.NewWriter(w, 2, 2, 1, '\t', 0)}
}

// compile check
var _ io.Writer = &logtabwriter{}

func (l *logtabwriter) Write(p []byte) (int, error) {
	n, err := l.tab.Write(p)
	if err != nil {
		return n, err
	}
	return n, l.tab.Flush()
}

func snakify(input string) string {
	out := bytes.NewBuffer(nil)
	lastWasSnake := true

	for _, r := range []rune(input) {

		switch {
		case unicode.IsLetter(r):
			out.WriteRune(r)
			lastWasSnake = false
		case lastWasSnake:
			// skip it
		default:
			out.WriteRune('_')
		}
	}
	return out.String()
}

func camelize(input string) string {
	out := bytes.NewBuffer(nil)
	needCamel := true
	for _, r := range []rune(input) {

		switch {
		case unicode.IsLetter(r) && needCamel:
			out.WriteRune(unicode.ToUpper(r))
			needCamel = false
		case unicode.IsLetter(r) && !needCamel:
			out.WriteRune(r)
		default:
			needCamel = true
		}
	}
	return out.String()
}

var filetempl = template.Must(template.New("file").Parse(`package staticfs

import (
    "bufio"
    "bytes"
    "compress/gzip"
    "encoding/base64"
    "io"
    "io/ioutil"
    "log"
)

// Get{{.RootName}} will lookup the static assets. It returns an io.ReadSeeker
// and true if found, false otherwise. The static assets contain exactly the
// following entries:
// {{range $name, $data := .RootMap}}
//   {{$name}}{{end}}
//
func Get{{.RootName}}(filename string) (io.ReadSeeker, bool) {
    data, ok := decompressed{{.RootName}}[filename]
    return bufio.NewReader(data), ok
}

var decompressed{{.RootName}} = make(map[string][]byte)

func init() {

    var compressed = [...]struct {
        name   string
        gzip64 string
    }{ {{range $name, $data := .RootMap}}
        {"{{$name}}", "{{$data}}"},{{end}}
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
        decompressed{{.RootName}}[file.name] = data
    }
}
`))
