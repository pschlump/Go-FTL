//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1254
//

//
// Socket.IO connector to Redis pub/sub
//

// TODO:
// 		1. xyzzy - this could be a MicroService at this point - durable to a destination. call/response or call/ignore response
// 		msgTo            string                     // -- must be a config item
//		"MessagePrefix":      { "type":[ "string" ], "default":"sio:%{Id%}" },
//		"MessageRespPrefix":  { "type":[ "string" ], "default":"r:listen" },
//
//
//
// 	0. xyzzy - Return MIME type! -> JSON
//
//	http://localhost:16001/api/sio/list-client
//	http://localhost:16001/
//
//	4. Login required for external access? -- Could be implemented by "LoginRequried" // xyzzyLoginReq
//
//	5. Add time of last message to lookup of clients -- add tyep of request GET/POST - poling, websocket etc.	// xyzzyTime
//		Also # of messages - Logging Info
//
// --done--
//

package SocketIO

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	JsonX "github.com/pschlump/JSONx"

	"github.com/Sirupsen/logrus"
	"github.com/pschlump/Go-FTL/server/cfg"
	"github.com/pschlump/Go-FTL/server/goftlmux"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/Go-FTL/server/mid"
	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/MicroServiceLib"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/Tracer2/lib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/pubsub"
	"github.com/pschlump/radix.v2/redis"
	"github.com/pschlump/socketio"
)

// --------------------------------------------------------------------------------------------------------------------------

//func init() {
//
//	// normally identical
//	initNext := func(next http.Handler, gCfg *cfg.ServerGlobalConfigType, ppCfg interface{}, serverName string, pNo int) (rv http.Handler, err error) {
//		pCfg, ok := ppCfg.(*SocketIOHandlerType)
//		if ok {
//			pCfg.SetNext(next)
//			rv = pCfg
//		} else {
//			err = mid.FtlConfigError
//			logrus.Errorf("Invalid type passed at: %s", godebug.LF())
//		}
//		gCfg.ConnectToRedis()
//		pCfg.gCfg = gCfg
//
//		pCfg.msgTo = "chat-bot" // ms := MicroServiceLib.NewMsCfgType("qr-img1", "qr-img1-reply") //xyzzy - need to set the reply template
//		MyId := sizlib.UUIDAsStrPacked()
//		pCfg.ms = MicroServiceLib.NewMsCfgType(pCfg.msgTo, pCfg.msgTo+":"+MyId)
//		pCfg.myId = MyId
//		// ms.ConnectToRedis()                                        // Create the redis connection pool, alternative is ms.SetRedisPool(pool) // ms . SetRedisPool(pool *pool.Pool)
//		pCfg.ms.SetRedisConnectInfo(gCfg.RedisConnectHost, gCfg.RedisConnectPort, gCfg.RedisConnectAuth)
//		pCfg.ms.SetRedisPool(gCfg.RedisPool)
//		pCfg.ms.SetupListen()
//		pCfg.ms.ListenFor()
//		return
//	}
//
//	postInit := func(h interface{}, cfgData map[string]interface{}, callNo int) error {
//		// fmt.Printf("In postInitValidation, h=%v\n", h)
//		hh, ok := h.(*SocketIOHandlerType)
//		if !ok {
//			fmt.Fprintf(os.Stderr, "%sError: Wrong data type passed, Line No:%d\n%s", MiscLib.ColorRed, hh.LineNo, MiscLib.ColorReset)
//			return mid.ErrInternalError
//		} else {
//
//			hh.mutex.Lock()
//			hh.LookupIds = make(map[string]*SocketIdType)
//			hh.LookupTrxIds = make(map[string]string)
//			hh.mutex.Unlock()
//
//			hh.apiEnableIR, _ = lib.ParseBool(hh.ApiEnableIR)
//			hh.apiEnableRR, _ = lib.ParseBool(hh.ApiEnableRR)
//
//			server, err := socketio.NewServer(nil)
//			if err != nil {
//				logrus.Errorf("Error: %s\n", err)
//				return mid.ErrInternalError
//			}
//			hh.server = server
//
//			client, err := redis.Dial("tcp", cfg.ServerGlobal.RedisConnectHost+":"+cfg.ServerGlobal.RedisConnectPort)
//			if err != nil {
//				log.Fatal(err)
//			}
//			if cfg.ServerGlobal.RedisConnectAuth != "" {
//				err = client.Cmd("AUTH", cfg.ServerGlobal.RedisConnectAuth).Err
//				if err != nil {
//					log.Fatal(err)
//				} else {
//					fmt.Fprintf(os.Stderr, "Success: Connected to redis-server with AUTH.\n")
//				}
//			} else {
//				fmt.Fprintf(os.Stderr, "Success: Connected to redis-server.\n")
//			}
//
//			hh.subClient = pubsub.NewSubClient(client) // subClient *pubsub.SubClient
//			hh.subChan = make(chan *pubsub.SubResp)
//
//			sr := hh.subClient.Subscribe(hh.MessageRespPrefix)
//			if sr.Err != nil {
//				fmt.Fprintf(os.Stderr, "%sError: subscribe, %s.%s\n", MiscLib.ColorRed, sr.Err, MiscLib.ColorReset)
//			}
//
//			fmt.Printf("\nListening for messages to send to client at [%s]\n\n", hh.MessageRespPrefix)
//
//			hh.InitChatServer()
//			hh.ListenFor(hh.server)
//
//			hh.server.On("error", func(so socketio.Socket, err error) {
//				fmt.Printf("Error: %s, %s\n", err, godebug.LF())
//			})
//
//		}
//		return nil
//	}
//
//	// normally identical
//	createEmptyType := func() interface{} { return &SocketIOHandlerType{} }
//
//	cfg.RegInitItem2("SocketIO", initNext, createEmptyType, postInit, `{
//		}`)
//}
//
//// normally identical
//func (hdlr *SocketIOHandlerType) SetNext(next http.Handler) {
//	hdlr.Next = next
//}

