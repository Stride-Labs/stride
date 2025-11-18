package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/Stride-Labs/stride/v30/utils"
)

/*
	This file is intended to replace defaults in config.toml and app.toml
	by overwriting the values in the respective files.

	The actions of this file can be disabled with the --reject-config-defaults flag.

	This file is taken almost entirely from the Osmosis repo, with minor modifications.
	Their original file can be found here:
	https://github.com/osmosis-labs/osmosis/blob/e5895ce4a460a585c0afb29873de9c7de826b690/cmd/osmosisd/cmd/root.go#L1
	Thank you to the excellent engineers at Osmosis.
*/

const (
	FlagRejectConfigDefaults = "reject-config-defaults"
)

type SectionKeyValue struct {
	Section string
	Key     string
	Value   any
}

var (
	recommendedAppTomlValues = []SectionKeyValue{
		{
			Section: "",
			Key:     "minimum-gas-prices",
			Value:   "0.0005ustrd,0.001stuosmo,0.0001stuatom,20000000000staevmos,500000000stinj,0.01stutia,15000000000stadydx,15000000000stadym,0.01stusaga,0.001ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34,0.0001ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2,0.01ibc/BF3B4F53F3694B66E13C23107C84B6485BD2B96296BB7EC680EA77BBA75B4801,20000000000ibc/4B322204B4F59D770680FE4D7A565DDC3F37BFF035474B717476C66A4F83DD72,500000000ibc/A7454562FF29FE068F42F9DE4805ABEF54F599D1720B345D6518D9B5C64EA6D2,15000000000ibc/561C70B20188A047BFDE6F9946BDDC5D8AC172B9BE04FF868DFABF819E5A9CCE,15000000000ibc/E1C22332C083574F3418481359733BA8887D171E76C80AD9237422AEABB66018,0.01ibc/520D9C4509027DE66C737A1D6A6021915A3071E30DBA8F758B46532B060D7AA5",
		},
	}

	recommendedConfigTomlValues = []SectionKeyValue{
		{
			Section: "consensus",
			Key:     "timeout_commit",
			Value:   "1s",
		},
	}
)

// overwriteConfigTomlValues overwrites config.toml values. Returns error if config.toml does not exist
//
// Currently, overwrites:
// - timeout_commit
//
// Also overwrites the respective viper config value.
//
// Silently handles and skips any error/panic due to write permission issues.
// No-op otherwise.
func overwriteConfigTomlValues(serverCtx *server.Context) error {
	// Get paths to config.toml and config parent directory
	rootDir := serverCtx.Viper.GetString(tmcli.HomeFlag)

	configParentDirPath := filepath.Join(rootDir, "config")
	configFilePath := filepath.Join(configParentDirPath, "config.toml")

	fileInfo, err := os.Stat(configFilePath)
	if err != nil {
		// something besides a does not exist error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read in %s: %w", configFilePath, err)
		}
	} else {
		// config.toml exists

		// Check if each key is already set to the recommended value
		// If it is, we don't need to overwrite it and can also skip the app.toml overwrite
		var sectionKeyValuesToWrite []SectionKeyValue

		// Set aside which keys need to be updated in the config.toml
		for _, rec := range recommendedConfigTomlValues {
			currentValue := serverCtx.Viper.Get(rec.Section + "." + rec.Key)
			if currentValue != rec.Value {
				// Current value in config.toml is not the recommended value
				// Set the value in viper to the recommended value
				// and add it to the list of key values we will overwrite in the config.toml
				serverCtx.Viper.Set(rec.Section+"."+rec.Key, rec.Value)
				sectionKeyValuesToWrite = append(sectionKeyValuesToWrite, rec)
			}
		}

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("failed to write to %s: %s\n", configFilePath, err)
			}
		}()

		// Check if the file is writable
		if fileInfo.Mode()&os.FileMode(0200) != 0 {
			// It will be re-read in server.InterceptConfigsPreRunHandler
			// this may panic for permissions issues. So we catch the panic.
			// Note that this exits with a non-zero exit code if fails to write the file.

			// Write the new config.toml file
			if len(sectionKeyValuesToWrite) > 0 {
				err := OverwriteWithCustomConfig(configFilePath, sectionKeyValuesToWrite)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("config.toml is not writable. Cannot apply update. Please consider manually changing to the following: %v\n", recommendedConfigTomlValues)
		}
	}
	return nil
}

