package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type GeeRegistry struct {
	mu      sync.Mutex
	servers map[string]*ServerItem
	timeout time.Duration
}
type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultRegistryPath = "/_geerpc_/registy"
	defaultTimeout      = 5 * time.Second
)

func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultRegistry = New(defaultTimeout)

// 添加服务器地址在仓库
func (r *GeeRegistry) Registry(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var item = r.servers[addr]
	if item == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		r.servers[addr].start = time.Now()
	}
}

// 返回仓库中存在的服务器地址
func (r *GeeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var s []string
	for addr, item := range r.servers {
		if r.timeout == 0 || item.start.Add(r.timeout).After(time.Now()) {
			s = append(s, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	//这里就不能直接返回s吗，奇怪
	sort.Strings(s)
	return s
}
func (r *GeeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("Wu-Geerpc-Servers", strings.Join(r.aliveServers(), ","))
	case "SET":
		addr := req.Header.Get("Wu-Geerpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Registry(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
func (r *GeeRegistry) HandleHttp(defaultregistrypath string) {
	http.Handle(defaultRegistryPath, r)
	log.Println("rpc registry on", defaultregistrypath)
}
func HandleHttp() {
	DefaultRegistry.HandleHttp(defaultRegistryPath)
}
func HeartBeat(register, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Minute*time.Duration(1)
	}
	var err error
	err = SendHeartBeat(register, addr)
	for err == nil {
		<-time.After(duration)
		err = SendHeartBeat(register, addr)
	}
}
func SendHeartBeat(register, addr string) error {
	log.Println(addr, "send heart beat to registry", register)
	//创建http客户端
	httpClient := &http.Client{}
	//构造http请求
	req, _ := http.NewRequest("SET", register, nil)
	//设置http请求头
	req.Header.Set("Wu-Geerpc-Server", addr)
	//发送http请求
	_, err := httpClient.Do(req)
	if err != nil {
		log.Println("rpc server heart beat err:", err)
		return err
	}
	return nil
}
