package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/another/trying/structures"
	"github.com/gorilla/mux"

	"strconv"
	"time"
)

const (
	apiUrl string = "https://api.telegram.org/bot" + structures.Tocken // Подвинул токен в конфиг
)

// Изменяемые штуки (надо)
var (
	Bot_Name string = "Prikol"
)

func main() {
	go UpdateLoop()
	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	http.ListenAndServe("localhost:8080", router)
}

func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	var R structures.MainStru

	resp, err := http.Get(apiUrl + "/getMe")

	if err != nil {
		fmt.Println(err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))

	err = json.Unmarshal(respBody, &R) // заполнили перемнную р
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	R.Result.Abilites = append(R.Result.Abilites, "reacting to command /privet")

	respReady, err := json.Marshal(R.Result)
	if err != nil {
		panic(err)
	}

	w.Write([]byte(respReady))

	println("НАШИ ДАННЫЕ ПРОЧИТАНЫ! ПОЛНАЯ ГОТОВНОСТЬ У НАС ГОСТИ!")

	w.Write([]byte("Вывод успешно произведён!"))
}

func UpdateLoop() {
	lastId := 0
	for {
		lastId = Update(lastId)
		time.Sleep(3 * time.Second)
	}
}

func Update(lastId int) int {
	raw, err := http.Get(apiUrl + "/getUpdates?offset=" + strconv.Itoa(lastId))
	if err != nil {
		panic(err)
	}
	body, _ := io.ReadAll(raw.Body)

	var v structures.UpdateResponse
	err = json.Unmarshal(body, &v)
	if err != nil {
		panic(err)
	}

	if len(v.Result) > 0 {
		ev := v.Result[len(v.Result)-1]
		txt := ev.Message.Text 

		if txt == "/privet" {
			txtmsg := structures.SendMessage{
				ChId:                ev.Message.Chat.Id,
				Text:                "ИДИ ОТ СЮДА, ЧИТАЙ ОПИСАНИЕ!",
				Reply_To_Message_Id: ev.Message.Id,
			}

			bytemsg, _ := json.Marshal(txtmsg)
			_, err = http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))
			if err != nil {
				fmt.Println(err)
				return lastId
			} else {
				return ev.Id + 1
			}
		}

		if txt == "/SayMyName" {

			txtmsg := structures.SendMessage{
				ChId:                ev.Message.Chat.Id,
				Text:                Bot_Name,
				Reply_To_Message_Id: ev.Message.Id,
			}

			bytemsg, _ := json.Marshal(txtmsg)
			_, err = http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))
			if err != nil {
				fmt.Println(err)
				return lastId
			} else {
				return ev.Id + 1
			}

		}

		if strings.Contains(txt, "/ChangeName"){

			if len(strings.Split(txt, " ")) > 1 {

				newName := strings.Split(txt, " ")[1]

				Bot_Name = newName

				txtmsg := structures.SendMessage{
					ChId:                ev.Message.Chat.Id,
					Text:                "New Bot Name is set to: " + Bot_Name,
					Reply_To_Message_Id: ev.Message.Id,
				}

				bytemsg, _ := json.Marshal(txtmsg)
				_, err = http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))
				if err != nil {
					fmt.Println(err)
					return lastId
				} else {
					return ev.Id + 1
				}
			}

		}		

		if txt == "/easter_egg" {
			txtmsg := structures.SendMessage{
				ChId:                ev.Message.Chat.Id,
				Text:                "https://www.youtube.com/watch?v=lIxM2rGKEV4",
				Reply_To_Message_Id: ev.Message.Id,
			}

			bytemsg, _ := json.Marshal(txtmsg)
			_, err = http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))
			if err != nil {
				fmt.Println(err)
				return lastId
			} else {
				return ev.Id + 1
			}
		}
	}

	return lastId
}
