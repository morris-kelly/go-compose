package compose

import (
	"time"
	"fmt"
	"encoding/json"
	"strings"
	"strconv"
)

// Models the `docker inspect` command output.
type Container struct {
	ID              string           `json:"Id"`
	Name            string           `json:"Name,omitempty"`
	Created         time.Time        `json:"Created,omitempty"`
	Config          *Config          `json:"Config,omitempty"`
	State           State            `json:"State,omitempty"`
	Image           string           `json:"Image,omitempty"`
	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty"`
}

// Models the config section of the `docker inspect` command output.
type Config struct {
	Hostname          string              `json:"Hostname,omitempty"`
	ExposedPorts      map[string]struct{} `json:"ExposedPorts,omitempty"`
	Env               []string            `json:"Env,omitempty"`
	Cmd               []string            `json:"Cmd"`
	Image             string              `json:"Image,omitempty"`
	Labels            map[string]string   `json:"Labels,omitempty"`
}

// Models the state section of the `docker inspect` command.
type State struct {
	Running    bool      `json:"Running,omitempty"`
	Paused     bool      `json:"Paused,omitempty"`
	Restarting bool      `json:"Restarting,omitempty"`
	OOMKilled  bool      `json:"OOMKilled,omitempty"`
	Pid        int       `json:"Pid,omitempty"`
	ExitCode   int       `json:"ExitCode,omitempty"`
	Error      string    `json:"Error,omitempty"`
	StartedAt  time.Time `json:"StartedAt,omitempty"`
	FinishedAt time.Time `json:"FinishedAt,omitempty"`
}

// Models the network settings section of the `docker inspect` command.
type NetworkSettings struct {
	Ports                  map[string][]PortBinding     `json:"Ports,omitempty"`
}

// Models a port binding in the network settings section of the `docker inspect command.
type PortBinding struct {
	HostIP   string `json:"HostIP,omitempty"`
	HostPort string `json:"HostPort,omitempty"`
}

const (
	DefaultRetryCount = 10 	// Default number of retries for the Connect operation.
	DefaultRetryDelay = 500 * time.Millisecond // Default delay between retries for the Connect operation.
)

// Inspects a container using the `docker inspect` command and returns a parsed version of its output.
func Inspect(id string) (*Container, error) {
	stdout, _, err := runCmd("docker", "inspect", id)
	if err != nil {
		return nil, fmt.Errorf("compose: error inspecting container: %v", err)
	}

	inspect := make([]*Container, 0)
	if err := json.Unmarshal([]byte(stdout), &inspect); err != nil {
		return nil, fmt.Errorf("compose: error parsing inspect output: %v", err)
	}
	if len(inspect) != 1 {
		return nil, fmt.Errorf("compose: inspect returned %v results, 1 expected", len(inspect))
	}

	return inspect[0], nil
}

// Like Inspect, but panics on error.
func MustInspect(id string) *Container {
	container, err := Inspect(id)
	if err != nil {
		panic(err)
	}
	return container
}

// Attempts to connect to a container using the given connector function.
// The given exposedPort is automatically mapped to the corresponding public port.
// Use retryCount and retryDelay to configure the number of retries and the time waited between them
func (c *Container) Connect(exposedPort uint32, proto string, retryCount int, retryDelay time.Duration, connector func (publicPort uint32) error) error {
	publicPort, err := c.GetFirstPublicPort(exposedPort, proto)
	if err != nil {
		return err
	}

	for i := 0; i < retryCount; i++ {
		err = connector(publicPort)
		if err == nil {
			return nil
		}
		time.Sleep(retryDelay)
	}

	return err
}

// Like Connect, but panics on error.
func (c *Container) MustConnect(exposedPort uint32, proto string, retryCount int, retryDelay time.Duration, connector func (publicPort uint32) error) {
	if err := c.Connect(exposedPort, proto, retryCount, retryDelay, connector); err != nil {
		panic(err)
	}
}

// Like Connect, with default values for retryCount and retryDelay.
func (c *Container) ConnectWithDefaults(exposedPort uint32, proto string, connector func (publicPort uint32) error) error {
	return c.Connect(exposedPort, proto, DefaultRetryCount, DefaultRetryDelay, connector)
}

// Like ConnectWithDefaults, but panics on error.
func (c *Container) MustConnectWithDefaults(exposedPort uint32, proto string, connector func (publicPort uint32) error) {
	if err := c.ConnectWithDefaults(exposedPort, proto, connector); err != nil {
		panic(err)
	}
}

// If found, returns the first public public port mapped to the given exposedPort, for the given proto ("tcp", "udp", etc.).
func (c *Container) GetFirstPublicPort(exposedPort uint32, proto string) (uint32, error) {
	if c.NetworkSettings == nil {
		return 0, fmt.Errorf("compose: no network settings for container '%v'", c.Name)
	}

	portSpec := fmt.Sprintf("%v/%v", exposedPort, strings.ToLower(proto))
	mapping, ok := c.NetworkSettings.Ports[portSpec]
	if !ok || len(mapping) == 0 {
		return 0, fmt.Errorf("compose: no public port for %v", portSpec)
	}

	port, err := strconv.ParseUint(mapping[0].HostPort, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("compose: error parsing port '%v'", mapping[0].HostPort)
	}

	return uint32(port), nil
}

// Like GetFirstPublicPort, but panics on error.
func (c *Container) MustGetFirstPublicPort(exposedPort uint32, proto string) uint32 {
	port, err := c.GetFirstPublicPort(exposedPort, proto)
	if err != nil {
		panic(err)
	}
	return port
}