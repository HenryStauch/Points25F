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

// "id": "103844305",
//         "group_id": "103844305",
//         "name": "Points",
//         "phone_number": "+1 8022391308",
//         "type": "private",
//         "description": "",
//         "image_url": null,
//         "creator_user_id": "86593728",
//         "created_at": 1728262604,
//         "updated_at": 1728500569,
//         "muted_until": null,
//         "audio_message_disabled": false,
//         "messages": {
//             "count": 6,
//             "last_message_id": "172826267188882649",
//             "last_message_created_at": 1728262671,
//             "last_message_updated_at": 1728262671,
//             "preview": {
//                 "nickname": "Ben Sontag",
//                 "text": "-100 @Rishav Chakravarty being a bumass damnass shitass fuckass fucjer",
//                 "image_url": "",
//                 "attachments": [
//                     {
//                         "loci": [
//                             [
//                                 5,
//                                 19
//                             ]
//                         ],
//                         "type": "mentions",
//                         "user_ids": [
//                             "55200205"
//                         ]
//                     }
//                 ]
//             }
//         },
//         "max_members": 5000,
//         "theme_name": null,
//         "like_icon": null,
//         "requires_approval": false,
//         "show_join_question": false,
//         "join_question": null,
//         "message_deletion_period": 2147483647,
//         "message_deletion_mode": [
//             "admin",
//             "sender"
//         ],
//         "children_count": 0,
//         "share_url": "https://groupme.com/join_group/103844305/BZUKM8nE",
//         "share_qr_code_url": "https://image.groupme.com/qr/join_group/103844305/BZUKM8nE/preview",
//         "directories": [],
//         "members": [
//             {
//                 "user_id": "86593728",
//                 "nickname": "Ben Sontag",
//                 "image_url": "",
//                 "id": "1019914981",
//                 "muted": false,
//                 "autokicked": false,
//                 "roles": [
//                     "admin",
//                     "owner"
//                 ],
//                 "name": "Ben Sontag"
//             },
//             {
//                 "user_id": "55200205",
//                 "nickname": "Rishav Chakravarty",
//                 "image_url": "https://i.groupme.com/979x979.jpeg.ad5c263dd9b648fc8cec4ad5f4a1a612",
//                 "id": "1019915005",
//                 "muted": false,
//                 "autokicked": false,
//                 "roles": [
//                     "admin"
//                 ],
//                 "name": "Rishav Chakravarty"
//             },
//             {
//                 "user_id": "90105187",
//                 "nickname": "Jack Macy",
//                 "image_url": "https://i.groupme.com/743x743.jpeg.ba6a0d4cbad5408b83c5617d3f73837e",
//                 "id": "1020810625",
//                 "muted": false,
//                 "autokicked": false,
//                 "roles": [
//                     "admin"
//                 ],
//                 "name": "Jack Macy"
//             }
//         ],
//         "members_count": 3,
//         "locations": [],
//         "visibility": "hidden",
//         "category_ids": null,
//         "active_call_participants": null,
//         "unread_count": null,
//         "last_read_message_id": null,
//         "last_read_at": null
//     }

func GetAllUsers() ([]User, error) {
	req_str := fmt.Sprintf("%s/groups/%s?token=%s", BASE_URL, os.Getenv("GROUP_ID"), os.Getenv("API_KEY"))
	resp, err := http.Get(req_str)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(body))

	var group Group
	err = json.Unmarshal(body, &group)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", group)
	fmt.Println(group.Id)

	return nil, nil
}

func LikeMessage(ConversationId string, MessageId string) error {
	req_str := fmt.Sprintf("%s/messages/%s/%s/like", ConversationId, MessageId)
	_, err := http.Get(req_str)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
