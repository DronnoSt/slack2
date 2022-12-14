package fsadapter

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var _ FS = &ZIP{}

type ZIP struct {
	zw   *zip.Writer
	mu   sync.Mutex
	f    *os.File
	seen map[string]bool // seen holds the list of seen directories.
}

func (z *ZIP) String() string {
	return fmt.Sprintf("<zip archive: %s>", z.f.Name())
}

func NewZIP(zw *zip.Writer) *ZIP {
	return &ZIP{zw: zw, seen: make(map[string]bool)}
}

func NewZipFile(filename string) (*ZIP, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	zw := zip.NewWriter(f)
	return &ZIP{zw: zw, f: f, seen: make(map[string]bool)}, nil
}

func (*ZIP) normalizePath(p string) string {
	split := strings.Split(filepath.Clean(p), string(os.PathSeparator))
	return path.Join(split...)
}

func (z *ZIP) Create(filename string) (io.WriteCloser, error) {
	// reassemble path in correct format for ZIP file
	// in case it uses OS specific path.
	filename = z.normalizePath(filename)

	z.mu.Lock() // mutex will be unlocked, when the user calls Close.
	w, err := z.create(filename)
	if err != nil {
		return nil, err
	}
	return &syncWriter{w: w, mu: &z.mu}, nil
}

func (z *ZIP) create(filename string) (io.Writer, error) {
	if err := z.ensureDir(filename); err != nil {
		return nil, err
	}
	header := &zip.FileHeader{
		Name:     filename,
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	return z.zw.CreateHeader(header)
}

func (z *ZIP) ensureDir(filename string) error {
	if z.seen == nil {
		z.seen = make(map[string]bool, 0)
	}
	var ensureFn = func(dir string) error {
		if _, seen := z.seen[dir]; seen {
			return nil
		}
		// not seen, create an empty directory.
		if _, err := z.zw.Create(dir); err != nil {
			return err
		}
		z.seen[dir] = true
		return nil
	}
	dir, _ := path.Split(filename)
	for _, d := range z.dirpath(dir) {
		if err := ensureFn(d); err != nil {
			return err
		}
	}
	return nil
}

func (*ZIP) dirpath(dir string) []string {
	const sep = "/"
	if len(dir) == 0 {
		return nil
	}
	var ret []string
	d := strings.TrimRight(dir, sep)
	for len(d) > 0 {
		ret = append([]string{strings.TrimRight(d, sep) + sep}, ret...)
		d, _ = path.Split(strings.TrimRight(d, sep))
	}
	return ret
}

func (z *ZIP) WriteFile(filename string, data []byte, _ os.FileMode) error {
	z.mu.Lock()
	defer z.mu.Unlock()
	zf, err := z.create(filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(zf, bytes.NewReader(data))
	return err

}

// Close closes the underlying zip writer and the file handle.  It is only necessary if
// ZIP was initialised using NewZipFile
func (z *ZIP) Close() error {
	if !z.ourHandles() {
		// we don't own the handles, so just bail out.
		return nil
	}
	z.mu.Lock()
	defer z.mu.Unlock()

	return z.closeHandles()
}

func (z *ZIP) closeHandles() error {
	if err := z.zw.Close(); err != nil {
		return err
	}
	if z.f == nil {
		return nil
	}
	return z.f.Close()
}

func (z *ZIP) ourHandles() bool {
	return z.f != nil
}

type syncWriter struct {
	w io.Writer // underlying writer

	// zip writer can only process one file at a time, so any process that wants
	// to Create the file will have to wait until Close is called:
	//
	// From zip.Create doc:  The file's contents must be written to the
	// io.Writer before the next call to Create, CreateHeader, or Close.
	mu *sync.Mutex
}

func (sw *syncWriter) Write(p []byte) (int, error) {
	return sw.w.Write(p)
}

func (sw *syncWriter) Close() error {
	sw.mu.Unlock()
	return nil
}
