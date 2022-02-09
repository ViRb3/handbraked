# handbraked

>  Watch and convert videos in a directory using [Handbrake](https://handbrake.fr)

## Introduction

This tool was created to solve a practical problem when using [Topaz Video Enhance AI](https://www.topazlabs.com/video-enhance-ai) with large collections of old shows. The upscaled videos were encoded very poorly, resulting in enormous file sizes, often 10x larger than the original, even though the upscale was only 2x or so.

handbraked ("d" for [daemon](https://en.wikipedia.org/wiki/Daemon_(computing))) is designed to run in parallel to the upscale program, picking up upscaled videos as soon as they're ready and re-encoding them using a high-efficiency algorithm, replacing the old inefficient video when done. Since both programs run at the same time, given that you have a powerful enough machine, you save both time and storage.

While initially designed for this use case, handbraked is universal enough to work with any upscaling program, and probably completely different workflows as well.

## Usage

1. Install [Handbrake CLI](https://handbrake.fr/downloads2.php) and make sure it is added to your PATH, so that it can be executed by calling `HandBrakeCLI` from any terminal.

2. Prepare a Handbrake preset file with your desired encoding settings - a x265 example tuned for cartoons on a MacBook Pro M1 Max can be found under [preset.json](preset.json). You will likely want to experiemnt with the quality settings so that the encoding does not slow down your upscaling (or vice versa).

3. Start your upscaling program and make it output the finished videos to a directory of your choice, let's call it `CONVERTED`.

4. Run handbraked and make it watch the same directory:

   ```bash
   ./handbraked -p /path/to/preset.json -d /path/to/CONVERTED/
   ```

5. Wait for everything to finish.

## Help

```bash
$ handbraked -help
```

```bash
Usage:
	handbraked -d WATCH_DIR -p PRESET_PATH [OPTIONS]
Options:
  -b, --buffer-size int    Number of pending videos to always keep intact before starting to convert (default 0)
  -i, --interval int       Interval in seconds between checking for new videos (default 5)
  -m, --min-size int       Minimum converted file size in bytes, will otherwise error and terminate (default 1000000)
  -p, --preset string      Path to Handbrake preset used for conversion
  -s, --suffix string      Suffix to add to converted videos. Matching files will be excluded from conversion (default "-x265")
  -v, --verbose            Enable verbose logging
  -t, --wait-time int      Time in seconds to wait since a video was last modified before starting conversion (default 30)
  -d, --watch-dir string   Directory to watch and automatically convert new videos
```
