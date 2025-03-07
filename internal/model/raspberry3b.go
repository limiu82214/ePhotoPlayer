package model

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/limiu82214/ePhotoPlayer/internal"
)

type Raspberry3B struct {
	volumesPath string
}

func NewRaspberry3B() internal.System {
	// 假設 Raspberry Pi 上的 USB 裝置會掛載到 /media/$user
	username := "pi" // default value
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Error getting current user: %v", err)
	} else {
		username = currentUser.Username
	}

	volumesPath := fmt.Sprintf("/media/%s", username)
	return &Raspberry3B{
		volumesPath: volumesPath,
	}
}

// GetVolumePaths 返回 /media/pi 下的所有目錄（假設這些都是可移動裝置）
func (l *Raspberry3B) GetVolumePaths() ([]string, error) {
	entries, err := os.ReadDir(l.volumesPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %w", l.volumesPath, err)
	}

	var volumePaths []string
	for _, entry := range entries {
		// 忽略非目錄及以 "." 開頭的隱藏項目
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		volumePath := filepath.Join(l.volumesPath, entry.Name())
		// 這裡簡單假設 /media/pi 下的目錄都是可用的 USB 裝置
		volumePaths = append(volumePaths, volumePath)
	}

	return volumePaths, nil
}

// ListFilePaths 利用 ls -R 遞迴列出指定路徑下所有檔案及子目錄
func (l *Raspberry3B) ListFilePaths(path string) ([]string, error) {
	cmd := exec.Command("ls", "-R", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error listing files in %s: %w", path, err)
	}

	// 將 ls 的結果依行切割，並將每行組合成完整路徑
	lines := strings.Split(string(output), "\n")
	res := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		res = append(res, filepath.Join(path, line))
	}

	return res, nil
}

// MonitorVolumes 利用輪詢方式，每秒檢查 /media/pi 目錄內容變化
// 返回兩個 channel，分別傳送新增與移除的裝置路徑
func (l *Raspberry3B) MonitorVolumes() (<-chan string, <-chan string, error) {
	addedChan := make(chan string)
	removedChan := make(chan string)
	volumesPath := l.volumesPath

	// 檢查 /media/pi 是否存在
	if _, err := os.Stat(volumesPath); err != nil {
		return nil, nil, err
	}

	// readVolumes 讀取 /media/pi 下所有目錄，並以 map 儲存名稱
	readVolumes := func() (map[string]bool, error) {
		current := make(map[string]bool)
		entries, err := os.ReadDir(volumesPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				current[entry.Name()] = true
			}
		}
		return current, nil
	}

	// 取得初始狀態
	previous, err := readVolumes()
	if err != nil {
		return nil, nil, err
	}

	// 啟動 goroutine，每秒輪詢目錄內容
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			current, err := readVolumes()
			if err != nil {
				log.Println("Error reading volumes:", err)
				continue
			}
			// 找出新增的項目
			for name := range current {
				if !previous[name] {
					addedChan <- filepath.Join(volumesPath, name)
				}
			}
			// 找出移除的項目
			for name := range previous {
				if !current[name] {
					removedChan <- filepath.Join(volumesPath, name)
				}
			}
			previous = current
		}
	}()

	return addedChan, removedChan, nil
}
