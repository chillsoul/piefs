package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"piefs/storage"
	"piefs/storage/volume"
	"piefs/util"
	"time"
)

func (m *Master) GetNeedle(w http.ResponseWriter, r *http.Request) {
	var (
		ok   bool
		vid  uint64
		nid  uint64
		host string = "not found"
		port int    = 65535
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
	host = m.volumeStatusListMap[vid][randInt].ApiHost
	port = m.volumeStatusListMap[vid][randInt].ApiPort
	http.Redirect(w, r, fmt.Sprintf("http://%s:%d/GetNeedle?vid=%d&nid=%d", host, port, vid, nid), http.StatusFound)
}
func (m *Master) Monitor(w http.ResponseWriter, r *http.Request) {
	if !util.IsMethodAllowed(w, r, "POST") {
		return
	}
	if !m.isAuthPassed(w, r) {
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	status := &storage.Status{}
	if err = json.Unmarshal(body, status); err != nil {
		fmt.Println(err)
		return
	}
	m.statusLock.Lock()
	defer m.statusLock.Unlock()
	flag := false
	for i := 0; i < len(m.storageStatusList); i++ {
		//update storage status
		if m.storageStatusList[i].ApiHost == status.ApiHost && m.storageStatusList[i].ApiPort == status.ApiPort {
			m.storageStatusList[i] = status
			flag = true
		}
	}
	if !flag { //first heartbeat
		m.storageStatusList = append(m.storageStatusList, status)
	}
	for _, vs := range status.VolumeStatusList {
		flag = false
		vsList := m.volumeStatusListMap[vs.ID]
		if vsList == nil { //new volume
			m.volumeStatusListMap[vs.ID] = []*volume.Status{vs}
			continue
		}
		for i, vs_ := range vsList {
			if vs_.ApiHost == vs.ApiHost && vs_.ApiPort == vs.ApiPort {
				m.volumeStatusListMap[vs.ID][i] = vs //update volume status
				flag = true
			}
		}
		if !flag { //the storage of an existed volume first appear
			m.volumeStatusListMap[vs.ID] = append(m.volumeStatusListMap[vs.ID], vs)
		}
	}
	fmt.Printf("receive %s:%d heartbeat\n", status.ApiHost, status.ApiPort)

}
func (m *Master) isAuthPassed(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("password") != m.password {
		http.Error(w, "permission denied", http.StatusUnauthorized)
		return false
	}
	return true
}
