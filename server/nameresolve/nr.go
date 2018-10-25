//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2018
//
// Convert names into handlers for sert of names
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1295
//

package nameresolve

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/godebug"
)

/*

	Addr: ::1/128
	Addr: 127.0.0.1/8
	Addr: fe80::1/64
	Addr: fe80::6a5b:35ff:fe91:644f/64
	Addr: 192.168.0.158/16
	Addr: 192.168.0.157/24

	2 alternatives for serves
		1. localhost:Port - Listen to :Port
		2. IP:Port - Listen to a single port

*/

/*
var ignore map[string]bool = map[string]bool{
	"::1": true,
}

var localhost map[string]bool = map[string]bool{
	"127.0.0.1": true,
	"fe80::1":   true,
	"":          true,
	"localhost": true,
}

type IpInfo struct {
	Addr        string
	IpV4        bool
	IsLocalhost bool
}

var listenToAddrs map[string]IpInfo

func init() {
	listenToAddrs = make(map[string]IpInfo)
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		//	fmt.Printf("Addrs: %s\n", godebug.SVarI(addrs))
		for _, addr := range addrs {
			fmt.Printf("runtimeInit: Addr: %s\n", addr)
			a := addr.String()
			p := strings.Index(a, "/")
			if p > 0 {
				a = a[0:p]
			}
			if _, ig := ignore[a]; !ig {
				if isLocalhost := localhost[a]; !isLocalhost {
					isIpV4 := strings.Index(a, ".") > -1
					listenToAddrs[a] = IpInfo{Addr: a, IpV4: isIpV4}
				} else {
					fmt.Printf("A Localhost found: %s\n", a)
					isIpV4 := strings.Index(a, ".") > -1
					listenToAddrs[a] = IpInfo{Addr: a, IpV4: isIpV4, IsLocalhost: true}
				}
			}
		}
	} else {
		fmt.Printf("Error: %s, Unable to initialize the listeners - failed\n", err)
		os.Exit(1)
	}

	fmt.Printf("listenToAddrs: %s\n", lib.SVarI(listenToAddrs))
}
*/

//

//
// Take names like www.test1.com and convert them into sets of IP address.
// also *.www.test1.com resolve and *.test1.com resolve and *.*.test1.com
// also there is a default handler if no resolution
//

type IpToHostPort struct {
	Host    string
	Port    string
	Proto   string
	IPAddr  string
	Handler http.Handler
	Id      int
}

type NameResolve struct {
	IpLookup  map[string]map[string]*IpToHostPort //
	RawLookup map[string]*IpToHostPort
	Debug1    bool       //
	Debug2    bool       // Use In: func GetProtoHostPort(name string) (proto, host, port string, err error)
	Debug3    bool       //
	Debug4    bool       //
	Debug5    bool       //
	mutex     sync.Mutex //
}

func (nr *NameResolve) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	name := "http://" + req.Host
	ii, ok := nr.GetHandler(name)
	if !ok {
		http.NotFound(www, req)
		return
	}
	ii.Handler.ServeHTTP(www, req)
}

// IPs, err := net.LookupIP(host) // xyzzy - fast enough - need to cache?
type CacheIPData struct {
	IP   []net.IP
	err  error
	When time.Time // When saved
}

type CacheIP struct {
	Data  map[string]CacheIPData
	mutex sync.Mutex
}

var CIp CacheIP

func init() {
	CIp.Data = make(map[string]CacheIPData)
}

func (cc *CacheIP) CachedLookupIP(host string) (ip []net.IP, err error) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	if it, ok := cc.Data[host]; ok {
		if it.When.Add(time.Duration(60) * time.Second).Before(time.Now()) {
			delete(cc.Data, host)
		} else {
			return it.IP, it.err
		}
	}
	if db44 {
		fmt.Printf("**************** Lookup of host - this is place to replace/mark hosts that map to local system, %s\n", host)
	}
	ip, err = net.LookupIP(host)
	cc.Data[host] = CacheIPData{ip, err, time.Now()}
	return
}

