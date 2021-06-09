package googlecal

import (
	"encoding/json"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

//DomainName name of domain
type DomainName string

//CalendarConfig config for calendar. Maps domains to configs
type CalendarConfig map[DomainName]*jwt.Config

//revive:disable:var-naming

//UnmarshalJSON -
func (c *CalendarConfig) UnmarshalJSON(b []byte) (err error) {
	//Parse config structure and convert
	//credentials section to an actual JWT config

	//credentials structure
	type Temp map[string]struct {
		Scopes                      []string `json:",omitempty"`
		Type                        string
		Project_id                  string
		Private_key_id              string
		Private_key                 string
		Client_email                string
		Client_id                   string
		Auth_uri                    string
		Token_uri                   string
		Auth_provider_x509_cert_url string
		Client_x509_cert_url        string
	}

	temp := Temp{}
	json.Unmarshal(b, &temp)

	//initialize config and get jwtconfig for each credential
	*c = CalendarConfig{}
	for account, v := range temp {
		scopes := v.Scopes
		v.Scopes = []string{}

		data, err := json.Marshal(v)
		if err != nil {
			c = nil
			return err
		}

		config, err := google.JWTConfigFromJSON(data, scopes...)
		config.Subject = "REPLACEME"
		if err != nil {
			c = nil
			return err
		}
		(*c)[DomainName(account)] = config

	}
	return nil
}

//revive:enable:var-naming

//ParseConfig parses config
func ParseConfig(b []byte) (config CalendarConfig, err error) {
	err = json.Unmarshal(b, &config)
	return
}
