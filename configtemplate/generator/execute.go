package generator

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type Executor struct {
	metdataBytes               []byte
	baseDirectory              string
	doNotIncludeProductVersion bool
	includeErrands             bool
	sizeOfCollections          int
	userSetSizeOfCollections   bool
}

func NewExecutor(metadataBytes []byte, baseDirectory string, doNotIncludeProductVersion, includeErrands bool, sizeOfCollections int, userSetSizeOfCollections bool) *Executor {
	return &Executor{
		metdataBytes:               metadataBytes,
		baseDirectory:              baseDirectory,
		doNotIncludeProductVersion: doNotIncludeProductVersion,
		includeErrands:             includeErrands,
		sizeOfCollections:          sizeOfCollections,
		userSetSizeOfCollections:   userSetSizeOfCollections,
	}
}

func (e *Executor) Generate() error {
	metadata, err := NewMetadata(e.metdataBytes)
	if err != nil {
		return err
	}
	productVersion := metadata.ProductVersion()
	if productVersion == "" {
		return errors.New("version in metadata is blank")
	}

	productName := metadata.ProductName()

	targetDirectory := path.Join(e.baseDirectory, productName)
	if !e.doNotIncludeProductVersion {
		targetDirectory = path.Join(targetDirectory, productVersion)
	}
	if err = e.createDirectory(targetDirectory); err != nil {
		return err
	}

	featuresDirectory := path.Join(targetDirectory, "features")
	if err = e.createDirectory(featuresDirectory); err != nil {
		return err
	}

	optionalDirectory := path.Join(targetDirectory, "optional")
	if err = e.createDirectory(optionalDirectory); err != nil {
		return err
	}

	networkDirectory := path.Join(targetDirectory, "network")
	if err = e.createDirectory(networkDirectory); err != nil {
		return err
	}

	resourceDirectory := path.Join(targetDirectory, "resource")
	if err = e.createDirectory(resourceDirectory); err != nil {
		return err
	}

	template, err := e.CreateTemplate(metadata)
	if err != nil {
		return err
	}

	template.ProductName = productName
	if err = e.writeYamlFile(path.Join(targetDirectory, "product.yml"), template); err != nil {
		return err
	}

	networkOpsFiles, err := CreateNetworkOpsFiles(metadata)
	if err != nil {
		return err
	}

	if len(networkOpsFiles) > 0 {
		for name, contents := range networkOpsFiles {
			if err = e.writeYamlFile(path.Join(networkDirectory, fmt.Sprintf("%s.yml", name)), contents); err != nil {
				return err
			}
		}
	}

	resourceVars := CreateResourceVars(metadata)

	if err = e.writeYamlFile(path.Join(targetDirectory, "resource-vars.yml"), resourceVars); err != nil {
		return err
	}

	var errandVars map[string]interface{}

	if e.includeErrands {
		errandVars = CreateErrandVars(metadata)
	}

	if err = e.writeYamlFile(path.Join(targetDirectory, "errand-vars.yml"), errandVars); err != nil {
		return err
	}

	resourceOpsFiles, err := CreateResourceOpsFiles(metadata)
	if err != nil {
		return err
	}

	if len(resourceOpsFiles) > 0 {
		for name, contents := range resourceOpsFiles {
			if err = e.writeYamlFile(path.Join(resourceDirectory, fmt.Sprintf("%s.yml", name)), contents); err != nil {
				return err
			}
		}
	}

	productPropertyVars, err := GetDefaultPropertyVars(metadata)
	if err != nil {
		return err
	}

	if err = e.writeYamlFile(path.Join(targetDirectory, "default-vars.yml"), productPropertyVars); err != nil {
		return err
	}

	requiredVars, err := GetRequiredPropertyVars(metadata)
	if err != nil {
		return err
	}
	if err = e.writeYamlFile(path.Join(targetDirectory, "required-vars.yml"), requiredVars); err != nil {
		return err
	}

	productPropertyOpsFiles, err := CreateProductPropertiesFeaturesOpsFiles(metadata)
	if err != nil {
		return err
	}

	if len(productPropertyOpsFiles) > 0 {
		for name, contents := range productPropertyOpsFiles {
			if err = e.writeYamlFile(path.Join(featuresDirectory, fmt.Sprintf("%s.yml", name)), contents); err != nil {
				return err
			}
		}
	}

	productPropertyOptionalOpsFiles, err := CreateProductPropertiesOptionalOpsFiles(metadata, e.sizeOfCollections, e.userSetSizeOfCollections)
	if err != nil {
		return err
	}

	if len(productPropertyOptionalOpsFiles) > 0 {
		for name, contents := range productPropertyOptionalOpsFiles {
			if err = e.writeYamlFile(path.Join(optionalDirectory, fmt.Sprintf("%s.yml", name)), contents); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Executor) CreateTemplate(metadata *Metadata) (*Template, error) {
	template := &Template{}
	if len(metadata.JobTypes) > 0 {
		template.NetworkProperties = CreateNetworkProperties(metadata)
		template.ResourceConfig = CreateResourceConfig(metadata)
	}

	if metadata.UsesOpsManagerSyslogProperties() {
		template.SyslogProperties = CreateSyslogProperties(metadata)
	}

	productProperties, err := GetAllProductProperties(metadata)
	if err != nil {
		return nil, err
	}
	template.ProductProperties = productProperties
	if e.includeErrands {
		template.ErrandConfig = CreateErrandConfig(metadata)
	}
	return template, nil
}

func (e *Executor) createDirectory(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %v", path, err)
		}
	}

	return nil
}

func (e *Executor) writeYamlFile(targetFile string, dataType interface{}) error {
	if dataType != nil {
		data, err := yaml.Marshal(dataType)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(targetFile, data, 0755)
	} else {
		return ioutil.WriteFile(targetFile, nil, 0755)
	}
}