func init() {
	CreateEmpty := func(name string) mid.GoFTLMiddleWare {
		x := &SocketIOHandlerType{}
		meta := make(map[string]JsonX.MetaInfo)
		JsonX.SetDefaults(&x, meta, "", "", "") // xyzzy - report errors in 'meta'
		return x
	}
	mid.RegInitItem3("SocketIO", CreateEmpty, `{
		"Paths":              { "type":["string","filepath"], "isarray":true, "required":true },
		"InternalRequestPw":  { "type":[ "string" ] },
		"ListClientApi":      { "type":[ "string" ], "default":"/api/sio/list-client" },
		"ApiEnableIR":        { "type":[ "string" ] },
		"ApiEnableRR":        { "type":[ "string" ] },
		"MessagePrefix":      { "type":[ "string" ], "default":"sio:%{Id%}" },
		"MessageRespPrefix":  { "type":[ "string" ], "default":"r:listen" },
		"SioRoutes":          { "type":[ "hash" ], "isarray":true },
		"LineNo":             { "type":[ "int" ], "default":"1" }
		}`)
	// "SocketIOLibrary":    { "type":["string"], "default":"socket.io.js" },
}

func (hdlr *SocketIOHandlerType) InitializeWithConfigData(next http.Handler, gCfg *cfg.ServerGlobalConfigType, serverName string, pNo, callNo int) (err error) {
	hdlr.Next = next
	//hdlr.CallNo = callNo // 0 if 1st init
	gCfg.ConnectToRedis()
	hdlr.gCfg = gCfg

	hdlr.msgTo = "chat-bot" // ms := MicroServiceLib.NewMsCfgType("qr-img1", "qr-img1-reply") //xyzzy - need to set the reply template
	MyId := sizlib.UUIDAsStrPacked()
	hdlr.ms = MicroServiceLib.NewMsCfgType(hdlr.msgTo, hdlr.msgTo+":"+MyId)
	hdlr.myId = MyId
	// ms.ConnectToRedis()                                        // Create the redis connection pool, alternative is ms.SetRedisPool(pool) // ms . SetRedisPool(pool *pool.Pool)
	hdlr.ms.SetRedisConnectInfo(gCfg.RedisConnectHost, gCfg.RedisConnectPort, gCfg.RedisConnectAuth)
	hdlr.ms.SetRedisPool(gCfg.RedisPool)
	hdlr.ms.SetupListen()
	hdlr.ms.ListenFor()
	return
}

