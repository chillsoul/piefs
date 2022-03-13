package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"piefs/storage/needle"
	"piefs/storage/volume"
	"piefs/util"
	"strconv"
)

func (s *Storage) AddVolume(w http.ResponseWriter, r *http.Request) {
	var (
		ok  bool
		err error
		vid uint64
		v   *volume.Volume
	)
	if !util.IsMethodAllowed(w, r, "POST") {
		return
	}
	if ok, vid = util.GetVidFromFormValue(w, r); !ok {
		return
	}
	if v, err = volume.NewVolume(vid, s.storeDir); err != nil {
		http.Error(w, fmt.Sprintf("create new volume for vid %s in dir %s error(%v)", r.FormValue("vid"), s.storeDir, err),
			http.StatusInternalServerError)
		return
	}
	s.directory.GetVolumeMap()[vid] = v
	w.WriteHeader(http.StatusCreated)
	return
}
func (s *Storage) GetNeedle(w http.ResponseWriter, r *http.Request) {
	var (
		ok  bool
		err error
		vid uint64
		nid uint64
		n   *needle.Needle
	)
	//request check
	if !util.IsMethodAllowed(w, r, "GET") {
		return
	}
	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
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
		ok  bool
		err error
		vid uint64
		nid uint64
	)
	if !s.isAuthPassed(w, r) {
		return
	}
	if !util.IsMethodAllowed(w, r, "POST") {
		return
	}
	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
		return
	}
	err = s.directory.Del(vid, nid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusInternalServerError)
	}
}
func (s *Storage) PutNeedle(w http.ResponseWriter, r *http.Request) {
	var (
		ok  bool
		err error
		vid uint64
		nid uint64
	)
	if !s.isAuthPassed(w, r) {
		return
	}
	if !util.IsMethodAllowed(w, r, "POST") {
		return
	}
	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
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

func (s *Storage) isAuthPassed(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("password") != s.password {
		http.Error(w, "permission denied", http.StatusUnauthorized)
		return false
	}
	return true
}
