package utils_test

import (
	"fmt"
	"testing"

	"github.com/yzletter/go-postery/utils"
)

func TestSlug(t *testing.T) {
	s := "Golang学习"
	fmt.Println(utils.Slugify(s))

	s = "Golang*学习"
	fmt.Println(utils.Slugify(s))

	s = "go*Lang学习"
	fmt.Println(utils.Slugify(s))

	s = "golang*学习"
	fmt.Println(utils.Slugify(s))
}

// go test -v ./utils -run=^TestSlug$ -count=1
