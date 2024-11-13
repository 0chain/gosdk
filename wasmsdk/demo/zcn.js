/*
 * This file is part of the 0chain @zerochain/0chain distribution
 * (https://github.com/0chain/client-sdk). Copyright (c) 2018 0chain LLC.
 *
 * 0chain @zerochain/0chain program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

'use strict'

const g =  window

function hexStringToByte(str) {
  if (!str) return new Uint8Array()

  const a = []
  for (let i = 0, len = str.length; i < len; i += 2) {
    a.push(parseInt(str.substr(i, 2), 16))
  }

  return new Uint8Array(a)
}

function blsSign(hash, secretKey) {
  const { jsProxy } = g.__zcn_wasm__

  if (!jsProxy || !secretKey) {
    const errMsg = 'err: bls.secretKey is not initialized'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  const bytes = hexStringToByte(hash)
  const sk = bls.deserializeHexStrToSecretKey(secretKey)
  const sig = sk.sign(bytes)

  if (!sig) {
    const errMsg = 'err: wasm blsSign function failed to sign transaction'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  return sig.serializeToHexStr()
}

async function createObjectURL(buf, mimeType) {
  var blob = new Blob([buf], { type: mimeType })
  return URL.createObjectURL(blob)
}


const readChunk = (offset, chunkSize, file) =>
  new Promise((res,rej) => {
    const fileReader = new FileReader()
    const blob = file.slice(offset, chunkSize+offset)
    fileReader.onload = e => {
      const t = e.target
      if (t.error == null) {
        res({
          size: t.result.byteLength,
          buffer: new Uint8Array(t.result)
        })
      }else{
        rej(t.error)
      }
    }

    fileReader.readAsArrayBuffer(blob)
  })


/**
 * Sleep is used when awaiting for Go Wasm to initialize.
 * It uses the lowest possible sane delay time (via requestAnimationFrame).
 * However, if the window is not focused, requestAnimationFrame never returns.
 * A timeout will ensure to be called after 50 ms, regardless of whether or not
 * the tab is in focus.
 *
 * @returns {Promise} an always-resolving promise when a tick has been
 *     completed.
 */
const sleep = (ms = 1000) =>
  new Promise(res => {
    requestAnimationFrame(res)
    setTimeout(res, ms)
  })



/**
 * The maximum amount of time that we would expect Wasm to take to initialize.
 * If it doesn't initialize after this time, we send a warning to console.
 * Most likely something has gone wrong if it takes more than 3 seconds to
 * initialize.
 */
const maxTime = 10 * 1000

// Initialize __zcn_wasm__
g.__zcn_wasm__ = g.__zcn_wasm_ || {
  glob:{
    index:0,
  },
  jsProxy: {
    secretKey: null,
    publicKey: null,
    sign: blsSign,
    verify: blsVerify,
    verifyWith: blsVerifyWith,
    createObjectURL,
    sleep,
  },
  sdk: {}, //proxy object for go to expose its methods
}

/**
 * bridge is an easier way to refer to the Go WASM object.
 */
const bridge = g.__zcn_wasm__

// bulk upload files with FileReader
// objects: the list of upload object
//  - allocationId: string
//  - remotePath: string
//  - file: File
//  - thumbnailBytes: []byte
//  - encrypt: bool
//  - isUpdate: bool
//  - isRepair: bool
//  - numBlocks: int
//  - callback: function(totalBytes,completedBytes,error)
async function bulkUpload(options) {
  const start = bridge.glob.index
  const opts = options.map(obj=>{
    const i = bridge.glob.index;
    bridge.glob.index++
    const readChunkFuncName = "__zcn_upload_reader_"+i.toString()
    const callbackFuncName = "__zcn_upload_callback_"+i.toString()
    var md5HashFuncName = ""
    g[readChunkFuncName] =  async (offset,chunkSize) => {
      const chunk = await readChunk(offset,chunkSize,obj.file)
      return chunk.buffer
    }
    if (obj.file.size > 25*1024*1024) {
      md5HashFuncName = "__zcn_md5_hash_"+i.toString()
      const md5Res = md5Hash(obj.file)
      g[md5HashFuncName] = async () => {
      const hash = await md5Res
      return hash
      }
  }

    if(obj.callback) {
      g[callbackFuncName] =  async (totalBytes,completedBytes,error)=> obj.callback(totalBytes,completedBytes,error)
    }

    return {
      allocationId:obj.allocationId,
      remotePath:obj.remotePath,
      readChunkFuncName:readChunkFuncName,
      fileSize: obj.file.size,
      thumbnailBytes:obj.thumbnailBytes?obj.thumbnailBytes.toString():"",
      encrypt:obj.encrypt,
      webstreaming:obj.webstreaming,
      isUpdate:obj.isUpdate,
      isRepair:obj.isRepair,
      numBlocks:obj.numBlocks,
      callbackFuncName:callbackFuncName,
      md5HashFuncName:md5HashFuncName,
    }
  })

  // md5Hash(options[0].file).then(hash=>{
  //   console.log("md5 hash: ",hash)
  // }).catch(err=>{
  //   console.log("md5 hash error: ",err)
  // })

  const end =  bridge.glob.index
  const result = await bridge.__proxy__.sdk.multiUpload(JSON.stringify(opts))
  for (let i=start; i<end;i++){
    g["__zcn_upload_reader_"+i.toString()] = null;
    g["__zcn_upload_callback_"+i.toString()] =null;
    g["__zcn_md5_hash_"+i.toString()] = null;
  }
  return result
}


