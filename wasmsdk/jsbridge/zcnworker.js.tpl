importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.21.5/misc/wasm/wasm_exec.js','https://cdn.jsdelivr.net/gh/herumi/bls-wasm@v1.1.1/browser/bls.js');

const go = new Go();
go.argv = {{.ArgsToJS}}
go.env = {{.EnvToJS}}
const bls = self.bls
bls.init(bls.BN254).then(()=>{})

 (async () => {
    let source = await fetch("{{.Path}}")
      .then(res => res)
      .catch(err => err)
    // fallback to our server where the app would be hosted
    if (!source?.ok) {
      source = await fetch("{{.FallbackPath}}")
    }

   WebAssembly.instantiate(source, go.importObject).then((result) => {
    go.run(result.instance);
    }).catch((err) => {
    console.error(err);
    });

  })();

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