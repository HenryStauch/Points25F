package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const BASE_URL = "https://api.groupme.com/v3"

type Group struct {
	Id      int64  `json:"id"`
	Members []User `json:"members"`
}

type GroupResponse struct {
	Response struct {
		Name    string `json:"name"`
		Members []User `json:"members"`
	} `json:"response"`
}

func GetAllUsers() ([]User, error) {
	req_str := fmt.Sprintf("%s/groups/%s?token=%s", BASE_URL, os.Getenv("GROUP_ID"), os.Getenv("API_KEY"))
	// fmt.Println(req_str)
	resp, err := http.Get(req_str)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ERROR: could not read group info")
		return nil, err
	}
	// fmt.Println(string(body))

	var group GroupResponse
	err = json.Unmarshal(body, &group)
	if err != nil {
		fmt.Println("ERROR: could not parse group info")
		return nil, err
	}

	// fmt.Printf("%+v\n", group.Response)
	// fmt.Println("Members: ")
	// fmt.Println(group.Response.Members)

	return group.Response.Members, nil
}

func LikeMessage(ConversationId string, MessageId string) error {
	req_str := fmt.Sprintf("%s/messages/%s/%s/like?token=%s", BASE_URL, ConversationId, MessageId, os.Getenv("API_KEY"))
	fmt.Println(req_str)
	_, err := http.Get(req_str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func SendMessage(text string) {
	values := map[string]string{"bot_id": os.Getenv("BOT_ID"), "text": text}
	json_data, err := json.Marshal(values)

	if err != nil {
		return
	}

	resp, err := http.Post("https://api.groupme.com/v3/bots/post", "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		return
	}

	defer resp.Body.Close()
}
