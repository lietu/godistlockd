package messages

import (
	"strings"
	"time"
	"strconv"
)

type MessageConstructor func(args []string) (Message, error)

var messageTypes = map[string]map[string]MessageConstructor{}

func ParseMessage(src []byte) (keyword string, args []string, err error) {
	parts := strings.Split(string(src), " ")

	if len(parts) == 0 {
		err = ERR_INVALID_MESSAGE
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

func LoadMessage(category string, src []byte) (keyword string, message Message, err error) {
	keyword, args, err := ParseMessage(src)

	if err != nil {
		return
	}

	if _, ok := messageTypes[category]; !ok {
		err = ERR_INVALID_CATEGORY
		return
	}

	constructor, ok := messageTypes[category][keyword]

	if !ok {
		err = ERR_INVALID_KEYWORD
		return
	}

	message, err = constructor(args)

	return
}
