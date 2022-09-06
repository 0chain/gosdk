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

function blsSign(hash) {
  console.log('blsSign: ', hash)

  const { jsProxy } = g.__zcn_wasm__

  if (!jsProxy || !jsProxy.secretKey) {
    const errMsg = 'err: bls.secretKey is not initialized'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  const bytes = hexStringToByte(hash)

  const sig = jsProxy.secretKey.sign(bytes)

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
  jsProxy: {
    secretKey: null,
    publicKey: null,
    sign: blsSign,
    verify: blsVerify,
    createObjectURL,
    sleep,
  },
  sdk: {}, //proxy object for go to expose its methods
}

/**
 * bridge is an easier way to refer to the Go WASM object.
 */
const bridge = g.__zcn_wasm__

async function blsSign(hash) {
  if (!bridge.jsProxy && !bridge.jsProxy.secretKey) {
    const errMsg = 'err: bls.secretKey is not initialized'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  const bytes = hexStringToByte(hash)

  const sig = bridge.jsProxy.secretKey.sign(bytes)

  if (!sig) {
    const errMsg = 'err: wasm blsSign function failed to sign transaction'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  return sig.serializeToHexStr()
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

async function setWallet(bls, clientID, sk, pk) {
  if (!bls) throw new Error('bls is undefined, on wasm setWallet fn')
  if (!sk) throw new Error('secret key is undefined, on wasm setWallet fn')
  if (!pk) throw new Error('public key is undefined, on wasm setWallet fn')

  if (bridge.walletId != clientID) {
    console.log('setWallet: ', clientID, sk, pk)
    bridge.jsProxy.bls = bls
    bridge.jsProxy.secretKey = bls.deserializeHexStrToSecretKey(sk)
    bridge.jsProxy.publicKey = bls.deserializeHexStrToPublicKey(pk)

    // use proxy.sdk to detect if sdk is ready
    await bridge.__proxy__.sdk.setWallet(clientID, pk)
    bridge.walletId = clientID
  }
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
    await fetch('zcn.wasm'),
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
    {},
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
    setWallet: setWallet,
    sdk: sdkProxy, //expose sdk methods for js
    jsProxy, //expose js methods for go
  }

  bridge.__proxy__ = proxy

  return proxy
}



