package sdk

import (
	"strconv"
	"strings"
)

// buildFfmpegArgs build ffmpeg arguments for darwin: https://ffmpeg.org/ffmpeg-devices.html#avfoundation
func buildFfmpegArgs(fileName string, delay int) []string {

	if strings.HasSuffix(fileName, ".m3u8") {
		return []string{
			//	"-thread_queue_size", "50",
			"-f", "avfoundation",
			"-framerate", "30",
			"-i", "default:default",
			"-r", "30",
			"-flags", "+cgop",
			"-g", "30",
			"-hls_time", strconv.Itoa(delay),

			fileName, //*.m3u8
		}
	}

	//mp4, avi...etc
	return []string{
		//"-thread_queue_size", "50",
		"-f", "avfoundation",
		"-framerate", "30",
		"-i", "default:default",
		"-c:v", "libx264",
		"-crf", "18",
		"-preset", "ultrafast",
		"-r", "30",
		"-map", "0",
		"-f", "segment",
		"-segment_time", strconv.Itoa(delay),

		fileName,
	}

}