async function md5Hash(file) {
  const result = new Promise((resolve, reject) => {
    const worker = new Worker('md5worker.js')
    worker.postMessage(file)
    worker.onmessage = e => {
      resolve(e.data)
      worker.terminate()
    }
    worker.onerror = reject
  })
  return result
}


async function blsSign(hash, secretKey) {
  if (!bridge.jsProxy && !secretKey) {
    const errMsg = 'err: bls.secretKey is not initialized'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  const bytes = hexStringToByte(hash)
  const sk = bls.deserializeHexStrToSecretKey(secretKey)
  const sig = sk.sign(bytes)

  if (!sig) {
    const errMsg = 'err: wasm blsSign function failed to sign transaction'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  return sig.serializeToHexStr()
}

async function blsVerifyWith(pk, signature, hash) {
  const publicKey = bls.deserializeHexStrToPublicKey(pk);
  const bytes = hexStringToByte(hash)
  const sig = bls.deserializeHexStrToSignature(signature)
  return publicKey.verify(sig, bytes)
}

async function blsVerify(signature, hash) {
  if (!bridge.jsProxy && !bridge.jsProxy.publicKey) {
    const errMsg = 'err: bls.publicKey is not initialized'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  const bytes = hexStringToByte(hash)
  const sig = bridge.jsProxy.bls.deserializeHexStrToSignature(signature)
  return bridge.jsProxy.publicKey.verify(sig, bytes)
}

async function setWallet(bls,
  clientID,
  clientKey,
  peerPublicKey,
  sk,
  pk,
  mnemonic,
  isSplit) {
  if (!bls) throw new Error('bls is undefined, on wasm setWallet fn')
  if (!sk) throw new Error('secret key is undefined, on wasm setWallet fn')
  if (!pk) throw new Error('public key is undefined, on wasm setWallet fn')

  console.log('setWallet: ', clientID, sk, pk)
  bridge.jsProxy.bls = bls
  bridge.jsProxy.secretKey = bls.deserializeHexStrToSecretKey(sk)
  bridge.jsProxy.publicKey = bls.deserializeHexStrToPublicKey(pk)

  // use proxy.sdk to detect if sdk is ready
  await bridge.__proxy__.sdk.setWallet(clientID, clientKey, peerPublicKey, pk, sk, mnemonic, isSplit)
  bridge.walletId = clientID
}

async function loadWasm(go) {
  // If instantiateStreaming doesn't exists, polyfill/create it on top of instantiate
  if (!WebAssembly?.instantiateStreaming) {
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
      const source = await (await resp).arrayBuffer()
      return await WebAssembly.instantiate(source, importObject)
    }
  }

  const result = await WebAssembly.instantiateStreaming(
    await fetch('test/zcn.wasm'),
    go.importObject
  )

  setTimeout(() => {
    if (g.__zcn_wasm__?.__wasm_initialized__ !== true) {
      console.warn(
        'wasm window.__zcn_wasm__ (zcn.__wasm_initialized__) still not true after max time'
      )
    }
  }, maxTime)

  go.run(result.instance)
}

async function createWasm() {
  if (bridge.__proxy__) {
    return bridge.__proxy__
  }

  const go = new g.Go()

  loadWasm(go)

  const sdkGet =
    (_, key) =>
    (...args) =>
      // eslint-disable-next-line
      new Promise(async (resolve, reject) => {
        if (!go || go.exited) {
          return reject(new Error('The Go instance is not active.'))
        }

        while (bridge.__wasm_initialized__ !== true) {
          await sleep(1000)
        }

        if (typeof bridge.sdk[key] !== 'function') {
          resolve(bridge.sdk[key])

          if (args.length !== 0) {
            reject(
              new Error(
                'Retrieved value from WASM returned function type, however called with arguments.'
              )
            )
          }
          return
        }

        try {
          let resp = bridge.sdk[key].apply(undefined, args)

          // support wasm.BindAsyncFunc
          if (resp && typeof resp.then === 'function') {
            resp = await Promise.race([resp])
          }

          if (resp && resp.error) {
            reject(resp.error)
          } else {
            resolve(resp)
          }
        } catch (e) {
          reject(e)
        }
      })

  const sdkProxy = new Proxy(
    {

    },
    {
      get: sdkGet,
    }
  )

  const jsProxy = new Proxy(
    {},
    {
      get: (_, key) => bridge.jsProxy[key],
      set: (_, key, value) => {
        bridge.jsProxy[key] = value
      },
    }
  )

  const proxy = {
    bulkUpload: bulkUpload,
    setWallet: setWallet,
    sdk: sdkProxy, //expose sdk methods for js
    jsProxy, //expose js methods for go
  }

  bridge.__proxy__ = proxy

  return proxy
}