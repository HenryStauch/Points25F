package src

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/montanaflynn/stats"
)

// Chat Data Type Object
// This is the data that gets passed to the API
// for every chat message
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

// Handle an incoming message
func ReceiveChat(c *gin.Context) {
	chat := ChatDTO{}
	err := c.BindJSON(&chat)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	first_char := []rune(chat.Text)[0]
	if first_char == '!' {
		// If the message is from an admin and starts with an exclamation point,
		// handle dev commands

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
		// Parse and handle each dev command
		switch command {
			
			// Ping command: testing the bot
			//
			// takes no arguments
		case "ping":
			SendMessage("pong")

			// This isn't really useful any more,
			// if needed it is called on demand
			// in !add_brother and !add_pledge
		case "attendance":
			GetAllUsers()

			// Add one or more members to DB marked as brothers
			// The provided names should match the groupme nicknames
			// i.e. if a user changes their GM name, use the new name
			//
			// !add_brother Firstname_Lastname Firstname_Lastname ...
		case "add_brother":
			if !args_found {
				return
			}
			all_users, err := GetAllUsers()
			if err != nil {
				return
			}
			// This is a silly way of doing it but it works
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

			// Add one or more members to DB marked as pledges
			// The provided names should match the groupme nicknames
			// i.e. if a user changes their GM name, use the new name
			//
			// !add_pledge Firstname_Lastname Firstname_Lastname ...
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

			// Not yet implemented and I don't plan to...
			// Haven't needed to have anybody as an admin
			// except for the person running the bot
		case "make_admin":
			break
		case "take_admin":
			break

			// This isn't a real command (yet)!
			// There is a timeout field in the DB but I
			// haven't cared to implement timeout checking
			// (although it would be easy).
			// So far it has been useful as a threat.
			// They don't know it doesn't do anything.
			// 
			// !timeout Firstname_Lastname
		case "timeout":
			if args_found {
				return
			}
			_ = args[0]

			// Tally command
			// Tallies the points of the week
			// and adds them to a total with a curve
			// to handle inflation.
			// Only places the point values into the DB,
			// does not display them.
			//
			// takes no arguments
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

			rows.Close()

			// Stats to determine how much to curve points
			// Currently using the average of the median and the mode
			point_val_data := stats.LoadRawData(all_point_values)
			points_mode_arr, _ := point_val_data.Mode()
			points_median, _ := point_val_data.Median()
			mode_arr_data := stats.LoadRawData(points_mode_arr)
			mode, _ := stats.Mean(mode_arr_data)
			log_factor := (mode + points_median) / 2
			fmt.Println(log_factor)

			// If we supply a manual curving factor, use that instead
			if args_found {
				arg_log, err := strconv.Atoi(argstr)
				if err != nil {
					return
				}
				log_factor = float64(arg_log)
			}

			// This is an arbitrary factor
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
				// Arbitrary check:
				// Even with the curving, is the point addition/deduction too high?
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

			// Leaderboard command
			// Displays a leaderboard of the currently saved points
			// and sends it to the chat,
			// not including the points that have yet to be tallied
			// and curved.
			//
			// takes no arguments
		case "leaderboard":
			fmt.Println("Sending Leaderboard!")
			var msgs []string
			msgs = append(msgs, "==Leaderboard==")
			var pledges []Pledge
			result := DB.Order("points desc").Find(&pledges)

			if result.RowsAffected == 0 || result.Error != nil {
				fmt.Println("Error getting pledge points!")
				return
			}

			for i, pledge := range pledges {
				row_str := fmt.Sprintf("%d: %s (%d points)", i+1, pledge.Name, pledge.Points)
				msgs = append(msgs, row_str)
				if i%5 == 0 || i == len(pledges)-1 {
					msg := strings.Join(msgs, "\n")
					SendMessage(msg)
					// fmt.Println(msg)
					msgs = []string{}
					time.Sleep(500 * time.Millisecond)
				}
			}

			// At the end of the term, goodbye points bot!
		case "bye":
			SendMessage("goodbye : )")

		default:
			return
		}

		// Debugging measure: if you want to have your (admin's)
		// account like all of the messages to make sure they were successful,
		// uncomment this
		LikeMessage(chat.GroupId, chat.Id)
	} else if first_char == '+' || first_char == '-' {
		// If the message starts with + or -
		// Handle adding/subtracting points
		// If at any point the parsing was unsuccessful,
		// exit without doing anything

		// Validate that a brother sent this
		var brother = Brother{}
		// This checks by the user ID so doesn't matter if brother changed their name
		result := DB.First(&brother, "brother_id = ?", chat.UserId)
		if result.Error != nil {
			// A non-brother (or timeout/non-registered brother) is trying to assign points
			return
		}

		points_str, rest, _ := strings.Cut(chat.Text, " ")

		// Get the points requested
		points, err := strconv.Atoi(points_str[1:])
		if err != nil {
			// non-int points format, parsing failed
			return
		}
		if points_str[0] == '-' {
			// Point deduction handling
			points = -points
		}

		// Parse the rest of the string, separating between usernames and later text
		// This is a pain to work with
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
				// If we already have part of a name,
				// add a space to the name string to separate
				// from the rest of the name (e.g. firstname lastname)
				iter_name += " "
			}
			iter_name += cur_word
			look_for_name = false

			var pledge = Pledge{}
			// TODO: also check for matching nicknames
			// This handles new GM @ format which allows people to
			// @ by either the nickname or the proper name.
			// In this case, also change the DB to store pledge nicknames
			result := DB.Model(Pledge{}).Limit(1).Find(&pledge, "name = ?", iter_name[1:])
			if result.Error == nil && result.RowsAffected > 0 {
				// Pledge exists with this name

				// Saves the points directly to the pledge's tally,
				// ignoring the curving function and removing the need
				// for a tally. Could be useful for testing/if something breaks
				// pledge.Points = pledge.Points + points
				// DB.Save(&pledge)
				var point Point = Point{
					PointsGiven: points,
					PledgeId:    pledge.ID,
				}
				DB.Create(&point)

				look_for_name = true
				iter_name = ""
				found_name = true
			} // if no pledge exists with this name, keep looking
		}

		if found_name {
			// Debugging measure: if you want to have your (admin's)
			// account like all of the messages to make sure they were successful,
			// uncomment this
			LikeMessage(chat.GroupId, chat.Id)
		}
	}
}
