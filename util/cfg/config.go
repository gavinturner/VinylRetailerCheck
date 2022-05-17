package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
)

var adHocSettingsCacheGlobal = map[string]string{}
var initConfigOnce sync.Once

func InitConfig() {
	initConfigOnce.Do(func() {
		err := initAdhocConfig()
		if err != nil {
			fmt.Printf("Error in initAdhocConfig: %v\n", err)
		}
	})
}

// findDefaultAppConfigPath
//   finds the default location of a possible `application.json`
//   based on the gopath and repo name that the code's currently running in
//   only called in dev and test and never called in production, so expects a dev environment
func findDefaultAppConfigPath() (string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", fmt.Errorf("GOPATH not set")
	}
	if len(os.Args) == 0 {
		return "", fmt.Errorf("Could not determine binary name")
	}
	binaryName := os.Args[0]

	if re := regexp.MustCompile(`github[.]com\/ScriptRock\/([^\/]+)\/`).FindStringSubmatch(binaryName); re != nil {
		binaryName = re[1]
	}
	binaryName = filepath.Base(binaryName)

	possibleAppConfigPath := filepath.Join(goPath, "src/github.com/ScriptRock", binaryName, "cfg/application.json")
	if _, err := os.Stat(possibleAppConfigPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Could not locate default config file '%v'", possibleAppConfigPath)
	}
	return possibleAppConfigPath, nil
}

// initAdhocConfig pre-populates config settings found in an existing ad-hoc configuration file.
// The relative/absolute path to the configuration file is identified by env_var CONFIG_PATH
// The contents of this file are assumed to be a single layer json object with name value pairs, each being a string.
func initAdhocConfig() (err error) {

	// clear out the cache so it can be re-filled
	adHocSettingsCacheGlobal = map[string]string{}

	// determine the location of the configuration file if it exists and read it.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath, err = findDefaultAppConfigPath()
		if err != nil {
			configPath = "./cfg/application.json"
		}
	}
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		// no config file found. return an error
		return fmt.Errorf("Config file could not be found at '%v'", configPath)
	}

	// unmarshal the json config file contents. should be name-value pairs
	config := map[string]interface{}{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return fmt.Errorf("Failed to process contents of '%v'. Bad JSON format: '%v'", configPath, err.Error())
	}

	// get any and all settings as strings into the ad-hoc cache
	for key, value := range config {
		adHocSettingsCacheGlobal[key] = fmt.Sprintf("%v", value)
	}

	return err
}

// SetConfigValue manually sets the value of a config entry in the adhoc config table
func SetConfigValue(name string, value string) (err error) {
	if name == "" {
		return fmt.Errorf("Cannot set a config entry with empty name")
	}
	adHocSettingsCacheGlobal[name] = value
	return nil
}

// StringSetting retrieves an ad-hoc string setting by name. Settings may come from a companion config file
// or from an env_var. Env_var's entries take precendence over identically named values from the config file.
// @see initAdhocConfig()
func StringSetting(name string) (value string, err error) {

	// determine if an ENV_VAR value exists
	value = os.Getenv(name)
	if value != "" {
		return value, nil
	}
	value = adHocSettingsCacheGlobal[name]
	if value == "" {
		err = fmt.Errorf("No configuration item '%v' found in settings", name)
	}
	return value, err
}

// IntSetting retrieves an ad-hoc int setting by name. Settings may come from a companion config file
// or from an env_var. Env_var's entries take precendence over identically named values from the config file.
// All settings int he config file are assumed to be strings initially. This method automatically performs
// the integer (base 10) conversion before returning.
// @see initAdhocConfig()
func IntSetting(name string) (intValue int, err error) {

	// determine if an ENV_VAR value exists
	value, err := StringSetting(name)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, fmt.Errorf("No configuration item '%v' found in settings", name)
	}
	return strconv.Atoi(value)
}
