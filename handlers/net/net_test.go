package time

import (
	"io/ioutil"
	"os"
	"testing"

	"arp242.net/sconfig"
)

// testfile will write data to a temporary file and will return the full file
// path. It is the caller's responsibility to clean the file.
func testfile(data string) (filename string) {
	fp, err := ioutil.TempFile(os.TempDir(), "sconfigtest")
	if err != nil {
		panic(err)
	}
	defer func() { _ = fp.Close() }()

	_, err = fp.WriteString(data)
	if err != nil {
		panic(err)
	}
	return fp.Name()
}

type config struct {
}

func TestTime(t *testing.T) {
	test := `
	`

	c := config{}
	f := testfile(test)
	defer os.Remove(f)
	err := sconfig.Parse(&c, f, nil)
	if err != nil {
		t.Error(err)
	}
}
