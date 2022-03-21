package master

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"math/rand"
	"net/http"
	"piefs/protobuf/master_pb"
	"piefs/protobuf/storage_pb"
	"piefs/util"
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
	rand.Seed(time.Now().UnixNano())
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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20+1<<19) //1.5MB
	nid = util.UniqueId()
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
			conn, err := grpc.Dial(vs.Url, grpc.WithTransportCredentials(insecure.NewCredentials()))
			defer conn.Close()
			client := storage_pb.NewStorageClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err = client.WriteNeedleBlob(ctx, &storage_pb.WriteNeedleBlobRequest{
				VolumeId:   vs.Id,
				NeedleId:   nid,
				NeedleData: data,
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
	result := []byte(fmt.Sprintf("{'vid':%d,\n'nid':%d}", vsList[0].Id, nid))
	w.Write(result)
	//space check
}
