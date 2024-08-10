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

type Endpoint struct {
	Id int `json:"Id"`
}

const (
	AUTH       = "/api/auth"
	ENDPOINTS  = "/api/endpoints"
	CONTAINERS = "/docker/containers/"
	PULLIMAGE  = "/docker/images/create?fromImage="
)

// --- exposed functions ---
// UpdateContainer update the given container ID with a new image ID in a given domain
// The domain should expose the docker API endpoints under `/endpoints/{endpointID}/docker`
// The request should be authenticated with the given username and password, by first creating an auth token then issuing the request.
//
// Parameters:
//   - username is the username to authenticate with
//   - password is the password to authenticate with
//   - domain is the domain to issue the request to
//   - containerID is the ID of the container to update
//   - NewImageID is the ID of the new image to update the container with
func UpdateContainer(username, password, domain, containerID, NewImageID string) (map[string]interface{}, error) {
	sdkLogger.Info("generating authtoken")
	authToken, err := getToken(username, password, domain)
	if err != nil {
		return nil, err
	}

	// get endpoint ID
	id, err := getEndpointId(authToken, domain)
	if err != nil {
		return nil, err
	}
	endpointID := fmt.Sprintf("/%d", id)

	sdkLogger.Info("pulling the new image...")
	url := domain + ENDPOINTS + endpointID + PULLIMAGE + NewImageID
	_, err = pullImage(authToken, domain, url)
	if err != nil {
		return nil, err
	}

	// stopContainer
	sdkLogger.Info("stopping the container..")
	url = domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/stop"
	_, err = stopContainer(authToken, domain, url)
	if err != nil {
		return nil, err
	}

	sdkLogger.Info("getting container Object")
	container, err := getContainer(authToken, domain, endpointID, containerID)
	if err != nil {
		return nil, err
	}
	containerName := container.Name
	containerReq, err := generateContainerObj(container)
	if err != nil {
		return nil, err
	}

	// renameContainer
	sdkLogger.Info("renaming container...")
	url = domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/rename?name=" + containerName + "-old"
	_, err = renameContainer(authToken, url)
	if err != nil {
		return nil, err
	}

	// createContainer
	sdkLogger.Info("creating a new container with image: ", NewImageID)
	containerReq.Image = NewImageID
	url = domain + ENDPOINTS + endpointID + CONTAINERS + "/create?name=" + containerName
	newContainer, err := createContainer(authToken, url, containerReq)
	if err != nil {
		return nil, err
	}
	newContainerID := newContainer["Id"].(string)

	// startContainer //204 no content
	sdkLogger.Info("starting starting the new container", newContainerID)
	_, err = startContainer(authToken, domain, newContainerID, endpointID)
	if err != nil {
		return nil, err
	}

	// delete old container
	sdkLogger.Info("deleting old container..")
	err = deleteContainer(authToken, domain, containerID, endpointID)
	if err != nil {
		return nil, err
	}

	return newContainer, nil
}

// GetContainers returns all the running containers in a given domain exposing the `/endpoints/{endpointID}/docker/containers/json` endpoint
// The request should be authenticated with the given username and password, by first creating an auth token then issuing the request.
// The response is a list of containers in JSON format.
//
// Parameters:
//   - username is the username to authenticate with
//   - password is the password to authenticate with
//   - domain is the domain to issue the request to
func GetContainers(username, password, domain string) ([]*map[string]interface{}, error) {
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
	body, status, err := doGetRequest(authToken, url)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return nil, err
	}
	if status == http.StatusOK {
		var containers []*map[string]interface{}
		err = json.Unmarshal(body, &containers)
		if err != nil {
			sdkLogger.Error("Error decoding JSON:", err)
			return nil, err
		}
		return containers, nil
	}
	return nil, fmt.Errorf("returned response %d. Body: %s", status, string(body))
}

// SearchContainer search a container with a given name in a given domain exposing the `/endpoints/{endpointID}/docker/containers/json` endpoint
// The request should be authenticated with the given username and password, by first creating an auth token then issuing the request.
// The response is a list of containers in JSON format that match the given name.
//
// Parameters:
//   - username is the username to authenticate with
//   - password is the password to authenticate with
//   - domain is the domain to issue the request to
//   - name is the name of the container to search for
func SearchContainer(username, password, domain, name string) ([]*map[string]interface{}, error) {
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

	// the search regex start with ^/blobber_ because the blobber config files reside in blobber folder
	// https://github.com/0chain/zcnwebappscripts/blob/main/chimney.sh#L18
	url := domain + ENDPOINTS + endpointID + CONTAINERS + fmt.Sprintf("json?all=1&filters={\"name\":[\"^/%s*\"]}", name)
	return searchContainerInternal(authToken, url)
}

// --- helper functions ----

