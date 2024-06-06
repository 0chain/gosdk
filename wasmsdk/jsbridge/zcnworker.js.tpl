importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.21.5/misc/wasm/wasm_exec.js');
importScripts('https://herumi.github.io/bls-wasm/browser/bls.js')

const go = new Go();
go.argv = {{.ArgsToJS}}
go.env = {{.EnvToJS}}
const bls = self.bls
bls.init(bls.BN254).then(()=>{})
WebAssembly.instantiateStreaming(fetch("http://localhost:3430/zcn.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
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