func (hdlr *SocketIOHandlerType) PreValidate(gCfg *cfg.ServerGlobalConfigType, cfgData map[string]interface{}, serverName string, pNo, callNo int) (err error) {
	hdlr.mutex.Lock()
	hdlr.LookupIds = make(map[string]*SocketIdType)
	hdlr.LookupTrxIds = make(map[string]string)
	hdlr.mutex.Unlock()

	hdlr.apiEnableIR, _ = lib.ParseBool(hdlr.ApiEnableIR)
	hdlr.apiEnableRR, _ = lib.ParseBool(hdlr.ApiEnableRR)

	server, err := socketio.NewServer(nil)
	if err != nil {
		logrus.Errorf("Error: %s\n", err)
		return mid.ErrInternalError
	}
	hdlr.server = server

	client, err := redis.Dial("tcp", cfg.ServerGlobal.RedisConnectHost+":"+cfg.ServerGlobal.RedisConnectPort)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.ServerGlobal.RedisConnectAuth != "" {
		err = client.Cmd("AUTH", cfg.ServerGlobal.RedisConnectAuth).Err
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintf(os.Stderr, "Success: Connected to redis-server with AUTH.\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "Success: Connected to redis-server.\n")
	}

	hdlr.subClient = pubsub.NewSubClient(client) // subClient *pubsub.SubClient
	hdlr.subChan = make(chan *pubsub.SubResp)

	sr := hdlr.subClient.Subscribe(hdlr.MessageRespPrefix)
	if sr.Err != nil {
		fmt.Fprintf(os.Stderr, "%sError: subscribe, %s.%s\n", MiscLib.ColorRed, sr.Err, MiscLib.ColorReset)
	}

	fmt.Printf("\nListening for messages to send to client at [%s]\n\n", hdlr.MessageRespPrefix)

	hdlr.InitChatServer()
	hdlr.ListenFor(hdlr.server)

	hdlr.server.On("error", func(so socketio.Socket, err error) {
		fmt.Printf("Error: %s, %s\n", err, godebug.LF())
	})

	return
}

var _ mid.GoFTLMiddleWare = (*SocketIOHandlerType)(nil)

// --------------------------------------------------------------------------------------------------------------------------

type SioRoutesType struct {
	RouteName       string // sio://<NAME>/<Path>
	RouteDest       string // chat-bot			-- template, chat-bot:%{Id%}
	RouteReply      string // r:chat-bot		-- item to listen to for replies from this server
	PingMsg         string // /chat-bot-ping	-- Path to send to server to ask if alive
	MsgIfNoPingMsg  string // r:no-server		-- Timeout on PingMessage - results in...
	MsgIfNoPingBody string // NoServerAvailable -- and this for a body
	PingTimeout     uint64 //					-- Time to wait for a ping-reply
}

type SocketIdType struct {
	Id    string
	TrxId string
	So    socketio.Socket
}

type SocketIOHandlerType struct {
	Next              http.Handler                //
	Paths             []string                    // Paths that this will work for
	InternalRequestPw string                      //
	ListClientApi     string                      //
	ApiEnableIR       string                      //
	ApiEnableRR       string                      //
	MessagePrefix     string                      // Template using sizlib.Qt ( quick template ) - the destination to publish to
	MessageRespPrefix string                      // Responce Listen to - where to get back responses - xyzzy - not a template yet - dont't know how to use patterns for listen
	SioRoutes         []SioRoutesType             //
	LineNo            int                         //
	server            *socketio.Server            //
	gCfg              *cfg.ServerGlobalConfigType //
	subClient         *pubsub.SubClient           //
	subChan           chan *pubsub.SubResp        //
	apiEnableIR       bool                        //
	apiEnableRR       bool                        //
	LookupIds         map[string]*SocketIdType    //
	LookupTrxIds      map[string]string           //
	mutex             sync.RWMutex                //
	ms                *MicroServiceLib.MsCfgType  //	MicroServiceLib related	-------------------------------------------------- xyzzy --------------------------------------
	msgTo             string                      //
	myId              string                      //	My ID for receiving replys from micro-services
	// SocketIOLibrary string           // Name of the parameter that is used to get the callback name, default "callback"
}

func NewSocketIOServer(n http.Handler, p []string, pf, pr string) *SocketIOHandlerType {
	return &SocketIOHandlerType{
		Next:              n,
		Paths:             p,
		MessagePrefix:     pf, // "sio:%{Id%}",
		MessageRespPrefix: pr, // "sio:%{Id%}",
		LookupIds:         make(map[string]*SocketIdType),
		LookupTrxIds:      make(map[string]string),
		// SocketIOLibrary: "socket.io.js",
	}
}

func (hdlr *SocketIOHandlerType) ServeHTTP(www http.ResponseWriter, req *http.Request) {
	if pn := lib.PathsMatchN(hdlr.Paths, req.URL.Path); pn >= 0 {
		if rw, ok := www.(*goftlmux.MidBuffer); ok {

			trx := mid.GetTrx(rw)
			trx.PathMatched(1, "SocketIO", hdlr.Paths, pn, req.URL.Path)

			TrxId := rw.Ps.ByName("X-Go-FTL-Trx-Id")
			Id := rw.Ps.ByName("sid")
			hdlr.AssociateTrxId(Id, TrxId)

			fmt.Fprintf(os.Stderr, "%sreq.URL.Path = %s, hdlr.LookupIds=%s, sid=%s, trxid=%s%s\n", MiscLib.ColorGreen, req.URL.Path, lib.SVarI(hdlr.LookupIds), Id, TrxId, MiscLib.ColorReset)
			fmt.Printf("req.URL.Path = %s, hdlr.LookupIds=%s, sid=%s, trxid=%s\n", req.URL.Path, lib.SVarI(hdlr.LookupIds), Id, TrxId)

			hdlr.server.ServeHTTP(rw, req)
			return

		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorRed, mid.ErrNonMidBufferWriter, MiscLib.ColorReset)
			fmt.Printf("%s\n", mid.ErrNonMidBufferWriter)
			www.WriteHeader(http.StatusInternalServerError)
		}
	} else if req.URL.Path == hdlr.ListClientApi && hdlr.apiEnableRR {
		//	4. Login required for external access? -- Could be implemented by "LoginRequried" // xyzzyLoginReq
		// xyzzy - Return MIME type! -> JSON
		s := hdlr.genClientList()
		fmt.Fprintf(www, "%s", s)
		return
	}
	hdlr.Next.ServeHTTP(www, req)

}

