package monitor

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType 文件事件类型
type EventType string

const (
	EventCreate EventType = "CREATE"
	EventWrite  EventType = "WRITE"
	EventRename EventType = "RENAME"
)

// FileEvent 文件事件
type FileEvent struct {
	Path      string
	Type      EventType
	Timestamp time.Time
	Size      int64
}

// EventHandler 事件处理回调
type EventHandler func(event FileEvent)

// Watcher 文件监控器
type Watcher struct {
	fsWatcher    *fsnotify.Watcher // 文件监控实例
	handler      EventHandler      // 自定义处理逻辑
	watchDir     string
	done         chan struct{}        // 停止信号通道
	wg           sync.WaitGroup       // 等待组
	mutex        sync.Mutex           // 锁
	recentEvents map[string]time.Time // 去重，避免短时间内重复事件
}

// NewWatcher 创建新的文件监控器
func NewWatcher(watchDir string, handler EventHandler) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher() // 创建文件监控实例
	if err != nil {
		return nil, fmt.Errorf("创建文件监控失败：%w", err)
	}
	w := &Watcher{
		fsWatcher:    fsWatcher,
		handler:      handler,
		watchDir:     watchDir,
		done:         make(chan struct{}),
		wg:           sync.WaitGroup{},
		mutex:        sync.Mutex{},
		recentEvents: make(map[string]time.Time),
	}
	return w, nil
}

// Start 开始监控
func (w *Watcher) Start() error {
	// 递归添加监控目录
	err := w.AddRecursive(w.watchDir)
	if err != nil {
		return err
	}
	w.wg.Add(1)      // 增加一个等待任务
	go w.EventLoop() // 启动事件循环

	w.wg.Add(1)        // 增加一个等待任务
	go w.CleanupLoop() // 启动清理循环

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("开始监控目录：%s\n", w.watchDir)
	return nil
}

// Stop 停止监控
func (w *Watcher) Stop() error {
	close(w.done)              // 关闭通道，停止信号
	w.wg.Wait()                // 等待所有任务完成
	return w.fsWatcher.Close() // 关闭文件监控实例
}

// AddRecursive 递归添加目录
func (w *Watcher) AddRecursive(dir string) error {
	// 递归遍历目录
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的路径
		}
		if info.IsDir() {
			if len(filepath.Base(path)) > 1 && filepath.Base(path)[0] == '.' {
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(path)
		}
		return nil
	})
}

// EventLoop 事件处理主循环
func (w *Watcher) EventLoop() {
	defer w.wg.Done() // 结束等待任务
	for {
		select {
		case <-w.done: // 情况1：停止信号
			return
		case event, ok := <-w.fsWatcher.Events: // 情况2：文件事件
			if !ok {
				return
			}
			w.processEvent(event)
		case err, ok := <-w.fsWatcher.Errors: // 情况3：错误事件
			if !ok {
				return
			}
			fmt.Printf("监控错误：%v\n", err)
		}
	}
}

func (w *Watcher) processEvent(event fsnotify.Event) {
	fmt.Printf("[Watcher] 原始事件: %s %v\n", event.Name, event.Op)

	var eventType EventType
	switch {
	case event.Op&fsnotify.Create != 0:
		eventType = EventCreate
		fmt.Printf("[Watcher] 检测到创建事件: %s\n", event.Name)
		// 新目录自动加入监控
		info, err := os.Stat(event.Name)
		if err == nil && info.IsDir() {
			fmt.Printf("[Watcher] 新目录，添加监控: %s\n", event.Name)
			_ = w.fsWatcher.Add(event.Name)
			return
		}
	case event.Op&fsnotify.Write != 0:
		eventType = EventWrite
		fmt.Printf("[Watcher] 检测到写入事件: %s\n", event.Name)
	case event.Op&fsnotify.Rename != 0:
		eventType = EventRename
		fmt.Printf("[Watcher] 检测到重命名事件: %s\n", event.Name)
	default:
		fmt.Printf("[Watcher] 忽略事件类型: %v\n", event.Op)
		return
	}
	// 去重检测
	w.mutex.Lock() // 加锁
	// 获取上一次事件时间
	lastTime, exists := w.recentEvents[event.Name]
	now := time.Now()
	// 忽略短时间内重复事件
	if exists && now.Sub(lastTime) < 500*time.Millisecond {
		w.mutex.Unlock()
		return
	}
	// 更新文件上一次事件时间
	w.recentEvents[event.Name] = now
	w.mutex.Unlock()

	// 获取文件信息
	var size int64
	info, err := os.Stat(event.Name)
	if err == nil {
		size = info.Size()
	}

	fileEvent := FileEvent{
		Path:      event.Name,
		Type:      eventType,
		Timestamp: now,
		Size:      size,
	}

	// 异步调用处理器,启动一个新的 goroutine 来处理事件
	go w.handler(fileEvent)
}

// CleanupLoop 定期清理去重缓存
func (w *Watcher) CleanupLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // 设置定时器，每30秒执行一次
	defer ticker.Stop()

	for {
		select {
		case <-w.done: // 停止信号
			return
		case <-ticker.C: // 定时器触发，每30秒触发一次
			w.mutex.Lock()
			now := time.Now()
			for path, t := range w.recentEvents {
				if now.Sub(t) > 5*time.Second { // // 如果事件时间距离现在超过5秒，判定为过期
					delete(w.recentEvents, path) // 删除数据
				}
			}
			w.mutex.Unlock()
		}
	}
}
