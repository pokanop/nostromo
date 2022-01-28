package config

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
)

// backupManifest at config path based on timestamp
func backupManifest(m *model.Manifest) error {
	// Before saving backup, prune old files
	pruneBackups(m)

	// Prevent backups if max count is 0
	if m.Config.BackupCount == 0 {
		return nil
	}

	// Create backups under base dir in a folder
	backupDir, err := ensureBackupDir()
	if err != nil {
		return err
	}

	// Copy existing manifest to backup path
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UnixNano()/int64(time.Millisecond))
	destinationFile := filepath.Join(backupDir, m.Name+"_"+ts+".yaml")
	sourceFile := pathutil.Abs(m.Path)

	// Check if manifest exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return nil
	}

	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func pruneBackups(m *model.Manifest) {
	backupDir, err := ensureBackupDir()
	if err != nil {
		return
	}

	// Read all files, sort by timestamp, and drop items > max count
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		log.Warningf("unable to read backup dir: %s\n", err)
		return
	}

	// Filter file list for matching names
	matches := []fs.FileInfo{}
	for _, file := range files {
		if isMatch, _ := regexp.MatchString(fmt.Sprintf("%s_\\d+\\.yaml", m.Name), file.Name()); isMatch {
			matches = append(matches, file)
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].ModTime().After(matches[j].ModTime())
	})

	// Add one more to backup count since a new backup will be created
	maxCount := m.Config.BackupCount - 1
	if maxCount < 0 {
		maxCount = 0
	}
	if len(matches) > maxCount {
		for _, file := range matches[maxCount:] {
			filename := filepath.Join(backupDir, file.Name())
			if err := os.Remove(filename); err != nil {
				log.Warningf("failed to prune backup file %s: %s\n", filename, err)
			}
		}
	}
}

func ensureBackupDir() (string, error) {
	backupDir := backupsPath()
	if err := pathutil.EnsurePath(backupDir); err != nil {
		return "", err
	}
	return pathutil.Abs(backupDir), nil
}