func NewNameResolve() *NameResolve {
	return &NameResolve{
		// First string is IP address only
		// Second string is NamePattern:port
		IpLookup:  make(map[string]map[string]*IpToHostPort),
		RawLookup: make(map[string]*IpToHostPort),
	}
}

//
// This is the name matcher for HTTP requests - it returns the correct handler.
// This means that if we are to dynamically configure - then this is the place to do it -
// We could use the "Id" to re-lookup (index) the handler from a table and pull it out
// Every time - that alone would lead to all new connections are based on new config.
//
// name may be bob.test1.com - and match to bob.test1.com, if not look for *.test1.com, then *.*.com, then *.com
func (nr *NameResolve) GetHandler(name string) (rv *IpToHostPort, ok bool) {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	if nr.Debug5 {
		fmt.Printf("\nlookup for %s, %s\n", name, godebug.LF())
	}

	// split "name" into its parts, Protocal, Host, Port
	proto, host, port, err := GetProtoHostPort(name)
	if err != nil {
		fmt.Printf("Unable to get protocal, host, port from %s, %s, %s, %s\n", proto, host, port, name)
		ok = false
		return
	}
	if port == "" {
		port = "80"
	}
	if nr.Debug5 {
		fmt.Printf("protocal, host, port from %s, %s, %s, name=%s, %s\n", proto, host, port, name, godebug.LF())
	}
	// IPs, err := net.LookupIP(host) // xyzzy - fast enough - need to cache?

	// Convert from host to the set of IPs that the host can represent.
	IPs, err := CIp.CachedLookupIP(host)
	if err != nil {
		if nr.Debug4 {
			fmt.Printf("Unable to get IPs from %s\n", name)
		}
		goto next
	}

	if nr.Debug5 {
		fmt.Printf("IPs for %s are %s, %s\n", IPs, host, godebug.LF())
		fmt.Printf("*** Data nr:%s, IPs:%s\n", lib.SVarI(nr), lib.SVarI(IPs))
	}

	// Search the set of IPs for the name, looking for best match
	for jj, ww := range IPs {
		_ = jj
		ips := fmt.Sprintf("%s:%s", ww, port)
		// xyzzy - xyzzy - IP6 - use [%s]:%s --------------------------------------------- <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
		if nr.Debug5 {
			fmt.Printf("for %d in IPs, host:port == %s, %s\n", jj, ips, lib.LF())
		}
		name_port := host + ":" + port // name plus port to look on
		if strings.HasPrefix(name_port, "http://") {
			name_port = name_port[7:]
		} else if strings.HasPrefix(name_port, "https://") {
			name_port = name_port[8:]
		}
		if nr.Debug5 {
			fmt.Printf("for %d in IPs, name:port == %s, %s\n", jj, name_port, godebug.LF())
		}
		if tt, xok := nr.IpLookup[ips]; xok {
			if rv, ok = tt[name_port]; ok {
				if nr.Debug5 {
					fmt.Printf("Found!!!!!!!!!!!!!!!!!\n")
				}
				return
			}
		}
	}

	if nr.Debug5 {
		fmt.Printf("First search failed, looking into 2nd one now, %s\n", godebug.LF())
	}

	// xyzzy - xyzzyElse
	// xyzzy - if IP address and IP/port is unique - then -

next:

	if nr.Debug1 {
		fmt.Printf("Part 2 - search ----------------------------------------------------, %s\n", godebug.LF())
	}
	comp := strings.Split(host, ".")
	n_comp := len(comp)
	for i := 0; i < n_comp-1; i++ {
		comp[i] = "*"
		nn := strings.Join(comp, ".")
		noStar := strings.Join(comp[i+1:], ".")
		if nr.Debug1 {
			fmt.Printf("new comp at [%s] = %s, %s\n", nn, noStar, godebug.LF())
		}

		// IPs, err := net.LookupIP(noStar) // xyzzy - fast enough - need to cache?
		IPs, err := CIp.CachedLookupIP(noStar)
		if err != nil {
			if nr.Debug4 {
				fmt.Printf("Unable to get IPs from %s\n", name)
			}
		} else {

			name_port := nn + ":" + port
			for jj, ww := range IPs {
				ips := fmt.Sprintf("%s:%s", ww, port)
				if nr.Debug1 {
					fmt.Printf("Lookup IP [%s], pass %d, nn=%s host=%s port=%s %s\n", ips, jj, nn, host, port, godebug.LF())
				}
				// xyzzy - xyzzy - IP6 - use [%s]:%s --------------------------------------------- <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
				if nr.Debug5 {
					fmt.Printf("for %d in IPs, host:port == %s, %s\n", jj, ips, lib.LF())
					fmt.Printf("for %d in IPs, name:port == %s, %s\n", jj, name_port, lib.LF())
				}
				if tt, xok := nr.IpLookup[ips]; xok {
					if rv, ok = tt[name_port]; ok {
						if nr.Debug5 {
							fmt.Printf("Found!!!!!!!!!!!!!!!!! in pass 2\n")
						}
						return
					}
				}
			}

		}
	}

	// xyzzy - xyzzyElse
	if nr.Debug1 {
		fmt.Printf("Part 3 - wildcard with 0.0.0.0 ----------------------------------------------------, %s\n", godebug.LF())
	}
	if tt, xok := nr.IpLookup["0.0.0.0:"+port]; xok {
		if rv, ok = tt["0.0.0.0:"+port]; ok {
			return
		}
		if rv, ok = tt["0.0.0.0:*"]; ok {
			return
		}
		if rv, ok = tt["*:*"]; ok {
			return
		}
	}

	ok = false
	return
}

