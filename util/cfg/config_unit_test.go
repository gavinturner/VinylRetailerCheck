//go:build unit_test
// +build unit_test

package cfg

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigInit(t *testing.T) {

	// not parallel as adhocConfig map is a shared resource
	InitConfig()
}

func TestAdHocSettings_EnvVarOnly(t *testing.T) {

	// not parallel as adhocConfig map is a shared resource
	os.Setenv("VALUE_1", "myvalue")
	os.Setenv("VALUE_2", "32")
	defer func() {
		os.Setenv("VALUE_1", "")
		os.Setenv("VALUE_2", "")
	}()

	initAdhocConfig()

	value1, err := StringSetting("VALUE_1")
	assert.Equal(t, "myvalue", value1, "Failed to retrieve ad-hoc config from env var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string env_var config")

	value2, err := StringSetting("VALUE_2")
	assert.Equal(t, "32", value2, "Failed to retrieve ad-hoc config from env var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string value from env_var config")

	value3, err := IntSetting("VALUE_2")
	assert.Equal(t, 32, value3, "Failed to retrieve ad-hoc integer value from env var config")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string env_var config")

	_, err = IntSetting("VALUE_1")
	assert.NotNil(t, err, "Expected error retrieving int value from invalid string representation")

	_, err = IntSetting("MISSING")
	assert.NotNil(t, err, "Expected error retrieving missing config value")
}

func TestAdHocSettings_ConfigFile(t *testing.T) {
	// not parallel as adhocConfig map is a shared resource

	filename := "/tmp/config1.json"
	os.Setenv("CONFIG_PATH", filename)
	d1 := []byte(`{ "VALUE_A": "myvalue", "VALUE_B": "32", "VALUE_C": 24 }`)
	err := ioutil.WriteFile(filename, d1, 0644)
	assert.Nil(t, err, "Failed to write test json file")

	// cleanup the file when we are done.
	defer func() {
		_ = os.Remove(filename)
	}()

	initAdhocConfig()

	value1, err := StringSetting("VALUE_A")
	assert.Equal(t, "myvalue", value1, "Failed to retrieve ad-hoc config from file")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	value2, err := StringSetting("VALUE_B")
	assert.Equal(t, "32", value2, "Failed to retrieve ad-hoc config from file")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string value from file config")

	value3, err := IntSetting("VALUE_B")
	assert.Equal(t, 32, value3, "Failed to retrieve ad-hoc integer value from file config")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	value4, err := IntSetting("VALUE_C")
	assert.Equal(t, 24, value4, "Failed to retrieve ad-hoc integer value from file config")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	_, err = IntSetting("VALUE_A")
	assert.NotNil(t, err, "Expected error retrieving int value from invalid string representation")

	_, err = IntSetting("MISSING")
	assert.NotNil(t, err, "Expected error retrieving missing config value")
}

func TestAdHocSettings_ConfigFileAndEnv(t *testing.T) {
	// not parallel as adhocConfig map is a shared resource

	filename := "/tmp/config2.json"
	os.Setenv("CONFIG_PATH", filename)
	d1 := []byte(`{ "VALUE_X": "myvalue", "VALUE_Z": "something" }`)
	err := ioutil.WriteFile(filename, d1, 0644)
	assert.Nil(t, err, "Failed to write test json file")

	os.Setenv("VALUE_Y", "32")
	os.Setenv("VALUE_Z", "something else")

	// cleanup the file when we are done.
	defer func() {
		_ = os.Remove(filename)
		os.Setenv("VALUE_Y", "")
		os.Setenv("VALUE_Z", "")
	}()

	initAdhocConfig()

	value1, err := StringSetting("VALUE_X")
	assert.Equal(t, "myvalue", value1, "Failed to retrieve ad-hoc config from file var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	value2, err := StringSetting("VALUE_Y")
	assert.Equal(t, "32", value2, "Failed to retrieve ad-hoc config from env var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string value from env_var config")

	value3, err := StringSetting("VALUE_Z")
	assert.Equal(t, "something else", value3, "Failed to retrieve ad-hoc config from overriding env var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string value from env_var overriding file config")
}

func TestAdHocSettings_Override(t *testing.T) {

	// not parallel as adhocConfig map is a shared resource
	initAdhocConfig()

	err := SetConfigValue("", "VALUE")
	assert.NotNil(t, err, "Expected error manually setting adhoc config value with no name")

	err = SetConfigValue("VALUE_X", "VALUE")
	assert.Nil(t, err, "Unexpected error manually setting adhoc config value")
	value1, err := StringSetting("VALUE_X")
	assert.Equal(t, "VALUE", value1, "Failed to retrieve ad-hoc config from file var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	err = SetConfigValue("VALUE_X", "VALUE2")
	assert.Nil(t, err, "Unexpected error manually setting adhoc config value")
	value1, err = StringSetting("VALUE_X")
	assert.Equal(t, "VALUE2", value1, "Failed to retrieve ad-hoc config from file var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	err = SetConfigValue("VALUE_X", "")
	assert.Nil(t, err, "Unexpected error manually setting adhoc config value")
	_, err = StringSetting("VALUE_X")
	assert.NotNil(t, err, "Exected error retrieving removed ad-hoc string file config")

	err = SetConfigValue("VALUE_X", "32")
	assert.Nil(t, err, "Unexpected error manually setting adhoc config value")
	value3, err := IntSetting("VALUE_X")
	assert.Equal(t, 32, value3, "Failed to retrieve ad-hoc config from file var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	err = SetConfigValue("VALUE_X", "33")
	assert.Nil(t, err, "Unexpected error manually setting adhoc config value")
	value3, err = IntSetting("VALUE_X")
	assert.Equal(t, 33, value3, "Failed to retrieve ad-hoc config from file var")
	assert.Nil(t, err, "Unexected error retrieving ad-hoc string file config")

	err = SetConfigValue("VALUE_X", "")
	assert.Nil(t, err, "Unexpected error manually setting adhoc config value")
	_, err = IntSetting("VALUE_X")
	assert.NotNil(t, err, "Exected error retrieving removed ad-hoc string file config")
}

