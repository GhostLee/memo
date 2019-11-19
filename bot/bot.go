package main

import (
	"context"
	"errors"
	"github.com/GhostLee/memo/bot/plugins"
	proto "github.com/GhostLee/memo/bot/proto"
	qqbotapi "github.com/catsworld/qq-bot-api"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"
)

type Bot struct {
	sendEnable bool
	instance *qqbotapi.BotAPI
	cache chan *proto.Message
	rpc *grpc.Server
	//cache *queue.MessageQueue
	wait time.Duration
	plugins map[string]*plugins.Plugin
	ev *qqbotapi.Ev
	secret string
	quit chan bool
}
func(s *Bot) Send(ctx context.Context, req *proto.SendRequest) (*proto.SendResponse, error){
	if req.Secret != s.secret{
		return &proto.SendResponse{Status: false}, errors.New("验证失败")
	}
	if !s.sendEnable {
		return &proto.SendResponse{Status: false}, errors.New("不允许发送消息")
	}
	for _, msg := range req.Messages{
		//log.Println("接收到消息...")
		s.cache <- msg
	}
	return &proto.SendResponse{Status: true}, nil
}
func(s *Bot) PluginRegister(ctx context.Context, req *proto.PluginRegisterRequest) (*proto.PluginRegisterResponse, error){
	if req.Secret != s.secret {
		return &proto.PluginRegisterResponse{Status:false}, nil
	}
	plug := plugins.NewPluginClient(
		req.Name,
		req.Host,
		int(req.Port),
		req.NoticeEvents,
		req.MessageEvents,
		req.ApplicationEvents,
		s.ev)
	err := plug.Install()
	if err != nil {
		log.Println(err)
		plug = nil
		return &proto.PluginRegisterResponse{
			Status:               false,
		}, nil
	}
	if old, exist := s.plugins[req.Name]; exist {
		log.Println("即将覆盖插件")
		_ = old.Uninstall()
		s.plugins[req.Name] = plug
	} else {
		s.plugins[req.Name] = plug
	}
	log.Printf("加载插件%v", req.Name)
	log.Printf("已注册插件%v", req.Name)
	for _, e := range req.MessageEvents{
		log.Printf("注册消息事件: %v", e)
	}
	for _, e := range req.ApplicationEvents{
		log.Printf("注册请求事件: %v", e)
	}
	for _, e := range req.NoticeEvents{
		log.Printf("注册通知事件: %v", e)
	}
	return &proto.PluginRegisterResponse{
		Status:               true,
	},nil
}
func(s *Bot) PluginUnregister(ctx context.Context, req *proto.PluginUnregisterRequest) (*proto.PluginUnregisterResponse, error){
	if old, exist := s.plugins[req.Name]; exist {
		log.Printf("卸载插件%v", req.Name)
		_ = old.Uninstall()
		delete(s.plugins, req.Name)
	} else {
		log.Printf("未找到插件%v", req.Name)
	}
	return &proto.PluginUnregisterResponse{
		Status:               true,
	},nil
}
func(s *Bot) HandleFriendApplication(ctx context.Context, req *proto.HandleFriendApplicationRequest) (*proto.HandleFriendApplicationResponse, error){
	if req.Secret != s.secret {
		return &proto.HandleFriendApplicationResponse{}, nil
	}
	rsp, err:=s.instance.HandleFriendRequest(req.Flag, req.Approve, req.Remark)
	if err != nil {
		log.Println(err)
	}
	log.Println(rsp)
	return &proto.HandleFriendApplicationResponse{}, nil
}
func(s *Bot) HandleGroupApplication(ctx context.Context, req *proto.HandleGroupApplicationRequest) (*proto.HandleGroupApplicationResponse, error){
	if req.Secret != s.secret {
		return &proto.HandleGroupApplicationResponse{}, nil
	}
	var typ string
	switch req.Type {
	case proto.GroupApplicationType_Add:
		typ = "add"
	case proto.GroupApplicationType_Invite:
		typ = "invite"
	default:
		return &proto.HandleGroupApplicationResponse{}, nil
	}
	rsp, err := s.instance.HandleGroupRequest(req.Flag, typ, req.Approve, req.Reson)
	if err != nil {
		log.Println(err)
	}
	log.Println(rsp)
	return &proto.HandleGroupApplicationResponse{}, nil
}
func (s *Bot) sender() {
	for  {
		select {
		case msg,ok := <- s.cache:
			if !ok{
				log.Println("缓存通道关闭")
				return
			}
			switch msg.GetType() {
			case proto.MessageType_Private:
				_,_ = s.instance.SendMessage(msg.Id, "private", msg.Text)
			case proto.MessageType_Group:
				_,_ = s.instance.SendMessage(msg.Id, "group", msg.Text)
			case proto.MessageType_Discuss:
				_,_ = s.instance.SendMessage(msg.Id, "discuss", msg.Text)
			}
			time.Sleep(s.wait * time.Duration(rand.Intn(2)))
			break
		case <-s.quit:
			log.Println("准备退出...")
			return
		}
	}

}

func (s *Bot) Quit() {
	close(s.cache)
	log.Println("准备关闭所有等待...")
	s.rpc.GracefulStop()
	s.quit <- true
}
func AddFriend(update qqbotapi.Update)  {
	log.Println(update.Comment)
	//bot.instance.HandleGroupRequest()
	//VerifyAddFriendRequest(update.UserID, update.Comment)
}
func (s *Bot) Run() {
	var err error
	instance, err := qqbotapi.NewBotAPI("mBPNKvMx12mJ", "ws://120.78.79.117:5700", "")
	if err != nil {
		log.Fatal(err)
	}
	instance.Debug = true
	u := qqbotapi.NewUpdate(0)
	u.PreloadUserInfo = true
	s.instance = instance
	updates, err := s.instance.GetUpdatesChan(u)

	if err != nil {
		log.Fatal(err)
	}
	s.ev = qqbotapi.NewEv(updates)
	// ---------------------------------------------------------------
	//s.ev.On("message.group.normal")(s.Echo)
	//s.ev.On("request.friend")(AddFriend)

	proto.RegisterBotServer(s.rpc, s)

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	log.Println("运行中...")
	go s.rpc.Serve(lis)
	go s.sender()
	<-s.quit
	log.Println("准备停止...")
}
var bot Bot

	//ev.On("message.group.normal")(Echo)
	//// Function Log will get triggered on receiving an update with
	//// PostType `message`
	//ev.On("message")(Log)
	//
	//ev.On("request.friend")(AddFriend)

func main(){
	log.SetFlags(log.Lshortfile)
	c := make(chan os.Signal)
	// 监听所有信号
	signal.Notify(c)
	bot = Bot{
		sendEnable: true,
		rpc: grpc.NewServer(),
		cache: make(chan *proto.Message, 5000),
		plugins: map[string]*plugins.Plugin{},
		wait:     time.Millisecond * 5,
		secret:   "94QfvK6HPYi8UDeE",
		quit: make(chan bool),
	}

	go bot.Run()
	<-c
	log.Println("执行退出前清理程序")
	defer bot.Quit()
}

//func main() {
//	puppeteer.GetInstance().Register(user.Topic, user.Instance)
//	puppeteer.GetInstance().Register(leosys.Topic, leosys.Instance)
//	go puppeteer.GetInstance().Listen()
//	server.Run()
//}