var ErrDuplicateNameResolve = errors.New("Each name must resolve to a unique set of rules - duplciate found")
var ErrUnableToParse = errors.New("Unable to parse url")

// Add name creates a table of IP:Port - with a list of name:Port for name matching

func (nr *NameResolve) AddName(namePattern string, hdlr http.Handler, id int, addrIfNone string) (e error) {
	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	for strings.HasSuffix(namePattern, "/") {
		namePattern = namePattern[0 : len(namePattern)-1]
	}

	if nr.Debug5 {
		fmt.Printf("***************\nAddName called with %s -- id %d\n******************\n", namePattern, id)
	}

	proto, host, port, err := GetProtoHostPort(namePattern)
	if err != nil {
		fmt.Printf("Unable to get protocal, host, port from %s, %s, %s, %s\n", proto, host, port, namePattern)
		return ErrUnableToParse
	}

	// xyzzy - default for http, https on name - 80/443!

	// IPs, err1 := net.LookupIP(NoWild(host))
	IPs, err := CIp.CachedLookupIP(NoWild(host))
	if err != nil {
		if addrIfNone != "" {
			IPs = make([]net.IP, 1, 1)
			IPs[0] = net.ParseIP(addrIfNone) // []byte(addrIfNone)
			if nr.Debug5 {
				fmt.Printf("Unable to get IP addresses from %s, %s, %s, %s - domain server using %s, %s\n", proto, host, port, namePattern, addrIfNone, IPs)
			}
		} else {
			fmt.Printf("Unable to get IP addresses from %s, %s, %s, %s\n", proto, host, port, namePattern)
			return ErrUnableToParse
		}
	}

	if nr.Debug5 {
		fmt.Printf("AddName: IPs for %s are: %s, Port=%s\n", namePattern, IPs, port)
	}

	for jj, ww := range IPs {
		ips := ""
		if IsIp4(ww.String()) {
			ips = fmt.Sprintf("%s:%s", ww, port)
		} else if ww.String() == "fe80::1" {
			ips = fmt.Sprintf("[::1]:%s", port)
		} else {
			ips = fmt.Sprintf("[%s]:%s", ww, port)
		}
		if nr.Debug5 {
			fmt.Printf("AddName: in loop at %d, ips=%s, port=%s\n", jj, ips, port)
		}
		xx, ok := nr.IpLookup[ips]
		if !ok {
			if nr.Debug3 {
				fmt.Printf("Creating nr.IpLookup[%s], %s\n", ips, godebug.LF())
			}
			nr.IpLookup[ips] = make(map[string]*IpToHostPort)
			xx = nr.IpLookup[ips]
		}
		hp := fmt.Sprintf("%s:%s", host, port)
		if _, ok := xx[hp]; ok {
			fmt.Printf("Error - dup entry, host:port = %s, id=%d\n", hp, id) // xyzzy - Should go to log
			return ErrDuplicateNameResolve
		} else {
			// fmt.Printf("Adding new item, %s, %s, %s\n", hp, id, godebug.LF())
			xx[hp] = &IpToHostPort{
				Host:    host,
				Port:    port,
				Proto:   proto,
				IPAddr:  ips,
				Handler: hdlr,
				Id:      id,
			}
			nr.RawLookup[namePattern] = xx[hp]
		}
	}
	return
}

