package messages

import (
	"strings"
	"time"
	"strconv"
	"log"
)

type MessageConstructor func(args []string) (Message, error)

var messageTypes = map[string]map[string]MessageConstructor{}

func ParseMessage(src []byte) (keyword string, args []string, err error) {
	parts := strings.Split(string(src), " ")

	if len(parts) == 0 {
		err = ErrInvalidMessage
		return
	}

	keyword = parts[0]
	args = parts[1:]
	err = nil

	return
}

func ToBytes(keyword string, args []string) []byte {
	words := []string{keyword, strings.Join(args, " ")}
	return []byte(strings.Join(words, " "))
}

func DurationToString(duration time.Duration) string {
	count := int64(time.Duration(duration) / time.Millisecond)
	return strconv.FormatInt(count, 10)
}

func StringToDuration(src string) (duration time.Duration, err error) {
	count, err := strconv.Atoi(src)
	duration = time.Millisecond * time.Duration(count)
	return
}

func RegisterMessageType(category string, keyword string, constructor MessageConstructor) {
	if _, ok := messageTypes[category]; !ok {
		messageTypes[category] = map[string]MessageConstructor{}
	}

	messageTypes[category][keyword] = constructor
}

func printTypes() {
	log.Print("Currently registered message types")
	for category, keywords := range messageTypes {
		log.Printf("%s:", category)
		for keyword, _ := range keywords {
			log.Printf(" - %s", keyword)
		}
		log.Println("")
	}
}


func LoadMessage(category string, src []byte) (keyword string, message Message, err error) {
	keyword, args, err := ParseMessage(src)

	if err != nil {
		return
	}

	if _, ok := messageTypes[category]; !ok {
		err = ErrInvalidCategory
		return
	}

	constructor, ok := messageTypes[category][keyword]

	if !ok {
		err = InvalidKeyword(keyword)
		printTypes()
		return
	}

	message, err = constructor(args)

	return
}
