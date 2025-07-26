package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tgagor/java-tuner/pkg/tuner"
)

func TestJavaVersion_1_8(t *testing.T) {
	output := `openjdk version "1.8.0_462"
OpenJDK Runtime Environment Corretto-8.462.08.1 (build 1.8.0_462-b08)
OpenJDK 64-Bit Server VM Corretto-8.462.08.1 (build 25.462-b08, mixed mode)`
	version, err := tuner.JavaVersion(output)
	assert.NoError(t, err)
	assert.Equal(t, "1.8.0+462", version)
}

func TestJavaVersion_11(t *testing.T) {
	output := `openjdk version "11.0.28" 2025-07-15 LTS
OpenJDK Runtime Environment Corretto-11.0.28.6.1 (build 11.0.28+6-LTS)
OpenJDK 64-Bit Server VM Corretto-11.0.28.6.1 (build 11.0.28+6-LTS, mixed mode)`
	version, err := tuner.JavaVersion(output)
	assert.NoError(t, err)
	assert.Equal(t, "11.0.28", version)
}

func TestJavaVersion_17(t *testing.T) {
	output := `openjdk version "17.0.16" 2025-07-15 LTS
OpenJDK Runtime Environment Corretto-17.0.16.8.1 (build 17.0.16+8-LTS)
OpenJDK 64-Bit Server VM Corretto-17.0.16.8.1 (build 17.0.16+8-LTS, mixed mode, sharing)`
	version, err := tuner.JavaVersion(output)
	assert.NoError(t, err)
	assert.Equal(t, "17.0.16", version)
}

func TestJavaVersion_21(t *testing.T) {
	output := `openjdk version "21.0.8" 2025-07-15 LTS
OpenJDK Runtime Environment Corretto-21.0.8.9.1 (build 21.0.8+9-LTS)
OpenJDK 64-Bit Server VM Corretto-21.0.8.9.1 (build 21.0.8+9-LTS, mixed mode, sharing)`
	version, err := tuner.JavaVersion(output)
	assert.NoError(t, err)
	assert.Equal(t, "21.0.8", version)
}

func TestJavaVersion_NotFound(t *testing.T) {
	output := `OpenJDK Runtime Environment Corretto-21.0.8.9.1 (build 21.0.8+9-LTS)`
	version, err := tuner.JavaVersion(output)
	assert.Error(t, err)
	assert.Empty(t, version)
}