// overwriteAppTomlValues overwrites app.toml values. Returns error if app.toml does not exist
//
// Currently, overwrites:
// - minimum-gas-prices
//
// Also overwrites the respective viper config value.
//
// Silently handles and skips any error/panic due to write permission issues.
// No-op otherwise.
func overwriteAppTomlValues(serverCtx *server.Context) error {
	// Get paths to app.toml and config parent directory
	rootDir := serverCtx.Viper.GetString(tmcli.HomeFlag)

	configParentDirPath := filepath.Join(rootDir, "config")
	appFilePath := filepath.Join(configParentDirPath, "app.toml")

	fileInfo, err := os.Stat(appFilePath)
	if err != nil {
		// something besides a does not exist error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read in %s: %w", appFilePath, err)
		}
	} else {
		// app.toml exists

		// Check if each key is already set to the recommended value
		// If it is, we don't need to overwrite it and can also skip the app.toml overwrite
		var sectionKeyValuesToWrite []SectionKeyValue

		for _, rec := range recommendedAppTomlValues {
			currentValue := serverCtx.Viper.Get(rec.Section + "." + rec.Key)
			if currentValue != rec.Value {
				// Current value in app.toml is not the recommended value
				// Set the value in viper to the recommended value
				// and add it to the list of key values we will overwrite in the app.toml
				serverCtx.Viper.Set(rec.Section+"."+rec.Key, rec.Value)
				sectionKeyValuesToWrite = append(sectionKeyValuesToWrite, rec)
			}
		}

		// Check if the file is writable
		if fileInfo.Mode()&os.FileMode(0200) != 0 {
			// It will be re-read in server.InterceptConfigsPreRunHandler
			// this may panic for permissions issues. So we catch the panic.
			// Note that this exits with a non-zero exit code if fails to write the file.

			// Write the new app.toml file
			if len(sectionKeyValuesToWrite) > 0 {
				err := OverwriteWithCustomConfig(appFilePath, sectionKeyValuesToWrite)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("app.toml is not writable. Cannot apply update. Please consider manually changing to the following: %v\n", recommendedAppTomlValues)
		}
	}
	return nil
}

// OverwriteWithCustomConfig searches the respective config file for the given section and key and overwrites the current value with the given value.
func OverwriteWithCustomConfig(configFilePath string, sectionKeyValues []SectionKeyValue) error {
	// Open the file for reading and writing
	file, err := os.OpenFile(configFilePath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a map from the sectionKeyValues array
	// This map will be used to quickly look up the new values for each section and key
	configMap := make(map[string]map[string]string)
	for _, skv := range sectionKeyValues {
		// If the section does not exist in the map, create it
		if _, ok := configMap[skv.Section]; !ok {
			configMap[skv.Section] = make(map[string]string)
		}
		// Add the key and value to the section in the map
		// If the value is a string, add quotes around it
		switch v := skv.Value.(type) {
		case string:
			configMap[skv.Section][skv.Key] = "\"" + v + "\""
		default:
			configMap[skv.Section][skv.Key] = fmt.Sprintf("%v", v)
		}
	}

	// Read the file line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		// If the line is a section header, update the current section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
		} else if configMap[currentSection] != nil {
			// If the line is in a section that needs to be overwritten, check each key
			for _, key := range utils.StringMapKeys(configMap[currentSection]) {
				value := configMap[currentSection][key]
				// Split the line into key and value parts
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				// Trim spaces and compare the key part with the target key
				if strings.TrimSpace(parts[0]) == key {
					// If the keys match, overwrite the line with the new key-value pair
					line = key + " = " + value
					break
				}
			}
		}
		// Add the line to the lines slice, whether it was overwritten or not
		lines = append(lines, line)
	}

	// Check for errors from the scanner
	if err := scanner.Err(); err != nil {
		return err
	}

	// Seek to the beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Truncate the file to remove the old content
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	// Write the new lines to the file
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}
