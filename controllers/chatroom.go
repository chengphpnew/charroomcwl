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
	"container/list"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"

	"keyChat/models"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type Subscription struct {
	Archive []models.Event      // All the events from the archive.来自存档的所有事件。
	New     <-chan models.Event // New events coming in.新事件接踵而至
}

func newEvent(ep models.EventType, user, roomid, msg string) models.Event {
	var img []byte
	return models.Event{ep, user, roomid, int(time.Now().Unix()), msg, img}
}

func newEvent_byte(ep models.EventType, user string, roomid string, msg string, img []byte) models.Event {
	return models.Event{ep, user, roomid, int(time.Now().Unix()), msg, img}
}

func Join(user string, roomid string, ws *websocket.Conn) {
	if !isUserExist(subscribers, user) {
		beego.Info("New user:", user)

	} else {
		beego.Info("old user:", user)
		//publish <- newEvent(4, user, roomid, "") // Publish a LEAVE event.发布 重复事件

	}
	subscribe <- Subscriber{Name: user, RoomId: roomid, Conn: ws} // 发送 [用户姓名]，[用户房间名称] ，[用户浏览器的websock的连接号]，  {我实验的结果是 一个页面可以打开一个 websock的连接}   这三个值 数据结构体  到---》 管道  subscribe中
	//	(箭头的指向就是数据的流向)
	// Channel是Go中的一个核心类型，你可以把它看成一个管道，通过它并发核心单元就可以发送或者接收数据进行通讯(communication)。它的操作符是箭头 <- 。
}

func Leave(user string) {
	unsubscribe <- user
}

// 定义一个 数据结构体
type Subscriber struct {
	Name   string
	RoomId string
	Conn   *websocket.Conn //只适用于WebSocket用户;否则nil。 Only for WebSocket users; otherwise nil.
}

var (
	// Channel for new join users.新加入用户的管道。
	subscribe = make(chan Subscriber, 10) //声明一个  新加入用户的  管道。用来存放新加入进来的 用户

	// Channel for exit users.退出用户管道。
	unsubscribe = make(chan string, 10) //声明一个   退出用户  管道。

	// Send events here to publish them.发送事件到这里发布它们。
	publish = make(chan models.Event, 10) //声明一个发布管道,如果 有  事件要  广播发布，就走这个管道
	// Long polling waiting list.轮询等候名单
	waitingList = list.New()
	subscribers = list.New() //list是一个双向链表。该结构具有链表的所有功能。

)

// This function handles all incoming chan messages.此函数处理所有传入的chan消息。

