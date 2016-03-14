package compose

import (
	"time"
	"fmt"
	"encoding/json"
	"strings"
	"strconv"
)

type Container struct {
	ID              string           `json:"Id"`
	Name            string           `json:"Name,omitempty"`
	Created         time.Time        `json:"Created,omitempty"`
	Config          *Config          `json:"Config,omitempty"`
	State           State            `json:"State,omitempty"`
	Image           string           `json:"Image,omitempty"`
	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty"`
}

type Config struct {
	Hostname          string              `json:"Hostname,omitempty"`
	ExposedPorts      map[string]struct{} `json:"ExposedPorts,omitempty"`
	Env               []string            `json:"Env,omitempty"`
	Cmd               []string            `json:"Cmd"`
	Image             string              `json:"Image,omitempty"`
	Labels            map[string]string   `json:"Labels,omitempty"`
}

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

type NetworkSettings struct {
	Ports                  map[string][]PortBinding     `json:"Ports,omitempty"`
}

type PortBinding struct {
	HostIP   string `json:"HostIP,omitempty"`
	HostPort string `json:"HostPort,omitempty"`
}

const (
	DefaultRetryCount = 10
	DefaultRetryDelay = 500 * time.Millisecond
)

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

func MustInspect(id string) *Container {
	container, err := Inspect(id)
	if err != nil {
		panic(err)
	}
	return container
}

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

func (c *Container) MustConnect(exposedPort uint32, proto string, retryCount int, retryDelay time.Duration, connector func (publicPort uint32) error) {
	if err := c.Connect(exposedPort, proto, retryCount, retryDelay, connector); err != nil {
		panic(err)
	}
}

func (c *Container) ConnectWithDefaults(exposedPort uint32, proto string, connector func (publicPort uint32) error) error {
	return c.Connect(exposedPort, proto, DefaultRetryCount, DefaultRetryDelay, connector)
}

func (c *Container) MustConnectWithDefaults(exposedPort uint32, proto string, connector func (publicPort uint32) error) {
	if err := c.ConnectWithDefaults(exposedPort, proto, connector); err != nil {
		panic(err)
	}
}

func (c *Container) GetFirstPublicPort(exposedPort uint32, proto string) (uint32, error) {
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

func (c *Container) MustGetFirstPublicPort(exposedPort uint32, proto string) uint32 {
	port, err := c.GetFirstPublicPort(exposedPort, proto)
	if err != nil {
		panic(err)
	}
	return port
}