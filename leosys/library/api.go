package library

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)
const PASSWORD = "leos3cr3t"

// 加密算法
func getHmacEncrypt(id, method string, timestamp int64) string {
	message := "seat::%v::%v::%v" // seat::id::timestamp::method
	h := hmac.New(sha256.New, []byte(PASSWORD))
	_, _ = io.WriteString(h, fmt.Sprintf(message, id, timestamp, method))
	return fmt.Sprintf("%x", h.Sum(nil))
}

type User struct {
	account  string
	password string
	token    string
	client   *http.Client
}

func NewLibUser(account, password string) *User {
	return &User{
		account:  account,
		password: password,
		client:   &http.Client{
			Transport:&http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}


//>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
//<<<<<<<<<<<<<<<<<<<<<<<<<<<<
//POST /rest/v2/freeBook HTTP/1.1
//user-agent: Dart/2.1 (dart:io)
//x-hmac-request-key: 8d971a4677e59e5eaa431e5574e3d87653dbacb7aa971f37457c7b85f2733141
//transfer-encoding: chunked
//x-request-date: 1557321743444
//				  1559402479000
//accept-encoding: gzip
//host: seat.ujn.edu.cn:8443
//content-type: application/x-www-form-urlencoded
//x-request-id: 50b1d140-5970-11e9-e085-bfcd8b9885ce
//
//4D
//token=A8TWODGP1C05082046&seat=22559&startTime=-1&endTime=1320&date=2019-05-08
//0

func request_call(client *http.Client, method string, request_url string, data url.Values, header http.Header) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("在请求", request_url, "时出现错误!")
		}
	}()
	req, err := http.NewRequest(method, request_url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.Header.Set("Host","seat.ujn.edu.cn:8443")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Type", "charset=UTF-8")
	req.Header.Set("user-agent", "Dart/2.1 (dart:io)",)
	req.Header.Set("Accept-Encoding", "gzip")
	//req.Header.Add("X-Forwarded-For", "10.167.159.118") // todo 原先服务器使用xff进行判断

	//todo: random request id
	id := uuid.NewV1()

	requestID := id.String()
	requestDate := time.Now().Unix() * 1000

	req.Header.Set("x-request-date", strconv.FormatInt(requestDate, 10))
	req.Header.Set("x-request-id", requestID)
	switch method {
	case http.MethodGet:
		req.Header.Set("x-hmac-request-key", getHmacEncrypt(requestID, http.MethodGet, requestDate),)
		break
	case http.MethodPost:
		req.Header.Set("x-hmac-request-key", getHmacEncrypt(requestID, http.MethodPost, requestDate),)
		break
	case http.MethodHead:
		req.Header.Set("x-hmac-request-key", getHmacEncrypt(requestID, http.MethodHead, requestDate),)
		break
	}
	// Warning: 保证请求之后断开连接
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (u *User) request(method string, request_url string, data url.Values) ([]byte, error) {
	header := http.Header{}
	//header.Add("X-Real-IP", "10.167.129.61")
	//header.Add("X-Forwarded-For", "10.167.129.61")
	header.Set("token", u.token)
	return request_call(u.client, method, request_url, data, header)
}

func (u *User) Login() (*TokenResponse, error) {
	body, err := u.request(http.MethodGet, fmt.Sprintf(GET_TOKEN_URL, u.account, u.password), nil)
	if err != nil {
		return nil, err
	}
	var response TokenResponse
	if err := json.Unmarshal(body, &response); err != nil{
		return nil, err
	}
	if response.Status != "success"{
		return &response, errors.New(response.Message)
	}
	u.token = response.Data.Token
	return &response, nil
}

/* User JSON
{
	"status": "success",
	"data": {
		"id": 7074,
		"enabled": true,
		"name": "我的名字",
		"username": "我的学号",
		"username2": null,
		"status": "NORMAL",
		"lastLogin": "2018-06-20T09:22:34.000",
		"checkedIn": false,
		"lastIn": null,
		"lastOut": null,
		"lastInBuildingId": null,
		"lastInBuildingName": null,
		"violationCount": 0
	},
	"message": "",
	"code": "0"
}*/
func (u *User) User() (*UserResponse, error) {
	body, err := u.request("GET", USER_URL, nil)
	if err != nil {
		return nil, err
	}
	var response UserResponse
	if err := json.Unmarshal(body, &response); err != nil{
		return nil, err
	}
	return &response, nil
}

/* Filters JSON
{
	"status": "success",
	"data": {
		"buildings": [
			[1, "东校区", 5],
			[2, "主校区", 8]
		],
		"rooms": [
			[8, "第五阅览室北区", 2, 6],
			[9, "第八阅览室北区", 2, 6],
			[11, "第二阅览室北区", 2, 3],
			[12, "第二阅览室中区", 2, 3],
			[13, "第二阅览室南区", 2, 3],
			[14, "第十一阅览室北区", 2, 3],
			[15, "第十一阅览室中区", 2, 3],
			...........................
			[RoomID, "room name", BuildingID, Floor]
			[62, "文化展厅（一楼北）", 1, 1]
		],
		"hours": 15,
		"dates": ["2018-06-21", "2018-06-22"]
	},
	"message": "",
	"code": "0"
}*/
func (u *User) Filters() ([]byte, error) {
	body, err := u.request("GET", FILTERS_URL, nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* RoomStats JSON
{
	"status": "success",
	"data": [{
		"roomId": 41,
		"room": "第一阅览室",
		"floor": 2,
		"maxHour": -1,
		"reserved": 1,
		"inUse": 133,
		"away": 0,
		"totalSeats": 136,
		"free": 2
	}, {
        ................
		................
        ................
	},	{
		"roomId": 25,
		"room": "第六阅览室南区",
		"floor": 7,
		"maxHour": -1,
		"reserved": 2,
		"inUse": 83,
		"away": 0,
		"totalSeats": 108,
		"free": 22
	}],
	"message": "",
	"code": "0"
}*/
func (u *User) RoomStats(buildingID int) ([]byte, error) {
	body, err := u.request("GET", fmt.Sprintf(ROOM_STATS_URL, buildingID), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* LayoutByDate JSON
{
	"status": "success",
	"data": {
		"id": 54,
		"name": "二楼大厅",
		"cols": 24,
		"rows": 8,
		"layout": {
			"0": {
				"type": "empty"
			},
			.........
			"2002": {
				"id": 41087,
				"name": "001",
				"type": "seat",
				"status": "FREE",
				"window": false,
				"power": false,
				"computer": false,
				"local": false
			},
			.........
			"3005": {
				"id": 4401,
				"name": "235",
				"type": "seat",
				"status": "IN_USE",
				"window": false,
				"power": false,
				"computer": false,
				"local": false
			},
			.........
			"3007": {
				"name": "架",
				"type": "word"
			},
			.........
			"20008": {
				"type": "empty"
			}
		}
	},
	"message": "",
	"code": "0"
}*/
func (u *User) LayoutByDate(roomID int, date string) ([]byte, error) {
	body, err := u.request("GET", fmt.Sprintf(LAYOUTBYDATE_URL, roomID, date), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* Seat start time JSON
{
	"status": "success",
	"data": {
		"startTimes": [{
			"id": "420",
			"value": "07:00"
		} ,{
			..............
		} ,{
			"id": "1260",
			"value": "21:00"
		}]
	},
	"message": "",
	"code": "0"
}
*/
func (u *User) SeatStartTime(seatID int, date string) ([]byte, error) {
	body, err := u.request("GET", fmt.Sprintf(SEAT_START_TIME_URL, seatID, date), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* Seat end time JSON
{
	"status": "success",
	"data": {
		"endTimes": [{
			"id": "480",
			"value": "08:00"
		} ,{
			..............
		} ,{
			"id": "1320",
			"value": "22:00"
		}]
	},
	"message": "",
	"code": "0"
}*/
func (u *User) SeatEndTime(seatID int, date, startTime string) ([]byte, error) {
	body, err := u.request("GET", fmt.Sprintf(SEAT_END_TIME_URL, seatID, date, startTime), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* Free Book JSON
{
	"status": "success",
	"data": {
		"id": 4309649,
		"receipt": "2177-649-4",
		"onDate": "2018 年 06 月 21 日",
		"begin": "20 : 00",
		"end": "22 : 00",
		"location": "东校区2层二楼大厅区二楼大厅，座位号024",
		"checkedIn": false,
		"checkInMsg": "当前没有可用预约"
	},
	"message": "",
	"code": "0"
}
*/
func (u *User) FreeBook(seatID int, startTime, endTime, date string) (*FreeBookResponse, error) {
	retryTimes := 0
	data := url.Values{}
	seat := strconv.Itoa(seatID)
	// data.Set("t", "1")
	data.Set("startTime", startTime)
	data.Set("endTime", endTime)
	data.Set("seat", seat)
	data.Set("date", date)
	// data.Set("t2", "2")
// todo: 重试次数
RetryLabel:
	body, err := u.request("POST", FREEBOOK_URL, data)
	if err != nil {
		if retryTimes <= 100{
			retryTimes ++
			time.Sleep(time.Millisecond*5)
			goto RetryLabel
		}
		return nil, err
	}
	var response FreeBookResponse
	if err := json.Unmarshal(body, &response); err != nil{
		return nil, err
	}
	if response.Status != "success"{
		return &response, errors.New(response.Message)
	}
	return &response, nil
}

/* Check in JSON
 */
func (u *User) CheckIn() (*CheckInResponse, error) {
	t1 := time.Now()
	body, err := u.request("GET", CHECKIN_URL, nil)
	if err != nil {
		return nil, err
	}
	var response CheckInResponse
	if err := json.Unmarshal(body, &response); err != nil{
		return nil, err
	}
	//todo 删除日志打印
	log.Println(response)
	if response.Status != "success"{
		return &response, errors.New(response.Message)
	}
	t2 := time.Now()
	log.Println(u.account, " CheckIn 耗时: ", t2.Sub(t1).Nanoseconds(), " ns")
	return &response, nil
}

/* Cancel JSON
{
	"status": "success",
	"data": null,
	"message": "",
	"code": "0"
}
*/
func (u *User) Cancel(reservationID int) ([]byte, error) {
	body, err := u.request("GET", fmt.Sprintf(CANCEL_URL, reservationID), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* Reservations JSON
{
	"status": "success",
	"data": null,
	"message": "",
	"code": "0"
}*/
func (u *User) RESERVATIONS() ([]byte, error) {
	body, err := u.request("GET", RESERVATIONS_URL, nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

/* History JSON
{
	"status": "success",
	"data": {
		"reservations": [{
			"id": 4300306,
			"date": "2018-6-21",
			"begin": "07:00",
			"end": "11:00",
			"awayBegin": null,
			"awayEnd": null,
			"loc": "主校区2层213室区第一阅览室018号",
			"stat": "RESERVE"
		}, {
			"id": 4300252,
			"date": "2018-6-21",
			"begin": "07:00",
			"end": "08:00",
			"awayBegin": null,
			"awayEnd": null,
			"loc": "主校区3层303室区第二阅览室北区002号",
			"stat": "CANCEL"
		}]
	},
	"message": "",
	"code": "0"
}
*/
func (u *User) History(page, count string) ([]byte, error) {
	body, err := u.request("GET", fmt.Sprintf(HISTORY_URL, page, count), nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (u *User) CheckAvailable() bool {
	req, err := http.NewRequest(http.MethodPost, FILTERS_URL, nil)
	if err != nil {
		return false
	}
	req.Header = http.Header{}
	req.Header.Set("token", u.token)
	req.Header.Set("Host","seat.ujn.edu.cn:8443")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Type", "charset=UTF-8")
	req.Header.Set("Accept-Encoding", "gzip")

	requestID := "d7a6afa0-5970-11e9-d32e-5bef610ac214"
	requestDate := time.Now().Unix() * 1000
	req.Header.Set("x-request-date", strconv.FormatInt(requestDate, 10))
	req.Header.Set("x-request-id", requestID)
	req.Header.Set("x-hmac-request-key", getHmacEncrypt(requestID, http.MethodPost, requestDate),)

	resp, err := u.client.Do(req)
	if err != nil {
		log.Println("检查服务器状态时出现问题! ", err)
		return false
	}

	response, err := ioutil.ReadAll(resp.Body)

	var UnmarshalResponse = FiltersResponse{}
	fmt.Println(string(response))
	err = json.Unmarshal(response, &UnmarshalResponse)
	if err != nil{
		log.Println("解析可预约日期时出现问题！")
		log.Println(err)
		log.Println(string(response))
	}

	// 检查执行状态
	now := time.Now()
	tomorrow := now.AddDate(0,0,1).Format("2006-01-02")
	if len(UnmarshalResponse.Data.Dates) ==2 &&  UnmarshalResponse.Data.Dates[1] == tomorrow && now.Unix() > time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location()).Unix(){
		return true
	} else {
		return false
	}
}
func (u *User) QuickCheckAvailable() bool {
	t1 := time.Now()
	req, err := http.NewRequest(http.MethodHead, FILTERS_URL, nil)
	if err != nil {
		return false
	}
	req.Header = http.Header{}
	req.Header.Set("token", u.token)
	req.Header.Set("Host","seat.ujn.edu.cn:8443")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Type", "charset=UTF-8")
	req.Header.Set("Accept-Encoding", "gzip")
	//req.Header.Add("X-Forwarded-For", "10.167.159.118")

	requestID := "d7a6afa0-5970-11e9-d32e-5bef610ac214"
	requestDate := time.Now().Unix() * 1000

	req.Header.Set("x-request-date", strconv.FormatInt(requestDate, 10))
	req.Header.Set("x-request-id", requestID)
	req.Header.Set("x-hmac-request-key", getHmacEncrypt(requestID, http.MethodHead, requestDate),)

	resp, err := u.client.Do(req)
	if err != nil {
		log.Println("检查服务器状态时出现问题! ", err)
		return false
	}
	t2 := time.Now()
	log.Println(u.account, " CheckAvailable 耗时: ", t2.Sub(t1).Nanoseconds(), " ns")

	//response, err := ioutil.ReadAll(resp.Body)
	//var length = 0
	length := resp.ContentLength
	//length = len(response)
	// 检查执行状态
	if length >=1590{
		log.Println("当前内容大小: ", resp.ContentLength)
		return true
	} else {
		log.Println("内容长度不足! ", resp.ContentLength)
		return false
	}
}