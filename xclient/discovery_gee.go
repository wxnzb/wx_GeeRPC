package xclient

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GeeRegistryDiscovery struct {
	*MultServers
	timeout      time.Duration
	lastUpdate   time.Time
	registryaddr string
	mu           sync.Mutex
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
	d.mu.Lock()
	defer d.mu.Unlock()
	// 先看看是否过期
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("GeeRegistryDiscovery Refresh", d.registryaddr)
	// 然后通过http拉取servers
	resp, err := http.Get(d.registryaddr)
	if err != nil {
		log.Println("GeeRegistryDiscovery Refresh err", err)
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

// 下面这两个都要调用refresh来保证仓库里的服务器地址是没有过期的
func (d *GeeRegistryDiscovery) Get(mod SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultServers.Get(mod)
}
func (d *GeeRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultServers.GetAll()
}

// 这个函数感觉还没有用到
func (d *GeeRegistryDiscovery) Updata(servers []string) error {
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}
