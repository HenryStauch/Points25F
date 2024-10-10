package src

import (
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
	fmt.Println(req_str)
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
	fmt.Println(string(body))

	var group GroupResponse
	err = json.Unmarshal(body, &group)
	if err != nil {
		fmt.Println("ERROR: could not parse group info")
		return nil, err
	}

	fmt.Printf("%+v\n", group.Response)
	fmt.Println("Members: ")
	fmt.Println(group.Response.Members)

	return nil, nil
}

func LikeMessage(ConversationId string, MessageId string) error {
	req_str := fmt.Sprintf("%s/messages/%s/%s/like", BASE_URL, ConversationId, MessageId)
	_, err := http.Get(req_str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
