package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type AppOperator struct {
	w          fyne.Window
	ctx        context.Context
	cancel     context.CancelFunc
	imagePaths []string
	mp3Paths   []string
}

func NewAppOperator(_ctx context.Context, a fyne.App) *AppOperator {
	ctx, cancel := context.WithCancel(_ctx)
	w := a.NewWindow("圖片輪播")
	return &AppOperator{ctx: ctx, cancel: cancel, w: w}
}

func (m *AppOperator) SetImagePath(imagePaths []string) {
	m.imagePaths = imagePaths
}
func (m *AppOperator) SetMp3Path(mp3Paths []string) {
	m.mp3Paths = mp3Paths
}
func (m *AppOperator) Close() {
	m.cancel()
	m.w.Close()
}

func (m *AppOperator) GetCtx() context.Context {
	return m.ctx
}

func (m *AppOperator) RunSlideShow() {
	// 建立 Fyne 應用程式
	w := m.w
	w.SetFullScreen(true)

	// 設定鍵盤事件，按下 ESC 時中止應用
	w.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		if event.Name == fyne.KeyEscape {
			m.Close()
		}
	})

	// 讀取第一張圖片
	img := canvas.NewImageFromFile(m.imagePaths[0])
	img.FillMode = canvas.ImageFillContain
	w.SetContent(img)
	w.Show()
	go m.RunMusicPlayer()

	// 開啟一個 goroutine 來進行圖片輪播，每隔固定秒數更換圖片
	go func() {
		idx := 0
		ticker := time.NewTicker(5 * time.Second) // 每 5 秒切換一次
		defer ticker.Stop()

		for range ticker.C {
			idx = (idx + 1) % len(m.imagePaths)
			// 更新圖片檔案路徑
			img.File = m.imagePaths[idx]
			img.Resource = nil // 清除舊資源，強制重新載入
			img.Refresh()      // 通知 Fyne 更新 UI
			select {
			case <-m.ctx.Done():
				return
			default:
			}
		}
	}()

	// 開啟一個 goroutine 來監聽應用程式結束事件
	go func() {
		select {
		case <-m.ctx.Done():
			m.w.Close()
		}
	}()

}

func (m *AppOperator) RunMusicPlayer() {
	if len(m.mp3Paths) == 0 {
		fmt.Println("沒有 MP3 檔案")
		return
	}

	for {
		for _, path := range m.mp3Paths {
			func() {
				f, err := os.Open(path)
				if err != nil {
					log.Printf("無法打開 %s: %v", path, err)
					return
				}
				defer f.Close()

				streamer, _, err := mp3.Decode(f)
				if err != nil {
					log.Printf("解碼 %s 失敗: %v", path, err)
					return
				}
				defer streamer.Close()

				// resampled := beep.Resample(4, format.SampleRate, targetSampleRate, streamer)

				done := make(chan bool)
				speaker.Play(beep.Seq(streamer, beep.Callback(func() {
					done <- true
				})))
				select {
				case <-done: // 等待播放完成
				case <-m.ctx.Done():
					speaker.Clear()
					return
				}
			}()
			select {
			case <-m.ctx.Done():
				return
			default:
			}
		}
	}
}
