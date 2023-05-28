//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
			ID                 int           `json:"Id"`
			ResourceID         string        `json:"ResourceId"`
			SubResourceIds     []string      `json:"SubResourceIds"`
			Type               int           `json:"Type"`
			UserAccesses       []interface{} `json:"UserAccesses"`
			TeamAccesses       []interface{} `json:"TeamAccesses"`
			Public             bool          `json:"Public"`
			AdministratorsOnly bool          `json:"AdministratorsOnly"`
			System             bool          `json:"System"`
		} `json:"ResourceControl"`
	} `json:"Portainer"`
	Ports  []interface{} `json:"Ports"`
	State  string        `json:"State"`
	Status string        `json:"Status"`
}

type Endpoint struct {
	Id int `json:"Id"`
}

const (
	AUTH       = "/portainer/api/auth"
	ENDPOINTS  = "/portainer/api/endpoints"
	CONTAINERS = "/docker/containers/"
	PULLIMAGE  = "/docker/images/create?fromImage="
)

func getEndpointId(authToken, domain string) (int, error) {
	url := domain + ENDPOINTS
	body, _, err := doGetRequest(authToken, url)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return 0, err
	}

	var endpoints []Endpoint
	err = json.Unmarshal(body, &endpoints)
	if err != nil {
		sdkLogger.Error("Error decoding endpoints:", err)
		return 0, err
	}

	if len(endpoints) > 0 {
		return endpoints[0].Id, nil
	}

	return 0, fmt.Errorf("empty endpoints returned")
}

// --- exposed functions ---
// UpdateContainer update the given container ID with a new image
func UpdateContainer(username, password, domain, containerID, NewImageID string) error {
	sdkLogger.Info("getting authtoken")
	authToken, err := getToken(username, password, domain)
	if err != nil {
		return err
	}

	// get endpoint ID
	id, err := getEndpointId(authToken, domain)
	if err != nil {
		return err
	}
	endpointID := fmt.Sprintf("/%d", id)

	sdkLogger.Info("pull the new image")
	url := domain + ENDPOINTS + endpointID + PULLIMAGE + NewImageID
	err = pullImage(authToken, domain, url)
	if err != nil {
		return err
	}

	// stopContainer
	sdkLogger.Info("stopContainer")
	url = domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/stop"
	err = stopContainer(authToken, domain, url)
	if err != nil {
		return err
	}

	sdkLogger.Info("SearchContainer")
	url = domain + ENDPOINTS + endpointID + CONTAINERS + fmt.Sprintf("json?all=1&filters={\"id\":[\"%s\"]}", containerID)
	containers, err := searchContainerInternal(authToken, url)
	if err != nil {
		return err
	}

	var containerName string
	var container *Container
	if len(containers) > 0 {
		container = containers[0]
		if len(container.Names) > 0 {
			containerName = container.Names[0]
		} else {
			return fmt.Errorf("could not find container name")
		}
	}

	// renameContainer
	sdkLogger.Info("renameContainer")
	url = domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/rename?name=" + containerName + "-old"
	err = renameContainer(authToken, url)
	if err != nil {
		return err
	}

	// createContainer
	sdkLogger.Info("CreateContainer")
	container.Image = NewImageID
	url = domain + ENDPOINTS + endpointID + CONTAINERS + "/create?name=" + containerName
	newContainerID, err := createContainer(authToken, url, container)
	if err != nil {
		return err
	}

	// startContainer //204 no content
	sdkLogger.Info("StartContainer", newContainerID)
	err = startContainer(authToken, domain, newContainerID, endpointID)
	if err != nil {
		return err
	}

	// delete old old container
	sdkLogger.Info("delete old container")
	err = deleteContainer(authToken, domain, containerID, endpointID)
	if err != nil {
		return err
	}

	return nil
}

func GetContainers(username, password, domain string) ([]*Container, error) {
	authToken, err := getToken(username, password, domain)
	if err != nil {
		return nil, err
	}

	id, err := getEndpointId(authToken, domain)
	if err != nil {
		return nil, err
	}

	endpointID := fmt.Sprintf("/%d", id)
	domain = strings.TrimRight(domain, "/")
	url := domain + ENDPOINTS + endpointID + CONTAINERS + "json?all=1"
	body, _, err := doGetRequest(authToken, url)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return nil, err
	}
	var containers []*Container
	err = json.Unmarshal(body, &containers)
	if err != nil {
		sdkLogger.Error("Error decoding JSON:", err)
		return nil, err
	}
	return containers, nil
}

// SearchContainer search a container with a given name
func SearchContainer(username, password, domain, name string) ([]*Container, error) {
	authToken, err := getToken(username, password, domain)
	if err != nil {
		return nil, err
	}

	id, err := getEndpointId(authToken, domain)
	if err != nil {
		return nil, err
	}
	endpointID := fmt.Sprintf("/%d", id)
	domain = strings.TrimRight(domain, "/")

	url := domain + ENDPOINTS + endpointID + CONTAINERS + fmt.Sprintf("json?all=1&filters={\"name\":[\"^/%s*\"]}", name)
	return searchContainerInternal(authToken, url)
}

func searchContainerInternal(authToken, url string) ([]*Container, error) {
	body, _, err := doGetRequest(authToken, url)
	if err != nil {
		return nil, err
	}
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}
	var containers []*Container
	err = json.Unmarshal(body, &containers)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil, err
	}
	return containers, err
}

// --- helper functions ----

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

	body, _, err := doHTTPRequest("POST", url, "", bytes.NewBuffer(jsonData))
	if err != nil {
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

func doGetRequest(authToken, url string) ([]byte, int, error) {
	return doHTTPRequest("GET", url, authToken, nil)
}

func doHTTPRequest(method, url, authToken string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		sdkLogger.Error("Error creating HTTP request:", err)
		return nil, 0, err
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sdkLogger.Error("Error sending HTTP request:", err)
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func doPostRequest(url, authToken string) error {

	body, statusCode, err := doHTTPRequest("POST", url, authToken, nil)
	if err != nil {
		return err
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusNotModified {
		return nil
	}

	var respMsg struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(body, &respMsg)
	if err != nil {
		return fmt.Errorf("got responsebody %s, with statuscode %d", string(body), statusCode)
	}

	return fmt.Errorf(respMsg.Message)
}

func deleteContainer(authToken, domain, containerID, endpointID string) error {
	url := domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "?force=true"

	body, status, err := doHTTPRequest("DELETE", url, authToken, nil)
	if err != nil {
		return err
	}
	if status == http.StatusOK || status == http.StatusNoContent {
		return nil
	}
	var respMsg struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(body, &respMsg)
	if err != nil {
		return fmt.Errorf("got responsebody %s, with statuscode %d", string(body), status)
	}

	return fmt.Errorf(respMsg.Message)
}

func pullImage(authToken, domain, url string) error {
	return doPostRequest(url, authToken)
}

func stopContainer(authToken, domain, url string) error {
	return doPostRequest(url, authToken)
}

func renameContainer(authToken, url string) error {
	return doPostRequest(url, authToken)
}

func startContainer(authToken, domain, containerID, endpointID string) error {
	url := domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/start"
	return doPostRequest(url, authToken)
}

func createContainer(authToken, url string, container *Container) (string, error) {

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