/*
//select基本用法
select {
case <- chan1:
// 如果chan1成功读到数据，则进行该case处理语句
case chan2 <- 1:
// 如果成功向chan2写入数据，则进行该case处理语句
default:
// 如果上面都没有成功，则进入default处理流程

*/
func chatroom() {
	o := orm.NewOrm()
	for { //无限循环，即for关键字后什么都没有
		select { //select就是用来监听和channel有关的IO操作，当 IO 操作发生时，触发相应的动作
		case sub := <-subscribe:
			fmt.Println(sub) //{丽莎（用户名） 非煤矿山（房间号） 0xc0422b82c0（websocker的 网络连接号）}

			if !isUserExist(subscribers, sub.Name) {
				subscribers.PushBack(sub) // Add user to the end of list.将用户添加到列表的末尾。
				// Publish a JOIN event.发布 加入房间 事件
				publish <- newEvent(models.EVENT_JOIN, sub.Name, sub.RoomId, "")
				beego.Info("New user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			} else {
				beego.Info("Old user:", sub.Name, ";WebSocket:", sub.Conn != nil)
				// fmt.Println(sub.RoomId) 房间名称非煤矿山

				//unsub := <-unsubscribe
				//fmt.Println(models.EVENT_LEAVE)  ---》1
				//fmt.Println(unsub)-----》成文龙
				//fmt.Println(sub.Value.(Subscriber).RoomId)---》烟花爆竹
				//old_username := sub.Name

				//如下是 循环 ，看看subscribers  这个链表里面 有没有 用户信息
				/*for sub := subscribers.Front(); sub != nil; sub = sub.Next() {

					if sub.Value.(Subscriber).Name == old_username {
						subscribers.Remove(sub)
						// Clone connection.克隆连接。
						ws := sub.Value.(Subscriber).Conn
						if ws != nil {
							ws.Close()
							beego.Error("WebSocket closed:", old_username)
						}

						//fmt.Println(models.EVENT_LEAVE)  ---》1
						//fmt.Println(unsub)-----》唐利莎
						//fmt.Println(sub.Value.(Subscriber).RoomId)---》烟花爆竹
						publish <- newEvent(models.EVENT_LEAVE, old_username, sub.Value.(Subscriber).RoomId, "") // Publish a LEAVE event.发布 离开事件
						break
					}
				}*/
			}
		case event := <-publish: // 不管是 新加入用户了，还是 用户发信息了，还是 用户离开了，都会走 这个方法
			// Notify waiting list.通知名单
			for ch := waitingList.Back(); ch != nil; ch = ch.Prev() {
				ch.Value.(chan bool) <- true
				waitingList.Remove(ch)
			}

			broadcastWebSocket(event)
			models.NewArchive(event)

			fmt.Println(event)
			fmt.Println(reflect.TypeOf(event.Content))
			/*
			   hisstory_id 	hisstory_content 群聊内容	hisstory_room_name 群聊房间名称	hisstory_time 群聊信息时间	hisstory_username 群聊发信息人用户名
			*/
			if event.Type == models.EVENT_IMG {
				var nothing string
				res, err := o.Raw("INSERT INTO im_hisstory (hisstory_content,hisstory_room_name,hisstory_time,hisstory_username,hisstory_img) values (?,?,?,?,?)", nothing, event.Roomid, event.Timestamp, event.User, string(event.Img)).Exec()
				fmt.Println(err)
				fmt.Println(reflect.TypeOf(event.Img))
				if err == nil {
					//当出现不等于nil的时候，说明出现某些错误了，需要我们对这个错误进行一些处理，而如果等于nil说明运行正常。
					//nil的意思是无，或者是零值。零值，zero value，是不是有点熟悉？在Go语言中，如果你声明了一个变量但是没有对它进行赋值操作，那么这个变量就会有一个类型的默认零值。这是每种类型对应的零值：
					num, _ := res.RowsAffected() // 获取 影响的  数据库 条数
					fmt.Println("mysql row affected nums: ", num)

					//b := []byte(event.Content) //string 转[]byte
					beego.Info("Message from", event.User, ";消息类型是图片:", reflect.TypeOf(event.Img))
				}

			}

			if event.Type == models.EVENT_MESSAGE {

				var map_content map[string]interface{}
				err := json.Unmarshal([]byte(event.Content), &map_content) //json转map
				//fmt.Println(err)
				if err == nil { //说明没有出现报错
					res, err := o.Raw("INSERT INTO im_hisstory (hisstory_content,hisstory_room_name,hisstory_time,hisstory_username) values (?,?,?,?)", map_content["content"], event.Roomid, event.Timestamp, event.User).Exec()
					if err == nil {
						//当出现不等于nil的时候，说明出现某些错误了，需要我们对这个错误进行一些处理，而如果等于nil说明运行正常。
						//nil的意思是无，或者是零值。零值，zero value，是不是有点熟悉？在Go语言中，如果你声明了一个变量但是没有对它进行赋值操作，那么这个变量就会有一个类型的默认零值。这是每种类型对应的零值：
						num, _ := res.RowsAffected() // 获取 影响的  数据库 条数
						fmt.Println("mysql row affected nums: ", num)
						beego.Info("Message from", event.User, ";消息类型是文字:", reflect.TypeOf(map_content["content"]))

					}
				}
				//fmt.Println(mapResult)

			}

		case unsub := <-unsubscribe: //取消订阅
			for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
				if sub.Value.(Subscriber).Name == unsub {
					subscribers.Remove(sub)
					// Clone connection.克隆连接。
					ws := sub.Value.(Subscriber).Conn
					if ws != nil {
						ws.Close()
						beego.Error("WebSocket closed:", unsub)
					}

					//fmt.Println(models.EVENT_LEAVE)  ---》1
					//fmt.Println(unsub)-----》唐利莎
					//fmt.Println(sub.Value.(Subscriber).RoomId)---》烟花爆竹
					publish <- newEvent(models.EVENT_LEAVE, unsub, sub.Value.(Subscriber).RoomId, "") // Publish a LEAVE event.发布 离开事件
					break
				}
			}
		}
	}
}

/*基于计数器的for循环
基于计数器的迭代，基本形式为：
for 初始化语句; 条件语句; 修饰语句 {
    //循环语句
}
  for i := 0; i < 10; i++ {
        sum += i
    }
最后一部分为修饰语句 i++，一般用于增加或减少计数器。
*/

func init() {
	go chatroom() //轻松开启高并发，一直都是golang语言引以为豪的功能点。
	//golang通过goroutine实现高并发，而开启goroutine的钥匙正是go关键字。开启并发执行的语法格式是： go  funcName()
}

//循环这个  里面有没有 重复用户
func isUserExist(subscribers *list.List, user string) bool {
	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
		if sub.Value.(Subscriber).Name == user {
			return true
		}
	}
	return false
}
