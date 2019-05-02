package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Profile holds a cx configuration profile which makes using it against multiple environments easier
type Profile struct {
	ApiURL       string `json:"api_url" yaml:"api_url"`
	BaseURL      string `json:"base_url" yaml:"base_url"`
	FayeEndpoint string `json:"faye_endpoint" yaml:"faye_endpoint"`
	ClientID     string `json:"client_id" yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
	Organization string `json:"organization" yaml:"organization"`
	Name         string `json:"name" yaml:"name"`
	TokenFile    string `json:"token_file" yaml:"token_file"`
}

type Profiles struct {
	LastProfile string              `json:"last_profile" yaml:"last_profile"`
	Profiles    map[string]*Profile `json:"profiles" yaml:"last_profile"`

	path string
}

func ReadProfiles(path string) (*Profiles, error) {
	var profiles *Profiles

	if _, err := os.Stat(path); err != nil {
		// no cxprofile.json. create the base one
		profile := defaultProfile()

		profiles = &Profiles{
			path:        path,
			LastProfile: "default",
			Profiles:    map[string]*Profile{"default": profile},
		}

		if err := profiles.WriteProfiles(); err != nil {
			return nil, err
		}
	}

	reader, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(reader, &profiles)
	if err != nil {
		return nil, err
	}

	profiles.path = path

	return profiles, nil
}

func (p *Profiles) WriteProfiles() error {
	writer, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p.path, writer, 0644)
}

func defaultProfile() *Profile {
	return &Profile{
		ApiURL:       "https://app.cloud66.com",
		ClientID:     "d4631fd51633bef0c04c6f946428a61fb9089abf4c1e13c15e9742cafd84a91f",
		ClientSecret: "e663473f7b991504eb561e208995de15550f499b6840299df588cebe981ba48e",
		Name:         "default",
		FayeEndpoint: "https://sockets.cloud66.com/push",
		TokenFile:    "cx.json",
		BaseURL:      "https://app.cloud66.com",
	}
}
