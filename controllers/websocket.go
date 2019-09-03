// Copyright 2013 Beego Samples authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"

	"keyChat/models"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

// WebSocketController handles WebSocket requests.
type WebSocketController struct {
	baseController
}
type Res struct {
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
}

// Get method handles GET requests for WebSocketController.
func (this *WebSocketController) Get() {
	// Safe check.
	uname := this.GetString("uname")
	roomid := this.GetString("roomid")
	if len(uname) == 0 || len(roomid) == 0 {
		this.Redirect("/", 302)
		return
	}

	this.TplName = "websocket.html"
	this.Data["roomid"] = roomid
	this.Data["IsWebSocket"] = true
	this.Data["UserName"] = uname
}

// Join method handles WebSocket requests for WebSocketController.
func (this *WebSocketController) Join() {
	uname := this.GetString("uname")
	roomid := this.GetString("roomid")
	if len(uname) == 0 || len(roomid) == 0 {
		this.Redirect("/", 302)
		return
	}

	// Upgrade from http request to WebSocket.从http请求升级到WebSocket。
	// Upgrade 函数将 http 升级到 WebSocket 协议
	ws, err := websocket.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil, 1024, 1024)
	//如果请求不是有效的WebSocket握手，则Upgrade返回一个--->握手错误类型的错误。应用程序应该处理这个错误,  用HTTP错误码 响应客户端。
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(this.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}

	// Join chat room.
	Join(uname, roomid, ws)
	defer Leave(uname) //defer函数会在return之后被调用   ，用户退出操作

	// Message receive loop.消息接收循环
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			return
		}
		//publish <- newEvent(models.EVENT_MESSAGE, uname, roomid, string(p))
		//b := []byte(event.Content) //string 转[]byte
		fmt.Println(len(p))
		if len(p) > 10000 {
			fmt.Println("刚刚接收到的数据是图片类型---》", reflect.TypeOf(p))
			var empty_str string //声明一个空字符串
			publish <- newEvent_byte(models.EVENT_IMG, uname, roomid, empty_str, p)
		} else {
			fmt.Println("刚刚接收到的数据是 文本消息类型---》", reflect.TypeOf(p))
			//var img []byte
			publish <- newEvent(models.EVENT_MESSAGE, uname, roomid, string(p))
		}

	}
}

// broadcastWebSocket broadcasts messages to WebSocket users.  -->broadcastWebSocket向WebSocket用户广播消息。
func broadcastWebSocket(event models.Event) {
	data, err := json.Marshal(event) //Json Marshal：将struct 结构体数据编码成json字符串
	if err != nil {
		beego.Error("Fail to marshal event:", err)
		return
	}

	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
		// Immediately send event to WebSocket users.立即向WebSocket用户发送事件。
		ws := sub.Value.(Subscriber).Conn
		if ws != nil && event.Roomid == sub.Value.(Subscriber).RoomId {
			if ws.WriteMessage(websocket.TextMessage, data) != nil {
				// User disconnected.//用户断开连接。
				unsubscribe <- sub.Value.(Subscriber).Name
			}
		}
	}
}

func (this *WebSocketController) Fetch() {
	lastReceived, err := this.GetInt("lastReceived")
	if err != nil {
		return
	}

	events := models.GetEvents(int(lastReceived))
	if len(events) > 0 {
		this.Data["json"] = events
		this.ServeJSON()
		return
	}
	return
}

func (this *WebSocketController) Hisstory() {
	o := orm.NewOrm()
	room_name := this.GetString("room_name")

	var hisstory_maps []orm.Params
	hisstory_num, err := o.Raw("SELECT * FROM im_hisstory WHERE hisstory_room_name = ?  ORDER BY hisstory_id  desc  LIMIT  100", room_name).Values(&hisstory_maps)

	//now_time := time.Now().Unix() // 获取  时间戳 格式的 当前时间
	//fmt.Println(hisstory_num)
	if err == nil && hisstory_num > 0 {

		res_new := Res{}
		res_new.Data = hisstory_maps
		res_new.Msg = "ok"
		res_new.Code = 1
		this.Data["json"] = res_new
		this.ServeJSON()
	}

}
