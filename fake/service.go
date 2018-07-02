package fake

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ServiceService service

type Service struct {
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Path           string `json:"path"`
	Retries        int    `json:"retries"`
	ConnectTimeout int    `json:"connect_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	ReadTimeout    int    `json:"read_timeout"`
	URL            string `json:"url,omitempty"`
	ID             string `json:"id,omitempty"`
	CreationDate   int    `json:"created_at,omitempty"`
	UpdateDate     int    `json:"updated_at,omitempty"`
}

//Post Add new Kong Service /services
func (ss *ServiceService) Post(s *Service) (*http.Response, error) {

	obj, _ := json.Marshal(ss.ExpectedOuput)

	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(string(obj))),
	}, nil
}

//Delete Delete Kong Service based on ID
func (ss *ServiceService) Delete(serviceID string) (*http.Response, error) {
	return nil, nil
}

//Patch Update Kong Service based on ID
func (ss *ServiceService) Patch(serviceID string, s *Service) (*http.Response, error) {
	return nil, nil
}