func (hdlr *SocketIOHandlerType) InitChatServer() {
	hdlr.server.On("connection", func(so socketio.Socket) {
		hdlr.AssociateId(so.Id(), so)
		hdlr.Publish(so.Id(), `{"msg":"connection"}`)
		fmt.Printf("%sa user connected, Id=%s%s, %s\n", MiscLib.ColorGreen, so.Id(), MiscLib.ColorReset, godebug.LF())
		so.Join("chat")
		so.On("chat message", func(msg string) {
			fmt.Printf("%schat message, %s%s, %s\n", MiscLib.ColorGreen, msg, MiscLib.ColorReset, godebug.LF())
			so.BroadcastTo("chat", "chat message", msg)
		})
		so.On("msg", func(so socketio.Socket, msg string) {
			// xyzzy
			fmt.Fprintf(os.Stderr, "%smsg, %s%s, %s\n", MiscLib.ColorGreen, msg, MiscLib.ColorReset, godebug.LF())

			hdlr.Publish(so.Id(), msg)
		})
		//so.OnAll(func(name, msg string) {
		//	SendMessage ( name, so.Id(), msg )
		//})
		so.On("disconnect", func() {
			hdlr.Publish(so.Id(), `{"msg":"disconnect"}`)
			hdlr.DisassociateId(so.Id()) // xyzzy - hurled on this line -- so nil?
			fmt.Printf("%suser disconnect, Id=%s%s, %s\n", MiscLib.ColorYellow, so.Id(), MiscLib.ColorReset, godebug.LF())
		})
	})
}

