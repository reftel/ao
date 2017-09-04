package auroraconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

const InvalidConfigurationError = "Invalid configuration"

func GetContent(filename string, configuration *configuration.ConfigurationClass) (content string, version string, err error) {
	auroraConfig, err := GetAuroraConfig(configuration)
	if err != nil {
		return
	}
	var fileFound bool = false

	_, fileFound = auroraConfig.Files[filename]
	if fileFound {
		content = string(auroraConfig.Files[filename])
	}

	version = auroraConfig.Versions[filename]

	if !fileFound {
		return "", "", errors.New("Illegal file/folder")
	}
	return content, version, nil

}

func GetAllContent(outputFolder string, configuration *configuration.ConfigurationClass) (output string, err error) {
	auroraConfig, err := GetAuroraConfig(configuration)
	if err != nil {
		return
	}

	if outputFolder != "" {
		if fileutil.IsLegalFileFolder(outputFolder) == fileutil.SpecIllegal {
			err = errors.New("Illegal file/folder")
			return "", err

		}
		var content string
		for filename := range auroraConfig.Files {
			content = jsonutil.PrettyPrintJson(string(auroraConfig.Files[filename]))
			err = fileutil.WriteFile(outputFolder, filename, content)
			if err != nil {
				return "", err
			}
		}
	}

	outputBytes, err := json.Marshal(auroraConfig)
	output = jsonutil.PrettyPrintJson(string(outputBytes))
	return output, err

}

func GetFileList(configuration *configuration.ConfigurationClass) (filenames []string, err error) {
	auroraConfig, err := GetAuroraConfig(configuration)
	if err != nil {
		return
	}
	filenames = make([]string, len(auroraConfig.Files))

	var filenameIndex = 0
	for filename := range auroraConfig.Files {
		filenames[filenameIndex] = filename
		filenameIndex++
	}
	return filenames, nil
}

func GetAuroraConfig(configuration *configuration.ConfigurationClass) (auroraConfig serverapi.AuroraConfig, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/auroraconfig"

	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.OpenshiftConfig, configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		if !response.Success {
			output, err := serverapi.ResponsItems2MessageString(response)
			if err != nil {
				return auroraConfig, err
			}
			err = errors.New(output)
			return auroraConfig, err

		}

	}

	auroraConfig, err = serverapi.ResponseItems2AuroraConfig(response)

	return auroraConfig, nil
}

func PutAuroraConfig(auroraConfig serverapi.AuroraConfig, configuration *configuration.ConfigurationClass) (err error) {
	content, err := json.Marshal(auroraConfig)
	if err != nil {
		return err
	}

	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/auroraconfig"

	_, err = putContent(apiEndpoint, string(content), "", configuration)
	if err != nil {
		return err
	}
	return
}

func putContent(apiEndpoint string, content string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {

	var versionHeader = make(map[string]string)
	versionHeader["AuroraConfigFileVersion"] = version

	response, err := serverapi.CallApiWithHeaders(versionHeader, http.MethodPut, apiEndpoint, content, true,
		configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose,
		configuration.OpenshiftConfig, configuration.GetPersistentOptions().DryRun, configuration.GetPersistentOptions().Debug,
		configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)

	if err != nil {
		if !response.Success {
			validationMessages, _ := serverapi.ResponsItems2MessageString(response)
			return validationMessages, errors.New(InvalidConfigurationError)
		}
	}

	return
}

func PutFile(filename string, content string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/auroraconfigfile/" + filename

	return putContent(apiEndpoint, content, version, configuration)
}

func PutSecret(vaultname string, secretname string, secret string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/" + vaultname + "/secret/" + secretname

	encodedSecret := base64.StdEncoding.EncodeToString([]byte(secret))
	return putContent(apiEndpoint, encodedSecret, version, configuration)
}

func PutVault(vaultname string, vault serverapi.Vault, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/"

	content, err := json.Marshal(vault)

	return putContent(apiEndpoint, string(content), version, configuration)

}

func deleteContent(apiEndpoint string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {

	var versionHeader = make(map[string]string)
	versionHeader["AuroraConfigFileVersion"] = version

	response, err := serverapi.CallApiWithHeaders(versionHeader, http.MethodDelete, apiEndpoint, "", true,
		configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose,
		configuration.OpenshiftConfig, configuration.GetPersistentOptions().DryRun, configuration.GetPersistentOptions().Debug,
		configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)

	if err != nil {
		if !response.Success {
			validationMessages, _ := serverapi.ResponsItems2MessageString(response)
			return validationMessages, errors.New(validationMessages)
		}
	}

	return
}

func DeleteVault(vaultname string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/" + vaultname

	return deleteContent(apiEndpoint, "", configuration)

}
