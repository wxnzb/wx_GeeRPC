package xclient

import (
	"net/http"
	"strings"
	"time"
)

type GeeRegistryDiscovery struct {
	*MultServers
	timeout      time.Duration
	lastUpdate   time.Time
	registryaddr string
}

const defaultTimeout = time.Second * time.Duration(10)

func NewGeeRegistryDiscovery(registryaddr string, timeout time.Duration) *GeeRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &GeeRegistryDiscovery{
		MultServers:  NewDiscovery(make([]string, 0)),
		timeout:      timeout,
		registryaddr: registryaddr,
	}
}
func (d *GeeRegistryDiscovery) Refresh() error {
	// 先看看是否过期
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	// 然后通过http拉取servers
	resp, err := http.Get(d.registryaddr)
	if err != nil {
		return err
	}
	servers := strings.Split(resp.Header.Get("Wu-Geerpc-Servers"), ",")
	d.MultServers.servers = make([]string, 0, len(servers))
	for _, s := range servers {
		if strings.TrimSpace(s) != "" {
			d.MultServers.servers = append(d.MultServers.servers, s)
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *GeeRegistryDiscovery) Get(mod SelectMode) (string, error) {
	return d.MultServers.Get(mod)
}
func (d *GeeRegistryDiscovery) GetAll() ([]string, error) {
	return d.MultServers.GetAll()
}
