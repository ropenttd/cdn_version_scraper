package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
)

type VersionData struct {
	Version  string `yaml:"version"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
	Date     string `yaml:"date"`
	Folder   string `yaml:"folder"`
}
type LatestVersions struct {
	Versions []VersionData `yaml:"latest,flow"`
}

var versionStability = []string{
	// This map is sorted by order of stability, with the most stable lowest in the list.
	// This is so that stabilities are tested in order, least to most stable.
	"testing",
	"stable",
}

// getStabilities returns a slice of stabilities of the requested level or higher, in order.
// i.e passing "testing" will return a slice of "testing" and "stable"
func getStabilities(channel *string) []string {
	var validStabilities []string
	var desiredStabilityMet bool

	for _, v := range versionStability {
		if v == *channel || desiredStabilityMet {
			// setting desiredStabilityMet means that the channel check is ignored, putting everything in the output
			desiredStabilityMet = true
			validStabilities = append(validStabilities, v)
		}
	}
	return validStabilities
}

// findTargetBuildVersion takes a list of VersionData and the desired channel, and
// returns the VersionData of the desired stability level or higher.
func findTargetBuildVersion(allVersions []VersionData, channel *string) (ret VersionData, err error) {
	for _, stability := range getStabilities(channel) {
		// getStabilities returns us an array sorted in order of stability (least to most), so we can just range it
		for _, ver := range allVersions {
			if ver.Category == "openttd" && ver.Name == stability {
				// We found the target
				// we are, however, making an assumption here that there isn't more than one entry
				return ver, nil
			}
		}
	}
	// We couldn't find the target channel in the manifest :(
	return VersionData{}, errors.New("no valid OpenTTD versions found on CDN of desired stability level or higher")

}

// main
func main() {
	channel := os.Getenv("PLUGIN_CHANNEL")

	if channel == "" {
		if len(os.Args) != 2 {
			fmt.Println("Usage:", os.Args[0], versionStability)
			os.Exit(1)
		}
		channel = os.Args[1]
	}

	// Pull the manifest from cdn.openttd.org
	response, err := http.Get("https://cdn.openttd.org/latest.yaml")

	if err != nil {
		log.Fatalf("Failed to read CDN: %v", err)
	}
	defer response.Body.Close()

	// Unmarshal the yaml into our struct
	fingerData := LatestVersions{}
	decoder := yaml.NewDecoder(response.Body)
	err = decoder.Decode(&fingerData)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Sanity check
	if len(fingerData.Versions) == 0 {
		log.Fatal("error: no OpenTTD versions found on CDN - check CDN sanity")
	}

	// Find the appropriate target
	openttdVersion, err := findTargetBuildVersion(fingerData.Versions, &channel)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// This returns a value that can be parsed by bash or whatever shell you choose
	envString := fmt.Sprintf("OPENTTD_VERSION=\"%v\"", openttdVersion.Version)
	fmt.Printf(envString)

	outputFile := os.Getenv("PLUGIN_OUTPUT_FILE")
	if outputFile != "" {
		// Also write the version out to the given file
		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		defer file.Close()

		_, err = io.WriteString(file, envString)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		err = file.Sync()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}

}
