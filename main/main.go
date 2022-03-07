package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func img(w http.ResponseWriter, r *http.Request) {
	file, err := os.OpenFile("./resources/testfile/gofactory.jpg", os.O_RDONLY, 0777)
	defer file.Close()
	if err != nil {
		fmt.Println("error open file")
	}
	buff, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("error read file")
	}
	w.Write(buff)
}
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/trueimg", http.StatusFound)
}
func main() {
	http.HandleFunc("/fakeimg", redirect)
	http.HandleFunc("/trueimg", img)
	err := http.ListenAndServe(":80", nil)
	fmt.Println(err)
}
