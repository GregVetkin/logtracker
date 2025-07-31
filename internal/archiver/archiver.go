package archiver

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Logger интерфейс для логгирования
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type ZipArchiver struct {
	outputPath string
	logger     Logger
}

func NewZipArchiver(outputPath string, logger Logger) *ZipArchiver {
	return &ZipArchiver{
		outputPath: outputPath,
		logger:     logger,
	}
}

func (za *ZipArchiver) Archive(files []string) error {
	zipFile, err := os.Create(za.outputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := za.addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	za.logger.Info("Archive created", "path", za.outputPath)
	return nil
}

func (za *ZipArchiver) addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(filepath.Base(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}