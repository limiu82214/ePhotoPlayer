package model

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/limiu82214/ePhotoPlayer/internal"
)

type Mac struct {
	volumesPath string
}

func NewMac() internal.System {
	return &Mac{
		volumesPath: "/Volumes",
	}
}

// GetVolumePaths returns a list of paths to removable volumes.
func (m *Mac) GetVolumePaths() ([]string, error) {
	entries, err := os.ReadDir(m.volumesPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %w", m.volumesPath, err)
	}

	var volumePaths []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		volumePath := m.volumesPath + "/" + entry.Name()

		cmd := exec.Command("diskutil", "info", volumePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// should be skip if err
			return nil, fmt.Errorf("Error running diskutil info on %s: %w", volumePath, err)
		}
		outStr := string(output)

		reUSB := regexp.MustCompile(`Protocol:\s+USB`)
		reSDCard := regexp.MustCompile(`Protocol:\s+Secure Digital`)
		reRemovable := regexp.MustCompile(`Removable\s+Media:\s+Yes`)
		reExternal := regexp.MustCompile(`External:\s+Yes`)

		if reRemovable.MatchString(outStr) || reExternal.MatchString(outStr) || reUSB.MatchString(outStr) || reSDCard.MatchString(outStr) {
			volumePaths = append(volumePaths, volumePath)
		}

	}

	return volumePaths, nil
}

func (m *Mac) ListFilePaths(path string) ([]string, error) {
	cmd := exec.Command("ls", "-R", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error listing images in %s: %w", path, err)
	}

	fields := strings.Split(string(output), "\n")
	res := make([]string, 0, len(fields))
	for _, f := range fields {
		res = append(res, path+"/"+f)
	}

	return res, nil
}

func (m *Mac) MonitorVolumes() (<-chan string, <-chan string, error) {
	addedChan := make(chan string)
	removedChan := make(chan string)
	volumesPath := "/Volumes"

	// 檢查 /Volumes 是否存在
	if _, err := os.Stat(volumesPath); err != nil {
		return nil, nil, err
	}

	// readVolumes 讀取 /Volumes 目錄中所有目錄，並以 map 儲存名稱
	readVolumes := func() (map[string]bool, error) {
		current := make(map[string]bool)
		entries, err := os.ReadDir(volumesPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			// 只考慮目錄，因為 USB 裝置掛載通常都是目錄
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

	// 啟動 goroutine 每秒輪詢 /Volumes 目錄
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			current, err := readVolumes()
			if err != nil {
				log.Println("讀取 /Volumes 失敗:", err)
				continue
			}

			// 檢查新增：在 current 但不在 previous 中的
			for name := range current {
				if !previous[name] {
					addedChan <- filepath.Join(volumesPath, name)
				}
			}

			// 檢查移除：在 previous 但不在 current 中的
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
