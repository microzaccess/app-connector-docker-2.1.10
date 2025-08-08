package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	// "gvisor.dev/gvisor/pkg/log"
)

var baseURL = "http://51.79.142.144:31290/api/"

// var TenantID = "12"
var prevTime = make(map[string]string)

type ConfigValues struct {
	DeviceID  string `json:"id"`
	DeviceKey string `json:"key"`
}

type APIResponse struct {
	Userdata map[string]struct {
		Timestamp string   `json:"timestamp"`
		Urls      []string `json:"urls"`
	} `json:"user_data"`
	IpMapping []struct {
		IP       string `json:"ipAddress"`
		TenantID string `json:"tenant_id"`
		Username string `json:"username"`
	} `json:"ip_mapping"`
	TenantID string `json:"tenant_id"`
}

var ctx = context.Background()

type PdnsRedisManager struct {
	rdb *redis.Client
}

func NewPdnsRedisManager(addr string, db int) *PdnsRedisManager {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	return &PdnsRedisManager{rdb}
}

type UserDomainsRequest struct {
	TenantID               string   `json:"tenant_id"`
	UserName               string   `json:"user_name"`
	Blacklist_urls         []string `json:"blacklist_urls"`
	Blacklist_applications []string `json:"blacklist_applications"`
	Blacklist_categories   []string `json:"blacklist_categories"`
	Whitelist_urls         []string `json:"whitelist_urls"`
	Whitelist_applications []string `json:"whitelist_applications"`
	Whitelist_categories   []string `json:"whitelist_categories"`
	RestAll                string   `json:"rest_all"`
}

func main() {
	fmt.Println("Starting app-connector...")
	baseURL = GetURL() + "/tenant/config/"
	fmt.Println("Base URL=====:>", baseURL)
	ticker := time.NewTicker(30 * time.Second)
	p := NewPdnsRedisManager("localhost:6379", 0)
	Configdata, err := LoadAndUpdateConfig()
	if err != nil {
		fmt.Printf("Error in LoadAndUpdateConfig: %v", err)

	}
	// p.HandleUserData(Configdata)
	p.HandleAllTenantData()

	fmt.Println("Config data loaded:", Configdata)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				Configdata, err = LoadAndUpdateConfig()
				if err != nil {
					fmt.Printf("Error in LoadAndUpdateConfig: %v", err)

				}
				// p.HandleUserData(Configdata)
				p.HandleAllTenantData()

			}
		}
	}()
	select {} // Keep the main goroutine running

}

func GetURL() string {

	ConfigPath := filepath.Join("/home", "mza_connecter_id.json")
	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0755); err != nil {
		fmt.Println("Error creating config directory:", err)
		return "https://cosgridnetworks.in/api/v1"
	}
	fmt.Println("==============================>", ConfigPath)
	// log.Info(ConfigPath)

	// Read the JSON file
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return "https://cosgridnetworks.in/api/v1"
	}
	type BaseURL struct {
		BaseURL string `json:"base_url"`
	}
	var baseURL BaseURL
	if err := json.Unmarshal(data, &baseURL); err != nil {
		fmt.Println("Error unmarshalling config data:", err)
		return "https://cosgridnetworks.in/api/v1"
	}
	fmt.Println("Base URL:", baseURL.BaseURL)
	if baseURL.BaseURL == "" {
		fmt.Println("Base URL is empty, using default")
		return "https://cosgridnetworks.in/api/v1"
	}
	return baseURL.BaseURL
}

func LoadAndUpdateConfig() (ConfigValues, error) {
	// Ensure the config directory exists
	var ConfigDir string
	_, err := os.UserHomeDir()
	if err == nil {
		// /home/.config/cosgrid/config.json
		ConfigDir = "/home/.config/cosgrid"
		fmt.Println("INFO - Using default config directory:", ConfigDir)
	}

	// Prepare config path
	ConfigPath := filepath.Join(ConfigDir, "config.json")
	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0755); err != nil {
		return ConfigValues{}, err
	}
	// fmt.Println(ConfigPath)
	// log.Info(ConfigPath)

	// Read the JSON file
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return ConfigValues{}, err
	}
	// fmt.Println("Data read from config file:", string(data))
	var config ConfigValues
	if err := json.Unmarshal(data, &config); err != nil {
		return ConfigValues{}, err
	}

	return config, nil
}

