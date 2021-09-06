package sdk

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// buildFfmpegArgs build ffmpeg arguments for windows
// http://4youngpadawans.com/stream-camera-video-and-audio-with-ffmpeg/
// https://trac.ffmpeg.org/wiki/DirectShow

func buildFfmpegArgs(fileName string, delay int) []string {

	if strings.HasSuffix(fileName, ".m3u8") {
		return []string{
			"-f", "dshow",
			"-i", getInputDevices(),
			"-preset", "ultrafast",
			"-tune", "zerolatency",
			"-vcodec", "libx264",
			"-acodec", "aac",
			"-ac", "2",
			"-map", "0",
			"-hls_time", strconv.Itoa(delay),

			fileName,
		}
	}

	//-f dshow -i video=0:audio=0 -preset ultrafast -tune zerolatency -vcodec libx264 -acodec aac -ac 2 -map 0 -f segment -segment_time 5  f:\zbox\live%d.mp4
	return []string{
		"-f", "dshow",
		"-i", getInputDevices(),
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-vcodec", "libx264",
		"-acodec", "aac",
		"-ac", "2",
		"-map", "0",
		"-f", "segment",
		"-segment_time", strconv.Itoa(delay),

		fileName,
	}
}

var (
	inputDevices = ""
)

func getInputDevices() string {

	if inputDevices != "" {
		return inputDevices
	}

	cmd := exec.Command("ffmpeg", "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	defer func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()

		}
	}()
	// ffmpeg -list_devices true -f dshow -i dummy
	// ffmpeg version 4.4-essentials_build-www.gyan.dev Copyright (c) 2000-2021 the FFmpeg developers
	//   built with gcc 10.2.0 (Rev6, Built by MSYS2 project)
	//   configuration: --enable-gpl --enable-version3 --enable-static --disable-w32threads --disable-autodetect --enable-fontconfig --enable-iconv --enable-gnutls --enable-libxml2 --enable-gmp --enable-lzma --enable-zlib --enable-libsrt --enable-libssh --enable-libzmq --enable-avisynth --enable-sdl2 --enable-libwebp --enable-libx264 --enable-libx265 --enable-libxvid --enable-libaom --enable-libopenjpeg --enable-libvpx --enable-libass --enable-libfreetype --enable-libfribidi --enable-libvidstab --enable-libvmaf --enable-libzimg --enable-amf --enable-cuda-llvm --enable-cuvid --enable-ffnvcodec --enable-nvdec --enable-nvenc --enable-d3d11va --enable-dxva2 --enable-libmfx --enable-libgme --enable-libopenmpt --enable-libopencore-amrwb --enable-libmp3lame --enable-libtheora --enable-libvo-amrwbenc --enable-libgsm --enable-libopencore-amrnb --enable-libopus --enable-libspeex --enable-libvorbis --enable-librubberband
	//   libavutil      56. 70.100 / 56. 70.100
	//   libavcodec     58.134.100 / 58.134.100
	//   libavformat    58. 76.100 / 58. 76.100
	//   libavdevice    58. 13.100 / 58. 13.100
	//   libavfilter     7.110.100 /  7.110.100
	//   libswscale      5.  9.100 /  5.  9.100
	//   libswresample   3.  9.100 /  3.  9.100
	//   libpostproc    55.  9.100 / 55.  9.100
	// [dshow @ 000001c224f0dd40] DirectShow video devices (some may be both video and audio devices)
	// [dshow @ 000001c224f0dd40]  "Integrated Camera"
	// [dshow @ 000001c224f0dd40]     Alternative name "@device_pnp_\\?\usb#vid_04ca&pid_7058&mi_00#6&24bef87c&0&0000#{65e8773d-8f56-11d0-a3b9-00a0c9223196}\global"
	// [dshow @ 000001c224f0dd40] DirectShow audio devices
	// [dshow @ 000001c224f0dd40]  "麦克风阵列 (Realtek High Definition Audio)"
	// [dshow @ 000001c224f0dd40]     Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\wave_{2E8471BE-F843-412E-9E92-7C0E9E51E929}"
	// dummy: Immediate exit requested

	var reader io.Reader

	var bufOutput, bufErr bytes.Buffer
	cmd.Stderr = &bufErr
	cmd.Stdout = &bufOutput

	err := cmd.Run()
	if err != nil {
		reader = &bufErr
	} else {
		reader = &bufOutput
	}

	scanner := bufio.NewScanner(reader)

	readVedioDeviceName := false
	readAudioDeviceName := false

	videoDeviceName := ""
	audioDeviceName := ""

	for scanner.Scan() {

		if videoDeviceName != "" && audioDeviceName != "" {
			break
		}

		line := scanner.Text()

		if readVedioDeviceName && videoDeviceName == "" {
			items := strings.Split(line, "\"")
			if len(items) > 1 {
				videoDeviceName = strings.TrimSpace(items[1])
			}
		}

		if readAudioDeviceName && audioDeviceName == "" {
			items := strings.Split(line, "\"")
			if len(items) > 1 {
				audioDeviceName = strings.TrimSpace(items[1])
			}
		}

		if strings.Contains(line, "DirectShow video devices") {
			readVedioDeviceName = true
		}

		if strings.Contains(line, "DirectShow audio devices") {
			readAudioDeviceName = true
		}

	}

	inputDevices = "video=" + videoDeviceName + ":audio=" + audioDeviceName + ""

	return inputDevices
}