func getEndpointId(authToken, domain string) (int, error) {
	url := domain + ENDPOINTS
	body, status, err := doGetRequest(authToken, url)
	if err != nil {
		sdkLogger.Error("Error reading response body:", err)
		return 0, err
	}

	var endpoints []Endpoint
	if status == http.StatusOK {
		err = json.Unmarshal(body, &endpoints)
		if err != nil {
			sdkLogger.Error("Error decoding endpoints:", err)
			return 0, err
		}

		if len(endpoints) > 0 {
			return endpoints[0].Id, nil
		}
	}
	return 0, fmt.Errorf("returned response %d. Body: %s", status, string(body))
}

// getContainer gets a container object by ID
func getContainer(authToken, domain, endpointID, containerID string) (*GetContainerResp, error) {
	url := domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/json"
	body, status, err := doGetRequest(authToken, url)
	if err != nil {
		return nil, err
	}

	if status == http.StatusOK {
		var container GetContainerResp
		err = json.Unmarshal(body, &container)
		if err != nil {
			sdkLogger.Error("Error decoding JSON:", err)
			return nil, err
		}
		return &container, nil
	}
	return nil, fmt.Errorf("returned response %d. Body: %s", status, string(body))
}

func searchContainerInternal(authToken, url string) ([]*map[string]interface{}, error) {
	body, status, err := doGetRequest(authToken, url)
	if err != nil {
		return nil, err
	}
	if status == http.StatusOK {
		var containers []*map[string]interface{}
		err = json.Unmarshal(body, &containers)
		if err != nil {
			sdkLogger.Error("Error decoding JSON:", err)
			return nil, err
		}
		return containers, err
	}
	return nil, fmt.Errorf("returned response %d. Body: %s", status, string(body))
}

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

	body, status, err := doHTTPRequest("POST", url, "", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	if status == http.StatusOK {
		var response AuthResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			sdkLogger.Error("Error decoding JSON:", err)
			return "", err
		}

		jwt := response.Jwt
		return jwt, nil
	}
	return "", fmt.Errorf("returned response %d. Body: %s", status, string(body))
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
	req.Header.Set("Content-Type", "application/json")

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

