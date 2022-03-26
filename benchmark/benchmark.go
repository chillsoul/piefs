package benchmark

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"piefs/util"
	"sync"
	"time"
)

type result struct {
	concurrent  int
	num         int
	startTime   time.Time
	endTime     time.Time
	completed   int32
	failed      int32
	transferred uint64
}
type uploadResponse struct {
	Vid uint64
	Nid uint64
}

func Benchmark(masterHost string, masterPort int, concurrent int, num int, size int) {
	uploadResult := &result{
		concurrent: concurrent,
		num:        num,
		startTime:  time.Now(),
	}

	loop := make(chan int)
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	needleList := make([]*uploadResponse, concurrent*num)
	randBytes := make([]byte, size)
	rand.Read(randBytes)

	dataMd5 := md5.Sum(randBytes)
	util.PathMustExists("./_tmp_")
	testFile, _ := ioutil.TempFile("./_tmp_", "")
	testFile.Truncate(int64(size))
	io.Copy(testFile, bytes.NewReader(randBytes))
	defer os.Remove(testFile.Name())
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(j int) {
			var err error
			for b := range loop {
				needleList[b], err = Upload(masterHost, masterPort, testFile.Name(), bytes.NewReader(randBytes))
				mutex.Lock()
				if err == nil {
					uploadResult.completed += 1
				} else {
					uploadResult.failed += 1
					fmt.Println("write failed:", err.Error())
				}
				mutex.Unlock()
			}
			wg.Done()
		}(i)
	}

	for i := 0; i < num; i++ {
		loop <- i
	}
	close(loop)

	wg.Wait()
	uploadResult.endTime = time.Now()
	timeTaken := float64(uploadResult.endTime.UnixNano()-uploadResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("upload %d %dbyte file:\n\n", uploadResult.num, size)
	fmt.Printf("concurrent:             %d\n", uploadResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", uploadResult.completed)
	fmt.Printf("failed:                 %d\n", uploadResult.failed)
	fmt.Printf("transferred:            %d byte\n", uploadResult.completed*int32(size))
	fmt.Printf("request per second:     %.2f\n", float64(uploadResult.num)/timeTaken)
	fmt.Printf("transferred per second: %.2f MB \n", float64(uploadResult.completed)*float64(size)/timeTaken/1024/1024)

	readResult := &result{
		concurrent: concurrent,
		num:        num,
		startTime:  time.Now(),
	}
	loop = make(chan int)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(j int) {
			for b := range loop {
				data, err := Get(masterHost, masterPort, needleList[b].Vid, needleList[b].Nid)
				mutex.Lock()
				if err == nil && md5.Sum(data) == dataMd5 {
					readResult.completed += 1
				} else {
					readResult.failed += 1
					fmt.Println("read failed:", err.Error())
				}
				mutex.Unlock()
			}
			wg.Done()
		}(i)
	}

	for i := 0; i < num; i++ {
		loop <- i
	}
	close(loop)
	wg.Wait()

	readResult.endTime = time.Now()
	timeTaken = float64(readResult.endTime.UnixNano()-readResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("\n\nread %d %dbyte file:\n\n", readResult.num, size)
	fmt.Printf("concurrent:             %d\n", readResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", readResult.completed)
	fmt.Printf("failed:                 %d\n", readResult.failed)
	fmt.Printf("transferred:            %d byte\n", readResult.completed*int32(size))
	fmt.Printf("request per second:     %.2f\n", float64(readResult.num)/timeTaken)
	fmt.Printf("transferred per second: %.2f MB \n", float64(readResult.completed)*float64(size)/timeTaken/1024/1024)

	deleteResult := &result{
		concurrent: concurrent,
		num:        num,
		startTime:  time.Now(),
	}
	loop = make(chan int)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(j int) {
			for b := range loop {
				err := Delete(masterHost, masterPort, needleList[b].Vid, needleList[b].Nid)
				mutex.Lock()
				if err == nil {
					deleteResult.completed += 1
				} else {
					deleteResult.failed += 1
					fmt.Println("delete failed:", err.Error())
				}
				mutex.Unlock()
			}
			wg.Done()
		}(i)
	}

	for i := 0; i < num; i++ {
		loop <- i
	}
	close(loop)
	wg.Wait()

	deleteResult.endTime = time.Now()
	timeTaken = float64(deleteResult.endTime.UnixNano()-deleteResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("\n\ndelete %d %dbyte file:\n\n", deleteResult.num, size)
	fmt.Printf("concurrent:             %d\n", deleteResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", deleteResult.completed)
	fmt.Printf("failed:                 %d\n", deleteResult.failed)
	fmt.Printf("transferred:            %d byte\n", deleteResult.completed*int32(size))
	fmt.Printf("request per second:     %.2f\n", float64(deleteResult.num)/timeTaken)
	fmt.Printf("transferred per second: %.2f MB\n", float64(deleteResult.completed)*float64(size)/timeTaken/1024/1024)

}
func Upload(host string, port int, srcFilePath string, reader io.Reader) (*uploadResponse, error) {

	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	filePart, err := mPart.CreateFormFile("file", filepath.Base(srcFilePath))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(filePart, reader)
	if err != nil {
		return nil, err
	}

	mPart.Close()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/PutNeedle",
		host, port), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	buf, err := ioutil.ReadAll(resp.Body)
	result := &uploadResponse{}
	err = json.Unmarshal(buf, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func Get(host string, port int, vid, nid uint64) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/GetNeedle?vid=%v&nid=%v", host,
		port, vid, nid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAll(resp.Body)
	} else {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}

}
func Delete(host string, port int, vid, nid uint64) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/DelNeedle?vid=%v&nid=%v",
		host, port, vid, nid), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}

}
