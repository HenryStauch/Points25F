package src

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/montanaflynn/stats"
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
	first_char := []rune(chat.Text)[0]
	if first_char == '!' {
		// Validate that an admin brother sent this
		var brother = Brother{}
		result := DB.First(&brother, "brother_id = ? AND is_admin = ?", chat.UserId, true)
		if result.Error != nil {
			// Someone trying an admin command without admin privileges
			return
		}

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

		case "tally":
			fmt.Println("Tallying points!")
			rows, err := DB.Model(&Point{}).Rows()
			if err != nil {
				fmt.Println("Error getting DB rows!")
				return
			}
			var all_points []Point = []Point{}
			var all_point_values []int = []int{}

			for rows.Next() {
				var point Point
				DB.ScanRows(rows, &point)

				all_points = append(all_points, point)
				// For the point value normalization, want absolute value
				if point.PointsGiven < 0 {
					all_point_values = append(all_point_values, -point.PointsGiven)
				} else {
					all_point_values = append(all_point_values, point.PointsGiven)
				}
			}

			// Stats to determine how much to curve points
			point_val_data := stats.LoadRawData(all_point_values)
			points_mode_arr, _ := point_val_data.Mode()
			points_median_arr, _ := point_val_data.Median()
			mode_arr_data := stats.LoadRawData(points_mode_arr)
			median_arr_data := stats.LoadRawData(points_median_arr)
			mode, _ := stats.Mean(mode_arr_data)
			median, _ := stats.Mean(median_arr_data)
			log_factor := mode + median/2
			fmt.Println(log_factor)

			if args_found {
				arg_log, err := strconv.Atoi(argstr)
				if err != nil {
					return
				}
				log_factor = float64(arg_log)
			}

			// This is arbitrary
			// Multiplying factor so with int truncation it doesn't all turn to 0s
			const factor int = 50

			for _, point_el := range all_points {
				is_neg := point_el.PointsGiven < 0
				point_give_req := point_el.PointsGiven

				var points_to_give int
				if is_neg {
					point_float := math.Log2(float64(-point_give_req)) / math.Log2(log_factor)
					point_float *= float64(factor)
					points_to_give = -int(point_float)
				} else {
					point_float := math.Log2(float64(point_give_req)) / math.Log2(log_factor)
					point_float *= float64(factor)
					points_to_give = int(point_float)
				}
				if points_to_give == 0 {
					continue
				}
				if points_to_give > 3*factor {
					points_to_give = 3 * factor
				} else if points_to_give < -3*factor {
					points_to_give = -3 * factor
				}

				var pledge Pledge
				result := DB.First(&pledge, "id = ?", point_el.PledgeId)
				if result.Error != nil {
					// a shame.
					continue
				}
				pledge.Points += points_to_give
				// fmt.Println(points_to_give)
				DB.Save(&pledge)
			}

			DB.Model(Point{}).Delete(&Point{})

		default:
			return
		}

		LikeMessage(chat.GroupId, chat.Id)
	} else if first_char == '+' || first_char == '-' {
		// If the message starts with + or -
		// Handle adding/subtracting points
		// Validate that a brother sent this
		var brother = Brother{}
		result := DB.First(&brother, "brother_id = ?", chat.UserId)
		if result.Error != nil {
			// A non-brother (or timeout/non-registered brother) is trying to assign points
			return
		}

		points_str, rest, _ := strings.Cut(chat.Text, " ")

		// Get the points requested
		points, err := strconv.Atoi(points_str[1:])
		if err != nil {
			return
		}
		if points_str[0] == '-' {
			points = -points
		}

		// Parse the rest of the string, separating between usernames and later text
		words := strings.Split(rest, " ")
		iter_name := ""
		look_for_name := true
		found_name := false
		for word_i, cur_word := range words {
			if word_i >= len(words) {
				break
			}
			if len(cur_word) == 0 {
				// Handle double-space messages
				continue
			}
			if look_for_name && len(cur_word) > 0 && cur_word[0] != '@' {
				// If we're looking for a username and don't get one
				// stop early
				break
			}
			if len(iter_name) > 0 {
				iter_name += " "
			}
			iter_name += cur_word
			look_for_name = false

			var pledge = Pledge{}
			result := DB.Model(Pledge{}).Limit(1).Find(&pledge, "name = ?", iter_name[1:])
			if result.Error == nil && result.RowsAffected > 0 {
				// Pledge exists with this name
				look_for_name = true
				iter_name = ""
				found_name = true

				// pledge.Points = pledge.Points + points
				// DB.Save(&pledge)
				var point Point = Point{
					PointsGiven: points,
					PledgeId:    pledge.ID,
				}
				DB.Create(&point)
			} // Otherwise, keep looking
		}

		if found_name {
			LikeMessage(chat.GroupId, chat.Id)
		}
	}
}
