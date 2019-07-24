package generator

import "fmt"

type PropertyValue interface {
	Parameters() []string
	IsSelector() bool
}

type SimpleType interface {
	Parameters() []string
	IsSelector() bool
}

type SimpleString string

func (s SimpleString) Parameters() []string {
	return []string{fmt.Sprintf("%v", s)}
}

func (s SimpleString) IsSelector() bool {
	return false
}

type SimpleBoolean bool

func (s SimpleBoolean) Parameters() []string {
	return []string{fmt.Sprintf("%v", s)}
}

func (s SimpleBoolean) IsSelector() bool {
	return false
}

type SimpleInteger int

func (s SimpleInteger) Parameters() []string {
	return []string{fmt.Sprintf("%v", s)}
}

func (s SimpleInteger) IsSelector() bool {
	return false
}

type SelectorValue struct {
	Value string `yaml:"value"`
}

func (s *SelectorValue) Parameters() []string {
	return []string{fmt.Sprintf("selector -> %s", s.Value)}
}
func (s *SelectorValue) IsSelector() bool {
	return true
}

type MultiSelectorValue struct {
	Value []string `yaml:"value"`
}

func (s *MultiSelectorValue) Parameters() []string {
	return s.Value
}

func (s *MultiSelectorValue) IsSelector() bool {
	return false
}

type SimpleValue struct {
	Value string `yaml:"value"`
}

func (s *SimpleValue) Parameters() []string {
	return []string{s.Value}
}

func (s *SimpleValue) IsSelector() bool {
	return false
}

type SimpleCredentialValueHolder struct {
	Value *SimpleCredentialValue `yaml:"value"`
}

func (s *SimpleCredentialValueHolder) Parameters() []string {
	return []string{s.Value.Password, s.Value.Identity}
}

func (s *SimpleCredentialValueHolder) IsSelector() bool {
	return false
}

type SecretValueHolder struct {
	Value *SecretValue `yaml:"value"`
}

func (s *SecretValueHolder) Parameters() []string {
	return []string{s.Value.Value}
}

func (s *SecretValueHolder) IsSelector() bool {
	return false
}

type SecretValue struct {
	Value string `yaml:"secret"`
}

func (s *SecretValue) Parameters() []string {
	return []string{s.Value}
}

func (s *SecretValue) IsSelector() bool {
	return false
}

type SimpleCredentialValue struct {
	Identity string `yaml:"identity"`
	Password string `yaml:"password"`
}

type CertificateValueHolder struct {
	Value *CertificateValue `yaml:"value"`
}

func (s *CertificateValueHolder) Parameters() []string {
	return []string{s.Value.CertPem, s.Value.CertPrivateKey}
}

func (s *CertificateValueHolder) IsSelector() bool {
	return false
}

type CertificateValue struct {
	CertPem        string `yaml:"cert_pem"`
	CertPrivateKey string `yaml:"private_key_pem"`
}

func NewCertificateValue(propertyName string) *CertificateValue {
	return &CertificateValue{
		CertPem:        fmt.Sprintf("((%s_%s))", propertyName, "certificate"),
		CertPrivateKey: fmt.Sprintf("((%s_%s))", propertyName, "privatekey"),
	}
}

func (s *CertificateValue) Parameters() []string {
	return []string{s.CertPem, s.CertPrivateKey}
}

func (s *CertificateValue) IsSelector() bool {
	return false
}

type CollectionsPropertiesValueHolder struct {
	Value []map[string]SimpleType `yaml:"value"`
}

func (s *CollectionsPropertiesValueHolder) Parameters() []string {
	var parameters []string
	for _, paramMap := range s.Value {
		for _, param := range paramMap {
			parameters = append(parameters, param.Parameters()...)
		}
	}

	return parameters
}

func (s *CollectionsPropertiesValueHolder) IsSelector() bool {
	return false
}
