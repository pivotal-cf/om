package commands

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type ConfigureProduct struct {
	productsService productConfigurer
	jobsService     jobsConfigurer
	logger          logger
	Options         struct {
		ProductName       string `short:"n"  long:"product-name" description:"name of the product being configured"`
		ProductProperties string `short:"p" long:"product-properties" description:"properties to be configured in JSON format" default:""`
		NetworkProperties string `short:"pn" long:"product-network" description:"network properties in JSON format" default:""`
		ProductResources  string `short:"pr" long:"product-resources" description:"resource configurations in JSON format" default:"{}"`
	}
}

//go:generate counterfeiter -o ./fakes/product_configurer.go --fake-name ProductConfigurer . productConfigurer
type productConfigurer interface {
	StagedProducts() (api.StagedProductsOutput, error)
	Configure(api.ProductsConfigurationInput) error
}

//go:generate counterfeiter -o ./fakes/jobs_configurer.go --fake-name JobsConfigurer . jobsConfigurer
type jobsConfigurer interface {
	Jobs(productGUID string) (map[string]string, error)
	GetExistingJobConfig(productGUID, jobGUID string) (api.JobProperties, error)
	ConfigureJob(productGUID, jobGUID string, jobProperties api.JobProperties) error
}

func NewConfigureProduct(productConfigurer productConfigurer, jobsConfigurer jobsConfigurer, logger logger) ConfigureProduct {
	return ConfigureProduct{
		productsService: productConfigurer,
		jobsService:     jobsConfigurer,
		logger:          logger,
	}
}

func (cp ConfigureProduct) Execute(args []string) error {
	_, err := flags.Parse(&cp.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse configure-product flags: %s", err)
	}

	if cp.Options.ProductName == "" {
		return fmt.Errorf("error: product-name is missing. Please see usage for more information.")
	}

	cp.logger.Printf("configuring product...")

	if cp.Options.ProductProperties == "" && cp.Options.NetworkProperties == "" && cp.Options.ProductResources == "{}" {
		cp.logger.Printf("Provided properties are empty, nothing to do here")
		return nil
	}

	stagedProducts, err := cp.productsService.StagedProducts()
	if err != nil {
		return err
	}

	var productGUID string
	for _, sp := range stagedProducts.Products {
		if sp.Type == cp.Options.ProductName {
			productGUID = sp.GUID
			break
		}
	}

	if productGUID == "" {
		return fmt.Errorf(`could not find product "%s"`, cp.Options.ProductName)
	}

	if cp.Options.ProductProperties != "" {
		cp.logger.Printf("setting properties")
		err = cp.productsService.Configure(api.ProductsConfigurationInput{
			GUID:          productGUID,
			Configuration: cp.Options.ProductProperties,
		})
		if err != nil {
			return fmt.Errorf("failed to configure product: %s", err)
		}
		cp.logger.Printf("finished setting properties")
	}

	if cp.Options.NetworkProperties != "" {
		cp.logger.Printf("setting up network")
		err = cp.productsService.Configure(api.ProductsConfigurationInput{
			GUID:    productGUID,
			Network: cp.Options.NetworkProperties,
		})
		if err != nil {
			return fmt.Errorf("failed to configure product: %s", err)
		}
		cp.logger.Printf("finished setting up network")
	}

	if cp.Options.ProductResources != "{}" {
		var userProvidedConfig map[string]json.RawMessage
		err = json.Unmarshal([]byte(cp.Options.ProductResources), &userProvidedConfig)
		if err != nil {
			return fmt.Errorf("could not decode product-resource json: %s", err)
		}

		jobs, err := cp.jobsService.Jobs(productGUID)
		if err != nil {
			return fmt.Errorf("failed to fetch jobs: %s", err)
		}

		var names []string
		for name, _ := range userProvidedConfig {
			names = append(names, name)
		}

		sort.Strings(names)

		cp.logger.Printf("applying resource configuration for the following jobs:")
		for _, name := range names {
			jobProperties, err := cp.jobsService.GetExistingJobConfig(productGUID, jobs[name])
			if err != nil {
				return fmt.Errorf("could not fetch existing job configuration: %s", err)
			}

			err = json.Unmarshal(userProvidedConfig[name], &jobProperties)
			if err != nil {
				return err
			}

			if !reflect.DeepEqual(jobProperties, api.JobProperties{}) {
				cp.logger.Printf("\t%s", name)
				err = cp.jobsService.ConfigureJob(productGUID, jobs[name], jobProperties)
				if err != nil {
					return fmt.Errorf("failed to configure resources: %s", err)
				}
			}
		}
	}

	cp.logger.Printf("finished configuring product")

	return nil
}

func (cp ConfigureProduct) Usage() Usage {
	return Usage{
		Description:      "This authenticated command configures a staged product",
		ShortDescription: "configures a staged product",
		Flags:            cp.Options,
	}
}
