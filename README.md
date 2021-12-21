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

>  :warning: Due to the way Topaz Video Enhance AI works, it is hard to tell whether a video file is completely finished processing. To remidy this, a buffer of (by default) 2 videos is always maintained before re-encoding the next video. This works great, but when all of your videos are done, the last 2 will be left untouched. To fix this, re-run handbraked, but this time, set the buffer size to 0 with the argument:  `-b 0`.

