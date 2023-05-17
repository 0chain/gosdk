async function stopPlay({goWasm, videoElement}){

  if (!videoElement) {
    throw new Error('video element is required');
  }

  await goWasm.sdk.stop()

  videoElement.pause()
  URL.revokeObjectURL(videoElement.src);
}

async function startPlay({
  goWasm,
  allocationId,
  containerElement,
  videoElement,
  remotePath,
  authTicket = '',
  lookupHash = '',
  mimeType = '',
  isLive = false,
}) {
  if (!videoElement) {
    throw new Error('video element is required');
  }

  await goWasm.sdk.play(
      allocationId, remotePath, authTicket, lookupHash, isLive);

  if (isLive) {
    return playStream({goWasm, videoElement,allocationId,remotePath,authTicket, lookupHash});
  }

  // first segment
  const buf = await goWasm.sdk.getNextSegment();

  const {isFragmented, mimeCodecs} = getMimeCodecs({mimeType, buf});

  // const mimeCodecs = `${mimeType};
  // codecs="${muxjs.mp4.probe.tracks(buf).map(t => t.codec).join(",")}"`;
  console.log(
      `playlist: isFragmented:${isFragmented} mimeCodecs:${mimeCodecs}`);

  if (isFragmented && MediaSource.isTypeSupported(mimeCodecs)) {
    return playChunks({goWasm, videoElement, buf, mimeCodecs});
  }

  goWasm.sdk.stop();
  // allocationID, remotePath, authTicket, lookupHash string,
  // downloadThumbnailOnly, autoCommit bool
  const {url} = await goWasm.sdk.download(
      allocationId, remotePath, authTicket, lookupHash, false,10,"");
      videoElement.crossOrigin = 'anonymous';
      videoElement.src = url

  var promise = videoElement.play();
  if (promise !== undefined) {
      promise.catch(err => {
        // Auto-play was prevented
        while (videoElement.lastElementChild) {
          videoElement.removeChild(videoElement.lastElementChild);
        }
        videoElement.removeAttribute("src")

        const source = document.createElement("source")
        source.setAttribute("src", url)
        source.setAttribute("type", mimeType)
        videoElement.appendChild(source)

        videoElement.setAttribute("muted",true)
        videoElement.setAttribute("autoplay",true)
        videoElement.setAttribute("playsinline",true)
        videoElement.setAttribute("loop",true)
        setTimeout(function() {
              // weird fix for safari
              containerElement.innerHTML = videoElement.outerHTML;
        }, 100);
      }).then(() => {
          // Auto-play started
      });
  } 
}


async function playStream({
  goWasm,
  videoElement,
  allocationId,
  remotePath,
  authTicket,
  lookupHash
}) {
  await goWasm.sdk.play(
      allocationId, remotePath, authTicket, lookupHash, true);

  const mimeCodecs = 'video/mp4; codecs="mp4a.40.2,avc1.64001f"';

  if ('MediaSource' in window && MediaSource.isTypeSupported(mimeCodecs)) {
    let sourceBuffer;

    const transmuxer = new muxjs.mp4.Transmuxer();
    const mediaSource = new MediaSource();

    const getNextSegment = async () => {
      try {
        const buf = await goWasm.sdk.getNextSegment()

        if (buf?.length > 0) {
          transmuxer.push(new Uint8Array(buf))
          transmuxer.flush()
        }
      } catch (err) {
        getNextSegment()
      }
    };

    transmuxer.on('data', segment => {
      const data = new Uint8Array(
          segment.initSegment.byteLength + segment.data.byteLength);
      data.set(segment.initSegment, 0);
      data.set(segment.data, segment.initSegment.byteLength);
      // To inspect data use =>
      // console.log(muxjs.mp4.tools.inspect(data));
      sourceBuffer.appendBuffer(data);
    })

    mediaSource.addEventListener('sourceopen', async () => {
      sourceBuffer = mediaSource.addSourceBuffer(mimeCodecs);
      sourceBuffer.mode = 'sequence';
      sourceBuffer.addEventListener('updateend', getNextSegment);

      await getNextSegment();

      videoElement.play();

      URL.revokeObjectURL(videoElement.src);
    })

    videoElement.src = URL.createObjectURL(mediaSource);
    videoElement.crossOrigin = 'anonymous';
  } else {
    throw new Error('Unsupported MIME type or codec: ', mimeCodecs);
  }
}


async function playChunks({goWasm, videoElement, buf, mimeCodecs}) {
  let sourceBuffer;

  const mediaSource = new MediaSource();

  videoElement.src = URL.createObjectURL(mediaSource);
  videoElement.crossOrigin = 'anonymous';

  const getNextSegment = async () => {
    try {
      const buffer = await goWasm.sdk.getNextSegment()

      if (buffer?.length > 0) {
        sourceBuffer.appendBuffer(new Uint8Array(buffer))
      }
      else {
        if (!sourceBuffer.updating && mediaSource.readyState === 'open') {
          mediaSource.endOfStream()
        }
      }
    } catch (err) {
      getNextSegment()
    }
  };

  mediaSource.addEventListener('sourceopen', async () => {
    sourceBuffer = mediaSource.addSourceBuffer(mimeCodecs);
    sourceBuffer.mode = 'sequence'
    sourceBuffer.addEventListener('updateend', getNextSegment)
    sourceBuffer.appendBuffer(buf)
    videoElement.play()
  })
}


function detectMp4({mimeType, buf}) {
  const isFragmented =
      muxjs.mp4.probe.findBox(buf, ['moov', 'mvex']).length > 0 ? true : false;

  return {
    isFragmented,
        mimeCodecs: `${mimeType}; codecs="${
            muxjs.mp4.probe.tracks(buf).map(t => t.codec).join(',')}"`,
  }
}

function detectWebm({mimeType, buf}) {
  const decoder = new EBML.Decoder();
  const codecs =
      decoder.decode(buf)
          .filter(it => it.name == 'CodecID')
          .map(it => it.value.replace(/^(V_)|(A_)/, '').toLowerCase())
          .join(',')
  return {
    isFragmented: true, mimeCodecs: `${mimeType}; codecs="${codecs}"`
  }
}

const detectors = {
  'video/mp4': detectMp4,
  'audio/mp4': detectMp4,
  'video/webm': detectWebm,
  'audio/webm': detectWebm,
}

function getMimeCodecs({mimeType, buf}) {
  const detect = detectors[mimeType];
  if (detect) {
    return detect({mimeType, buf})
  }
  return mimeType
}