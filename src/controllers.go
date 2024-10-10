package src

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatDTO struct {
	Attachments []any  `json:"attachments"`
	AvatarUrl   string `json:"avatar_url"`
	CreatedAt   uint64 `json:"created_at"`
	GroupId     string `json:"group_id"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	SenderId    string `json:"sender_id"`
	SenderType  string `json:"sender_type"`
	SourceGUID  string `json:"source_guid"`
	System      bool   `json:"system"`
	Text        string `json:"text"`
	UserId      string `json:"user_id"`
}

func ReceiveChat(c *gin.Context) {
	chat := ChatDTO{}
	err := c.BindJSON(&chat)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// If the message is from an admin and starts with an exclamation point,
	// handle dev commands
	fmt.Println(chat.Text)
	first_char := []rune(chat.Text)[0]
	if first_char == '!' {
		// Validate that an admin brother sent this
		var brother = Brother{BrotherId: chat.UserId, IsAdmin: true}
		result := DB.First(&brother)
		if result.Error != nil {
			// Someone trying an admin command without admin privileges
			return
		}
		fmt.Println(brother.Name)

		command, argstr, args_found := strings.Cut(chat.Text, " ")
		args := []string{}
		if args_found {
			args = strings.Split(argstr, " ")
		}

		command = command[1:]
		switch command {
		case "ping":
			SendMessage("pong")

		case "attendance":
			GetAllUsers()

		case "add_brother":
			if !args_found {
				return
			}
			all_users, err := GetAllUsers()
			if err != nil {
				return
			}
			for _, brother := range args {
				for _, user := range all_users {
					brother_sanitized := strings.ReplaceAll(brother, "_", " ")
					if strings.EqualFold(brother_sanitized, user.Nickname) {
						var brother Brother = Brother{
							Name:      user.Name,
							BrotherId: user.UserId,
							IsAdmin:   false,
							IsTimeout: false,
						}
						DB.Create(&brother)
						fmt.Println("Added brother " + user.Nickname)
						break
					}
				}
			}
			break
		case "add_pledge":
			if !args_found {
				return
			}
			all_users, err := GetAllUsers()
			if err != nil {
				return
			}
			for _, pledge := range args {
				for _, user := range all_users {
					pledge_sanitized := strings.ReplaceAll(pledge, "_", " ")
					if user.Nickname == pledge_sanitized {
						var pledge Pledge = Pledge{
							Name:     user.Name,
							PledgeId: user.UserId,
							Points:   0,
						}
						DB.Create(&pledge)
						fmt.Println("Added pledge " + user.Nickname)
						break
					}
				}

			}

		case "make_admin":
			break
		case "take_admin":
			break
		case "timeout":
			if args_found {
				return
			}
			_ = args[0]
			break
		default:
			return
		}

		LikeMessage(chat.GroupId, chat.Id)
	} else if first_char == '+' || first_char == '-' {
		// If the message starts with + or -
		// Handle adding/subtracting points
		// Validate that a brother sent this
		var brother = Brother{BrotherId: chat.UserId}
		result := DB.First(&brother)
		if result.Error != nil {
			// A non-brother (or non-registered brother) is trying to assign points
			return
		}
		fmt.Println("Points from " + brother.Name)

		points_str, rest, _ := strings.Cut(chat.Text, " ")

		// Get the points requested
		points, err := strconv.Atoi(points_str[1:])
		if err != nil {
			return
		}
		if points_str[0] == '-' {
			points = -points
		}
		fmt.Println(points)

		// Parse the rest of the string, separating between usernames and later text
		words := strings.Split(rest, " ")
		word_i := 0
		iter_name := ""
		look_for_name := true
		for {
			if word_i >= len(words) {
				break
			}
			cur_word := words[word_i]
			if look_for_name && cur_word[0] != '@' {
				// If we're looking for a username and don't get one
				// stop early
				break
			}
			if len(iter_name) > 0 {
				iter_name += " "
			}
			iter_name += cur_word
			look_for_name = false

			var pledge = Pledge{Name: iter_name[1:]}
			result := DB.First(&pledge)
			if result.Error == nil {
				// if iter_name[1:] == "Jack Macy" || iter_name[1:] == "Miles" {
				// Pledge exists with this name
				fmt.Println(pledge.ID, pledge.Name)
				look_for_name = true
				iter_name = ""

				// pledge.Points = pledge.Points + points
				// DB.Save(&pledge)
				var point Point = Point{
					PointsGiven: points,
					PledgeId:    pledge.ID,
				}
				DB.Create(&point)

				fmt.Printf("Gave %d points to %s\n", points, pledge.Name)
			}

			word_i++
		}

		LikeMessage(chat.GroupId, chat.Id)
	}
}
