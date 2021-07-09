package sdk

import (
	"strconv"
)

//http://4youngpadawans.com/stream-camera-video-and-audio-with-ffmpeg/

// buildFfmpegArgs build ffmpeg arguments for windows
func buildFfmpegArgs(fileName string, delay int) []string {
	return []string{
		"-thread_queue_size", "50",
		"-f", "dshow",
		"-framerate", "30",
		"-i", "0:0",
		"-c:v", "libx264",
		"-crf", "18",
		"-preset", "ultrafast",
		"-r", "30",
		"-map", "0",
		"-f", "segment",
		"-segment_time", strconv.Itoa(delay),
		"-loglevel", "warning",
		fileName,
	}
}
