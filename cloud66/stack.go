package cloud66

import (
  "time"
  "strings"
  "errors"
)

var stackStatus = map[int]string {
  0: "Pending analysis",        //STK_QUEUED
  1: "Deployed successfully",   //STK_SUCCESS
  2: "Deployment failed",       //STK_FAILED
  3: "Analyzing",               //STK_ANALYSING
  4: "Analyzed",                //STK_ANALYSED
  5: "Queued for deployment",   //STK_QUEUED_FOR_DEPLOYING
  6: "Deploying",               //STK_DEPLOYING
  7: "Unable to analyze",       //STK_TERMINAL_FAILURE
}

var healthStatus = map[int]string {
  0: "Unknown",     //HLT_UNKNOWN
  1: "Building",    //HLT_BUILDING
  2: "Impaired",    //HLT_PARTIAL
  3: "Healthy",     //HLT_OK
  4: "Failed",      //HLT_BROKEN
}

type Stack struct {
  Uid             string     `json:"uid"`
  Name            string     `json:"name"`
  Git             string     `json:"git"`
  GitBranch       string     `json:"git_branch"`
  Environment     string     `json:"environment"`
  Cloud           string     `json:"cloud"`
  Fqdn            string     `json:"fqdn"`
  Language        string     `json:"language"`
  Framework       string     `json:"framework"`
  StatusCode      int        `json:"status"`
  HealthCode      int        `json:"health"`
  MaintenanceMode bool       `json:"maintenance_mode"`
  HasLoadBalancer bool       `json:"has_loadbalancer"`
  RedeployHook    *string    `json:"redeploy_hook"`
  LastActivity    *time.Time `json:"last_activity_iso"`
  UpdatedAt       time.Time  `json:"updated_at_iso"`
  CreatedAt       time.Time  `json:"created_at_iso"`
}

type StackSetting struct {
  Key               string        `json:"key"`
  Value             interface{}   `json:"value"`
  Readonly          bool          `json:"readonly"`
}

func (s Stack) Status() string {
  return stackStatus[s.StatusCode]
}

func (s Stack) Health() string {
  return healthStatus[s.HealthCode]
}

func (c *Client) StackList() ([]Stack, error) {
	req, err := c.NewRequest("GET", "/stacks.json", nil)
	if err != nil {
		return nil, err
	}

	var stacksRes []Stack
	return stacksRes, c.DoReq(req, &stacksRes)
}

func (c *Client) StackListWithFilter(filter filterFunction) ([]Stack, error) {
  req, err := c.NewRequest("GET", "/stacks.json", nil)
  if err != nil {
    return nil, err
  }

  var stacksRes []Stack
  err = c.DoReq(req, &stacksRes)
  if err != nil {
    return nil, err
  }

  var result []Stack
  for _, item := range stacksRes {
    if filter(item) {
      result = append(result, item)
    }
  }
  return result, nil
}

func (c *Client) StackInfo(stackName string) (*Stack, error) {
  stack, err := c.FindStackByName(stackName, "")
  if err != nil {
    return nil, err
  }

  uid := stack.Uid
  req, err := c.NewRequest("GET", "/stacks/" + uid + ".json", nil)
  if err != nil {
    return nil, err
  }

  var stacksRes *Stack
  return stacksRes, c.DoReq(req, &stacksRes)
}

func (c *Client) StackInfoWithEnvironment(stackName, environment string) (*Stack, error) {
  stack, err := c.FindStackByName(stackName, environment)
  if err != nil {
    return nil, err
  }

  uid := stack.Uid
  req, err := c.NewRequest("GET", "/stacks/" + uid + ".json", nil)
  if err != nil {
    return nil, err
  }

  var stacksRes *Stack
  return stacksRes, c.DoReq(req, &stacksRes)
}

func (c *Client) StackSettings(uid string) ([]StackSetting, error) {
  req, err := c.NewRequest("GET", "/stacks/" + uid + "/settings.json", nil)
  if err != nil {
    return nil, err
  }

  var settingsRes []StackSetting
  return settingsRes, c.DoReq(req, &settingsRes)
}

func (c *Client) FindStackByName(stackName, environment string) (*Stack, error) {
  stacks, err := c.StackList()

  for _, b := range stacks {
    if (strings.ToLower(b.Name) == strings.ToLower(stackName)) && (environment == "" || environment == b.Environment) {
      return &b, err
    }
  }

  return nil, errors.New("Stack not found")
}

func (c *Client) Servers(uid string) ([]Server, error) {
  req, err := c.NewRequest("GET", "/stacks/" + uid + "/servers.json", nil)
  if err != nil {
    return nil, err
  }

  var serversRes []Server
  return serversRes, c.DoReq(req, &serversRes)
}

func (c *Client) ManagedBackups(uid string) ([]ManagedBackup, error) {
  req, err := c.NewRequest("GET", "/stacks/" + uid + "/managed_backups.json", nil)
  if err != nil {
    return nil, err
  }

  var managedBackupsRes []ManagedBackup
  return managedBackupsRes, c.DoReq(req, &managedBackupsRes)
}

func (c *Client) Set(uid string, key string, value string) (*GenericResponse, error) {
  params := struct {
		Key   string `json:"setting_name"`
    Value string `json:"setting_value"`
	}{
		Key:   key,
    Value: value,
	}

  req, err := c.NewRequest("POST", "/stacks/" + uid + "/setting.json", params)
  if err != nil {
    return nil, err
  }

  var settingRes *GenericResponse
  return settingRes, c.DoReq(req, &settingRes)
}

func (c *Client) RestartStack(uid string) (*GenericResponse, error) {
  req, err := c.NewRequest("POST", "/stacks/" + uid + "/restart.json", nil)
  if err != nil {
    return nil, err
  }

  var stacksRes *GenericResponse
  return stacksRes, c.DoReq(req, &stacksRes)
}

func (c *Client) Lease(uid string, ipAddress *string, timeToOpen *int, port *int) (*GenericResponse, error) {
  params := struct {
    TimeToOpen   *int     `json:"time_to_open"`
    IpAddress    *string  `json:"ip_address"`
    Port         *int     `json:"port"`
  }{
    TimeToOpen: timeToOpen,
    IpAddress: ipAddress,
    Port: port,
  }

  req, err := c.NewRequest("POST", "/stacks/" + uid + "/lease.json", params)
  if err != nil {
    return nil, err
  }

  var leaseRes *GenericResponse
  return leaseRes, c.DoReq(req, &leaseRes)
}

func (c *Client) RedeployStack(uid string) (*GenericResponse, error) {
  req, err := c.NewRequest("POST", "/stacks/" + uid + "/redeploy.json", nil)
  if err != nil {
    return nil, err
  }

  var stacksRes *GenericResponse
  return stacksRes, c.DoReq(req, &stacksRes)
}

func (c *Client) ClearCachesStack(uid string) (*GenericResponse, error) {
  req, err := c.NewRequest("POST", "/stacks/" + uid + "/clear_caches.json", nil)
  if err != nil {
    return nil, err
  }
  
  var stacksRes *GenericResponse
  return stacksRes, c.DoReq(req, &stacksRes)
}


