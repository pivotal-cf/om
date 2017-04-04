package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pivotal-cf/om/flags"
)

type SetErrandState struct {
	errandsService       errandsService
	stagedProductsFinder stagedProductsFinder
	Options              struct {
		ProductName     string `short:"p" long:"product-name" description:"name of product"`
		ErrandName      string `short:"e" long:"errand-name" description:"name of errand"`
		PostDeployState string `long:"post-deploy-state" description:"desired errand state. (enabled/disabled/when-changed)"`
		PreDeleteState  string `long:"pre-delete-state" description:"desired errand state (enabled/disabled)"`
	}
}

var userToOMInputs = map[string]string{
	"enabled":      "true",
	"disabled":     "false",
	"when-changed": "when-changed",
	"default":      "default",
}

func NewSetErrandState(errandsService errandsService, stagedProductsFinder stagedProductsFinder) SetErrandState {
	return SetErrandState{
		errandsService:       errandsService,
		stagedProductsFinder: stagedProductsFinder,
	}
}

func (s SetErrandState) Execute(args []string) error {
	_, err := flags.Parse(&s.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse errands flags: %s", err)
	}

	if s.Options.ProductName == "" {
		return errors.New("error: product-name is missing. Please see usage for more information.")
	}

	if s.Options.ErrandName == "" {
		return errors.New("error: errand-name is missing. Please see usage for more information.")
	}

	findOutput, err := s.stagedProductsFinder.Find(s.Options.ProductName)
	if err != nil {
		return fmt.Errorf("failed to find staged product %q: %s", s.Options.ProductName, err)
	}

	var (
		postDeployState string
		errs            []string
	)

	if s.Options.PostDeployState != "" {
		var ok bool
		postDeployState, ok = userToOMInputs[s.Options.PostDeployState]
		if !ok {
			errs = append(errs, fmt.Sprintf("post-deploy-state %q is invalid", s.Options.PostDeployState))
		}
	}

	var preDeleteState string
	if s.Options.PreDeleteState != "" {
		var ok bool
		preDeleteState, ok = userToOMInputs[s.Options.PreDeleteState]
		if !ok {
			errs = append(errs, fmt.Sprintf("pre-delete-state %q is invalid", s.Options.PreDeleteState))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	err = s.errandsService.SetState(findOutput.Product.GUID, s.Options.ErrandName, postDeployState, preDeleteState)
	if err != nil {
		return fmt.Errorf("failed to set errand state: %s", err)
	}

	return nil
}

func (s SetErrandState) Usage() Usage {
	return Usage{
		Description:      "This authenticated command sets the state of a product's errand.",
		ShortDescription: "sets state for a product's errand",
		Flags:            s.Options,
	}
}
