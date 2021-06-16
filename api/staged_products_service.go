package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type StageProductInput struct {
	ProductName    string `json:"name"`
	ProductVersion string `json:"product_version"`
}

type StagedProductsOutput struct {
	Products []StagedProduct
}

type StagedProduct struct {
	GUID string
	Type string
}

type UnstageProductInput struct {
	ProductName string `json:"name"`
}

type UpdateStagedProductPropertiesInput struct {
	GUID       string
	Properties string
}

type UpdateStagedProductNetworksAndAZsInput struct {
	GUID           string
	NetworksAndAZs string
}

type SyslogConfiguration struct {
	SyslogConfiguration map[string]interface{} `json:"syslog_configuration"`
}

type UpdateSyslogConfigurationInput struct {
	GUID                string
	SyslogConfiguration string
}

type ResponseProperty struct {
	Value          interface{}
	SelectedOption string `yaml:"selected_option,omitempty"`
	Configurable   bool
	IsCredential   bool   `yaml:"credential"`
	Type           string `yaml:"type"`
}

func (r *ResponseProperty) isCollection() bool {
	return r.Type == "collection"
}

type UpgradeRequest struct {
	ToVersion string `json:"to_version"`
}

type ConfigurationRequest struct {
	Method        string
	URL           string
	Configuration string
}

func (a Api) Stage(input StageProductInput, deployedGUID string) error {
	stagedGUID, err := a.checkStagedProducts(input.ProductName)
	if err != nil {
		return err
	}

	var stReq *http.Request
	if deployedGUID == "" && stagedGUID == "" {
		stagedProductBody, err := json.Marshal(input)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("POST", "/api/v0/staged/products", bytes.NewBuffer(stagedProductBody))
		if err != nil {
			return err
		}
	} else if deployedGUID != "" {
		upgradeReq := UpgradeRequest{
			ToVersion: input.ProductVersion,
		}

		upgradeReqBody, err := json.Marshal(upgradeReq)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s", deployedGUID), bytes.NewBuffer(upgradeReqBody))
		if err != nil {
			return err
		}
	} else if stagedGUID != "" {
		upgradeReq := UpgradeRequest{
			ToVersion: input.ProductVersion,
		}

		upgradeReqBody, err := json.Marshal(upgradeReq)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s", stagedGUID), bytes.NewBuffer(upgradeReqBody))
		if err != nil {
			return err
		}
	}

	stReq.Header.Set("Content-Type", "application/json")
	stResp, err := a.client.Do(stReq)
	if err != nil {
		return fmt.Errorf("could not make %s api request to staged products endpoint: %w", stReq.Method, err)
	}
	defer stResp.Body.Close()

	if err = validateStatusOK(stResp); err != nil {
		return err
	}

	return nil
}

func (a Api) DeleteStagedProduct(input UnstageProductInput) error {
	stagedGUID, err := a.checkStagedProducts(input.ProductName)
	if err != nil {
		return err
	}
	if len(stagedGUID) == 0 {
		return fmt.Errorf("product is not staged: %s", input.ProductName)
	}

	resp, err := a.sendAPIRequest("DELETE", fmt.Sprintf("/api/v0/staged/products/%s", stagedGUID), []byte(`{}`))
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) ListStagedProducts() (StagedProductsOutput, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/products", nil)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not make request to staged-products endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return StagedProductsOutput{}, err
	}

	var stagedProducts []StagedProduct
	err = json.NewDecoder(resp.Body).Decode(&stagedProducts)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not unmarshal staged products response: %w", err)
	}

	return StagedProductsOutput{
		Products: stagedProducts,
	}, nil
}

func (a Api) GetStagedProductSyslogConfiguration(product string) (map[string]interface{}, error) {
	req := fmt.Sprintf("/api/v0/staged/products/%s/syslog_configuration", product)
	resp, err := a.sendAPIRequest("GET", req, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to staged product syslog_configuration endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, nil
	}

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var syslogConfig SyslogConfiguration
	err = json.NewDecoder(resp.Body).Decode(&syslogConfig)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal staged product syslog_configuration response: %w", err)
	}

	return syslogConfig.SyslogConfiguration, nil
}

