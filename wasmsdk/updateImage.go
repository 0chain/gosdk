//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type AuthRequest struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type AuthResponse struct {
	Jwt string `json:"jwt"`
}

type Container struct {
	Command    string `json:"Command"`
	Created    int64  `json:"Created"`
	HostConfig struct {
		NetworkMode string `json:"NetworkMode"`
	} `json:"HostConfig"`
	ID      string            `json:"Id"`
	Image   string            `json:"Image"`
	ImageID string            `json:"ImageID"`
	Labels  map[string]string `json:"Labels"`
	Mounts  []struct {
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
		Propagation string `json:"Propagation"`
		RW          bool   `json:"RW"`
		Source      string `json:"Source"`
		Type        string `json:"Type"`
	} `json:"Mounts"`
	Names           []string `json:"Names"`
	NetworkSettings struct {
		Networks map[string]struct {
			Aliases             interface{} `json:"Aliases"`
			DriverOpts          interface{} `json:"DriverOpts"`
			EndpointID          string      `json:"EndpointID"`
			Gateway             string      `json:"Gateway"`
			GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
			GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
			IPAMConfig          interface{} `json:"IPAMConfig"`
			IPAddress           string      `json:"IPAddress"`
			IPPrefixLen         int         `json:"IPPrefixLen"`
			IPv6Gateway         string      `json:"IPv6Gateway"`
			Links               interface{} `json:"Links"`
			MacAddress          string      `json:"MacAddress"`
			NetworkID           string      `json:"NetworkID"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
	Portainer struct {
		ResourceControl struct {
			ID                 int      `json:"Id"`
			ResourceID         string   `json:"ResourceId"`
			SubResourceIds     []string `json:"SubResourceIds"`
			Type               int      `json:"Type"`
			UserAccesses       []string `json:"UserAccesses"`
			TeamAccesses       []string `json:"TeamAccesses"`
			Public             bool     `json:"Public"`
			AdministratorsOnly bool     `json:"AdministratorsOnly"`
			System             bool     `json:"System"`
		} `json:"ResourceControl"`
	} `json:"Portainer"`
	Ports  []interface{} `json:"Ports"`
	State  string        `json:"State"`
	Status string        `json:"Status"`
}

const (
	AUTH       = "/portainer/api/auth"
	CONTAINERS = "/portainer/api/endpoints/1/docker/containers/"
	PULLIMAGE  = "/portainer/api/endpoints/1/docker/images/create?fromImage="
)

// GetContainers the containers present on the given hostmachine
func getToken(username, password, domain string) (string, error) {
	// get AuthToken
	authData := AuthRequest{
		Password: password,
		Username: username,
	}

	url := domain + AUTH
	jsonData, err := json.Marshal(authData)
	if err != nil {
		sdkLogger.Error("Error marshaling JSON:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		sdkLogger.Error("Error creating HTTP request:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return "", err
	}
	var response AuthResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		sdkLogger.Error("Error decoding JSON:", err)
		return "", err
	}

	jwt := response.Jwt
	return jwt, nil
}

func doGetRequest(authToken, url string) ([]Container, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sdkLogger.Error("Error creating HTTP request:", err)
		return nil, err
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return nil, err
	}

	var containers []Container
	err = json.Unmarshal(body, &containers)
	if err != nil {
		sdkLogger.Error("Error decoding JSON:", err)
		return nil, err
	}
	return containers, nil
}

func GetContainers(username, password, domain string) ([]Container, error) {
	authToken, err := getToken(username, password, domain)
	if err != nil {
		return nil, err
	}
	domain = strings.TrimRight(domain, "/")
	url := domain + CONTAINERS + "json?all=1"
	return doGetRequest(authToken, url)
}

func doPostRequest(url, authToken string) error {
	fmt.Println("url is", url)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		sdkLogger.Error("Error creating HTTP request:", err)
		return err
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request:", err)
		return err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request:", err)
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotModified {
		return nil
	}

	var respMsg struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(responseBody, &respMsg)
	if err != nil {
		return fmt.Errorf("got responsebody %s, with statuscode %d", string(responseBody), resp.StatusCode)
	}

	return fmt.Errorf(respMsg.Message)
}

func deleteContainer(authToken, domain, containerID string) error {
	url := domain + CONTAINERS + containerID + "?force=true"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request:", err)
		return err
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request:", err)
		return err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		sdkLogger.Error("Error reading response body:", err)
		return err
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	var respMsg struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(responseBody, &respMsg)
	if err != nil {
		return fmt.Errorf("got responsebody %s, with statuscode %d", string(responseBody), resp.StatusCode)
	}

	return fmt.Errorf(respMsg.Message)
}

func searchContainer(authToken, domain, containerID string) ([]Container, error) {
	url := domain + CONTAINERS + fmt.Sprintf("json?all=1&filters={\"id\":[\"%s\"]}", containerID)
	containers, err := doGetRequest(authToken, url)
	if err != nil {
		return nil, err
	}

	if len(containers) != 1 {
		return nil, fmt.Errorf("expected containers with ID: %s are 1. Current containers: %d", containerID, len(containers))
	}

	return containers, err
}

func pullImage(authToken, domain, NewImageID string) error {
	url := domain + PULLIMAGE + NewImageID
	return doPostRequest(url, authToken)
}

func stopContainer(authToken, domain, containerId string) error {
	url := domain + CONTAINERS + containerId + "/stop"
	return doPostRequest(url, authToken)
}

func renameContainer(authToken, domain, containerId, containerName string) error {
	url := domain + CONTAINERS + containerId + "/rename?name=" + containerName + "-old"
	return doPostRequest(url, authToken)
}

func startContainer(authToken, domain, containerId string) error {
	url := domain + CONTAINERS + containerId + "/start"
	return doPostRequest(url, authToken)
}

func createContainer(authToken, domain, containerName string, container Container) (string, error) {

	url := domain + CONTAINERS + "/create?name=" + containerName
	reqBodyJSON, err := json.Marshal(container)
	if err != nil {
		sdkLogger.Error("Error marshaling request body:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		sdkLogger.Error("Error creating HTTP request:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sdkLogger.Error("Error creating HTTP request:", err)
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		var respMsg struct {
			Id string `json:"Id"`
		}
		err = json.Unmarshal(responseBody, &respMsg)
		if err != nil {
			sdkLogger.Error("Error decoding JSON:", err)
			return "", err
		}
		return respMsg.Id, nil
	}
	var respMsg struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(responseBody, &respMsg)
	if err != nil {
		return "", fmt.Errorf("got responsebody %s, with statuscode %d", string(responseBody), resp.StatusCode)
	}
	return "", fmt.Errorf(respMsg.Message)
}

// UpdateContainer update the given container ID with a new image
func UpdateContainer(username, password, domain, containerID, NewImageID string) error {
	sdkLogger.Info("getting authtoken")
	authToken, err := getToken(username, password, domain)
	if err != nil {
		return err
	}

	// pull the new image
	sdkLogger.Info("pull the new image")
	err = pullImage(authToken, domain, NewImageID)
	if err != nil {
		return err
	}

	// stopContainer
	sdkLogger.Info("stopContainer")
	err = stopContainer(authToken, domain, containerID)
	if err != nil {
		return err
	}

	sdkLogger.Info("SearchContainer")
	containers, err := searchContainer(authToken, domain, containerID)
	if err != nil {
		return err
	}

	var containerName string
	container := containers[0]
	if len(container.Names) > 0 {
		containerName = container.Names[0]
	} else {
		return fmt.Errorf("could not find container name")
	}

	// renameContainer
	sdkLogger.Info("renameContainer")
	err = renameContainer(authToken, domain, containerID, containerName+"-old")
	if err != nil {
		return err
	}

	// createContainer
	sdkLogger.Info("CreateContainer")
	container.Image = NewImageID
	newContainerID, err := createContainer(authToken, domain, containerName, container)
	if err != nil {
		return err
	}

	// startContainer //204 no content
	sdkLogger.Info("StartContainer", newContainerID)
	err = startContainer(authToken, domain, newContainerID)
	if err != nil {
		return err
	}

	// delete old old container
	sdkLogger.Info("delete old container")
	err = deleteContainer(authToken, domain, containerID)
	if err != nil {
		return err
	}

	return nil
}
