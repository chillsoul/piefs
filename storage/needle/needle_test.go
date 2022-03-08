package needle

import (
	"fmt"
	"os"
	"testing"
)

var testfile, err = os.OpenFile("../../resources/testfile/gofactory.jpg", os.O_RDONLY, 0777)

func TestMarshal(t *testing.T) {
	fmt.Println(err)
}
func TestUnmarshal(t *testing.T) {

}
