package utilities

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
		targetThreshold := 1024 * 1024 * 24000 // 24GiB
		dontTouchLastFiles := 10
		totalFileLimit := 400
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

		log.Info("wait 120s in case something is still in progress")
		time.Sleep(120 * time.Second)

		workingName := "working"
		workingPath := filepath.Join(basePath, fmt.Sprintf("%s.hnsw.commitlog.d", workingName))
		backupPath := filepath.Join(basePath, fmt.Sprintf("main.hnsw.commitlog.d.%d.bak", time.Now().Unix()))

		err = os.MkdirAll(workingPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Fatal("Failed to create working folder")
		}
		err = os.MkdirAll(backupPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Fatal("Failed to create backup folder")
		}

		selectedFiles, err := selectCommitLogs(commitLogPath, dontTouchLastFiles, totalFileLimit)
		if err != nil {
			log.WithError(err).Fatal("Failed to select commit logs")
		}

		log.Infof("start copying into working path: %s", workingPath)
		err = copyCommitLogs(selectedFiles, commitLogPath, workingPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to copy commit logs into working dir")
		}

		log.Infof("start copying into backup path: %s", backupPath)
		err = copyCommitLogs(selectedFiles, commitLogPath, backupPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to copy commit logs into backup")
		}

		logger := log.New()
		commitLogger, err := hnsw.NewCommitLogger(basePath, workingName, logger,
			cyclemanager.NewCallbackGroupNoop(),
			hnsw.WithCommitlogThresholdForCombining(int64(targetThreshold)),
			hnsw.WithCommitlogThreshold(int64(targetThreshold/5)))
		if err != nil {
			log.WithError(err).Fatal("Failed to create commit logger")
		}

		i := 0
		for {
			var ok1 bool
			var ok2 bool
			var err error

			ok := true
			for ok {
				ok, err = commitLogger.CombineLogs()
				if ok {
					ok1 = true
				}
				if err != nil {
					log.WithError(err).Fatal("Failed to combine commit logs")
				}
			}

			ok = true
			for ok {
				ok, err = commitLogger.CondenseOldLogs()
				if ok {
					ok2 = true
				}
				if err != nil {
					log.WithError(err).Fatal("Failed to condense commit logs")
				}
			}

			i++
			ok = ok1 || ok2
			if !ok {
				// never entered either loop, we are done!
				log.Infof("completing combine and condense loop after %d iterations", i)
				break
			}
		}

		err = commitLogger.Flush()
		if err != nil {
			log.WithError(err).Fatal("Failed to flush commit logger")
		}

		err = commitLogger.Shutdown(context.Background())
		if err != nil {
			log.WithError(err).Fatal("Failed to shutdown commit logger")
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

func selectCommitLogs(path string, dontTouchLastFiles int, totalLimit int) ([]string, error) {
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

	// Exclude the last dontTouchLastFiles files or return an empty list if there are 10 or fewer files
	if len(filteredFiles) <= dontTouchLastFiles {
		return []string{}, nil
	}
	filteredFiles = filteredFiles[:len(filteredFiles)-dontTouchLastFiles]

	// for _, file := range filteredFiles {
	// 	log.WithField("file", file).Info("Selected commit log")
	// }

	if len(filteredFiles) > totalLimit {
		log.Infof("Found %d eligibile files, but limit is set to %d, ignoring remaining files", len(filteredFiles), totalLimit)
		return filteredFiles[:totalLimit], nil
	}

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
