package plugins

import (
	"context"
	"fmt"
	plugin "github.com/GhostLee/memo/bot/plugins/proto"
	qqbotapi "github.com/catsworld/qq-bot-api"
	"google.golang.org/grpc"
	"log"
	"time"
)

type Plugin struct {
	InstallAt time.Time
	UpdateAt  time.Time
	name      string
	port      int
	host      string
	enable    bool
	instance  plugin.PluginClient
	noticeEvents map[string]qqbotapi.Unsubscribe
	messageEvents map[string]qqbotapi.Unsubscribe
	applicationEvents map[string]qqbotapi.Unsubscribe
	eventManager *qqbotapi.Ev
}

func (p *Plugin) RemoteAddress() string {
	return fmt.Sprintf("%v:%v", p.host, p.port)
}

func NewPluginClient(name, host string, port int, noticeEvents, messageEvents, applicationEvents []string, ev *qqbotapi.Ev) *Plugin {
	var notice = make(map[string]qqbotapi.Unsubscribe)
	for _, e := range noticeEvents {
		notice[e] = nil
	}
	var message = make(map[string]qqbotapi.Unsubscribe)
	for _, e := range messageEvents {
		message[e] = nil
	}
	var application = make(map[string]qqbotapi.Unsubscribe)
	for _, e := range applicationEvents {
		application[e] = nil
	}
	return &Plugin{
		name:     name,
		host:     host,
		port:     port,

		instance: nil,
		noticeEvents: notice,
		messageEvents:message,
		applicationEvents:application,
		eventManager: ev,
	}
}

func (p *Plugin) Install() error {
	conn, err := grpc.Dial(p.RemoteAddress(), grpc.WithInsecure())
	if err != nil {
		log.Printf("grpc.Dial err: %v", err)
		return err
	}
	p.InstallAt = time.Now()
	p.instance = plugin.NewPluginClient(conn)
	for e := range p.messageEvents {
		p.messageEvents[e] = p.eventManager.On(e)(p.HandleMessage)
	}
	for e := range p.applicationEvents {
		p.applicationEvents[e] = p.eventManager.On(e)(p.HandleApplication)
	}
	for e := range p.noticeEvents {
		p.noticeEvents[e] = p.eventManager.On(e)(p.HandleNotice)
	}
	return nil
}
func (p *Plugin) Log(v interface{}) {
	log.Printf("In Plugin :%v----> %v", p.name, v)
}
func (p *Plugin) HandleMessage(update qqbotapi.Update) {
	var typ plugin.Type
	var subtype plugin.SubType
	p.Log(update)
	switch update.MessageType {
	case "private":
		typ = plugin.Type_PrivateChat
		switch update.SubType {
		case "friend":
			subtype = plugin.SubType_Friend
		case "group":
			subtype = plugin.SubType_Group
		case "discuss":
			subtype = plugin.SubType_Discuss
		default:
			subtype = plugin.SubType_Other
		}
	case "group":
		typ = plugin.Type_GroupChat
		switch update.SubType {
		case "normal":
			subtype = plugin.SubType_Normal
		case "anonymous":
			subtype = plugin.SubType_Anonymous
		case "notice":
			subtype = plugin.SubType_Notice
		default:
			subtype = plugin.SubType_Other
		}
	case "discuss":
		typ = plugin.Type_DiscussChat
	default:
		typ = plugin.Type_UnknownChat
	}
	req := plugin.HandleMessageRequest{
		Bot:       0,
		Time:      update.Time,
		MessageId: update.MessageID,
		Type:      typ,
		SubType:   subtype,
		UserId:    update.UserID,
		GroupId:   update.GroupID,
		DiscussId: update.DiscussID,
		Text:      update.Text,
	}

	//todo 对应的错误处理
	_, _ = p.instance.HandleMessage(context.Background(), &req)
}
func (p *Plugin) HandleApplication(update qqbotapi.Update) {
	var typ plugin.ApplicationType
	p.Log(update)
	switch update.RequestType {
	case "friend":
		typ = plugin.ApplicationType_FriendAdd
	case "group":
		switch update.SubType {
		case "add":
			typ = plugin.ApplicationType_GroupAdd
		case "invite":
			typ = plugin.ApplicationType_GroupInvite
		default:
			return
		}
	default:
		return
	}
	req := plugin.HandleApplicationRequest{
		Type:    typ,
		UserId:  update.UserID,
		GroupId: update.GroupID,
		Comment: update.Comment,
		Flag:    update.Flag,
	}
	//todo 对应的错误处理
	_, _ = p.instance.HandleApplication(context.Background(), &req)
}
func (p *Plugin) HandleNotice(update qqbotapi.Update) {
	p.Log(update)
	log.Println("Not Impl")
}
func (p *Plugin) Uninstall() error {
	for e, unsubscribe := range p.messageEvents {
		log.Printf("退订%v插件%v事件", p.name, e)
		unsubscribe()
	}
	for e, unsubscribe := range p.applicationEvents {
		log.Printf("退订%v插件%v事件", p.name, e)
		unsubscribe()
	}
	for e, unsubscribe := range p.noticeEvents {
		log.Printf("退订%v插件%v事件", p.name, e)
		unsubscribe()
	}
	return nil
}
