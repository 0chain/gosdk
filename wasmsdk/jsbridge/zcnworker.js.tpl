importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.21.5/misc/wasm/wasm_exec.js','https://cdn.jsdelivr.net/gh/herumi/bls-wasm@v1.1.1/browser/bls.js');

const go = new Go();
go.argv = {{.ArgsToJS}}
go.env = {{.EnvToJS}}
const bls = self.bls
bls.init(bls.BN254).then(()=>{})

async function getWasmModule() {
  const cache = await caches.open('wasm-cache');
  let response = await cache.match("{{.CachePath}}");
  if(!response?.ok) {
    response = await fetch("{{.Path}}").then(res => res).catch(err => err);
    if (!response?.ok) {
    response = await fetch("{{.FallbackPath}}").then(res => res).catch(err => err);
    }
    if (!response.ok) {
      throw new Error(`Failed to fetch WASM: ${response.statusText}`);
    }
  }
  const bytes = await response.arrayBuffer();
  return WebAssembly.instantiate(bytes, go.importObject);
}

getWasmModule().then(result => {
  go.run(result.instance);
}).catch(error => {
  console.error("Failed to load WASM:", error);
});

function hexStringToByte(str) {
    if (!str) return new Uint8Array()
  
    const a = []
    for (let i = 0, len = str.length; i < len; i += 2) {
      a.push(parseInt(str.substr(i, 2), 16))
    }
  
    return new Uint8Array(a)
  }

self.__zcn_worker_wasm__ = {
    sign: async (hash, secretKey) => {
        if (!secretKey){
            const errMsg = 'err: wasm blsSign function requires a secret key'
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
    },
    initProxyKeys: initProxyKeys,
    verify: blsVerify,
    verifyWith: blsVerifyWith,
    addSignature: blsAddSignature
}  

async function initProxyKeys(publicKey, privateKey) {
  const pubKey = bls.deserializeHexStrToPublicKey(publicKey)
  const privKey = bls.deserializeHexStrToSecretKey(privateKey)
  bls.publicKey = pubKey
  bls.secretKey = privKey
}

async function blsVerify(signature, hash) {
  const bytes = hexStringToByte(hash)
  const sig = bls.deserializeHexStrToSignature(signature)
  return jsProxy.publicKey.verify(sig, bytes)
}

async function blsVerifyWith(pk, signature, hash) {
  const publicKey = bls.deserializeHexStrToPublicKey(pk)
  const bytes = hexStringToByte(hash)
  const sig = bls.deserializeHexStrToSignature(signature)
  return publicKey.verify(sig, bytes)
}

async function blsAddSignature(secretKey, signature, hash) {
  const privateKey = bls.deserializeHexStrToSecretKey(secretKey)
  const sig = bls.deserializeHexStrToSignature(signature)
  var sig2 = privateKey.sign(hexStringToByte(hash))
  if (!sig2) {
    const errMsg =
      'err: wasm blsAddSignature function failed to sign transaction'
    console.warn(errMsg)
    throw new Error(errMsg)
  }

  sig.add(sig2)

  return sig.serializeToHexStr()
}
