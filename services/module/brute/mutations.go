package brute

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	numberRegex    = regexp.MustCompile("[0-9]+")
	colorMutations = []string{
		"red",
		"blue",
		"green",
		"white",
		"black",
		"gray",
		"silver",
		"purple",
		"yellow",
	}
)

// NumberMutation finds and mutates all numbers +/- 5 and returns the
// subdomain without any numbers (so it can be used in unique checks)
func NumberMutation(in string) []string {
	//results := numberRegex.FindAll([]byte(in), -1)
	indexRanges := numberRegex.FindAllStringIndex(in, -1)
	if len(indexRanges) == 0 {
		return []string{}
	}
	//mutatedNumbers([][]string, 0)
	mutations := make([]string, 0)

	for _, index := range indexRanges {
		number, _ := strconv.Atoi(in[index[0]:index[1]])
		start := in[0:index[0]]
		end := in[index[1]:len(in)]
		mutatedNumbers := MutateNumber(number, 5)
		for _, mutation := range mutatedNumbers {
			mutations = append(mutations, start+mutation+end)
		}
	}

	return mutations
}

// UnmutateNumber removes all numbers and returns it so we can do uniqueness checks
func UnmutateNumber(in string) string {
	result := numberRegex.Split(in, -1)
	return strings.Join(result, "")
}

func MutateNumber(number, amount int) []string {
	numbers := make([]string, 0)
	for i := number + 1; i < number+amount+1; i++ {
		numbers = append(numbers, strconv.Itoa(i))
	}

	for i := number - 1; i > number-amount-1; i-- {
		if i < 0 {
			break
		}
		numbers = append(numbers, strconv.Itoa(i))
	}
	return numbers
}
