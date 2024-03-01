package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/fasibio/portainer-api-cli/logger"
)

type PortainerApi struct {
	PortainerUrl string
	jwt          string
}

type DeployNewStackInformation struct {
	Env              []Env  `json:"Env"`
	Name             string `json:"Name"`
	StackFileContent string `json:"StackFileContent"`
	SwarmID          string `json:"SwarmID"`
}

type PortainerError struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

type StackDeployFeedback struct {
	ID          int64  `json:"Id"`
	Name        string `json:"Name"`
	Type        int64  `json:"Type"`
	EndpointID  int64  `json:"EndpointId"`
	SwarmID     string `json:"SwarmId"`
	EntryPoint  string `json:"EntryPoint"`
	Env         []Env  `json:"Env"`
	ProjectPath string `json:"ProjectPath"`
}

func (p *PortainerApi) Login(username, password string) error {

	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return err
	}
	resp, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("%s/api/auth", p.PortainerUrl), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	resp.Header.Set("Content-Type", "application/json")
	defer func() {
		if err = resp.Body.Close(); err != nil {
			logger.Get().Error(err)
		}
	}()
	decoder := json.NewDecoder(resp.Body)
	var body map[string]string
	err = decoder.Decode(&body)
	if err != nil {
		return err
	}
	p.jwt = body["jwt"]
	return nil
}

type CreateConfigBody struct {
	Name string `json:"Name,omitempty"`
	Data string `json:"Data,omitempty"`
}

func (p *PortainerApi) CreateConfig(name, content, endpoint string) (string, error) {
	client := &http.Client{}

	base64Content := base64.RawStdEncoding.EncodeToString([]byte(content))
	log.Println("dfssfd", base64Content)
	b, err := json.Marshal(&CreateConfigBody{Name: name, Data: base64Content})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("%s/api/endpoints/%s/docker/configs/create", p.PortainerUrl, endpoint), bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", p.jwt))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Get().Error(err)
		}
	}()
	resb, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return fmt.Sprintf("%d: %s", res.StatusCode, string(resb)), nil

}

func (p *PortainerApi) DeployNewApp(deployInfo DeployNewStackInformation, endpoint string) (StackDeployFeedback, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(deployInfo)
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("%s/api/stacks?endpointId=%s&method=string&type=1", p.PortainerUrl, endpoint), bytes.NewBuffer(requestBody))
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", p.jwt))
	res, err := client.Do(req)
	if err != nil {
		return StackDeployFeedback{}, err
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Get().Error(err)
		}
	}()
	decoder := json.NewDecoder(res.Body)
	if res.StatusCode != 200 {
		var errorMsg PortainerError
		err := decoder.Decode(&errorMsg)
		if err != nil {
			return StackDeployFeedback{}, err
		}
		return StackDeployFeedback{}, fmt.Errorf("Get error from server Status: %s Message: %s Details  %s", res.Status, errorMsg.Message, errorMsg.Details)
	}
	var result StackDeployFeedback
	err = decoder.Decode(&result)
	return result, err
}

func (p *PortainerApi) GetStackIDByName(name string) (int64, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("%s/api/stacks", p.PortainerUrl), nil)
	if err != nil {
		return -1, err
	}
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", p.jwt))
	res, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	if res.StatusCode != 200 {
		return -1, fmt.Errorf("Error by request Status: %s", res.Status)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Get().Error(err)
		}
	}()
	decoder := json.NewDecoder(res.Body)
	var result []StackDeployFeedback
	err = decoder.Decode(&result)
	if err != nil {
		return -1, err
	}

	var id int64 = -1

	for _, one := range result {
		if one.Name == name {
			id = one.ID
		}
	}

	if id == -1 {
		err = fmt.Errorf("Can not find stack with name %s", name)
	}
	return id, err
}

type UpdateStackInfo struct {
	Env              []Env  `json:"Env"`
	StackFileContent string `json:"StackFileContent"`
	Prune            bool   `json:"Prune"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (p *PortainerApi) UpdateStack(info UpdateStackInfo, stackid int64, endpoint string) (StackDeployFeedback, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(info)
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), "PUT", fmt.Sprintf("%s/api/stacks/%d?endpointId=%s&methode=string&type=1", p.PortainerUrl, stackid, endpoint), bytes.NewBuffer(requestBody))
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", p.jwt))
	res, err := client.Do(req)
	if err != nil {
		return StackDeployFeedback{}, err
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Get().Error(err)
		}
	}()
	decoder := json.NewDecoder(res.Body)
	if res.StatusCode != 200 {
		var errorMsg PortainerError
		err := decoder.Decode(&errorMsg)
		if err != nil {
			return StackDeployFeedback{}, err
		}
		return StackDeployFeedback{}, fmt.Errorf("Get error from server Status: %s Message: %s Details  %s", res.Status, errorMsg.Message, errorMsg.Details)
	}
	var result StackDeployFeedback
	err = decoder.Decode(&result)
	return result, err
}
