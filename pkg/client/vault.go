package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type (
	Secrets map[string]string

	// TODO: rename to response
	AuroraVaultInfo struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		Secrets     Secrets  `json:"secrets"`
		HasAccess   bool     `json:"hasAccess"`
	}

	// TODO: rename to request
	AuroraSecretVault struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		Secrets     Secrets  `json:"secrets"`
	}

	VaultFileResource struct {
		Contents string `json:"contents"`
	}
)

func NewAuroraSecretVault(name string) *AuroraSecretVault {
	return &AuroraSecretVault{
		Name:        name,
		Secrets:     make(Secrets),
		Permissions: []string{},
	}
}

func (api *ApiClient) GetVaults() ([]*AuroraVaultInfo, error) {
	endpoint := fmt.Sprintf("/vault/%s", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var vaults []*AuroraVaultInfo
	err = response.ParseItems(&vaults)
	if err != nil {
		return nil, err
	}

	return vaults, nil
}

func (api *ApiClient) GetVault(vaultName string) (*AuroraSecretVault, error) {
	endpoint := fmt.Sprintf("/vault/%s/%s", api.Affiliation, vaultName)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var vault AuroraSecretVault
	err = response.ParseFirstItem(&vault)
	if err != nil {
		return nil, err
	}

	return &vault, nil
}

func (api *ApiClient) DeleteVault(vaultName string) error {
	endpoint := fmt.Sprintf("/vault/%s/%s", api.Affiliation, vaultName)

	response, err := api.Do(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Message)
	}

	return nil
}

func (api *ApiClient) SaveVault(vault AuroraSecretVault) error {
	endpoint := fmt.Sprintf("/vault/%s", api.Affiliation)

	data, err := json.Marshal(vault)
	if err != nil {
		return err
	}

	response, err := api.Do(http.MethodPut, endpoint, data)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Message)
	}

	return nil
}

func (api *ApiClient) GetSecretFile(vault, secret string) (string, string, error) {
	endpoint := fmt.Sprintf("/vault/%s/%s/%s", api.Affiliation, vault, secret)

	bundle, err := api.DoWithHeader(http.MethodGet, endpoint, nil, nil)
	if err != nil || bundle == nil {
		return "", "", err
	}

	if !bundle.BooberResponse.Success {
		return "", "", errors.New(bundle.BooberResponse.Message)
	}

	var vaultFile VaultFileResource
	err = bundle.BooberResponse.ParseFirstItem(&vaultFile)
	if err != nil {
		return "", "", nil
	}

	data, err := base64.StdEncoding.DecodeString(vaultFile.Contents)
	if err != nil {
		return "", "", err
	}

	eTag := bundle.HttpResponse.Header.Get("ETag")

	return string(data), eTag, nil
}

func (api *ApiClient) UpdateSecretFile(vault, secret, eTag string, content []byte) error {
	endpoint := fmt.Sprintf("/vault/%s/%s/%s", api.Affiliation, vault, secret)

	encoded := base64.StdEncoding.EncodeToString(content)

	header := map[string]string{
		"If-Match": eTag,
	}

	vaultFile := VaultFileResource{
		Contents: encoded,
	}

	data, err := json.Marshal(vaultFile)
	if err != nil {
		return err
	}

	bundle, err := api.DoWithHeader(http.MethodPut, endpoint, header, data)
	if err != nil || bundle == nil {
		return err
	}

	if !bundle.BooberResponse.Success {
		return errors.New(bundle.BooberResponse.Message)
	}

	return nil
}

func (s Secrets) GetSecret(name string) (string, error) {
	secret, found := s[name]
	if !found {
		return "", errors.Errorf("Did not find secret %s", name)
	}
	data, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s Secrets) AddSecret(name, content string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	s[name] = encoded
}

func (s Secrets) RemoveSecret(name string) {
	delete(s, name)
}
