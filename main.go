package main

import (
	"encoding/json"
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var watchDir string
var watchInterval int
var handbrakePreset string
var convertedSuffix string
var minimumSize int
var bufferSize int
var waitTime int
var verbose bool

var handbrakePresetName string

func main() {
	flag.Usage = func() {
		fmt.Print(`
handbraked: watch and convert videos in a directory using Handbrake
Usage:
	handbraked -d WATCH_DIR -p PRESET_PATH [OPTIONS]
Options:
`[1:])
		flag.PrintDefaults()
	}
	flag.StringVarP(&watchDir, "watch-dir", "d", "", "Directory to watch and automatically convert new videos")
	flag.IntVarP(&watchInterval, "interval", "i", 5, "Interval in seconds between checking for new videos")
	flag.StringVarP(&convertedSuffix, "suffix", "s", "-x265", "Suffix to add to converted videos. "+
		"Matching files will be excluded from conversion")
	flag.StringVarP(&handbrakePreset, "preset", "p", "", "Path to Handbrake preset used for conversion")
	flag.IntVarP(&minimumSize, "min-size", "m", 1_000_000, "Minimum converted file size in bytes, will otherwise error and terminate")
	flag.IntVarP(&bufferSize, "buffer-size", "b", 2, "Number of pending videos to always keep intact before starting to convert")
	flag.IntVarP(&waitTime, "wait-time", "t", 10, "Time in seconds to wait since a video's modification time before starting conversion")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	flag.Parse()
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}
	if err := work(); err != nil {
		log.Fatalln(err)
	}
}

type PresetFormat struct {
	PresetList []struct {
		PresetName string `json:"PresetName"`
	} `json:"PresetList"`
}

func work() error {
	if err := validateFlags(); err != nil {
		return err
	}
	if err := parsePresetName(); err != nil {
		return errors.New("parse preset name: " + err.Error())
	}
	log.Println("Started daemon")
	for {
		if err := workLoop(); err != nil {
			log.Println("ERROR: " + err.Error())
		}
		<-time.After(time.Duration(watchInterval) * time.Second)
	}
}

func parsePresetName() error {
	presetFile, err := os.Open(handbrakePreset)
	if err != nil {
		return err
	}
	defer presetFile.Close()
	presetBytes, err := io.ReadAll(presetFile)
	if err != nil {
		return err
	}
	var preset PresetFormat
	if err := json.Unmarshal(presetBytes, &preset); err != nil {
		return err
	}
	if len(preset.PresetList) < 1 {
		return errors.New("no presets found in file")
	} else if len(preset.PresetList) > 1 {
		log.Println("Found more than one preset in file, using first")
	}
	handbrakePresetName = preset.PresetList[0].PresetName
	return nil
}

func validateFlags() error {
	for name, item := range map[string]string{
		"--watch-dir": watchDir,
		"--preset":    handbrakePreset,
	} {
		if item == "" {
			flag.Usage()
			return errors.New("invalid " + name + " flag")
		}
		if _, err := os.Stat(item); os.IsNotExist(err) {
			return errors.New(name + " does not exist: " + item)
		} else if err != nil {
			return errors.New("unknown error: " + err.Error())
		}
	}
	if _, err := exec.LookPath("HandbrakeCLI"); err != nil {
		return err
	}
	return nil
}

func workLoop() error {
	if verbose {
		log.Println("Checking watch directory")
	}
	dirs, err := os.ReadDir(watchDir)
	if err != nil {
		return err
	}
	videoMap := map[os.DirEntry]os.FileInfo{}
	var videoKeys []os.DirEntry
	for _, dir := range dirs {
		if dir.IsDir() || dir.Name()[0] == '.' || strings.HasSuffix(removeExtension(dir.Name()), convertedSuffix) {
			continue
		}
		info, err := dir.Info()
		if err != nil {
			return err
		}
		if info.ModTime().Add(time.Duration(waitTime) * time.Second).After(time.Now()) {
			continue
		}
		videoMap[dir] = info
		videoKeys = append(videoKeys, dir)
	}
	if len(videoMap) < bufferSize+1 {
		if verbose {
			log.Println("Nothing to process")
		}
		return nil
	}
	sort.Slice(videoKeys, func(i, j int) bool {
		return videoMap[videoKeys[i]].ModTime().After(videoMap[videoKeys[j]].ModTime())
	})
	var combinedError error
	for _, video := range videoKeys[bufferSize:] {
		log.Println("Processing " + video.Name())
		if err := handbrake(filepath.Join(watchDir, video.Name())); err != nil {
			if combinedError == nil {
				combinedError = err
			} else {
				combinedError = errors.New(combinedError.Error() + " , " + err.Error())
			}
		}
		log.Println("Done processing")
	}
	return combinedError
}

func removeExtension(inputPath string) string {
	return inputPath[:len(inputPath)-len(filepath.Ext(inputPath))]
}

func handbrake(inputPath string) error {
	outputPath := removeExtension(inputPath) + convertedSuffix + filepath.Ext(inputPath)
	command := exec.Command("HandbrakeCLI", "-i", inputPath, "-o", outputPath, "--preset-import-file", handbrakePreset, "--preset", handbrakePresetName)
	output, err := command.CombinedOutput()
	if err != nil {
		return errors.New("Handbrake errored:\n" + string(output))
	}
	stat, err := os.Stat(outputPath)
	if err != nil {
		return errors.New("error reading converted video: " + err.Error())
	}
	if stat.Size() < int64(minimumSize) {
		return errors.New("converted video too small")
	}
	if err := os.Remove(inputPath); err != nil {
		return errors.New("error removing old video: " + err.Error())
	}
	return nil
}