func (p *PdnsRedisManager) HandleUserData(Configdata ConfigValues) {
	// fmt.Println("handle-userdata is called")
	baseURL := baseURL + "webfilter-endpoint/app-connector-data/"
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	prevTimeJson, err := json.Marshal(prevTime)
	if err != nil {
		fmt.Println("Failed Marshal map-previous-time ", prevTimeJson)
	}

	prevTimeBody := []byte(fmt.Sprintf(`{"user_timestamps":%s, "device_id":"%s"}`, prevTimeJson, Configdata.DeviceID))
	fmt.Println(string(prevTimeBody))
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(prevTimeBody))
	if err != nil {
		fmt.Printf("Failed to create Post request: %v", err)

		return
	}
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("req", req)

	postResp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to create Post request: %v", err)
		return
	}
	defer postResp.Body.Close()
	userDataResponse, err := io.ReadAll(postResp.Body)
	if err != nil {
		fmt.Printf("Failed to read the response body previous-time response: %v", err)
	}

	fmt.Println("Current api Response state: ", string(userDataResponse))

	var currApiResponse APIResponse
	err = json.Unmarshal(userDataResponse, &currApiResponse)
	if err != nil {
		fmt.Println("cannot unmarshal previous time response", err)
	}
	updatePreviousTimeMapAndData(currApiResponse)

	p.SetUserDomains(currApiResponse, currApiResponse.TenantID)
	p.SetIPMapping(currApiResponse)
	// fmt.Println(currApiResponse)
}

var AllTenantpreviousTime = make(map[string]map[string]string)
var AllTenant = []string{"12", "13"}

func (p *PdnsRedisManager) HandleAllTenantData() {
	for _, TenantID := range AllTenant {
		AllTenantpreviousTime[TenantID] = make(map[string]string)
	}
	// fmt.Println("handle-userdata is called")
	baseURL := baseURL + "webfilter-endpoint/saas-connector-data/"
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	prevTimeJson, err := json.Marshal(AllTenantpreviousTime)
	if err != nil {
		fmt.Println("Failed Marshal map-previous-time ", prevTimeJson)
	}

	prevTimeBody := []byte(fmt.Sprintf(`{"user_timestamps":%s}`, prevTimeJson))
	fmt.Println("===========>>>>>>>>>>>>>", string(prevTimeBody))
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(prevTimeBody))
	if err != nil {
		fmt.Printf("Failed to create Post request: %v", err)

		return
	}
	req.Header.Set("Content-Type", "application/json")
	// fmt.Println("req", req)

	postResp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to create Post request: %v", err)
		return
	}
	defer postResp.Body.Close()
	userDataResponse, err := io.ReadAll(postResp.Body)
	if err != nil {
		fmt.Printf("Failed to read the response body previous-time response: %v", err)
	}

	fmt.Println("Current api Response state: ", string(userDataResponse))

	var currApiResponse []APIResponse
	err = json.Unmarshal(userDataResponse, &currApiResponse)
	if err != nil {
		fmt.Println("cannot unmarshal previous time response", err)
	}
	fmt.Println("Current saas Response:", currApiResponse)
	for _, apiResponse := range currApiResponse {

		updatePreviousTimeMapAndData(apiResponse)
		p.SetUserDomains(apiResponse, apiResponse.TenantID)
		p.SetIPMapping(apiResponse)
		fmt.Println("TenantID:", apiResponse.TenantID, "Userdata:", apiResponse.Userdata)
	}

}

func updatePreviousTimeMapAndData(currApiResponse APIResponse) {
	for k, items := range currApiResponse.Userdata {
		prevTime[k] = items.Timestamp
		// fmt.Println(ApplicationsData[k], "application Data")
	}
}

func (p *PdnsRedisManager) SetUserDomains(currentApiResponse APIResponse, TenantID string) {

	for user, data := range currentApiResponse.Userdata {
		// fmt.Println("User:", user, "Data:", data)
		key := fmt.Sprintf("tenant:%s:user:%s", TenantID, user)
		urls := data.Urls
		// fmt.Println("urls:", urls)
		value, _ := json.Marshal(map[string]interface{}{
			"whitelist_urls": urls})

		err := p.rdb.Set(ctx, key, value, 0).Err()
		if err != nil {
			fmt.Printf("Failed to set user domains for %s: %v", user, err)
			continue
		}
		fmt.Printf("User domains set for %s, key: %s , value : %s", user, key, value)
	}
}

func (p *PdnsRedisManager) SetIPMapping(currentApiResponse APIResponse) {
	for _, req := range currentApiResponse.IpMapping {

		// fmt.Println("IP Mapping:", req.IP, "TenantID:", req.TenantID, "Username:", req.Username)
		key := fmt.Sprintf("ip:%s", req.IP)
		value, _ := json.Marshal(map[string]interface{}{
			"tenant_id": req.TenantID,
			"user_id":   req.Username,
			"group":     "",
		})

		err := p.rdb.Set(ctx, key, value, 0).Err()
		if err != nil {
			fmt.Printf("Failed to set IP mapping for %s: %v", req.IP, err)
			return
		}
		fmt.Printf("IP mapping set for %s, key: %s, value: %s", req.IP, key, value)
	}
}
