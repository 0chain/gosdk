package sdk

import (
	"strconv"
)

// buildFfmpegArgs build ffmpeg arguments for freebsd
func buildFfmpegArgs(fileName string, delay int) []string {
	if strings.HasSuffix(fileName, ".m3u8") {
		return []string{
			//	"-thread_queue_size", "50",
			"-f", "v4l2",
			"-i", "/dev/video0",
			"-f", "alsa",
			"-i", "hw:0",
			"-preset", "ultrafast",
			"-tune", "zerolatency",
			"-vcodec", "libx264",
			"-r", "30",
			"-b:v", "512k",
			"-acodec", "aac",
			"-strict", "-2",
			"-ac", "2",
			"-hls_time", strconv.Itoa(delay),

			fileName, //*.m3u8
		}
	}

	return []string{
		//"-thread_queue_size", "10",
		"-f", "v4l2",
		"-i", "/dev/video0",
		"-f", "alsa",
		"-i", "hw:0",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-vcodec", "libx264",
		"-r", "30",
		"-b:v", "512k",
		"-acodec", "aac",
		"-strict", "-2",
		"-ac", "2",
		"-ab", "32k",
		"-ar", "44100",
		"-map", "0",
		"-map", "1",
		"-f", "segment",
		"-segment_time", strconv.Itoa(delay),

		fileName,
	}
}
