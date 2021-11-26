package bls

// function _emscripten_memcpy_big(dest, src, num) {
// 	HEAPU8.copyWithin(dest, src, src + num);
//    }

//    function _emscripten_resize_heap(requestedSize) {
// 	var oldSize = HEAPU8.length;
// 	requestedSize = requestedSize >>> 0;
// 	return false;
//    }

// var buffer, HEAP8, HEAPU8, HEAP16, HEAPU16, HEAP32, HEAPU32, HEAPF32, HEAPF64;

// function updateGlobalBufferAndViews(buf) {
//  buffer = buf;
//  Module["HEAP8"] = HEAP8 = new Int8Array(buf);
//  Module["HEAP16"] = HEAP16 = new Int16Array(buf);
//  Module["HEAP32"] = HEAP32 = new Int32Array(buf);
//  Module["HEAPU8"] = HEAPU8 = new Uint8Array(buf);
//  Module["HEAPU16"] = HEAPU16 = new Uint16Array(buf);
//  Module["HEAPU32"] = HEAPU32 = new Uint32Array(buf);
//  Module["HEAPF32"] = HEAPF32 = new Float32Array(buf);
//  Module["HEAPF64"] = HEAPF64 = new Float64Array(buf);
// }

// var asmLibraryArg = {
// 	"a": _emscripten_memcpy_big,
// 	"b": _emscripten_resize_heap
//    };

// function createWasm() {
// 	var info = {
// 	 "a": asmLibraryArg
// 	};
// 	function receiveInstance(instance, module) {
// 	 var exports = instance.exports;
// 	 Module["asm"] = exports;
// 	 wasmMemory = Module["asm"]["c"];
// 	 updateGlobalBufferAndViews(wasmMemory.buffer);
// 	 wasmTable = Module["asm"]["Td"];
// 	 addOnInit(Module["asm"]["d"]);
// 	 removeRunDependency("wasm-instantiate");
// 	}
// 	addRunDependency("wasm-instantiate");
// 	function receiveInstantiationResult(result) {
// 	 receiveInstance(result["instance"]);
// 	}
// 	function instantiateArrayBuffer(receiver) {
// 	 return getBinaryPromise().then(function(binary) {
// 	  var result = WebAssembly.instantiate(binary, info);
// 	  return result;
// 	 }).then(receiver, function(reason) {
// 	  err("failed to asynchronously prepare wasm: " + reason);
// 	  abort(reason);
// 	 });
// 	}
