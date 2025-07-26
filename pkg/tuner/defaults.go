package tuner

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

var Defaults []ConfigSet = []ConfigSet{
	{
		minVersion: "v1.7",
		maxVersion: "v9.0",
		// older versions of Java preferred both initial and max RAM to be the same
		maxRamPercentage:     80.0,
		initialRamPercentage: 80.0,
		maxRamFlags: []string{
			"-Xmx=%.0fm",
		},
		initialRamFlags: []string{
			"-Xms=%.0fm",
		},
		opts: []string{
			"-XX:+AlwaysActAsServerClassMachine",     // Always use server JVM
			"-Dnetworkaddress.cache.ttl=10",          // DNS cache
			"-Dnetworkaddress.cache.negative.ttl=10", // Negative DNS cache
			"-XX:+UseStringDeduplication",            // Enable string deduplication
			"-Xshare:off",
		},
	},
	{
		minVersion: "v10.0",
		maxVersion: "v25.0",
		// Java 10+ prefers MaxRAMPercentage and InitialRAMPercentage
		// instead of -Xmx and -Xms
		maxRamPercentage:     70.0,
		initialRamPercentage: 25.0,
		maxRamFlags: []string{
			"-XX:MaxRAMPercentage=%.1f",
		},
		initialRamFlags: []string{
			"-XX:InitialRAMPercentage=%.1f",
			"-XX:MinRAMPercentage=%.1f",
		},
		opts: []string{
			"-XX:+AlwaysActAsServerClassMachine",     // Always use server JVM
			"-Dnetworkaddress.cache.ttl=10",          // DNS cache
			"-Dnetworkaddress.cache.negative.ttl=10", // Negative DNS cache
			"-XX:+UseStringDeduplication",            // Enable string deduplication
			"-Xshare:off",
		},
	},
}

type ConfigSet struct {
	minVersion           string
	maxVersion           string
	maxRamPercentage     float64
	initialRamPercentage float64
	maxRamFlags          []string
	initialRamFlags      []string
	opts                 []string
}

// JavaVersion parses the Java version from the output string.
// It would detect the version and return it in the SemVer format
// vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]]
func JavaVersion(versionOutput string) (string, error) {
	// The output may be in stderr or stdout, so combine both if needed
	lines := strings.Split(versionOutput, "\n")
	var versionLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "openjdk version") || strings.HasPrefix(line, "java version") {
			versionLine = line
			break
		}
	}
	if versionLine == "" {
		return "", fmt.Errorf("version of Java not found in output")
	}

	// Extract the quoted version string
	re := regexp.MustCompile(`"(.*?)"`)
	matches := re.FindStringSubmatch(versionLine)
	if len(matches) < 2 {
		return "", fmt.Errorf("version of Java not found in the string")
	}
	rawVersion := matches[1]

	// Convert to SemVer format
	// Remove any trailing build or patch info after first space
	parts := strings.Fields(rawVersion)
	semver := parts[0]

	// Parse trailing underscores as builds (e.g., 1.8.0_462 -> 1.8.0+462)
	semver = strings.ReplaceAll(semver, "_", "+")

	return semver, nil
}

func GetDefaults(javaVersion string) ConfigSet {
	for _, set := range Defaults {
		if (set.minVersion == "" || semver.Compare(javaVersion, set.minVersion) >= 0) &&
			(set.maxVersion == "" || semver.Compare(javaVersion, set.maxVersion) <= 0) {
			return set
		}
	}
	return ConfigSet{} // Return empty if no match found
}
