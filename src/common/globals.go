package main

// Global configuration variables
var (
	BaseClusterName string
	NoClusters      int
	AccessToken     string
	RumToken        string
	ApiToken        string
	InputDir        string
	InputFile       string
	OutputDir       string
)

// Struct for parsing config.yaml
type Config struct {
	BaseClusterName string `yaml:"base_cluster_name"`
	NoClusters      int    `yaml:"no_clusters"`
	AccessToken     string `yaml:"access_token"`
	RumToken        string `yaml:"rum_token"`
	ApiToken        string `yaml:"api_token"`
	InputDir        string `yaml:"input_dir"`
	InputFile       string `yaml:"input_file"`
	OutputDir       string `yaml:"output_dir"`
}
