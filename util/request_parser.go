package util

import (
	"fmt"
	"net/http"
	"strconv"
)

func IsMethodAllowed(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
func GetVidNidFromFormValue(w http.ResponseWriter, r *http.Request) (ok bool, vid, nid uint64) {
	var err error
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return false, 0, 0
	}
	if nid, err = strconv.ParseUint(r.FormValue("nid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusNotFound)
		return false, 0, 0
	}
	return true, vid, nid
}
