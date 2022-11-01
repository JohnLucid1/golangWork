package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/another/trying/structures"
	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
)

const (
	apiUrl string = "https://api.telegram.org/bot" + structures.Tocken // Подвинул токен в конфиг
)

// Изменяемые штуки (надо)
var (
	Bot_Name string = "Prikol"
)

func main() {

	sql.Register("sqlite3_with_extensions",
		&sqlite3.SQLiteDriver{
			Extensions: []string{
				"sqlite3_mod_regexp",
			},
		})

	go UpdateLoop()
	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/botName", NameHandler) 
	http.ListenAndServe("localhost:8080", router)
}

func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	var R structures.MainStru

	Ping()
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

	w.Write([]byte("Вывод успешно произведён!"))
}

func NameHandler(w http.ResponseWriter, _ *http.Request) {
	db, err := sql.Open("sqlite3", "APIBOTSTATUS.sql")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var gotname string
	var resp sql.NullString // для результата
	err = db.QueryRow("SELECT name FROM bot_status").Scan(&resp)
	if err != nil {
		fmt.Println(err)
	}
	if resp.Valid { // если результат валид
		gotname = resp.String // берём оттуда обычный string
	}
	w.Write([]byte(gotname))
}

func UpdateLoop() {
    db, err := sql.Open("sqlite3", "APIBOTSTATUS.sql")
    if err != nil {
        panic(err)
    }
    
    defer db.Close()
	lastId := 0
    var nickname1 string
    err = db.QueryRow(`select name from bot_status`).Scan(&nickname1)

    if err != nil {
        fmt.Println(err)
    }

	for {
        newId := Update(lastId, &nickname1)
        if lastId != newId {
            lastId = newId
            db.Exec(`update bot_status set lastid = $1`, lastId)
        }
        time.Sleep(5 * time.Millisecond)
	}
}

func Update(lastId int, nickname *string) int {
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

		if strings.Split(txt, ", ")[0] == Bot_Name {

			switch strings.Split(strings.Split(txt, ", ")[1], ": ")[0] {
			case "anekdot":
				{
					txt = ""
					return Anek(lastId, ev)
				}
			case "random number":
				{
					return RandGen(lastId, ev, txt)
				}

			case "change name to":
				{
					if strings.Contains(txt, ":") {
						return ChangeName(lastId, ev, txt)

					} else {
						return SomeMessage(lastId, ev, "Wrong")
					}
				}
			case "privet":
				{
					fmt.Println("worsk")
					return SomeMessage(lastId, ev, "Hey loser")
				}
			case "name":
				{
					fmt.Println("worsk")
					return SayMyName(lastId, ev)
				}
			}
		}
	}

	return lastId
}

func Anek(lastID int, ev structures.UpdateStruct) int {
	txtmsg := structures.SendMessage{
		ChId: ev.Message.Chat.Id,
		Text: "https://www.youtube.com/watch?v=tvkxupwbFLk&ab_channel=Corpax",
	}

	bytemsg, _ := json.Marshal(txtmsg)
	_, err := http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))
	if err != nil {
		fmt.Println(err)
		return lastID
	} else {
		return ev.Id + 1
	}
}

func RandGen(lastID int, ev structures.UpdateStruct, txt string) int {
	retotal := strings.Split(txt, "до ")[1]
	s, err := strconv.Atoi(retotal)
	if err != nil {
		panic(err)
	}
	num := strconv.Itoa(rand.Intn(s))
	txtmsg := structures.SendMessage{
		ChId: ev.Message.Chat.Id,
		Text: "Сгенерированное число: " + num,
	}

	bytemsg, _ := json.Marshal(txtmsg)
	_, err = http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))

	if err != nil {
		fmt.Println(err)
		return lastID
	} else {
		return ev.Id + 1
	}
}

func ChangeName(lastId int, ev structures.UpdateStruct, txt string) int {
	newap := strings.Split(txt, "change name to:")
	Bot_Name = newap[1]

	txtmsg := structures.SendMessage{
		ChId: ev.Message.Chat.Id,
		Text: "Обращение изменено на: " + Bot_Name,
	}

	bytemsg, _ := json.Marshal(txtmsg)
	_, err := http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))

	if err != nil {
		fmt.Println(err)
		return lastId
	} else {
		return ev.Id + 1
	}

}

func SomeMessage(lastId int, ev structures.UpdateStruct, txt string) int {
	txtmsg := structures.SendMessage{
		ChId: ev.Message.Chat.Id,
		Text: txt,
	}

	bytemsg, _ := json.Marshal(txtmsg)
	_, err := http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))

	if err != nil {
		fmt.Println(err)
		return lastId
	} else {
		return ev.Id + 1
	}
}

func SayMyName(lastId int, ev structures.UpdateStruct) int {
	txtmsg := structures.SendMessage{
		ChId: ev.Message.Chat.Id,
		Text: Bot_Name,
	}

	bytemsg, _ := json.Marshal(txtmsg)
	_, err := http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))

	if err != nil {
		fmt.Println(err)
		return lastId
	} else {
		return ev.Id + 1
	}

}

func Ping() {
	txtmsg := structures.SendMessage{
		ChId: 802708579,
		Text: "Страница посещена",
	}

	bytemsg, _ := json.Marshal(txtmsg)
	_, err := http.Post(apiUrl+"/sendMessage", "application/json", bytes.NewReader(bytemsg))
	if err != nil {
		fmt.Println(err)
	}
}
