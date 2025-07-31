package tracker

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
	"path/filepath"
)

// Logger интерфейс для логгирования
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type Config struct {
	FilesToWatch []string
}

type FileTracker struct {
	config      Config
	logger      Logger
	fileHandles map[string]*os.File
	filePos     map[string]int64
	newFiles    map[string]string
	mu          sync.Mutex
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

func New(config Config, logger Logger) *FileTracker {
	return &FileTracker{
		config:      config,
		logger:      logger,
		fileHandles: make(map[string]*os.File),
		filePos:     make(map[string]int64),
		newFiles:    make(map[string]string),
		stopChan:    make(chan struct{}),
	}
}

func (ft *FileTracker) Start() error {
	for _, file := range ft.config.FilesToWatch {
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		info, err := f.Stat()
		if err != nil {
			f.Close()
			return err
		}

		ft.fileHandles[file] = f
		ft.filePos[file] = info.Size()
		ft.newFiles[file] = "/tmp/" + filepath.Base(file)

		ft.wg.Add(1)
		go ft.monitorFile(file)
	}
	return nil
}

func (ft *FileTracker) monitorFile(filename string) {
	defer ft.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ft.checkNewLines(filename)
		case <-ft.stopChan:
			return
		}
	}
}

func (ft *FileTracker) checkNewLines(filename string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	// Получаем текущий размер файла
	info, err := ft.fileHandles[filename].Stat()
	if err != nil {
		ft.logger.Error("Failed to get file stats", "file", filename, "error", err)
		return
	}

	// Если файл уменьшился (например, был перезаписан), сбрасываем позицию
	if info.Size() < ft.filePos[filename] {
		ft.filePos[filename] = 0
	}

	// Перемещаемся к последней известной позиции
	if _, err := ft.fileHandles[filename].Seek(ft.filePos[filename], io.SeekStart); err != nil {
		ft.logger.Error("Seek failed", "file", filename, "error", err)
		return
	}

	// Читаем новые данные
	scanner := bufio.NewScanner(ft.fileHandles[filename])
	var newLines []byte

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 { // Пропускаем полностью пустые строки
			continue
		}
		newLines = append(newLines, line...)
	}

	if len(newLines) > 0 {
		// Открываем файл для добавления с буферизацией
		outputFile, err := os.OpenFile(
			ft.newFiles[filename],
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			ft.logger.Error("Failed to open output file", "file", filename, "error", err)
			return
		}
		defer outputFile.Close()

		writer := bufio.NewWriter(outputFile)
		if _, err := writer.Write(newLines); err != nil {
			ft.logger.Error("Write failed", "file", filename, "error", err)
			return
		}
		writer.Flush()

		// Обновляем позицию
		ft.filePos[filename] = ft.filePos[filename] + int64(len(newLines))
	}
}

func (ft *FileTracker) Stop() {
	close(ft.stopChan)
	ft.wg.Wait()
	for _, f := range ft.fileHandles {
		f.Close()
	}
}

func (ft *FileTracker) NewFiles() []string {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	files := make([]string, 0, len(ft.newFiles))
	for _, file := range ft.newFiles {
		files = append(files, file)
	}
	return files
}