var errNotFound = errors.New("element not found")

func getString(element interface{}, key string) (string, error) {
	value, err := get(element, key)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", errNotFound
	}
	strVal, ok := value.(string)
	if ok {
		return strVal, nil
	}
	return "", fmt.Errorf("element %v (%T) with key %q is not a string", element, value, key)
}

func set(element interface{}, key, value string) error {
	mapString, ok := element.(map[string]interface{})
	if ok {
		mapString[key] = value
		return nil
	}
	mapInterface, ok := element.(map[interface{}]interface{})
	if ok {
		mapInterface[key] = value
		return nil
	}
	return fmt.Errorf("Unexpected type %v", element)
}

func get(element interface{}, key string) (interface{}, error) {
	mapString, ok := element.(map[string]interface{})
	if ok {
		return mapString[key], nil
	}
	mapInterface, ok := element.(map[interface{}]interface{})
	if ok {
		return mapInterface[key], nil
	}
	return nil, fmt.Errorf("Unexpected type %v", element)
}

func collectionElementGUID(propertyName, elementName string, configuredProperties map[string]ResponseProperty) (string, error) {
	collection := configuredProperties[propertyName].Value
	collectionArray := collection.([]interface{})
	for _, collectionElement := range collectionArray {
		element, err := get(collectionElement, "name")
		if err != nil {
			return "", err
		}
		currentElement, err := getString(element, "value")
		if err != nil {
			return "", err
		}
		if currentElement == elementName {
			guidElement, err := get(collectionElement, "guid")
			if err != nil {
				return "", err
			}
			guid, err := getString(guidElement, "value")
			if err != nil {
				return "", err
			}
			return guid, nil
		}

	}
	return "", nil
}

