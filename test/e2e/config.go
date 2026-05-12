package e2e

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	awsinternal "github.com/aws/eks-hybrid/internal/aws"
	"github.com/aws/eks-hybrid/test/e2e/constants"
)

type TestConfig struct {
	ClusterName     string `yaml:"clusterName"`
	ClusterRegion   string `yaml:"clusterRegion"`
	NodeadmUrlAMD   string `yaml:"nodeadmUrlAMD"`
	NodeadmUrlARM   string `yaml:"nodeadmUrlARM"`
	SetRootPassword bool   `yaml:"setRootPassword"`
	NodeK8sVersion  string `yaml:"nodeK8SVersion"`
	LogsBucket      string `yaml:"logsBucket"`
	Endpoint        string `yaml:"endpoint"`
	// ArtifactsFolder is the local path where the test will store the artifacts.
	ArtifactsFolder string `yaml:"artifactsFolder"`
	DNSSuffix       string `yaml:"dnsSuffix"`
	EcrAccount      string `yaml:"ecrAccount"`
	ManifestURL     string `yaml:"manifestUrl"`
}

// ReadConfig reads the configuration from the specified file path and unmarshals it into the TestConfig struct.
func ReadConfig(configPath string) (*TestConfig, error) {
	config := &TestConfig{}
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading tests configuration file %s: %w", configPath, err)
	}

	if err = yaml.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("unmarshaling test configuration: %w", err)
	}

	if config.ArtifactsFolder == "" {
		config.ArtifactsFolder = "/tmp"
	}

	if config.DNSSuffix == "" {
		// Auto-detect DNS suffix from region
		partition := awsinternal.GetPartitionFromRegionFallback(config.ClusterRegion)
		config.DNSSuffix = awsinternal.GetPartitionDNSSuffix(partition)
	}

	if config.EcrAccount == "" {
		// Auto-detect ECR account based on region
		config.EcrAccount = GetEcrAccountForRegion(config.ClusterRegion)
	}

	return config, nil
}

// GetEcrAccountForRegion returns the appropriate ECR account ID for the given region.
// For China regions (aws-cn partition), it uses the China-specific ECR account (constants.ChinaEcrAccountId).
// For all other regions, it returns the standard test ECR account (constants.EcrAccountId).
func GetEcrAccountForRegion(region string) string {
	if awsinternal.GetPartitionFromRegionFallback(region) == "aws-cn" {
		return constants.ChinaEcrAccountId
	}
	return constants.EcrAccountId
}
