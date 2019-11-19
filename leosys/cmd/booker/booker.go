package main

import (
	"context"
	"fmt"
	proto "github.com/GhostLee/deno/rpc/service/leosys_booker/proto"
	"github.com/GhostLee/memo/leosys/library"
	"github.com/tencentyun/scf-go-lib/cloudevents/scf"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var server = os.Getenv("server")
var bookerID = os.Getenv("id")
//var bookerID = "47e7c236-6a82-4f8b-9eec-122d3ef936e5"
var manager proto.LeoSysClient
var tasks []*Task
var results []*proto.BookResult
var MaxRetry = 5
var checkDuration = time.Millisecond * 200
var tomorrow = time.Now().AddDate(0,0,1).Format("2006-01-02")
var timeout = struct {
	Hour int
	Minute int
	Second int
}{7, 4, 0}
type Task struct {
	proto.Account
	stop bool
	result uint
	message string
	client    *library.User
}

// 设置停止
func (t *Task) SetStop(code uint, msg string)  {
	t.stop = true
	t.result = code
	t.message = msg
}
// 登陆
func (t *Task) Login() {
	var retry = 0
	//设置初值
	log.Printf("正在为用户%v的%v登陆,准备预约座位ID: %v的%v点到%v点", t.User, t.Username, t.Seat, t.Start, t.End)
	t.client = library.NewLibUser(t.Username, t.Password)
	// ---------------------------------Login--------------------------------------------
LOGIN:
	rsp, err := t.client.Login()
	if err != nil {
		if rsp != nil {
			code, err := strconv.ParseUint(rsp.Code, 10, 64)
			if err != nil {
				code = 2
			}
			log.Printf("账号%v登陆时出现异常,错误信息:%v", t.Username, rsp.Message)
			t.SetStop(uint(code), rsp.Message)
			return
		} else {
			retry++
			if retry < MaxRetry {
				goto LOGIN
			}
			log.Printf("账号%v登陆时出现异常,错误信息:%v", t.Username, err.Error())
			t.SetStop(0, err.Error())
			return
		}
	}
	return
}
// 预约
func (t *Task) Book() {
	rsp, err := t.client.FreeBook(
		int(t.Seat),
		strconv.FormatUint(uint64(t.Start*60), 10),
		strconv.FormatUint(uint64(t.End*60), 10),
		tomorrow,
		)
	if err != nil {
		log.Println(err)
		t.SetStop(0, err.Error())
		return
	}
	if rsp == nil{
		t.SetStop(0, "解析预约结果返回值出现错误!请登录APP查看预约结果!")
		return
	}
	if rsp.Status != "success" {
		code, err := strconv.ParseUint(rsp.Code, 10, 64)
		if err != nil {
			code = 2
		}
		t.SetStop(uint(code), rsp.Message)
	} else {
		//"onDate": "2018 年 06 月 21 日",
		//	"begin": "20 : 00",
		//	"end": "22 : 00",
		//	"location": "东校区2层二楼大厅区二楼大厅，座位号024",
		t.SetStop(1, fmt.Sprintf("为账号%v预约%v %v从%v到%v成功!",t.Username, rsp.Data.OnDate, rsp.Data.Location, rsp.Data.Begin, rsp.Data.End))
	}
}
func (t *Task) IsReady() bool {
	if t.stop {
		log.Printf("账号%v被停止\n", t.Username)
		return false
	}
	time.Sleep(checkDuration+time.Duration(rand.Intn(50)) * time.Millisecond)
	return t.client.CheckAvailable()
}
// 中途是否出现失败
func (t *Task) Stopped() bool{
	return t.stop
}

func (t *Task) Result() *proto.BookResult {
	var code proto.Status
	switch t.result {
	case 0:
		code = proto.Status_Fail
	case 1:
		code = proto.Status_Success
	default:
		code = proto.Status_UserError
	}
	return &proto.BookResult{
		Username:             t.Username,
		UserId:               t.User,
		Qq:                   t.Qq,
		SeatIndex:            t.Seat,
		Result:               code,
		Msg:                  t.message,

	}
}
// 进行初始化操作
func handler(ctx context.Context, event scf.APIGatewayProxyRequest) error {
	return worker()
}

func worker() error {
	var err error
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("grpc.Dial err: %v", err)
		return err
	}
	manager = proto.NewLeoSysClient(conn)
	log.Printf("当前Booker使用的公网IP为%v", getExternalIP())
	// 执行操作逻辑
	Run()
	_ = conn.Close()
	return nil
}
func main() {
	cloudfunction.Start(handler)
	//log.SetFlags(log.Lshortfile)
	//log.Println(worker())
}
// 将manager返回的account加工成task
func generateTask(account proto.Account)  {
	tasks = append(tasks, &Task{Account:account, stop: false})
}
// 全局检查可预约状态
func CheckReady() bool {
	// 检查登陆失败的账号
	var stopped = 0
	for _, task := range tasks  {
		if task.Stopped(){
			stopped ++
		}
	}
	if stopped == len(tasks){
		log.Println("该节点所有预约账号均存在登陆异常状况")
		return false
	}
	for  {
		for _, task := range tasks  {
			var now = time.Now()
			log.Printf("正在使用账号%v检查图书馆可预约状态!检查时间:%v", task.Username, now.Format("2006-01-02 15:04:05"))
			if task.IsReady(){
				return true
			}

			if now.Hour() == timeout.Hour && now.Minute() == timeout.Minute && now.Second() == timeout.Second{
				log.Println("检查可预约状态超时,图书馆服务出现问题,正在设置账户状态以通知用户")
				for _, t := range tasks {
					t.SetStop(2, "检查可预约状态超时,图书馆预约服务可能存在宕机问题,请登陆APP查看利昂软件是否存在异常!")
				}
				return false
			}
		}
	}

}

func Run() {
	rsp, err := manager.GetTasks(context.Background(), &proto.GetTasksRequest{
		Id:                   bookerID,
	})
	if err != nil {
		log.Println(err)
		// todo: 处理请求错误
	}

	accounts := rsp.Accounts
	if len(accounts) == 0 {
		log.Println("当日无预约任务!")
		return
	}
	for _, account := range accounts  {
		generateTask(*account)
	}
	for _, task := range tasks{
		task.Login()
	}
	if !CheckReady(){
		log.Println("检查状态出现意外!准备提交失败结果!")
	} else {
		var wg sync.WaitGroup
		for _, task := range tasks{
			go func(t *Task) {
				wg.Add(1)
				defer wg.Done()
				t.Book()
			}(task)
		}
		wg.Wait()
	}
	for _, task := range tasks{
		results = append(results, task.Result())
	}
	_, err = manager.Finish(context.Background(), &proto.FinishRequest{
		Id:                   bookerID,
		Results:              results,
	})
	log.Printf("本次预约共计预约账号%v个,成功%v个,失败%v个", len(tasks), )
	if err != nil {
		log.Println("上传结果时出现错误!错误详情: ", err.Error())
	} else {
		log.Println("预约结果已上传到Manager!")
	}
	//for _, result := range results{
	//	log.Printf("用户%v的账号%v预约座位ID%v的预约状态为%v\n即将通过%v推送结果\n%v",
	//		result.UserID,
	//		result.Username,
	//		result.SeatIndex,
	//		result.Code,
	//		result.QQ,
	//		result.Msg,
	//		)
	//}
}

//获取公网ip
func getExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(resp.Body)
	//s := buf.String()
	return string(content)
}