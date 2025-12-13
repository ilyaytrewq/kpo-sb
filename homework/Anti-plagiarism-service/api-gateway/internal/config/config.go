package config

import (
	"fmt"
	"os"
	"strconv"
)


type FileStoringConfig struct {
	Host string
	Port int
	Url string
}

func LoadFileStoringConfig() (*FileStoringConfig, error) {
	host := os.Getenv("FILE_STORING_HOST")
	if host == "" {
		return nil, fmt.Errorf("FILE_STORING_HOST environment variable is not set")
	}

	portStr := os.Getenv("FILE_STORING_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("FILE_STORING_PORT environment variable is not set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &FileStoringConfig{
		Host: host,
		Port: port,
		Url: fmt.Sprintf("http://%s:%d", host, port),
	}, nil
}

type FileAnalysisConfig struct {
	Host string
	Port int
	Url string
}

func LoadFileAnalysisConfig() (*FileAnalysisConfig, error) {
	host := os.Getenv("FILE_ANALYSIS_HOST")
	if host == "" {
		return nil, fmt.Errorf("FILE_ANALYSIS_HOST environment variable is not set")
	}

	portStr := os.Getenv("FILE_ANALYSIS_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("FILE_ANALYSIS_PORT environment variable is not set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &FileAnalysisConfig{
		Host: host,
		Port: port,
		Url: fmt.Sprintf("http://%s:%d", host, port),
	}, nil
}