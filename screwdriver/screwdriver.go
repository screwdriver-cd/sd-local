package screwdriver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	apiVersion        = "v4"
	validatorEndpoint = "validator"
	tokenEndpoint     = "auth/token"
)

// API has method to get job
type API interface {
	Job(jobName, filePath string) (Job, error)
	JWT() string
}

type sdAPI struct {
	HTTPClient *http.Client
	UserToken  string
	APIURL     string
	SDJWT      string
}

var _ API = (*sdAPI)(nil)

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

type jobs map[string][]Job

type validatorResponse struct {
	Jobs   jobs     `json:"jobs"`
	Errors []string `json:"errors"`
}

type tokenResponse struct {
	JWT string `json:"token"`
}

// New creates a API
func New(apiURL, token string) (API, error) {
	s := &sdAPI{
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

func (sd *sdAPI) makeURL(endpoint string) (*url.URL, error) {
	u, err := url.Parse(sd.APIURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, apiVersion, endpoint)

	return u, nil
}

func (sd *sdAPI) request(method, path string, body io.Reader) (*http.Response, error) {
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

func (sd *sdAPI) jwt() (string, error) {
	fullpath, err := sd.makeURL(tokenEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to make request url: %v", err)
	}

	query := fullpath.Query()
	query.Set("api_token", sd.UserToken)
	fullpath.RawQuery = query.Encode()

	res, err := sd.request(http.MethodGet, fullpath.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get JWT: StatusCode %d", res.StatusCode)
	}
	defer res.Body.Close()

	tokenResponse := new(tokenResponse)
	err = json.NewDecoder(res.Body).Decode(tokenResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse JWT response: %v", err)
	}

	return tokenResponse.JWT, nil
}

func readScrewdriverYAML(filePath string) (string, error) {
	yaml, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read screwdriver.yaml: %v", err)
	}
	return string(yaml), nil
}

func (sd *sdAPI) validate(filePath string) (jobs, error) {
	fullpath, err := sd.makeURL(validatorEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to make request url: %v", err)
	}

	yaml, err := readScrewdriverYAML(filePath)
	if err != nil {
		return nil, err
	}

	escapedYaml := strings.ReplaceAll(yaml, "\"", "\\\"")
	body := fmt.Sprintf(`{"yaml": "%s"}`, strings.ReplaceAll(escapedYaml, "\n", "\\n"))

	res, err := sd.request(http.MethodPost, fullpath.String(), strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to post validator: StatusCode %d", res.StatusCode)
	}

	v := new(validatorResponse)
	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		return nil, fmt.Errorf("failed to parse validator response: %v", err)
	}

	if v.Errors != nil {
		return nil, fmt.Errorf("failed to parse screwdriver.yaml: %v", v.Errors)
	}

	return v.Jobs, nil
}

// Job returns job represented by "jobName"
func (sd *sdAPI) Job(jobName, filepath string) (Job, error) {
	jobs, err := sd.validate(filepath)
	if err != nil {
		return Job{}, err
	}

	job, ok := jobs[jobName]
	if !ok {
		return Job{}, fmt.Errorf("not found '%s' in parsed screwdriver.yaml", jobName)
	}

	return job[0], nil
}

// JWT returns JWT token for screwdriver API
func (sd *sdAPI) JWT() string {
	return sd.SDJWT
}
