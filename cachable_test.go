package polygonio

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type TestIoer struct {
	atomicWrite func(dir string, filename string, write func(w io.Writer) error) error
	read        func(filepath string) (io.ReadCloser, error)
}

func (ioer *TestIoer) AtomicWrite(dir string, filename string, write func(w io.Writer) error) error {
	return ioer.atomicWrite(dir, filename, write)
}

func (ioer *TestIoer) Read(filepath string) (io.ReadCloser, error) {
	return ioer.read(filepath)
}

var googleResponse = `HTTP/1.1 200 OK
Connection: close
Cache-Control: private, max-age=0
Content-Type: text/html; charset=ISO-8859-1
Date: Sat, 25 Apr 2020 23:27:09 GMT
Expires: -1
P3p: CP="This is not a P3P policy! See g.co/p3phelp for more info."
Server: gws
X-Frame-Options: SAMEORIGIN
X-Xss-Protection: 0

<!doctype html>`

var googleBody = `<!doctype html>`

func TestFileCacher_Save(t *testing.T) {

	testIoEr := &TestIoer{}

	fc := FileCacher{
		Dir:          "Test",
		FileCacherIo: testIoEr,
	}

	r, _ := http.NewRequest("GET", "http://www.google.com", nil)
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(googleResponse)), r)
	if err != nil {
		panic(err)
	}

	atomicWriteCalled := false
	testIoEr.atomicWrite = func(dir string, filename string, write func(w io.Writer) error) error {
		atomicWriteCalled = true
		return nil
	}

	if err := fc.Save(r, resp); err != nil {
		panic(err)
	}

	if !atomicWriteCalled {
		t.Fatal("expected atomicWrite to be called")
	}
}

func TestFileCacher_Get(t *testing.T) {

	testIoEr := &TestIoer{}

	fc := FileCacher{
		Dir:          "Test",
		FileCacherIo: testIoEr,
	}

	r, _ := http.NewRequest("GET", "http://www.google.com", nil)
	readCalled := false
	testIoEr.read = func(filepath string) (io.ReadCloser, error) {
		readCalled = true
		return ioutil.NopCloser(strings.NewReader(googleResponse)), nil
	}

	out, err := fc.Get(r)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Body.Close()

	if !readCalled {
		t.Error("readCalled expected")
	}

	outBytes, err := ioutil.ReadAll(out.Body)
	if err != nil {
		t.Fatal(err)
	}

	stringBody := string(outBytes)
	if stringBody != googleBody {
		t.Fatal("outBytes != googleResponse")
	}

}
