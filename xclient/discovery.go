package xclient

import (
	"errors"
	"math/rand"
	"sync"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type MultServers struct {
	servers []string
	r       *rand.Rand
	index   int
	mu      sync.RWMutex
}
type Discovery interface {
	Refresh() error //怎么感觉这个函数没啥作用呀
	Updata([]string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// 实现Discovery接口
var _ Discovery = (*MultServers)(nil)

// 实现接口实例
func (ms *MultServers) Refresh() error {
	return nil
}
func (ms *MultServers) Updata(servers []string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	//这里为啥不能用copy函数呢
	ms.servers = servers
	return nil
}
func (ms *MultServers) Get(mode SelectMode) (string, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	n := len(ms.servers)
	switch mode {
	case RandomSelect:
		return ms.servers[ms.r.Intn(n)], nil
	case RoundRobinSelect:
		s := ms.servers[ms.index]
		ms.index = ms.index + 1
		return s, nil
	}
	return "", errors.New("no such select mode")
}
func (ms *MultServers) GetAll() ([]string, error) {
	servers := make([]string, len(ms.servers), len(ms.servers))
	copy(servers, ms.servers)
	return servers, nil
}
