package sdk

// buildFfmpegArgs build ffmpeg arguments for freebsd
func buildFfmpegArgs(fileName string, delay int) []string {
	return []string{
		"-thread_queue_size", "50",
		"-f", "bktr",
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
