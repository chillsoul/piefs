package master

//func (m *Master) GetNeedle(w http.ResponseWriter, r *http.Request) {
//	var (
//		ok  bool
//		vid uint64
//		nid uint64
//	)
//	if ok, vid, nid = util.GetVidNidFromFormValue(w, r); !ok {
//		return
//	}
//	rand.Seed(time.Now().UnixNano())
//	vsList := m.volumeStatusListMap[vid]
//	if vsList == nil {
//		http.Error(w, "volume not found", http.StatusNotFound)
//		return
//	}
//	randInt := rand.Intn(len(m.volumeStatusListMap[vid]))
//	http.Redirect(w, r, fmt.Sprintf("http://%s/GetNeedle?vid=%d&nid=%d", vsList[randInt].Url, vid, nid), http.StatusFound)
//}

//func (m *Master) HandOutNeedle(w http.ResponseWriter, r *http.Request) {
//	var (
//		err error
//		nid uint64
//	)
//	if !m.isAuthPassed(w, r) {
//		return
//	}
//	if !util.IsMethodAllowed(w, r, "POST") {
//		return
//	}
//
//	file, header, err := r.FormFile("file")
//	if err != nil {
//		http.Error(w, "r.FromFile: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	defer file.Close()
//
//	vsList, err := m.getWritableVolumes(uint64(header.Size))
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	r.Body = http.MaxBytesReader(w, r.Body, 1<<20+1<<19) //1.5MB
//	nid = util.UniqueId()
//	data, err := ioutil.ReadAll(file)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	wg := sync.WaitGroup{}
//	var uploadErr []string
//	for _, vStatus := range vsList {
//		wg.Add(1)
//		go func(vs *volume.Status) {
//			defer wg.Done()
//			//给该vid对应的所有volume上传文件
//			err = vs.UploadFile(nid, &data, header.Filename, m.password)
//			if err != nil {
//				uploadErr = append(uploadErr, fmt.Sprintf("host: %s port: %d error: %s", vs.ApiHost, vs.ApiPort, err))
//			}
//		}(vStatus)
//	}
//	wg.Wait()
//	if len(uploadErr) != 0 {
//		http.Error(w, strings.Join(uploadErr, "\n"), http.StatusInternalServerError)
//		return
//	}
//	w.WriteHeader(http.StatusCreated)
//	result := []byte(fmt.Sprintf("{'vid':%d,\n'nid':%d}", vsList[0].ID, nid))
//	w.Write(result)
//	//space check
//}
//
//// Monitor
//// Deprecated
//func (m *Master) Monitor(w http.ResponseWriter, r *http.Request) {
//	if !util.IsMethodAllowed(w, r, "POST") {
//		return
//	}
//	if !m.isAuthPassed(w, r) {
//		return
//	}
//	body, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	status := &storage.Status{}
//	if err = json.Unmarshal(body, status); err != nil {
//		fmt.Println(err)
//		return
//	}
//	m.statusLock.Lock()
//	defer m.statusLock.Unlock()
//	flag := false
//	for i := 0; i < len(m.storageStatusList); i++ {
//		//update storage status
//		if m.storageStatusList[i].ApiHost == status.ApiHost && m.storageStatusList[i].ApiPort == status.ApiPort {
//			m.storageStatusList[i] = status
//			flag = true
//		}
//	}
//	if !flag { //first heartbeat
//		m.storageStatusList = append(m.storageStatusList, status)
//	}
//	for _, vs := range status.VolumeStatusList {
//		flag = false
//		vsList := m.volumeStatusListMap[vs.ID]
//		if vsList == nil { //new volume
//			m.volumeStatusListMap[vs.ID] = []*volume.Status{vs}
//			continue
//		}
//		for i, vs_ := range vsList {
//			if vs_.ApiHost == vs.ApiHost && vs_.ApiPort == vs.ApiPort {
//				m.volumeStatusListMap[vs.ID][i] = vs //update volume status
//				flag = true
//			}
//		}
//		if !flag { //the storage of an existed volume first appear
//			m.volumeStatusListMap[vs.ID] = append(m.volumeStatusListMap[vs.ID], vs)
//		}
//	}
//	fmt.Printf("receive %s:%d heartbeat\n", status.ApiHost, status.ApiPort)
//
//}
//func (m *Master) isAuthPassed(w http.ResponseWriter, r *http.Request) bool {
//	if r.Header.Get("password") != m.password {
//		http.Error(w, "permission denied", http.StatusUnauthorized)
//		return false
//	}
//	return true
//}