func (hdlr *SocketIOHandlerType) AssociateId(Id string, so socketio.Socket) {
	hdlr.mutex.Lock()
	vv, ok := hdlr.LookupIds[Id]
	if ok {
		vv.So = so
		vv.Id = Id
		hdlr.LookupIds[Id] = vv
	} else {
		hdlr.LookupIds[Id] = &SocketIdType{Id: Id, So: so}
		fmt.Fprintf(os.Stderr, "%sError: missing trx id associated with id=%s, %s%s\n", MiscLib.ColorRed, Id, godebug.LF(), MiscLib.ColorReset)
	}
	hdlr.mutex.Unlock()
}

func (hdlr *SocketIOHandlerType) AssociateTrxId(Id, TrxId string) {
	if Id == "" {
		return
	}
	hdlr.mutex.Lock()
	vv, ok := hdlr.LookupIds[Id]
	if ok {
		vv.TrxId = TrxId
		vv.Id = Id
		hdlr.LookupIds[Id] = vv
	} else {
		hdlr.LookupIds[Id] = &SocketIdType{Id: Id, TrxId: TrxId}
	}
	hdlr.LookupTrxIds[TrxId] = Id
	hdlr.mutex.Unlock()

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Infof(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
		return
	}
	defer hdlr.gCfg.RedisPool.Put(conn)
	list, err := conn.Cmd("SMEMBERS", "trx|list").List() // Set key to value
	if err != nil || !lib.InArray(TrxId, list) {
		err = conn.Cmd("SADD", "trx|list", TrxId).Err
	}
	if err != nil {
		logrus.Infof(`{"msg":"Error %s Unable to SADD to trx|list in redis.","LineFile":%q}`+"\n", err, godebug.LF())
		return
	}
}

