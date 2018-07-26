package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type SetErrandState struct {
	service setErrandStateService
	Options struct {
		ProductName     string `long:"product-name"      short:"p" required:"true" description:"name of product"`
		ErrandName      string `long:"errand-name"       short:"e" required:"true" description:"name of errand"`
		PostDeployState string `long:"post-deploy-state"                           description:"desired errand state. (enabled/disabled/when-changed)"`
		PreDeleteState  string `long:"pre-delete-state"                            description:"desired errand state (enabled/disabled)"`
	}
}

//go:generate counterfeiter -o ./fakes/set_errand_state_service.go --fake-name SetErrandStateService . setErrandStateService
type setErrandStateService interface {
	GetStagedProductByName(productName string) (api.StagedProductsFindOutput, error)
	UpdateStagedProductErrands(productID, errandName string, postDeployState, preDeleteState interface{}) error
}

var userToOMInputs = map[string]interface{}{
	"enabled":      true,
	"disabled":     false,
	"when-changed": "when-changed",
	"default":      "default",
}

func NewSetErrandState(service setErrandStateService) SetErrandState {
	return SetErrandState{
		service: service,
	}
}

func (s SetErrandState) Execute(args []string) error {
	if _, err := jhanda.Parse(&s.Options, args); err != nil {
		return fmt.Errorf("could not parse set-errand-state flags: %s", err)
	}

	findOutput, err := s.service.GetStagedProductByName(s.Options.ProductName)
	if err != nil {
		return fmt.Errorf("failed to find staged product %q: %s", s.Options.ProductName, err)
	}

	var (
		preDeleteState  interface{}
		postDeployState interface{}
		errs            []string
	)

	if s.Options.PostDeployState != "" {
		var ok bool
		postDeployState, ok = userToOMInputs[s.Options.PostDeployState]
		if !ok {
			errs = append(errs, fmt.Sprintf("post-deploy-state %q is invalid", s.Options.PostDeployState))
		}
	}

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

	err = s.service.UpdateStagedProductErrands(findOutput.Product.GUID, s.Options.ErrandName, postDeployState, preDeleteState)
	if err != nil {
		return fmt.Errorf("failed to set errand state: %s", err)
	}

	return nil
}

func (s SetErrandState) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command sets the state of a product's errand.",
		ShortDescription: "**DEPRECATED** (use configure-product instead) sets state for a product's errand",
		Flags:            s.Options,
	}
}
