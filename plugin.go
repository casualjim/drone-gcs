package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"mime"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

const maxUploadFilePerTime = 100
const maxRetryUploadTimes = 5

type (
	Repo struct {
		Owner string
		Name  string
	}

	Build struct {
		Event  string
		Number int
		Commit string
		Branch string
		Author string
		Status string
		Link   string
		Tag    string
	}

	Metadata struct {
		data map[string]string
	}

	Config struct {
		AuthKey      string
		Source       string
		Target       string
		Ignore       string
		Acl          []string
		Gzip         []string
		CacheControl string
		Metadata     string
	}

	Plugin struct {
		Repo   Repo
		Build  Build
		Config Config
	}
)

var (
	fatalf        = log.Fatalf
	printf        = log.Printf
	upload_config Config
	bucket        *storage.BucketHandle
	sleep         = time.Sleep
	ecodeMu       sync.Mutex
	ecode         int
)

// errorf sets exit code to a non-zero value and outputs using printf.
func errorf(format string, args ...interface{}) {
	ecodeMu.Lock()
	ecode = 1
	ecodeMu.Unlock()
	printf(format, args...)
}

func (p Plugin) Exec() error {
	upload_config = p.Config
	//Temporay, read gcs auth key
	file, e := ioutil.ReadFile(p.Config.AuthKey)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	auth, err := google.JWTConfigFromJSON([]byte(file), storage.ScopeFullControl)
	if err != nil {
		fatalf("auth: %v", err)
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx, cloud.WithTokenSource(auth.TokenSource(ctx)))
	if err != nil {
		fatalf("storage client: %v", err)
	}
	return p.run(client)
}

// walkFiles creates a complete set of files to upload
// by walking source recursively.
//
// It excludes files matching ignore_pattern.
// The ignore_pattern is matched using filepath.Match against a partial
// file name, relative to source.
func walkFiles() ([]string, error) {
	var items []string
	err := filepath.Walk(upload_config.Source, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		rel, err := filepath.Rel(upload_config.Source, p)
		if err != nil {
			return err
		}
		var ignore bool
		if upload_config.Ignore != "" {
			ignore, err = filepath.Match(upload_config.Ignore, rel)
		}
		if err != nil || ignore {
			return err
		}
		items = append(items, p)
		return nil
	})
	return items, err
}

// run is the actual entry point called from main.
func (p Plugin) run(client *storage.Client) error {
	// extract bucket name from the target path
	paths := strings.SplitN(upload_config.Target, "/", 2)
	bname := paths[0]
	if len(paths) == 1 {
		upload_config.Target = ""
	} else {
		upload_config.Target = paths[1]
	}
	bucket = client.Bucket(strings.Trim(bname, "/"))

	// create a list of files to upload
	src, err := walkFiles()
	if err != nil {
		fatalf("local files: %v", err)
	}

	// result contains upload result of a single file
	type result struct {
		name string
		err  error
	}

	// upload all files in a goroutine, maxUploadFilePerTime at a time
	buf := make(chan struct{}, maxUploadFilePerTime)
	res := make(chan *result, len(src))
	for _, f := range src {
		fmt.Println(f)
		buf <- struct{}{} // alloc one slot
		go func(f string) {
			rel, err := filepath.Rel(upload_config.Source, f)
			if err != nil {
				res <- &result{f, err}
				return
			}
			fmt.Println("begin uploading")
			err = retryUpload(path.Join(upload_config.Target, rel), f, maxRetryUploadTimes)
			if err != nil {
				fmt.Println(err)
			}
			res <- &result{rel, err}
			<-buf // free up
		}(f)
	}

	// wait for all files to be uploaded or stop at first error
	for _ = range src {
		r := <-res
		if r.err != nil {
			fatalf("%s: %v", r.name, r.err)
			return r.err
		}
		printf(r.name)
	}

	return nil
}

// retryUpload calls uploadFile until the latter returns nil
// or the number of invocations reaches n.
// It blocks for a duration of seconds exponential to the iteration between the calls.
func retryUpload(dst, file string, n int) error {
	var err error
	for i := 0; i <= n; i++ {
		if i > 0 {
			t := time.Duration((math.Pow(2, float64(i)) + rand.Float64()) * float64(time.Second))
			sleep(t)
		}
		if err = uploadFile(dst, file); err == nil {
			fmt.Println(err)
			fmt.Println("retry:", i)
			break
		}
	}
	return err
}

// uploadFile uploads the file to dst using global bucket.
// To get a more robust upload use retryUpload instead.
func uploadFile(dst, file string) error {
	r, gz, err := gzipper(file)
	if err != nil {
		return err
	}
	defer r.Close()
	rel, err := filepath.Rel(upload_config.Source, file)
	if err != nil {
		return err
	}
	name := path.Join(upload_config.Target, rel)
	w := bucket.Object(name).NewWriter(context.Background())
	w.CacheControl = upload_config.CacheControl
	var metadata interface{}
	err = json.Unmarshal([]byte(upload_config.Metadata), &metadata)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(metadata)
	w.Metadata = make(map[string]string)
	for k, v := range metadata.(map[string]interface{}) {
		switch vtype := v.(type) {
		case string:
			fmt.Println(vtype)
			w.Metadata[k] = v.(string)
		}
	}
	fmt.Println(w.Metadata)
	for _, s := range upload_config.Acl {
		a := strings.SplitN(s, ":", 2)
		if len(a) != 2 {
			return fmt.Errorf("%s: invalid ACL %q", name, s)
		}
		w.ACL = append(w.ACL, storage.ACLRule{
			Entity: storage.ACLEntity(a[0]),
			Role:   storage.ACLRole(a[1]),
		})
	}
	w.ContentType = mime.TypeByExtension(filepath.Ext(file))
	if w.ContentType == "" {
		w.ContentType = "application/octet-stream"
	}
	if gz {
		w.ContentEncoding = "gzip"
	}
	fmt.Println("TEST")
	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return w.Close()
}

// gzipper returns a stream of file and a boolean indicating
// whether the stream is gzip-compressed.
//
// The stream is compressed if upload_config.Gzip contains file extension.
func gzipper(file string) (io.ReadCloser, bool, error) {
	r, err := os.Open(file)
	if err != nil || !matchGzip(file) {
		return r, false, err
	}
	pr, pw := io.Pipe()
	w := gzip.NewWriter(pw)
	go func() {
		_, err := io.Copy(w, r)
		if err != nil {
			errorf("%s: io.Copy: %v", file, err)
		}
		if err := w.Close(); err != nil {
			errorf("%s: gzip: %v", file, err)
		}
		if err := pw.Close(); err != nil {
			errorf("%s: pipe: %v", file, err)
		}
		r.Close()
	}()
	return pr, true, nil
}

// matchGzip reports whether the file should be gzip-compressed during upload.
// Compressed files should be uploaded with "gzip" content-encoding.
func matchGzip(file string) bool {
	ext := filepath.Ext(file)
	if ext == "" {
		return false
	}
	ext = ext[1:]
	i := sort.SearchStrings(upload_config.Gzip, ext)
	return i < len(upload_config.Gzip) && upload_config.Gzip[i] == ext
}
