package storage

import (
	"fmt"
	"io"
	"net/http"
	"piefs/storage/needle"
	"piefs/util"
)

func (s *Storage) GetNeedle(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var (
		ok  bool
		err error
		vid uint64
		nid uint64
	)
	//request check
	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
		return
	}
	metadata, err := s.cache.GetNeedleMetadata(vid, nid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get Cache Needle of nid %d of volume vid %d error %v", nid, vid, err), http.StatusBadRequest)
		return
	}
	n, err := needle.Unmarshal(metadata)
	n.File = s.directory.GetVolumeMap()[vid].File
	//w.Header().Set("Content-Type", "image/*")
	//w.Header().Set("Accept-Ranges", "bytes")
	//w.Header().Set("ETag", fmt.Sprintf("%d", nid))
	//w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	_, err = io.CopyN(w, n, int64(n.Size))
	//w.Write(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read Needle data error %v", err), http.StatusInternalServerError)
		return
	}
}

//func getContentType(fileExt string) string {
//	contentType := "application/octet-stream"
//	if fileExt != "" && fileExt != "." {
//		if tmp := mime.TypeByExtension(fileExt); tmp != "" {
//			contentType = tmp
//		}
//	}
//	return contentType
//}
