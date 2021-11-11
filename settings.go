package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
)

// Settings Program settings definition
type Settings struct {
	ServerHost string
	ServerPort int
	TargetUrl  *url.URL
	ReplaceUrl *url.URL
}

// Load settings from the environment variables
func (s *Settings) Load() {
	// required target url
	targetUrl := os.Getenv("TARGET_URL")
	if targetUrl == "" {
		log.Fatal("TARGET_URL is required but missing in the env variables")
	}
	url, err := url.Parse(targetUrl)
	if err != nil {
		log.Fatal("Cannot parse TARGET_URL as valid url")
	}
	s.TargetUrl = url

	// required replace url
	replaceUrl := os.Getenv("REPLACE_URL")
	if replaceUrl == "" {
		log.Fatal("REPLACE_URL is required but missing in the env variables")
	}
	url, err = url.Parse(replaceUrl)
	if err != nil {
		log.Fatal("Cannot parse REPLACE_URL as valid url")
	}
	s.ReplaceUrl = url

	// optional params
	host := os.Getenv("HOST")
	if host != "" {
		s.ServerHost = host
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err == nil {
		s.ServerPort = port
	}
}
