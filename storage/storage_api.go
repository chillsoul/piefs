package storage

import (
	"fmt"
	"io"
	"io/ioutil"
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
	var (
		err error
		vid uint64
		nid uint64
	)
	if !s.isMaster(r) {
		http.Error(w, fmt.Sprintf("permission denied"), http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}

	if nid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusNotFound)
		return
	}
	err = s.directory.Del(vid, nid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusInternalServerError)
	}
}
func (s *Storage) PutNeedle(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		vid uint64
		nid uint64
	)
	if !s.isMaster(r) {
		http.Error(w, fmt.Sprintf("permission denied"), http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if vid, err = strconv.ParseUint(r.FormValue("vid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusBadRequest)
		return
	}
	if nid, err = strconv.ParseUint(r.FormValue("fid"), 10, 64); err != nil {
		http.Error(w, fmt.Sprintf("strconv.ParseInt(\"%s\") error(%v)", r.FormValue("vid"), err), http.StatusNotFound)
		return
	}
	v := s.directory.GetVolumeMap()[vid]
	if v == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}
	file, header, err := r.FormFile("file")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20+1<<19) //1.5MB
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := v.NewFile(nid, data, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.directory.Set(vid, nid, n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
func (s *Storage) isMaster(r *http.Request) bool {
	if r.RemoteAddr == fmt.Sprintf("%s:%d", s.masterHost, s.masterPort) {
		return true
	} else {
		return false
	}
}
