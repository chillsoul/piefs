package volume

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type Status struct {
	ApiHost     string
	ApiPort     int
	ID          uint64
	CurrentSize uint64
	//Writable bool
}

func (s *Status) IsWritable() bool {
	if s.CurrentSize < MaxVolumeSize {
		return true
	} else {
		return false
	}
}
func (s *Status) HasEnoughSpace(size uint64) bool {
	if s.CurrentSize+size <= MaxVolumeSize {
		return true
	} else {
		return false
	}
}
func (s *Status) UploadFile(nid uint64, data *[]byte, fileName string, password string) error {
	writerBuf := &bytes.Buffer{} //will auto allocate slice when write()
	mPart := multipart.NewWriter(writerBuf)
	filePart, err := mPart.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	_, err = filePart.Write(*data)
	if err != nil {
		return err
	}
	mPart.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/PutNeedle?vid=%d&nid=%d", s.ApiHost,
		s.ApiPort, s.ID, nid), writerBuf)
	if err != nil {
		return err
	}
	req.Header.Set("password", password)
	req.Header.Set("Content-Type", mPart.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	return nil
}
