package main

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	yaml "gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	"master",
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

// decodeGitReferenceVersionString decodes a version string like 20200412-master-g4bfa3ff619
// and returns its git hash.
func decodeGitReferenceVersionString(versionData *VersionData) (versionOrGitCommit string) {
	verArray := strings.Split(versionData.Version, "-")
	if len(verArray) != 3 {
		// this isn't a valid version string
		return versionData.Version
	}
	return verArray[2]
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

	var envVars map[string]string
	ver := decodeGitReferenceVersionString(&openttdVersion)
	sver, err := semver.NewVersion(ver)

	// this looks kind of a dumb way of doing this, but it saves doing a map merge
	// could also just be lazy and do two slices of string(k=v)?
	if err != nil {
		// Problem decoding version, just provide version string.
		envVars = map[string]string{
			"version": ver,
		}
	} else {
		envVars = map[string]string{
			"version":      ver,
			"semver_major": strconv.Itoa(int(sver.Major())),
			"semver_minor": strconv.Itoa(int(sver.Minor())),
			"semver_patch": strconv.Itoa(int(sver.Patch())),
		}
	}

	// Try to find an appropriate output file
	var outputFile string
	outputFile, ok := os.LookupEnv("PLUGIN_OUTPUT_FILE")
	if !ok {
		actionMode := os.Getenv("GITHUB_ACTIONS")
		if actionMode == "true" {
			// We're running as a Github Action.
			outputFile, ok = os.LookupEnv("GITHUB_OUTPUT")
			if !ok {
				fmt.Print("::error::No Github Actions Environment file value")
			}
		}
	}
	if outputFile != "" {
		// Also write the version out to the given file
		file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		defer file.Close()

		for k, v := range envVars {
			_, err = io.WriteString(file, fmt.Sprintf("%s=%s\n", k, v))
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		}
		err = file.Sync()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	} else {
		for k, v := range envVars {
			fmt.Printf("%s=%s\n", k, v)
		}
	}

}
