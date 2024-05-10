package utilities

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/adapters/repos/db/vector/hnsw"
	"github.com/weaviate/weaviate/entities/cyclemanager"

	"github.com/spf13/cobra"
)

var combineCommitLogCmd = &cobra.Command{
	Use:   "combine-commit-logs <path>",
	Short: "Combine HNSW commit logs to reduce startup time",
	Long:  `Combine HNSW commit logs to reduce startup time`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Running commit log combiner")
		basePath := filepath.Clean(args[0])
		name := "main"
		targetThreshold := 1024 * 1024 * 5000 // 5GB
		commitLogPath := fmt.Sprintf("%s/%s.hnsw.commitlog.d", basePath, name)

		log.WithField("path", basePath).Info("Path value")
		log.WithField("path", commitLogPath).Info("Commit log value")

		err := validatePath(commitLogPath)
		if err != nil {
			log.WithError(err).Fatal("Path validation failed")
		}

		err = createSentinelFile(commitLogPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to create sentinel file")
		}

		// TODO do we need to wait here if commit log in in progress?

		workingName := "working"
		workingPath := filepath.Join(basePath, fmt.Sprintf("%s.hnsw.commitlog.d", workingName))
		err = os.MkdirAll(workingPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Fatal("Failed to create working folder")
		}

		selectedFiles, err := selectCommitLogs(commitLogPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to select commit logs")
		}

		err = copyCommitLogs(selectedFiles, commitLogPath, workingPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to copy commit logs")
		}

		logger := log.New()
		commitLogger, err := hnsw.NewCommitLogger(basePath, workingName, logger,
			cyclemanager.NewCallbackGroupNoop(),
			hnsw.WithCommitlogThresholdForCombining(int64(targetThreshold)),
			hnsw.WithCommitlogThreshold(int64(targetThreshold/5)))
		if err != nil {
			log.WithError(err).Fatal("Failed to create commit logger")
		}

		err = commitLogger.CombineAndCondenseLogs()
		if err != nil {
			log.WithError(err).Fatal("Failed to combine and condense logs")
		}

		// Remove the selected files
		for _, file := range selectedFiles {
			filePath := filepath.Join(commitLogPath, file)
			err = os.Remove(filePath)
			if err != nil {
				log.WithError(err).WithField("file", file).Fatal("Failed to remove file")
			}
			log.WithField("file", file).Info("Removed commit log")
		}

		// Copy the combined files to the main folder
		var combinedFiles []string
		files, err := os.ReadDir(workingPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to read directory")
		}
		for _, file := range files {
			combinedFiles = append(combinedFiles, file.Name())
		}

		err = copyCommitLogs(combinedFiles, workingPath, commitLogPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to copy new combined commit logs")
		}

		// // Remove the working folder
		err = os.RemoveAll(workingPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to remove working folder")
		}

		err = removeSentinelFile(commitLogPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to remove sentinel file")
		}
	},
}

func NewCombineCommitLogCmd() *cobra.Command {
	return combineCommitLogCmd
}

func validatePath(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("path must be a folder: %s", path)
	}

	// Check if the path ends with "main.hnsw.commitlog.d/"
	if !strings.HasSuffix(filepath.ToSlash(path), "main.hnsw.commitlog.d") {
		return fmt.Errorf("path must end with 'main.hnsw.commitlog.d'")
	}

	return nil
}

func copyCommitLogs(selectedFiles []string, basePath string, workingPath string) error {

	// Copy the selected files to the new folder
	for _, file := range selectedFiles {
		srcPath := filepath.Join(basePath, file)
		dstPath := filepath.Join(workingPath, file)

		src, err := os.Open(srcPath)
		if err != nil {
			log.WithError(err).WithField("file", file).Panic("Failed to open source file")
			continue
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			log.WithError(err).WithField("file", file).Panic("Failed to create destination file")
			continue
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			log.WithError(err).WithField("file", file).Panic("Failed to copy file")
			continue
		}

		log.WithField("file", file).Info("Copied commit log")
	}

	return nil
}

func selectCommitLogs(path string) ([]string, error) {

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %s", err)
	}

	var filteredFiles []string
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.HasPrefix(fileName, "17") {
				filteredFiles = append(filteredFiles, fileName)
			}
		}
	}

	// Sort the filtered files by name
	sort.Strings(filteredFiles)

	// Exclude the last 10 files or return an empty list if there are 10 or fewer files
	if len(filteredFiles) <= 10 {
		return []string{}, nil
	}
	filteredFiles = filteredFiles[:len(filteredFiles)-10]

	// for _, file := range filteredFiles {
	// 	log.WithField("file", file).Info("Selected commit log")
	// }

	return filteredFiles, nil
}

func createSentinelFile(path string) error {
	// Check if the path is a folder
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("path must be a folder: %s", path)
	}

	// Create the disabled sentinel file
	disabledFilePath := filepath.Join(path, "disabled")
	_, err = os.Create(disabledFilePath)
	if err != nil {
		return fmt.Errorf("failed to create disabled sentinel file: %s", err)
	}

	log.WithField("path", disabledFilePath).Info("Created disabled sentinel file")
	return nil
}

func removeSentinelFile(path string) error {
	// Check if the path is a folder
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("path must be a folder: %s", path)
	}

	// Remove the disabled sentinel file
	disabledFilePath := filepath.Join(path, "disabled")
	err = os.Remove(disabledFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("Disabled sentinel file does not exist: %s\n", disabledFilePath)
			return nil
		}
		return fmt.Errorf("failed to remove disabled sentinel file: %s", err)
	}

	log.WithField("path", disabledFilePath).Info("Removed disabled sentinel file")
	return nil
}