func doPostRequest(url, authToken string, reqBody io.Reader) (map[string]interface{}, error) {
	body, statusCode, err := doHTTPRequest("POST", url, authToken, reqBody)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK || statusCode == http.StatusNoContent || statusCode == http.StatusNotModified {
		var resp map[string]interface{}
		err = json.Unmarshal(body, &resp)
		if err != nil {
			// ignore if unmarhalling fails
			sdkLogger.Info("failed to unmarshall:", err)
			return nil, nil
		}
		return resp, nil
	}

	var respMsg struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(body, &respMsg)
	if err != nil {
		return nil, fmt.Errorf("got responsebody %s, with statuscode %d", string(body), statusCode)
	}

	return nil, fmt.Errorf(respMsg.Message)
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

func pullImage(authToken, domain, url string) (map[string]interface{}, error) {
	return doPostRequest(url, authToken, nil)
}

func stopContainer(authToken, domain, url string) (map[string]interface{}, error) {
	return doPostRequest(url, authToken, nil)
}

func renameContainer(authToken, url string) (map[string]interface{}, error) {
	return doPostRequest(url, authToken, nil)
}

func startContainer(authToken, domain, containerID, endpointID string) (map[string]interface{}, error) {
	url := domain + ENDPOINTS + endpointID + CONTAINERS + containerID + "/start"
	return doPostRequest(url, authToken, nil)
}

type NetworkSettings struct {
	Networks map[string]EndpointConfig `json:"Networks"`
}

type GetContainerResp struct {
	Config          Body            `json:"Config"`
	HostConfig      HostConfig      `json:"HostConfig"`
	Name            string          `json:"Name"`
	NetworkSettings NetworkSettings `json:"NetworkSettings"`
}

func generateContainerObj(container *GetContainerResp) (*Body, error) {
	// var config = angular.copy($scope.config);
	var config Body
	config = container.Config
	config.HostConfig = container.HostConfig
	config.NetworkingConfig.EndpointsConfig = container.NetworkSettings.Networks
	return &config, nil
}

func createContainer(authToken, url string, container *Body) (map[string]interface{}, error) {
	reqBodyJSON, err := json.Marshal(container)
	if err != nil {
		sdkLogger.Error("Error marshaling request body:", err)
		return nil, err
	}

	return doPostRequest(url, authToken, bytes.NewBuffer(reqBodyJSON))
}

type Body struct {
	Hostname         string                 `json:"Hostname"`
	Domainname       string                 `json:"Domainname"`
	User             string                 `json:"User"`
	AttachStdin      bool                   `json:"AttachStdin"`
	AttachStdout     bool                   `json:"AttachStdout"`
	AttachStderr     bool                   `json:"AttachStderr"`
	Tty              bool                   `json:"Tty"`
	OpenStdin        bool                   `json:"OpenStdin"`
	StdinOnce        bool                   `json:"StdinOnce"`
	Env              []string               `json:"Env"`
	Cmd              []string               `json:"Cmd"`
	Entrypoint       string                 `json:"Entrypoint"`
	Image            string                 `json:"Image"`
	Labels           map[string]string      `json:"Labels"`
	Volumes          map[string]interface{} `json:"Volumes"`
	WorkingDir       string                 `json:"WorkingDir"`
	NetworkDisabled  bool                   `json:"NetworkDisabled"`
	MacAddress       string                 `json:"MacAddress"`
	ExposedPorts     map[string]interface{} `json:"ExposedPorts"`
	StopSignal       string                 `json:"StopSignal"`
	StopTimeout      int                    `json:"StopTimeout"`
	HostConfig       HostConfig             `json:"HostConfig"`
	NetworkingConfig NetworkingConfig       `json:"NetworkingConfig"`
}

type HostConfig struct {
	Binds                []string                 `json:"Binds"`
	Links                []string                 `json:"Links"`
	Memory               int                      `json:"Memory"`
	MemorySwap           int                      `json:"MemorySwap"`
	MemoryReservation    int                      `json:"MemoryReservation"`
	KernelMemory         int                      `json:"KernelMemory"`
	NanoCpus             int                      `json:"NanoCpus"`
	CpuPercent           int                      `json:"CpuPercent"`
	CpuShares            int                      `json:"CpuShares"`
	CpuPeriod            int                      `json:"CpuPeriod"`
	CpuRealtimePeriod    int                      `json:"CpuRealtimePeriod"`
	CpuRealtimeRuntime   int                      `json:"CpuRealtimeRuntime"`
	CpuQuota             int                      `json:"CpuQuota"`
	CpusetCpus           string                   `json:"CpusetCpus"`
	CpusetMems           string                   `json:"CpusetMems"`
	MaximumIOps          int                      `json:"MaximumIOps"`
	MaximumIOBps         int                      `json:"MaximumIOBps"`
	BlkioWeight          int                      `json:"BlkioWeight"`
	BlkioWeightDevice    []interface{}            `json:"BlkioWeightDevice"`
	BlkioDeviceReadBps   []interface{}            `json:"BlkioDeviceReadBps"`
	BlkioDeviceReadIOps  []interface{}            `json:"BlkioDeviceReadIOps"`
	BlkioDeviceWriteBps  []interface{}            `json:"BlkioDeviceWriteBps"`
	BlkioDeviceWriteIOps []interface{}            `json:"BlkioDeviceWriteIOps"`
	MemorySwappiness     int                      `json:"MemorySwappiness"`
	OomKillDisable       bool                     `json:"OomKillDisable"`
	OomScoreAdj          int                      `json:"OomScoreAdj"`
	PidMode              string                   `json:"PidMode"`
	PidsLimit            int                      `json:"PidsLimit"`
	PortBindings         map[string][]PortBinding `json:"PortBindings"`
	PublishAllPorts      bool                     `json:"PublishAllPorts"`
	Privileged           bool                     `json:"Privileged"`
	ReadonlyRootfs       bool                     `json:"ReadonlyRootfs"`
	Dns                  []string                 `json:"Dns"`
	DnsOptions           []string                 `json:"DnsOptions"`
	DnsSearch            []string                 `json:"DnsSearch"`
	VolumesFrom          []string                 `json:"VolumesFrom"`
	CapAdd               []string                 `json:"CapAdd"`
	CapDrop              []string                 `json:"CapDrop"`
	GroupAdd             []string                 `json:"GroupAdd"`
	RestartPolicy        RestartPolicy            `json:"RestartPolicy"`
	AutoRemove           bool                     `json:"AutoRemove"`
	NetworkMode          string                   `json:"NetworkMode"`
	Devices              []interface{}            `json:"Devices"`
	Ulimits              []interface{}            `json:"Ulimits"`
	LogConfig            LogConfig                `json:"LogConfig"`
	SecurityOpt          []interface{}            `json:"SecurityOpt"`
	StorageOpt           map[string]interface{}   `json:"StorageOpt"`
	CgroupParent         string                   `json:"CgroupParent"`
	VolumeDriver         string                   `json:"VolumeDriver"`
	ShmSize              int                      `json:"ShmSize"`
}

type PortBinding struct {
	HostPort string `json:"HostPort"`
}

type RestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

type LogConfig struct {
	Type   string                 `json:"Type"`
	Config map[string]interface{} `json:"Config"`
}

type NetworkingConfig struct {
	EndpointsConfig map[string]EndpointConfig `json:"EndpointsConfig"`
}

type EndpointConfig struct {
	IPAMConfig IPAMConfig `json:"IPAMConfig"`
	Links      []string   `json:"Links"`
	Aliases    []string   `json:"Aliases"`
}

type IPAMConfig struct {
	IPv4Address  string   `json:"IPv4Address"`
	IPv6Address  string   `json:"IPv6Address"`
	LinkLocalIPs []string `json:"LinkLocalIPs"`
}
