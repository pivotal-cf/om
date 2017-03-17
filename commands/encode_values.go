package commands

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type NetworksConfiguration struct {
	ICMP     bool `url:"infrastructure[icmp_checks_enabled],int" json:"icmp_checks_enabled"`
	Networks Networks
	CommonConfiguration
}

type Networks []NetworkConfiguration

type NetworkConfiguration struct {
	GUID           int      `url:"network_collection[networks_attributes][**network**][guid]"`
	Name           string   `url:"network_collection[networks_attributes][**network**][name]" json:"name"`
	ServiceNetwork *bool    `url:"network_collection[networks_attributes][**network**][service_network]" json:"service_network"`
	Subnets        []Subnet `json:"subnets"`
}

type Subnet struct {
	IAASIdentifier        string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][iaas_identifier]" json:"iaas_identifier"`
	CIDR                  string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][cidr]" json:"cidr"`
	ReservedIPRanges      string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][reserved_ip_ranges]" json:"reserved_ip_ranges"`
	DNS                   string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][dns]" json:"dns"`
	Gateway               string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][gateway]" json:"gateway"`
	AvailabilityZones     []string `json:"availability_zones,omitempty"`
	AvailabilityZoneGUIDs []string `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][availability_zone_references][]"`
}

func (n Networks) EncodeValues(key string, v *url.Values) error {
	var (
		networkingFields []reflect.Type
		networkingValues []reflect.Value
	)

	for index, config := range n {
		networkingFields = append(networkingFields, reflect.TypeOf(config))
		networkingValues = append(networkingValues, reflect.ValueOf(config))

		numNetworks := strconv.Itoa(index)

		for i, subnet := range config.Subnets {
			networkingFields = append(networkingFields, reflect.TypeOf(subnet))
			networkingValues = append(networkingValues, reflect.ValueOf(subnet))

			numSubnets := strconv.Itoa(i)

			assignIndex(networkingFields, networkingValues, numNetworks, numSubnets, v)
		}
	}

	return nil
}

func assignIndex(fields []reflect.Type, values []reflect.Value, numNetworks string, numSubnets string, urlValues *url.Values) {
	for index, v := range values {
		for i := 0; i < v.NumField(); i++ {
			field := fields[index].Field(i)
			value := v.Field(i)
			tag := field.Tag.Get("url")

			if tag == "" {
				continue
			}

			// add value nil check here to continue
			// only for bool right now but whatevs

			newTag := strings.Replace(tag, "**subnet**", numSubnets, -1)
			finalTag := strings.Replace(newTag, "**network**", numNetworks, -1)

			switch value.Kind() {
			case reflect.Int:
				urlValues.Set(finalTag, "0")
			case reflect.String:
				urlValues.Set(finalTag, value.String())
			case reflect.Bool:
				boolString := "0"
				if value.Bool() {
					boolString = "1"
				}
				urlValues.Set(finalTag, boolString)
			case reflect.Ptr:
				if !value.IsNil() {
					urlValues.Set(finalTag, convertBoolPtr(reflect.Indirect(value)))
				}
			case reflect.Slice:
				temp := *urlValues
				temp[finalTag] = value.Interface().([]string)
			}
		}
	}
}

func convertBoolPtr(value reflect.Value) string {
	if value.Bool() {
		return "1"
	}
	return "0"
}
