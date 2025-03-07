package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/limiu82214/ePhotoPlayer/internal"
	"github.com/limiu82214/ePhotoPlayer/internal/model"
	"github.com/limiu82214/ePhotoPlayer/internal/utils"

	"fyne.io/fyne/v2/app"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

func main() {
	runtime.LockOSThread() // for key fyne
	var system internal.System

	switch runtime.GOOS {
	case "darwin":
		fmt.Println("SYS: Running on macOS")
		system = model.NewMac()
	case "linux":
		fmt.Println("SYS: Running on Linux")
		system = model.NewRaspberry3B()
	default:
		fmt.Printf("未知作業系統: %s\n", runtime.GOOS)
	}

	addedChan, removedChan, err := system.MonitorVolumes()
	if err != nil {
		fmt.Println("Error monitoring volumes:", err)
		return
	}

	// 第一次直接拿可以拿到的
	firstChan := make(chan string)
	addedChan = utils.MixStrChan(addedChan, firstChan)

	paths, err := system.GetVolumePaths()
	if err != nil {
		fmt.Println("Error getting volume paths:", err)
		return
	}
	if len(paths) > 0 {
		firstChan <- paths[0]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mainApp := app.New()
	// 設定一個固定的 targetSampleRate，例如 44100 Hz
	targetSampleRate := beep.SampleRate(44100)
	// 初始化 speaker 使用固定的採樣率
	speaker.Init(targetSampleRate, targetSampleRate.N(time.Second/10))
	defer speaker.Close()

	var op *internal.AppOperator
	var opCtx context.Context = context.Background()
	var currentVolumePath string
	go func() {
		time.Sleep(1 * time.Second)
		for {
			select {
			case addedPath := <-addedChan:
				fmt.Println("SYS: USB added:", addedPath)
				if currentVolumePath == "" {
					currentVolumePath = addedPath
					var imagePaths, mp3Paths []string
					for range 5 {
						imagePaths, mp3Paths, err = getPath(system, currentVolumePath)
						if err != nil {
							fmt.Println("Error getting paths:", err)
							time.Sleep(1 * time.Second)
						} else {
							break
						}
					}
					if err != nil {
						continue
					}

					op = internal.NewAppOperator(ctx, mainApp)
					opCtx = op.GetCtx()
					op.SetImagePath(imagePaths)
					op.SetMp3Path(mp3Paths)
					go op.RunSlideShow()
					go op.RunMusicPlayer()
				}
			case removedPath := <-removedChan:
				fmt.Println("SYS: USB removed:", removedPath)
				if removedPath == currentVolumePath {
					op.Close()
					op = nil
					opCtx = context.Background()
					currentVolumePath = ""
				}
			case <-opCtx.Done():
				fmt.Println("SYS: OP closed")
				op.Close()
				op = nil
				opCtx = context.Background()
				currentVolumePath = ""
			}
		}
	}()

	// app
	fmt.Println("SYS: Running app")
	mainApp.NewWindow("ePhotoPlayer keepAlive").Show()
	mainApp.Run()
	fmt.Print("SYS: App closed")
}

func getPath(system internal.System, volumePath string) (imagePaths, mp3Paths []string, err error) {
	fmt.Println("SYS: Getting USB volume paths")
	// volume
	paths, err := system.GetVolumePaths()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting volume paths: %w", err)
	}
	if len(paths) == 0 {
		return nil, nil, errors.New("No removable volumes found")
	}

	// path
	basePath := paths[0]
	photosPath := filepath.Join(basePath, "photos")
	mp3Path := filepath.Join(basePath, "music")

	imagePaths, err = system.ListFilePaths(photosPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error listing files: %w", err)
	}
	imagePaths = utils.FilterImagePaths(imagePaths)
	utils.ShuffleStrings(imagePaths)
	fmt.Print("SYS: Image paths counts ", len(imagePaths))
	if len(imagePaths) > 0 {
		fmt.Println(" - first path:", imagePaths[0])
	} else {
		fmt.Println(" - no image files found")
	}

	mp3Paths, err = system.ListFilePaths(mp3Path)
	if err != nil {
		return nil, nil, fmt.Errorf("Error listing files: %w", err)
	}
	mp3Paths = utils.FilterMp3Paths(mp3Paths)
	utils.ShuffleStrings(mp3Paths)
	fmt.Print("SYS: MP3 paths counts ", len(mp3Paths))
	if len(mp3Paths) > 0 {
		fmt.Println(" - first path:", mp3Paths[0])
	} else {
		fmt.Println(" - no mp3 files found")
	}

	return imagePaths, mp3Paths, nil
}