func (hdlr *SocketIOHandlerType) DisassociateId(id string) {
	if id == "" {
		return
	}

	hdlr.mutex.Lock()
	vv, ok := hdlr.LookupIds[id]
	hdlr.mutex.Unlock()
	if !ok {
		return
	}
	delete(hdlr.LookupTrxIds, vv.TrxId)
	delete(hdlr.LookupIds, id)

	conn, err := hdlr.gCfg.RedisPool.Get()
	if err != nil {
		logrus.Infof(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
		return
	}
	err = conn.Cmd("SREM", "trx|list", vv.TrxId).Err
	if err != nil {
		logrus.Infof(`{"msg":"Error %s Unable to SREM to trx|list in redis.","LineFile":%q}`+"\n", err, godebug.LF())
		return
	}
	hdlr.gCfg.RedisPool.Put(conn)
}

func ProcessTemplate(tmpl string, subs ...string) (rv string) {
	mdata := make(map[string]string)
	for i := 0; i < len(subs); i += 2 {
		s := ""
		if i+1 < len(subs) {
			s = subs[i+1]
		}
		mdata[subs[i]] = s
	}
	rv = sizlib.Qt(tmpl, mdata)
	return
}

func (hdlr *SocketIOHandlerType) GetTrxId(Id string) (TrxId string) {
	hdlr.mutex.RLock()
	vv, ok := hdlr.LookupIds[Id]
	hdlr.mutex.RUnlock()
	if ok {
		TrxId = vv.TrxId
	}
	return
}

func (hdlr *SocketIOHandlerType) Publish(Id, msg string) {
	conn, err := hdlr.gCfg.RedisPool.Get()
	defer hdlr.gCfg.RedisPool.Put(conn)
	if err != nil {
		logrus.Infof(`{"msg":"Error %s Unable to get redis pooled connection.","LineFile":%q}`+"\n", err, godebug.LF())
		return
	}

	TrxId := hdlr.GetTrxId(Id)

	key := ""

	mm, err := lib.JsonStringToData(msg)
	if err != nil {
		fmt.Printf("Note: msg did not parse, %s, %s, %s\n", err, msg, godebug.LF())
		fmt.Fprintf(os.Stderr, "%sError: msg did not parse, %s, %s, %s%s\n", MiscLib.ColorRed, err, msg, godebug.LF(), MiscLib.ColorReset)
		return
	}

	if msg, ok := mm["msg"]; ok {
		switch msg {
		case "connection":
			// possibly add TrxId from set
			return
		case "disconnect":
			// possibly remove TrxId from set
			// hdlr.DisassociateId(Id)
			return
		}
	}

	fmt.Printf("AT: %s\n", godebug.LF())
	mm_To, err := tracerlib.GetFromMapInterface("To", mm)
	if err != nil {
		// {"msg":"connection"} -- can trigger this
		// Should check that chat-bot is live at this point? -- or ignore completely
		fmt.Fprintf(os.Stderr, "%sError: Missing destination, 'to' in [%s] %s%s\n", MiscLib.ColorRed, msg, godebug.LF(), MiscLib.ColorReset)
		return
	}
	u, err := url.Parse(mm_To)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Unable to parse >%s< as destination, err=%s, %s%s\n", MiscLib.ColorRed, mm_To, err, godebug.LF(), MiscLib.ColorReset)
		return
	}

	key = ProcessTemplate(hdlr.MessagePrefix, "Id", Id, "TrxId", TrxId, "Host", u.Host, "Scheme", u.Scheme, "Path", u.Path)

	for ii, vv := range u.Query() {
		fmt.Printf("AT: %s\n", godebug.LF())
		if len(vv) > 1 {
			mm[ii] = lib.SVar(vv)
		} else if len(vv) == 1 {
			mm[ii] = vv[0]
		} else {
			mm[ii] = ""
		}
	}

	mm["Id"] = Id
	mm["TrxId"] = TrxId
	mm["Path"] = u.Path
	mm["Host"] = u.Host
	mm["Scheme"] = u.Scheme

	// xyzzy - this could be a MicroService at this point - durable to a destination. call/response or call/ignore response
	// 1. Add "cmd" to this

	pmsg := lib.SVar(mm)

	if false {

		if db1 {
			fmt.Fprintf(os.Stdout, `PUBLISH "%s" "%s"`+"\n", key, pmsg)
			fmt.Fprintf(os.Stderr, `%sPUBLISH "%s" "%s"`+"%s\n", MiscLib.ColorGreen, key, pmsg, MiscLib.ColorReset)
		}

		err = conn.Cmd("PUBLISH", key, pmsg).Err
		if err != nil {
			logrus.Infof(`{"msg":"Error %s Unable to get publis to redis.","LineFile":%q}`+"\n", err, godebug.LF())
			return
		}

	} else {

		fmt.Printf("Micro Service Decode Version: AT: %s\n", godebug.LF())
		fmt.Fprintf(os.Stderr, `%sMs:Send "%s" "%s"`+"%s\n", MiscLib.ColorGreen, key, pmsg, MiscLib.ColorReset)

		Body := ""
		if t, ok := mm["body"]; ok {
			if s, ok := t.(string); ok {
				Body = s
			}
		}

		User := ""
		if t, ok := mm["user"]; ok {
			if s, ok := t.(string); ok {
				User = s
			}
		}

		// func (ms *MsCfgType) SetReplyFunc(fx func(replyMessager *MsMessageToSend)) {
		hdlr.ms.SetReplyFunc(func(replyMessage *MicroServiceLib.MsMessageToSend) {
			fmt.Printf("In Reply Function!!!, message = %s, IsTimeout=%v, AT:%s\n", godebug.SVarI(replyMessage), replyMessage.IsTimeout, godebug.LF())
			err := MicroServiceLib.GetParam(replyMessage.Params, "Error")
			data := MicroServiceLib.GetParam(replyMessage.Params, "Data")
			if err != "" {
				fmt.Printf(`{"status":"error","id":%q,"msg":%q}`, Id, err)
			} else {
				fmt.Printf(`{"status":"success","id":%q,"results":%q}`, Id, data)
			}
		})

		fmt.Printf("AT: %s\n", godebug.LF())

		id := sizlib.UUIDAsStrPacked()
		hdlr.ms.SendMessage(&MicroServiceLib.MsMessageToSend{
			To: hdlr.msgTo,
			Id: id,
			Params: []MicroServiceLib.MsMessageParams{
				MicroServiceLib.MsMessageParams{Name: "cmd", Value: "send"},
				MicroServiceLib.MsMessageParams{Name: "id", Value: Id},
				MicroServiceLib.MsMessageParams{Name: "TrxId", Value: TrxId},
				MicroServiceLib.MsMessageParams{Name: "Path", Value: u.Path},
				MicroServiceLib.MsMessageParams{Name: "Host", Value: u.Host},
				MicroServiceLib.MsMessageParams{Name: "Scheme", Value: u.Scheme},
				MicroServiceLib.MsMessageParams{Name: "Body", Value: Body},
				MicroServiceLib.MsMessageParams{Name: "User", Value: User},
				MicroServiceLib.MsMessageParams{Name: "To", Value: mm_To},
				// MicroServiceLib.MsMessageParams{Name: "NoReply", Value: "true"},
			},
		})

		fmt.Printf("Micro Service Version: AT: %s\n", godebug.LF())
		return
	}
}

