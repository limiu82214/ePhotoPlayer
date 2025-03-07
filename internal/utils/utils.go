package utils

import (
	"math/rand"
	"strings"
	"time"
)

func Path2URL(path string) string {
	return "file://" + path
}

func Paths2URLs(paths []string) []string {
	urls := make([]string, len(paths))
	for i, path := range paths {
		urls[i] = Path2URL(path)
	}
	return urls
}

func FilterImagePaths(paths []string) []string {
	var imagePaths []string
	for _, path := range paths {
		if strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg") || strings.HasSuffix(strings.ToLower(path), ".png") {
			imagePaths = append(imagePaths, path)
		}
	}
	return imagePaths
}

func FilterMp3Paths(paths []string) []string {
	var mp3Paths []string
	for _, path := range paths {
		if strings.HasSuffix(strings.ToLower(path), ".mp3") {
			mp3Paths = append(mp3Paths, path)
		}
	}
	return mp3Paths
}

func ShuffleStrings(s []string) {
	rand.Seed(time.Now().UnixNano()) // 初始化隨機種子
	for i := len(s) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) // 取得 0~i 之間的隨機數
		s[i], s[j] = s[j], s[i]
	}
}

func MixStrChan(a <-chan string, b <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			select {
			case v := <-a:
				c <- v
			case v := <-b:
				c <- v
			}
		}
	}()
	return c
}
