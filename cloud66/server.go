package cloud66

import (
  "time"
)

type Server struct {
	Uid		          string       `json:"uid"`
	VendorUid		    string       `json:"vendor_uid"`
	Name		         string       `json:"name"`
	Address          string       `json:"address"`
	Distro           string       `json:"distro"`
	DistroVersion		string       `json:"distro_version"`
	DnsRecord		    string       `json:"dns_record"`
	UserName		     string       `json:"user_name"`
	ServerType		   string       `json:"server_type"`
	ServerGroupId		int          `json:"server_group_id"`
  Roles            []string     `json:"server_roles"`
	StackUid		     string       `json:"stack_uid"`
	HasAgent		     bool         `json:"has_agent"`
	Params		       interface{}  `json:"params"`
	CreatedAt		    time.Time    `json:"created_at_iso"`
	UpdatedAt		    time.Time    `json:"updated_at_iso"`
	Region		       string       `json:"region"`
	AvailabilityZone string  	   `json:"availability_zone"`
	ExtIpV4	        string       `json:"ext_ipv4"`
  HealthCode       int          `json:"health_state"`
}

func (s Server) Health() string {
  return healthStatus[s.HealthCode]
}

func (c *Client) ServerSshPrivateKey(uid string) (string, error) {
  req, err := c.NewRequest("GET", "/servers/" + uid + "/ssh_private_key.json", nil)
  if err != nil {
    return "", err
  }

  type Ssh struct {
    Ok    bool    `json:"ok"`
    Key   string  `json:"private_key"`
  }

  var sshRes *Ssh
  return sshRes.Key, c.DoReq(req, &sshRes)
}

func (c *Client) ServerSettings(uid string) ([]StackSetting, error) {
  req, err := c.NewRequest("GET", "/servers/" + uid + "/settings.json", nil)

  if err != nil {
    return nil, err
  }

  var settingsRes []StackSetting
  return settingsRes, c.DoReq(req, &settingsRes)
}

func (c *Client) ServerSet(uid string, key string, value string) (*GenericResponse, error) {
  params := struct {
		Key   string `json:"setting_name"`
    Value string `json:"setting_value"`
	}{
		Key:   key,
    Value: value,
	}

  req, err := c.NewRequest("POST", "/servers/" + uid + "/setting.json", params)
  if err != nil {
    return nil, err
  }

  var settingRes *GenericResponse
  return settingRes, c.DoReq(req, &settingRes)
}

