package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

type ResponseProperty struct {
	Value        interface{}
	Configurable bool
	IsCredential bool   `yaml:"credential"`
	Type         string `yaml:"type"`
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

// TODO: extract to helper package?
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
		return fmt.Errorf("could not make %s api request to staged products endpoint: %s", stReq.Method, err)
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

	var req *http.Request
	req, err = http.NewRequest("DELETE", fmt.Sprintf("/api/v0/staged/products/%s", stagedGUID), strings.NewReader("{}"))

	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make %s api request to staged products endpoint: %s", req.Method, err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) ListStagedProducts() (StagedProductsOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/staged/products", nil)
	if err != nil {
		return StagedProductsOutput{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not make request to staged-products endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return StagedProductsOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StagedProductsOutput{}, err
	}

	var stagedProducts []StagedProduct
	err = json.Unmarshal(respBody, &stagedProducts)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not unmarshal staged products response: %s", err)
	}

	return StagedProductsOutput{
		Products: stagedProducts,
	}, nil
}

func (a Api) getString(element interface{}, key string) (string, error) {
	value, err := a.get(element, key)
	if err != nil {
		return "", err
	}
	strVal, ok := value.(string)
	if ok {
		return strVal, nil
	}
	return "", fmt.Errorf("element %v with key %s is not a string", element, key)
}

func (a Api) set(element interface{}, key, value string) error {
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

func (a Api) get(element interface{}, key string) (interface{}, error) {
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

func (a Api) collectionElementGUID(propertyName, elementName string, configuredProperties map[string]ResponseProperty) (string, error) {
	collection := configuredProperties[propertyName].Value
	collectionArray := collection.([]interface{})
	for _, collectionElement := range collectionArray {
		element, err := a.get(collectionElement, "name")
		if err != nil {
			return "", err
		}
		currentElement, err := a.getString(element, "value")
		if err != nil {
			return "", err
		}
		if currentElement == elementName {
			guidElement, err := a.get(collectionElement, "guid")
			if err != nil {
				return "", err
			}
			guid, err := a.getString(guidElement, "value")
			if err != nil {
				return "", err
			}
			return guid, nil
		}

	}
	return "", nil
}

func (a Api) UpdateStagedProductProperties(input UpdateStagedProductPropertiesInput) error {
	currentConfiguredProperties, err := a.GetStagedProductProperties(input.GUID)
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
			collectionValue, err := a.get(property, "value")
			if err != nil {
				return err
			}
			for _, collectionElement := range collectionValue.([]interface{}) {
				name, err := a.getString(collectionElement, "name")
				if err != nil {
					return err
				}
				guid, err := a.collectionElementGUID(propertyName, name, currentConfiguredProperties)
				if err != nil {
					return err
				}
				err = a.set(collectionElement, "guid", guid)
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
		return fmt.Errorf("could not make api request to staged product properties endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateStagedProductNetworksAndAZs(input UpdateStagedProductNetworksAndAZsInput) error {
	body := bytes.NewBufferString(fmt.Sprintf(`{"networks_and_azs": %s}`, input.NetworksAndAZs))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s/networks_and_azs", input.GUID), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to staged product networks_and_azs endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

//TODO consider refactoring to use fetchProductResource
func (a Api) GetStagedProductManifest(guid string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/manifest", guid), nil)
	if err != nil {
		return "", err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not make api request to staged products manifest endpoint: %s", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var contents struct {
		Manifest interface{} `json:"manifest"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = yaml.Unmarshal(body, &contents); err != nil {
		return "", fmt.Errorf("could not parse json: %s", err)
	}

	manifest, err := yaml.Marshal(contents.Manifest)
	if err != nil {
		return "", err // this should never happen, all valid json can be marshalled
	}

	return string(manifest), nil
}

func (a Api) GetStagedProductProperties(product string) (map[string]ResponseProperty, error) {
	respBody, err := a.fetchProductResource(product, "properties")
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, err
	}

	propertiesResponse := struct {
		Properties map[string]ResponseProperty
	}{}

	if err = yaml.Unmarshal(body, &propertiesResponse); err != nil {
		return nil, fmt.Errorf("could not parse json: %s", err)
	}

	return propertiesResponse.Properties, nil
}

func (a Api) GetStagedProductNetworksAndAZs(product string) (map[string]interface{}, error) {
	respBody, err := a.fetchProductResource(product, "networks_and_azs")
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	networksResponse := struct {
		Networks map[string]interface{} `json:"networks_and_azs"`
	}{}
	if err = json.NewDecoder(respBody).Decode(&networksResponse); err != nil {
		return nil, fmt.Errorf("could not parse json: %s", err)
	}

	return networksResponse.Networks, nil
}

func (a Api) fetchProductResource(guid, endpoint string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/%s", guid, endpoint), nil)
	if err != nil {
		return nil, err // un-tested
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil,
			fmt.Errorf("could not make api request to staged product properties endpoint: %s", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	return resp.Body, nil
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
