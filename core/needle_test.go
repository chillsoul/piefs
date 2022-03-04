package core

import (
	"fmt"
	"os"
	"testing"
)

var testfile, err = os.OpenFile("../resources/testfile/gofactory.jpg", os.O_RDONLY, 0777)

func TestNeedleMarshal(t *testing.T) {
	fmt.Println(err)
}
func TestNeedleUnmarshal(t *testing.T) {

}
