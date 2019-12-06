package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	resp, err := http.Post(fmt.Sprintf("%s/api/auth", p.PortainerUrl), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var body map[string]string
	err = decoder.Decode(&body)
	if err != nil {
		return err
	}
	p.jwt = body["jwt"]
	return nil
}

func (p *PortainerApi) DeployNewApp(deployInfo DeployNewStackInformation, endpoint string) (StackDeployFeedback, error) {
	client := &http.Client{}
	requestBody, err := json.Marshal(deployInfo)
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/stacks?endpointId=%s&method=string&type=1", p.PortainerUrl, endpoint), bytes.NewBuffer(requestBody))
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", p.jwt))
	res, err := client.Do(req)
	defer res.Body.Close()
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/stacks", p.PortainerUrl), nil)
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
	defer res.Body.Close()
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
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/stacks/%d?endpointId=%s&methode=string&type=1", p.PortainerUrl, stackid, endpoint), bytes.NewBuffer(requestBody))
	if err != nil {
		return StackDeployFeedback{}, err
	}
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", p.jwt))
	res, err := client.Do(req)
	defer res.Body.Close()
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
