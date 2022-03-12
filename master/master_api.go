package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"piefs/storage"
)

func (m *Master) Monitor(w http.ResponseWriter, r *http.Request) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("password") != m.password {
		http.Error(w, "permission denied", http.StatusUnauthorized)
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
	fmt.Printf("receive %s:%d heartbeat\n", status.ApiHost, status.ApiPort)
	flag := false
	for _, v := range m.storageList {
		if v.ApiHost == status.ApiHost && v.ApiPort == status.ApiPort {
			*v = *status
			flag = true
		}
	}
	if !flag {
		m.storageList = append(m.storageList, status)
	}
}
