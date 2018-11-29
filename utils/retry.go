package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

func Sqrt(f float64) (float64, error) {
	fmt.Println("start sqrt")
	if f < 0 {
		return 0, errors.New("math: square root of negative number")
	}
	fmt.Println("Ok")
	return 0, nil
}

func retryTest() {
	Retry(func() error {
		_, error := Sqrt(-1)
		return error
	}, "test", 5, 3)
}

func Retry(f func() error, describe string, attempts int, sleep int) error {
	err := f()
	if err == nil {
		return nil
	}
	if s, ok := err.(stop); ok {
		// Return the original error for later checking
		return s.error
	}
	if attempts--; attempts > 0 {
		// Add some randomness to prevent creating a Thundering Herd
		jitter := rand.Int63n(int64(sleep))
		fmt.Println(int(jitter))
		sleep = sleep + int(jitter/2)
		logrus.Errorf("执行:%s失败，错误为:%v，将在%d秒后重试，剩余%d次", describe, err.Error(), sleep, attempts)
		time.Sleep(time.Duration(sleep) * time.Second)
		return Retry(f, describe, attempts, 2*sleep)
	}
	return err
}

type stop struct {
	error
}
