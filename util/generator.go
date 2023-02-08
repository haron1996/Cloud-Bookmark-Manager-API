package util

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

func randomNumber(min, max int32) int32 {
	rand.Seed(time.Now().UnixNano())
	return min + int32(rand.Intn(int(max-min)))
}

func randomStringGen(charSet string, codeLenghth int32, stringChan chan string) {
	code := ""

	charSetLenghth := int32(len(charSet))

	for i := int32(0); i < codeLenghth; i++ {
		index := randomNumber(0, charSetLenghth)
		code += string(charSet[index])
	}

	stringChan <- code
}

func OTPGenerator() string {
	charSet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	stringChan := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		randomStringGen(charSet, 6, stringChan)
	}()

	wg.Wait()

	return <-stringChan
}

func RandomStringGenerator(strChan chan string) {
	charSet := "abcdefghijklmnopqrstuvwxyz-_"

	stringChan := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		randomStringGen(charSet, 33, stringChan)
	}()

	wg.Wait()

	s := <-stringChan

	strChan <- s
}

func GenFolderLabel(stringChan chan string) {
	charSet := "abcdefghiklmnopqrstvxyzABCDEFGHIKLMNOPQRSTVXYZ0123456789"

	stringChan1 := make(chan string, 1)

	stringChan2 := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		randomStringGen(charSet, 3, stringChan1)

	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		randomStringGen(charSet, 3, stringChan2)
	}()

	wg.Wait()

	stringChan <- strings.Join([]string{<-stringChan1, <-stringChan2}, "_")
}
