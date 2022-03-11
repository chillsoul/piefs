package storage

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"piefs/storage/needle"
	"strconv"
)

func (s *Storage) AddVolume(w http.ResponseWriter, r *http.Request) {

}
func (s *Storage) GetNeedle(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		vid uint64
		nid uint64
		n   *needle.Needle
	)
	//request check
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseUInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}
	if nid, err = strconv.ParseUint(r.FormValue("nid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseUInt(\"%s\") error(%v)", r.FormValue("nid"), err), http.StatusBadRequest)
		return
	}
	n, err = s.directory.Get(vid, nid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get Needle of nid %d of volume vid %d error %v", nid, vid, err), http.StatusBadRequest)
		return
	}
	n.File = s.directory.GetVolumeMap()[vid].File
	w.Header().Set("Content-Type", getContentType(n.FileExt))
	//w.Header().Set("Accept-Ranges", "bytes")
	//w.Header().Set("ETag", fmt.Sprintf("%d", nid))
	w.Header().Set("Content-Length", strconv.FormatUint(n.Size, 10))
	_, err = io.CopyN(w, n, int64(n.Size))
	if err != nil {
		http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusInternalServerError)
		return
	}
}
func (s *Storage) DelNeedle(w http.ResponseWriter, r *http.Request) {

}
func (s *Storage) PutNeedle(w http.ResponseWriter, r *http.Request) {

}
func getContentType(fileExt string) string {
	contentType := "application/octet-stream"
	if fileExt != "" && fileExt != "." {
		if tmp := mime.TypeByExtension(fileExt); tmp != "" {
			contentType = tmp
		}
	}
	return contentType
}
