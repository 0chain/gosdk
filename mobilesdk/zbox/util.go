package zbox

import (
	"regexp"
	"strconv"
)

func GetNumber(value string) int {
	re := regexp.MustCompile("[0-9]+")
	submatchall := re.FindAllString(value, -1)
	for _, element := range submatchall {
		res, _ := strconv.Atoi(element)
		return res
	}
	return -1
}