//			p1, err := bot.GetRawTopHandler ( listen );
func (nr *NameResolve) GetRawTopHandler(listen string) (Handler http.Handler, err error) {
	tmp, ok := nr.RawLookup[listen]
	if !ok {
		err = fmt.Errorf("Did not find %s listener\n", listen)
		return
	}
	Handler = tmp.Handler
	return
}

func IsIp4(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return true
		case ':':
			return false
		}
	}
	return false

}

func (nr *NameResolve) AddDefault(proto, port string, hdlr http.Handler, id int) (e error) {
	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	host := "0.0.0.0"
	if proto == "" {
		proto = "http:"
	}
	if port == "" {
		proto = "80"
	}

	ips := host
	xx, ok := nr.IpLookup[ips]
	if !ok {
		if nr.Debug3 {
			fmt.Printf("Creating/Default nr.IpLookup[%s], %s\n", ips, godebug.LF())
		}
		nr.IpLookup[ips] = make(map[string]*IpToHostPort)
		xx = nr.IpLookup[ips]
	}
	hp := fmt.Sprintf("%s:%s", host, port)
	if _, ok := xx[hp]; ok {
		fmt.Printf("Error - dup entry, host:port = %s, id=%d\n", hp, id) // xyzzy - Should go to log
		return ErrDuplicateNameResolve
	} else {
		// fmt.Printf("Adding new item, %s, %s, %s\n", hp, id, godebug.LF())
		xx[hp] = &IpToHostPort{
			Host:    host,
			Port:    port,
			Proto:   proto,
			IPAddr:  ips,
			Handler: hdlr,
			Id:      id,
		}
	}
	return
}

// remove leading *. from host pattern to get domain name for IP lookup
func NoWild(host string) (rv string) {
	rv = host
	for len(rv) > 2 && rv[0:2] == "*." {
		rv = rv[2:]
	}
	return
}

func GetProtoHostPort(name string) (proto, host, port string, err error) {
	proto = "http:"
	// fmt.Printf("Before [%s]\n", name)
	name = strings.TrimRight(name, "/")
	// fmt.Printf("After [%s]\n", name)
	if strings.HasPrefix(name, "http://") {
		port = "80"
		name = name[7:]
	} else if strings.HasPrefix(name, "https://") {
		proto = "https:"
		port = "443"
		name = name[8:]
	}
	/*
		host, port, err = net.SplitHostPort(name)
		if err != nil {
			if strings.HasPrefix(fmt.Sprintf("%s", err), "missing port in address") {
				if proto == "http:" {
					port = "80"
					host = name
					err = nil
				} else if proto == "https:" {
					port = "443"
					host = name
					err = nil
				}
			} else {
				fmt.Printf(" Error: %s, %s\n", err, godebug.LF())
			}
		} else if err != nil {
			fmt.Printf(" Error: %s, %s\n", err, godebug.LF())
		}
	*/
	urlPort := regexp.MustCompile(".*:.*")
	if urlPort.MatchString(name) {
		aa := strings.Split(name, ":")
		host, port = aa[0], aa[1]
	} else {
		host = name
	}
	if db4 {
		fmt.Printf("name ->%s<- proto [%s] host [%s] port [%s], %s\n", name, proto, host, port, godebug.LF())
	}
	return
}

const db4 = false
const db44 = false