func (a Api) UpdateStagedProductProperties(input UpdateStagedProductPropertiesInput) error {
	currentConfiguredProperties, err := a.GetStagedProductProperties(input.GUID, false)
	if err != nil {
		return err
	}

	newProperties := make(map[string]interface{})
	err = json.Unmarshal([]byte(input.Properties), &newProperties)
	if err != nil {
		return err
	}
	for propertyName, property := range newProperties {
		configuredProperty := currentConfiguredProperties[propertyName]
		if configuredProperty.isCollection() {
			collectionValue, err := get(property, "value")
			if err != nil {
				return err
			}
			for _, collectionElement := range collectionValue.([]interface{}) {
				name, err := getString(collectionElement, "name")
				if err != nil {
					if err == errNotFound {
						continue
					}
					return err
				}
				guid, err := collectionElementGUID(propertyName, name, currentConfiguredProperties)
				if err != nil {
					return err
				}
				err = set(collectionElement, "guid", guid)
				if err != nil {
					return err
				}
			}
		}
	}

	propertyJson, err := json.Marshal(newProperties)
	if err != nil {
		return err
	}
	body := bytes.NewBufferString(fmt.Sprintf(`{"properties": %s}`, propertyJson))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s/properties", input.GUID), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to staged product properties endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateStagedProductNetworksAndAZs(input UpdateStagedProductNetworksAndAZsInput) error {
	resp, err := a.sendAPIRequest("PUT",
		fmt.Sprintf("/api/v0/staged/products/%s/networks_and_azs", input.GUID),
		[]byte(fmt.Sprintf(`{"networks_and_azs": %s}`, input.NetworksAndAZs)),
	)
	if err != nil {
		return fmt.Errorf("could not make api request to staged product networks_and_azs endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateSyslogConfiguration(input UpdateSyslogConfigurationInput) error {
	resp, err := a.sendAPIRequest("PUT",
		fmt.Sprintf("/api/v0/staged/products/%s/syslog_configuration", input.GUID),
		[]byte(fmt.Sprintf(`{"syslog_configuration": %s}`, input.SyslogConfiguration)),
	)
	if err != nil {
		return fmt.Errorf("could not make api request to staged product syslog_configuration endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

// TODO consider refactoring to use fetchProductResource
func (a Api) GetStagedProductManifest(guid string) (string, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/manifest", guid), nil)
	if err != nil {
		return "", fmt.Errorf("could not make api request to staged products manifest endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return "", err
	}

	var contents struct {
		Manifest interface{}
	}
	err = yaml.NewDecoder(resp.Body).Decode(&contents)
	if err != nil {
		return "", fmt.Errorf("could not parse json: %w", err)
	}

	manifest, err := yaml.Marshal(contents.Manifest)
	if err != nil {
		return "", err
	}

	return string(manifest), nil
}

func (a Api) GetStagedProductProperties(product string, redact bool) (map[string]ResponseProperty, error) {
	endpoint := "properties?redact=false"
	if redact {
		endpoint = "properties?redact=true"
	}
	resp, err := a.fetchProductResource(product, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var propertiesResponse struct {
		Properties map[string]ResponseProperty
	}
	err = yaml.NewDecoder(resp.Body).Decode(&propertiesResponse)
	if err != nil {
		return nil, fmt.Errorf("could not parse json: %w", err)
	}

	return propertiesResponse.Properties, nil
}

func (a Api) GetStagedProductNetworksAndAZs(product string) (map[string]interface{}, error) {
	var networksResponse struct {
		Networks map[string]interface{} `json:"networks_and_azs"`
	}

	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/networks_and_azs", product), nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to staged product properties endpoint: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// TODO should be a log line here probably?
		return nil, nil
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(resp.Body).Decode(&networksResponse); err != nil {
		return nil, fmt.Errorf("could not parse json: %w", err)
	}

	return networksResponse.Networks, nil
}

func (a Api) fetchProductResource(guid, endpoint string) (*http.Response, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/%s", guid, endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to staged product properties endpoint: %w", err)
	}

	return resp, nil
}

func (a Api) checkStagedProducts(productName string) (string, error) {
	stagedProductsOutput, err := a.ListStagedProducts()
	if err != nil {
		return "", err
	}

	for _, product := range stagedProductsOutput.Products {
		if productName == product.Type {
			return product.GUID, nil
		}
	}

	return "", nil
}

func (a Api) GetStagedProductJobMaxInFlight(productGUID string) (ProductJobMaxInFlights map[string]interface{}, err error) {
	resp, err := a.sendAPIRequest(
		"GET",
		fmt.Sprintf("/api/v0/staged/products/%s/max_in_flight", productGUID),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)

	var payload map[string]interface{}
	err = json.Unmarshal(contents, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response from server: %v", err)
	}

	ProductJobMaxInFlights, _ = payload["max_in_flight"].(map[string]interface{})
	return ProductJobMaxInFlights, nil
}

func (a Api) UpdateStagedProductJobMaxInFlight(productGUID string, jobsToMaxInFlight map[string]interface{}) error {
	if len(jobsToMaxInFlight) == 0 {
		return nil
	}

	for job, maxInFlight := range jobsToMaxInFlight {
		if v, ok := maxInFlight.(string); ok {
			if !(strings.Contains(v, "%") || v == "default") {
				value, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("invalid max_in_flight value provided for job '%s': '%s'\nvalid options configurations include percentages ('50%%'), counts ('2'), and 'default'", job, v)
				}
				jobsToMaxInFlight[job] = value
			}
		}
	}

	payload, err := json.Marshal(jobsToMaxInFlight)
	if err != nil {
		return err
	}

	resp, err := a.sendAPIRequest("PUT",
		fmt.Sprintf("/api/v0/staged/products/%s/max_in_flight", productGUID),
		[]byte(fmt.Sprintf(`{"max_in_flight": %s}`, payload)),
	)
	if err != nil {
		return fmt.Errorf("could not make api request to staged product networks_and_azs endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
