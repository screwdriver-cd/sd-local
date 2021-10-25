package screwdriver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var regexpForUnmarshal = regexp.MustCompile(`"(.*?)" *: *"(.*?)"`)

const (
	apiVersion        = "v4"
	validatorEndpoint = "validator"
	tokenEndpoint     = "auth/token"
)

// API has method to get job
type API interface {
	Job(jobName, filePath string) (Job, error)
	JWT() string
	InitJWT() error
}

type sdAPI struct {
	HTTPClient *http.Client
	UserToken  string
	APIURL     string
	SDJWT      string
	UA         string
}

var _ API = (*sdAPI)(nil)

// Step is step entity struct
type Step struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// EnvVar is an environment slice
type EnvVar []struct {
	Key   string
	Value string
}

// UnmarshalJSON replaces JSON of a normal associative array to EnvVar
func (en *EnvVar) UnmarshalJSON(data []byte) error {
	for _, pair := range regexpForUnmarshal.FindAllStringSubmatch(string(data), -1) {
		*en = append(*en, struct {
			Key   string
			Value string
		}{
			pair[1],
			pair[2],
		})
	}
	return nil
}

// MarshalJSON replaces EnvVar to JSON of a normal associative array
func (en EnvVar) MarshalJSON() ([]byte, error) {
	outputArray := make([]string, len(en))
	for i, pair := range en {
		outputArray[i] = "{\"" + pair.Key + "\":\"" + pair.Value + "\"}"
	}
	output := "[" + strings.Join(outputArray, ",") + "]"
	return []byte(output), nil
}

// Get the newest Value whose Key is key
func (en EnvVar) Get(key string) string {
	s := ""
	for _, e := range en {
		if e.Key == key {
			s = e.Value
		}
	}
	return s
}

// Set adds (key, val)
func (en *EnvVar) Set(key string, val string) {
	*en = append(*en, struct {
		Key   string
		Value string
	}{
		key,
		val,
	})
}

// Merge en2 to en
func (en *EnvVar) Merge(en2 EnvVar) {
	for _, e := range en2 {
		en.Set(e.Key, e.Value)
	}
}

// Job is job entity struct
type Job struct {
	Steps       []Step `json:"commands"`
	Environment EnvVar `json:"environment"`
	Image       string `json:"image"`
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
func New(apiURL, token, ua string) API {
	s := &sdAPI{
		HTTPClient: http.DefaultClient,
		APIURL:     apiURL,
		UserToken:  token,
		UA:         ua,
	}

	return s
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

	req.Header.Add("User-Agent", sd.UA)

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

	escapedYaml := strconv.Quote(yaml)
	body := fmt.Sprintf(`{"yaml": %s}`, escapedYaml)

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

func (sd *sdAPI) InitJWT() error {
	jwt, err := sd.jwt()
	if err != nil {
		return err
	}

	sd.SDJWT = jwt

	return nil
}

// JWT returns JWT token for screwdriver API
func (sd *sdAPI) JWT() string {
	return sd.SDJWT
}
