importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.21.5/misc/wasm/wasm_exec.js','https://cdn.jsdelivr.net/gh/herumi/bls-wasm@v1.1.1/browser/bls.js');

const go = new Go();
go.argv = {{.ArgsToJS}}
go.env = {{.EnvToJS}}
const bls = self.bls
bls.init(bls.BN254).then(()=>{})

async function getWasmModule() {
  const cache = await caches.open('wasm-cache');
  let response = await cache.match("{{.Path}}");
  if(!response) {
    response = await cache.match("{{.FallbackPath}}")
    if (!response) {
    response = await fetch("{{.FallbackPath}}");
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
    }
}  