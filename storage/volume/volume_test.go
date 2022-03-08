package volume

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestNewVolume(t *testing.T) {
	var testfile, err = os.OpenFile("../../resources/testfile/gofactory.jpg", os.O_RDONLY, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer testfile.Close()
	volume, err := NewVolume(202103070001, "202103070001")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		volume.File.Close()
		os.RemoveAll(volume.Path)
	}()
	data, err := ioutil.ReadAll(testfile)
	if err != nil {
		t.Fatal(err)
	}
	needle, err := volume.NewFile(202103070002, data, ".jpg")
	var ndata []byte
	needle.ReadData(ndata)
	if reflect.DeepEqual(ndata, data) {
		t.Fatal("error needle data not equal")
	}
}
