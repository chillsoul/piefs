package master

import (
	"context"
	"fmt"
	"github.com/chillsoul/piefs/protobuf/master_pb"
	"github.com/chillsoul/piefs/protobuf/storage_pb"
	"github.com/chillsoul/piefs/util"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

func (m *Master) GetNeedle(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var (
		ok  bool
		vid uint64
		nid uint64
	)
	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
		return
	}

	vsList := m.volumeStatusListMap[vid]
	if vsList == nil {
		http.Error(w, "volume not found", http.StatusNotFound)
		return
	}
	randInt := rand.Intn(len(m.volumeStatusListMap[vid]))
	http.Redirect(w, r, fmt.Sprintf("http://%s/GetNeedle?vid=%d&nid=%d", vsList[randInt].Url, vid, nid), http.StatusFound)
}

func (m *Master) PutNeedle(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var (
		err error
		nid uint64
	)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20+1<<19) //1.5MB
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "r.FromFile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	vsList, err := m.getWritableVolumes(uint64(header.Size))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nid = m.snowflake.NextVal()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wg := sync.WaitGroup{}
	var uploadErr []string
	for _, vStatus := range vsList {
		wg.Add(1)
		go func(vs *master_pb.VolumeStatus) {
			defer wg.Done()
			//给该vid对应的所有volume上传文件
			conn := m.getSingletonConnection(vs.Url)
			//conn, err := grpc.Dial(vs.Url, grpc.WithTransportCredentials(insecure.NewCredentials()))
			//defer conn.Close()
			client := storage_pb.NewStorageClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_, err = client.WriteNeedleBlob(ctx, &storage_pb.WriteNeedleBlobRequest{
				VolumeId:   vs.Id,
				NeedleId:   nid,
				NeedleData: data,
				FileExt:    path.Ext(header.Filename),
			})
			if err != nil {
				uploadErr = append(uploadErr, fmt.Sprintf("storage: %s  error: %s", vs.Url, err))
			}

		}(vStatus)
	}
	wg.Wait()
	if len(uploadErr) != 0 {
		http.Error(w, strings.Join(uploadErr, "\n"), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	result := []byte(fmt.Sprintf("{\"vid\":%d,\"nid\":%d}", vsList[0].Id, nid))
	w.Write(result)
}
func (m *Master) DelNeedle(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var (
		ok  bool
		vid uint64
		nid uint64
	)
	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
		return
	}
	vsList := m.volumeStatusListMap[vid]
	if vsList == nil {
		http.Error(w, "volume not found", http.StatusNotFound)
		return
	}
	wg := sync.WaitGroup{}
	var deleteErr []string
	for _, vStatus := range vsList {
		wg.Add(1)
		go func(vs *master_pb.VolumeStatus) {
			defer wg.Done()
			//给该vid对应的所有volume删除文件
			conn := m.getSingletonConnection(vs.Url)
			client := storage_pb.NewStorageClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := client.DeleteNeedleBlob(ctx, &storage_pb.DeleteNeedleBlobRequest{
				VolumeId: vs.Id,
				NeedleId: nid,
			})
			if err != nil {
				deleteErr = append(deleteErr, fmt.Sprintf("storage: %s  error: %s", vs.Url, err))
			}

		}(vStatus)
	}
	wg.Wait()
	if len(deleteErr) != 0 {
		http.Error(w, strings.Join(deleteErr, "\n"), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("success"))
}
