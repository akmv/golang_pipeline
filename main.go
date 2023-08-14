package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync/atomic"
	"time"
)

const ElementCount = 10
const MaxElementValue = 10

func getEnvAsInt(name string, defaultVal int) int {
	valueStr, exists := os.LookupEnv(name)
	if !exists {
		return defaultVal
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}

	return valueInt
}

func genSlice(workerCount, genTimeout int) <-chan []int {
	out := make(chan []int, workerCount)
	go func() {
		for {
			arr := make([]int, ElementCount)
			for i := 0; i < ElementCount; i++ {
				arr[i] = rand.Intn(MaxElementValue)
			}
			out <- arr
			time.Sleep(time.Millisecond * time.Duration(genTimeout))
		}
	}()
	return out
}

func findThreeMaxNumbers(ch <-chan []int) <-chan []int {
	out := make(chan []int)
	go func() {
		for arr := range ch {
			sort.Ints(arr)
			out <- arr[len(arr)-3:]
		}
	}()

	return out
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
		os.Exit(1)
	}

	genTimeout := getEnvAsInt("PACKAGE_GENERATION_TIMEOUT", 500)
	printTimeout := getEnvAsInt("PRINT_TIMEOUT", 1)
	workerCount := getEnvAsInt("WORKER_COUNT", 5)
	fmt.Println("PACKAGE_GENERATION_TIMEOUT =", genTimeout, "PRINT_TIMEOUT =", printTimeout, "WORKER_COUNT =", workerCount)

	rand.Seed(time.Now().Unix())
	workerChannels := make([]<-chan []int, workerCount)
	genSliceChannel := genSlice(workerCount, genTimeout)
	for i := 0; i < workerCount; i++ {
		workerChannels[i] = findThreeMaxNumbers(genSliceChannel)
	}

	var sum uint64

	accumulate := func(ch <-chan []int) {
		for arr := range ch {
			localSum := 0
			for i := 0; i < len(arr); i++ {
				localSum += arr[i]
			}
			atomic.AddUint64(&sum, uint64(localSum))
		}
	}

	for _, ch := range workerChannels {
		go accumulate(ch)
	}

	for {
		fmt.Println(sum)
		time.Sleep(time.Second * time.Duration(printTimeout))
	}
}
