package WsServiceLib

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/pool" // Modified pool to have NewAuth for authorized connections
	"github.com/pschlump/radix.v2/pubsub"
	"github.com/pschlump/radix.v2/redis"
)

// Redis Connect Info ( 2 channels )
type MsCfgType struct {
	ServerId            string               `json:"Sid"`              //
	Name                string               `json:"QName"`            //	// Name of the Q to send stuff to //
	ReplyTTL            uint64               `json:"ReplyTTL"`         // how long will a reply last if not picked up.
	isRedisConnected    bool                 `json:"-"`                // If connect to redis has occured for wakie-wakie calls
	RedisConnectHost    string               `json:"RedisConnectHost"` // Connection infor for Redis Database
	RedisConnectPort    string               `json:"RedisConnectPort"` //
	RedisConnectAuth    string               `json:"RedisConnectAuth"` //
	RedisPool           *pool.Pool           `json:"-"`                // Pooled Redis Client connectioninformation
	Err                 error                `json:"-"`                // Error Holding Pen
	subClient           *pubsub.SubClient    `json:"-"`                //
	subChan             chan *pubsub.SubResp `json:"-"`                //
	timeout             chan bool            `json:"-"`                //
	DebugTimeoutMessage bool                 `json:"-"`                // turn on/off the timeout 1ce a second message
	TickInMilliseconds  int                  `json:"-"`                // # of miliseconds for 1 tick
}

type WorkFuncType func(arb map[string]interface{})

func NewMsCfgType(qName string) (ms *MsCfgType) {
	var err error
	ms = &MsCfgType{
		ServerId:           UUIDAsStrPacked(), //
		Err:                err,               //
		Name:               qName,             // Name of message Q that will be published on
		TickInMilliseconds: 100,               // 100 milliseconds
	}
	return
}

func (ms *MsCfgType) SetupListenServer(pattern string) {

	client, err := redis.Dial("tcp", ms.RedisConnectHost+":"+ms.RedisConnectPort)
	if err != nil {
		log.Fatal(err)
	}
	if ms.RedisConnectAuth != "" {
		err = client.Cmd("AUTH", ms.RedisConnectAuth).Err
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintf(os.Stderr, "Success: Connected to redis-server with AUTH.\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "Success: Connected to redis-server.\n")
	}

	ms.subClient = pubsub.NewSubClient(client) // subClient *pubsub.SubClient
	ms.subChan = make(chan *pubsub.SubResp)
	ms.timeout = make(chan bool, 1)

	sr := ms.subClient.PSubscribe(ms.Name, pattern)
	if sr.Err != nil {
		fmt.Fprintf(os.Stderr, "%sError: psubscribe(%s), %s.%s\n", MiscLib.ColorRed, sr.Err, pattern, MiscLib.ColorReset)
	}
}

func (hdlr *MsCfgType) ConnectToRedis() bool {
	// Note: best test for this is in the TabServer2 - test 0001 - checks that this works.
	var err error

	dflt := func(a string, d string) (rv string) {
		rv = a
		if rv == "" {
			rv = d
		}
		return
	}

	redis_host := dflt(hdlr.RedisConnectHost, "127.0.0.1")
	redis_port := dflt(hdlr.RedisConnectPort, "6379")
	redis_auth := hdlr.RedisConnectAuth

	if redis_auth == "" { // If Redis AUTH section
		hdlr.RedisPool, err = pool.New("tcp", redis_host+":"+redis_port, 20)
	} else {
		hdlr.RedisPool, err = pool.NewAuth("tcp", redis_host+":"+redis_port, 20, redis_auth)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Failed to connect to redis-server.%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		fmt.Printf("Error: Failed to connect to redis-server.\n")
		// goftlmux.G_Log.Info("Error: Failed to connect to redis-server.\n")
		// logrus.Fatalf("Error: Failed to connect to redis-server.\n")
		return false
	} else {
		if db3 {
			fmt.Fprintf(os.Stderr, "%sSuccess: Connected to redis-server.%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
	}

	return true
}

func (ms *MsCfgType) SetRedisConnectInfo(h, p, a string) {
	ms.RedisConnectHost = h
	ms.RedisConnectPort = p
	ms.RedisConnectAuth = a
}

/*
 */
func (ms *MsCfgType) ListenForServer(doWork WorkFuncType, wg *sync.WaitGroup) { // server *socketio.Server) {

	arb := make(map[string]interface{})
	arb["cmd"] = "at-top"
	doWork(arb)

	go func() {
		for {
			ms.subChan <- ms.subClient.Receive()
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(ms.TickInMilliseconds) * time.Millisecond)
			ms.timeout <- true
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var sr *pubsub.SubResp
		counter := 0
		threshold := 100 // xyzzy - from config!!!
		for {
			select {
			case sr = <-ms.subChan:
				if db1 {
					fmt.Fprintf(os.Stderr, "%s**** Got a message, sr=%+v, is --->>>%s<<<---, AT:%s%s\n", MiscLib.ColorGreen, godebug.SVar(sr), sr.Message, godebug.LF(), MiscLib.ColorReset)
				}

				arb := make(map[string]interface{})
				err := json.Unmarshal([]byte(sr.Message), &arb)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%sError: %s --->>>%s<<<--- AT: %s%s\n", MiscLib.ColorRed, err, sr.Message, godebug.LF(), MiscLib.ColorReset)
				} else {

					cmd := ""
					cmd_x, ok := arb["cmd"]
					if ok {
						cmd_s, ok := cmd_x.(string)
						if ok {
							cmd = cmd_s
						}
					}
					if cmd == "exit-now" {
						break
					}

					doWork(arb)

				}

			case <-ms.timeout: // the read from ms.subChan has timed out

				if ms.DebugTimeoutMessage {
					fmt.Fprintf(os.Stderr, "%s**** Got a timeout, AT:%s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
				}

				// If stuck doing work - may need to kill/restart - server side timeout.
				counter++
				if counter > threshold {
					if repoll_db {
						fmt.Fprintf(os.Stderr, "%s**** timeout - results in a call to doWork(), AT:%s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
					}
					arb := make(map[string]interface{})
					arb["cmd"] = "timeout-call"
					doWork(arb)
					counter = 0
				}

			}
		}
	}()

}

const db1 = false
const db3 = false
const repoll_db = false
