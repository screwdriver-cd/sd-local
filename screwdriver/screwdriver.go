package screwdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// API has method to get job
type API interface {
	Job(jobName, filePath string) (Job, error)
}

// SDAPI has methods for control Screwdriver.cd APIs
type SDAPI struct {
	HTTPClient *http.Client
	UserToken  string
	APIURL     string
	SDJWT      string
}

// Step is step entity struct
type Step struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Job is job entity struct
type Job struct {
	Steps       []Step            `json:"commands"`
	Environment map[string]string `json:"environment"`
	Image       string            `json:"image"`
}

// Jobs is Job entity map
type Jobs map[string][]Job

type validatorResponse struct {
	Jobs   Jobs     `json:"jobs"`
	Errors []string `json:"errors"`
}

type tokenResponse struct {
	JWT string `json:"token"`
}

// New creates a SDAPI
func New(apiURL, token string) (SDAPI, error) {
	s := SDAPI{
		HTTPClient: http.DefaultClient,
		APIURL:     apiURL,
		UserToken:  token,
	}

	jwt, err := s.jwt()
	if err != nil {
		return s, err
	}

	s.SDJWT = jwt

	return s, nil
}

func (sd *SDAPI) makeURL(path string) (*url.URL, error) {
	version := "v4"
	fullpath := fmt.Sprintf("%s/%s/%s", sd.APIURL, version, path)

	return url.Parse(fullpath)
}

func (sd *SDAPI) request(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	switch method {
	case http.MethodGet:
		{
			req.Header.Add("Accept", "application/json")
		}
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		{
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+sd.SDJWT)
		}
	}

	return sd.HTTPClient.Do(req)
}

func (sd *SDAPI) jwt() (string, error) {
	path := "auth/token?api_token=" + sd.UserToken
	fullpath, err := sd.makeURL(path)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	res, err := sd.request(http.MethodGet, fullpath.String(), nil)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get JWT: StatusCode %d", res.StatusCode)
	}
	defer res.Body.Close()

	tokenResponse := new(tokenResponse)
	err = json.NewDecoder(res.Body).Decode(tokenResponse)

	return tokenResponse.JWT, err
}

func readScrewdriverYAML(filePath string) (string, error) {
	yaml, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read screwdriver.yaml: %v", err)
	}
	return string(yaml), nil
}

func (sd *SDAPI) validate(filePath string) (*validatorResponse, error) {
	fullpath, err := sd.makeURL("validator")
	if err != nil {
		return &validatorResponse{}, fmt.Errorf("failed to create api endpoint URL: %v", err)
	}

	yaml, err := readScrewdriverYAML(filePath)
	if err != nil {
		return &validatorResponse{}, err
	}

	escapedYaml := strings.ReplaceAll(string(yaml), "\"", "\\\"")
	body := fmt.Sprintf(`{"yaml": "%s"}`, strings.ReplaceAll(string(escapedYaml), "\n", "\\n"))

	res, err := sd.request(http.MethodPost, fullpath.String(), bytes.NewBuffer([]byte(body)))
	if err != nil {
		return &validatorResponse{}, fmt.Errorf("failed to send request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return &validatorResponse{}, fmt.Errorf("failed to post validator: StatusCode %d", res.StatusCode)
	}

	v := new(validatorResponse)
	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		return &validatorResponse{}, fmt.Errorf("failed to parse validator response: %v", err)
	}

	return v, nil
}

// Job returns job represented by "jobName"
func (sd *SDAPI) Job(jobName, filepath string) (Job, error) {
	v, err := sd.validate(filepath)
	if err != nil {
		return Job{}, err
	}

	if v.Errors != nil {
		return Job{}, fmt.Errorf("failed to parse screwdriver.yaml: %v", v.Errors)
	}

	job, ok := v.Jobs[jobName]
	if !ok {
		return Job{}, fmt.Errorf("not found '%s' in parsed screwdriver.yaml", jobName)
	}

	return job[0], nil
}