func (hdlr *SocketIOHandlerType) LookupSocketFromId(id string) (so socketio.Socket, err error) {
	hdlr.mutex.RLock()
	t, ok := hdlr.LookupIds[id]
	hdlr.mutex.RUnlock()
	if !ok {
		err = errors.New("Unable to find socket Id")
		return
	}
	so = t.So
	return
}

// http://redis.io/commands/SMEMBERS
//	SADD, SREM

func (hdlr *SocketIOHandlerType) ListenFor(server *socketio.Server) {

	go func() {
		for {
			hdlr.subChan <- hdlr.subClient.Receive()
		}
	}()

	go func() {
		var sr *pubsub.SubResp
		for {
			select {
			case sr = <-hdlr.subChan:
				if db1 {
					// sr={"Type":3,"Channel":"listen","Pattern":"","SubCount":0,"Message":"abcF","Err":null}
					fmt.Printf("***************** Got a message, sr=%+v\n", godebug.SVar(sr))
					fmt.Printf("\nMessage --->>>%s<<<---\n\n", sr.Message)
					// Message --->>>{"msg":"msg", "Id":"UwjkIGvXWMmB1nB-bJ9t", "body":{"msg":"connection"} }<<<---
				}

				var mm map[string]interface{}
				err := json.Unmarshal([]byte(sr.Message), &mm)
				if err != nil {
					fmt.Printf("AT: %s\n", godebug.LF())
					fmt.Printf("Error: %s --->>>%s<<<--- AT: %s\n", err, sr.Message, godebug.LF())
					fmt.Fprintf(os.Stderr, "%sError: %s --->>>%s<<<--- AT: %s%s\n", MiscLib.ColorYellow, err, sr.Message, godebug.LF(), MiscLib.ColorReset)
				} else {

					if db1 {
						fmt.Printf("Sending to client data in message %s, %s\n", lib.SVar(mm), godebug.LF())
					}

					mm_To, err := tracerlib.GetFromMapInterface("To", mm)
					u, err := url.Parse(mm_To)
					if err != nil {
						fmt.Printf("Error parsing >%s< - invalid To destiation url, %s, %s\n", mm_To, err, godebug.LF())
					} else {

						q := u.Query()
						mm_Pw := q.Get("Pw")
						mm_TrxId := q.Get("TrxId")
						mm_Id := q.Get("Id")
						mm_ReplyTo := q.Get("ReplyTo")

						// xyzzy - parse body - combine
						if mm_TrxId == "" {
							mm_TrxId, _ = tracerlib.GetFromMapInterface("TrxId", mm)
						}
						if mm_Id == "" {
							mm_Id, _ = tracerlib.GetFromMapInterface("Id", mm)
						}

						fmt.Fprintf(os.Stderr, "Pw [%s] TrxId [%s] Id [%s] ReplyTo [%s] AT: %s\n", mm_Pw, mm_TrxId, mm_Id, mm_ReplyTo, godebug.LF())

						// xyzzy - shold lookup u.Scheme -> config to find real destination for sending stuff to - then template process it.

						// "ListClientApi":      { "type":[ "string" ], "default":"/api/sio/list-client" },
						// sio://server/api/sio/list-client?Pw=xxxx
						// if mm_Pw == hdlr.InternalRequestPw && u.Path == hdlr.ListClientApi && hdlr.apiEnableIR && u.Host == "server" && u.Scheme == "sio" {
						if mm_Pw == hdlr.InternalRequestPw && u.Path == hdlr.ListClientApi && u.Host == "server" && u.Scheme == "sio" {
							fmt.Printf("AT: %s\n", godebug.LF())
							fmt.Fprintf(os.Stderr, "Respond to /api call, %s\n", godebug.LF())
							s := hdlr.genClientList()
							hdlr.Publish(mm_ReplyTo, s)
						} else {
							fmt.Printf("AT: %s\n", godebug.LF())

							// sio://server/typing?Name=bob&TrxId=1111
							// sio://server/display?Name=bob&TrxId=1111&Msg=Yo+Dude!
							// sio://server/display   body==>{ "Name":..., "TrxId":..., "Msg":... }

							hdlr.mutex.RLock()
							Id, ok := hdlr.LookupTrxIds[mm_TrxId] // Lookup using TrxId
							hdlr.mutex.RUnlock()
							if !ok {
								fmt.Printf("AT: %s\n", godebug.LF())
								Id = mm_Id
							}

							fmt.Fprintf(os.Stderr, "Emit to Id=%s, AllData=%s, TrxId=%s, %s\n", Id, lib.SVar(mm), mm_TrxId, godebug.LF())

							so, err := hdlr.LookupSocketFromId(Id)
							if err != nil {
								fmt.Printf("AT: %s\n", godebug.LF())
								fmt.Fprintf(os.Stderr, "%sError: %s id=%s AT: %s%s\n", MiscLib.ColorRed, err, mm_Id, godebug.LF(), MiscLib.ColorReset)
								fmt.Printf("Error: %s id=%s AT: %s\n", err, mm_Id, godebug.LF()) // xyzzy Should log this
							} else {
								fmt.Printf("AT: %s\n", godebug.LF())
								fmt.Fprintf(os.Stderr, "Doing emit [%s] [%s]\n", u.Path, lib.SVar(mm)) // xyzzy Should log this
								so.Emit(u.Path, lib.SVar(mm))
							}

						}
					}
				}
			}
		}
	}()

}

func (hdlr *SocketIOHandlerType) genClientList() (s string) {
	s = `{ "data":[ ` + "\n"
	com := " "
	for _, vv := range hdlr.LookupIds {
		s += fmt.Sprintf(`    %s{ "Id":%q, "TrxId":%q }`+"\n", com, vv.Id, vv.TrxId) // xyzzyTime
		com = ","
	}
	s += `]` + "\n"
	// xyzzy - add in hdlr.LookupTrxIds -> this
	s += fmt.Sprintf(` , "trx":%s `, lib.SVarI(hdlr.LookupTrxIds))
	s += `}` + "\n"
	return
}

const db1 = true

/* vim: set noai ts=4 sw=4: */
