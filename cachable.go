package polygonio

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DoCache(client *http.Client, r *http.Request, cacheable bool, cacher Cacher) (*http.Response, error) {
	if r.Method == "GET" && cacheable && cacher != nil {
		resp, err := cacher.Get(r)
		if resp != nil && err == nil {
			return resp, err
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	if r.Method == "GET" && resp.StatusCode == 200 && cacheable && cacher != nil {
		if err := cacher.Save(r, resp); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

type FileCacher struct {
	Dir          string
	FileCacherIo FileCacherIo
}

type FileCacherIo interface {
	AtomicWrite(dir string, filename string, write func(w io.Writer) error) error
	Read(filepath string) (io.ReadCloser, error)
}

func (fc FileCacher) FilePath(request *http.Request) (dir string, fn string) {
	q := request.URL.Query()
	q.Set("apiKey", "X")
	path := []string{fc.Dir, request.URL.Scheme, request.URL.Host}
	path = append(path, strings.Split(request.URL.Path, "/")...)
	return filepath.Join(path...), q.Encode() + ".json"
}

func (fc FileCacher) Save(request *http.Request, response *http.Response) error {

	ogResponse := response.Body
	defer ogResponse.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(body)
	response.Body = ioutil.NopCloser(reader)
	defer func() {
		reader.Reset(body)
	}()

	dir, fn := fc.FilePath(request)

	return fc.FileCacherIo.AtomicWrite(dir, fn, func(w io.Writer) error {
		return response.Write(w)
	})
}

func (fc FileCacher) Get(request *http.Request) (*http.Response, error) {

	dir, fn := fc.FilePath(request)

	f, err := fc.FileCacherIo.Read(filepath.Join(dir, fn))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fb, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(bytes.NewReader(fb)), request)
}

type OsFileCacherIo struct{}

func (OsFileCacherIo) AtomicWrite(dir string, filename string, write func(w io.Writer) error) error {

	//create file destination directories
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}

	f, err := ioutil.TempFile("", "atomic-")
	if err != nil {
		panic(err)
	}
	defer func() {
		// Clean up (best effort) in case we are returning with an error:
		// Prevent file descriptor leaks.
		f.Close()
		// Remove the tempfile to avoid filling up the file system.
		os.Remove(f.Name())
	}()

	if err := write(f); err != nil {
		panic(err)
	}

	f.Chmod(0644)
	f.Sync()

	if err := f.Close(); err != nil {
		panic(err)
	}

	//rename temp file to destination file
	if err := os.Rename(f.Name(), filepath.Join(dir, filename)); err != nil {
		panic(err)
	}

	return nil

}

func (OsFileCacherIo) Read(filepath string) (io.ReadCloser, error) {
	return os.Open(filepath)
}
