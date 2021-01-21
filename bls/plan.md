# All the functions we need to port #

// From herumi's email
sig.Verify ; pairing, ellptic operations, hash to curve
sign.Add ; elliptic operation
GetMasterSecretKey ; field operation
GetPublicKey ; scalar multiplication

// Others we should plan for. Curated by me.
SetByCSPRNG
SetDecString
SerializeToHexStr
SetLittleEndian
GetLittleEndian
FrSub
blsInit / bls.Init

var sk bls.SecretKey
sk.Set(polynomial, &id)

## bls.Sign Serialize ##

Seems like this is a good old serialize to byte array.

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L454>
  // Serialize --
  func (sig Sign) Serialize() []byte {
    buf := make([]byte, 2048)
    // #nosec
    n := C.blsSignatureSerialize(unsafe.Pointer(&buf[0]), C.mclSize(len(buf)), &sig.v)
    if n == 0 {
      panic("err blsSignatureSerialize")
    }
    return buf[:n]
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L478>
  mclSize blsSignatureSerialize(void buf, mclSize maxBufSize, const blsSignature sig)
  {
    return cast(&sig->v)->serialize(buf, maxBufSize);
  }

## bls.PublicKey.DeserializeHexStr ##

This is the maze of deserialized. Lets start from the beginning...

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L382>
  // DeserializeHexStr --
  func (pub PublicKey) DeserializeHexStr(s string) error {
    a, err := hex2byte(s)
    if err != nil {
      return err
    }
    return pub.Deserialize(a)
  }

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L367>
  // Deserialize --
  func (pub PublicKey) Deserialize(buf []byte) error {
    // #nosec
    err := C.blsPublicKeyDeserialize(&pub.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf)))
    if err == 0 {
      return fmt.Errorf("err blsPublicKeyDeserialize %x", buf)
    }
    return nil
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L493>
  mclSize blsPublicKeyDeserialize(blsPublicKey *pub, const void *buf, mclSize bufSize)
  {
    return cast(&pub->v)->deserialize(buf, bufSize);
  }

## bls.SecretKey ##

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L154>
  // SecretKey --
  type SecretKey struct {
    v C.blsSecretKey
  }

  <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L56>
  typedef struct {
    mclBnFr v;
  } blsSecretKey;

And this is Fr, which is just FP with a modulo.

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L89>
  /*
    G1 and G2 are isomorphism to Fr
  */
  typedef struct {
    uint64_t d[MCLBN_FR_UNIT_SIZE];
  } mclBnFr;

## bls.PublicKey ##

Lets start from the beginning, what is the PublicKey struct? It's basically
just an alias for G1/G2 again...

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L334>
  // PublicKey --
  type PublicKey struct {
    v C.blsPublicKey
  }

  // PublicKeys ..
  type PublicKeys []PublicKey

  <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L60>
  typedef struct {
  #ifdef BLS_ETH
    mclBnG1 v;
  #else
    mclBnG2 v;
  #endif
  } blsPublicKey;

## bls.Sign.DeserializeHexStr ##

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L480>
  // DeserializeHexStr --
  func (sig Sign) DeserializeHexStr(s string) error {
    a, err := hex2byte(s)
    if err != nil {
      return err
    }
    return sig.Deserialize(a)
  }

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L465>
  // Deserialize --
  func (sig Sign) Deserialize(buf []byte) error {
    // #nosec
    err := C.blsSignatureDeserialize(&sig.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf)))
    if err == 0 {
      return fmt.Errorf("err blsSignatureDeserialize %x", buf)
    }
    return nil
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L498>
  mclSize blsSignatureDeserialize(blsSignature *sig, const void *buf, mclSize bufSize)
  {
    return cast(&sig->v)->deserialize(buf, bufSize);
  }

And according to SHIGEO, the serialize of Secret Key is:

  <https://github.com/herumi/mcl/blob/master/include/mcl/fp.hpp#L308-L322>
  ...?

And this part makes no sense to me.

## bls.Sign ##

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L449>
  // Sign  --
  type Sign struct {
    v C.blsSignature
  }

  <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L68>
  typedef struct {
  #ifdef BLS_ETH
    mclBnG2 v;
  #else
    mclBnG1 v;
  #endif
  } blsSignature;

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L96>
  /* G1 is defined over Fp */
  typedef struct {
    mclBnFp x, y, z;
  } mclBnG1;

So 'mclBnG1' just looks a lot like:

  <https://github.com/miracl/core/blob/master/go/ECP.go#L24>
  type ECP struct {
    x *FP
    y *FP
    z *FP
  }

Interestingly, SHIGEO also has G1 defined in bls library, while mclBnG1 is
defined in mcl library. they are really the same struct.

Here are further recursing down the code:

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L78>
  typedef struct {
    uint64_t d[MCLBN_FP_UNIT_SIZE];
  } mclBnFp;

  // MCLBN_FP_UNIT_SIZE is
  #error "define MCLBN_FP_UNIT_SIZE 4(, 6 or 8)"

## blsInit / bls.Init ##
GoSDK uses it as:

  core/zcncrypto/bls0chain.go:16: err := bls.Init(bls.CurveFp254BNb)

Entry point in C code is: <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L78>

The corresponding constant for 'CurveFp254BNb' is probably BN254.BN or BN254.CURVE_Order.

  <https://github.com/miracl/core/blob/master/go/BenchtestALL.go#L243>
  // And then search for -i "curve" in that function.

blsInit is important to get correct as it initializes getBasePoint(), which will be used for getPublicKey.
This is easy to miss: look for g_Q or g_P, as these are corresponding to getBasePoint.
--Scratch that, it just precomputes using getBasePoint().--

blsInit precomputes with precomputeG2 function. You can find it defined here:

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.hpp#L1689>
  ;

## sig.Verify
https://github.com/miracl/core/blob/fda3416694d153f900b617d7bc42038df34a2da6/go/BLS.go#L123
`func Core_Verify(SIG []byte, M []byte, W []byte) int { ... }`

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L553>
  // Verify --
  func (sig Sign) Verify(pub PublicKey, m string) bool {
    if sig == nil || pub == nil {
      return false
    }
    buf := []byte(m)
    // #nosec
    return C.blsVerify(&sig.v, &pub.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf))) == 1
  }

## sign.Add
Convert string to ECP, then perform Add.

### Step 1.
// Example: https://github.com/miracl/core/blob/fda3416694d153f900b617d7bc42038df34a2da6/go/BLS.go#L126
`D := ECP_fromBytes(SIG) `

// D has type 'ECP'.
// ECP_fromBytes declared here:
// https://github.com/miracl/core/blob/f9de005e0168f59a56afe177498b19f4d43f054f/go/ECP.go#L392
`func ECP_fromBytes(b []byte) *ECP { ... } `

### Step 2.
Perform add on 'sign'.

https://github.com/miracl/core/blob/f9de005e0168f59a56afe177498b19f4d43f054f/go/ECP.go#L678
`func (E *ECP) Add(Q *ECP) { ... }`

## SetByCSPRNG ##

Well the following is how it's done in the javascript library.
It's just setting every byte to a random value.

  exports.SecretKey = class extends Common {
    setByCSPRNG () {
      const a = new Uint8Array(BLS_SECRETKEY_SIZE)
      exports.getRandomValues(a)
      this.setLittleEndian(a)
    }
  }

  setLittleEndian (s) {
    this._setter(exports.mclBnFr_setLittleEndian, s)
  }

And this is the default RNG function.

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/randgen.hpp#L141>
	static RandGen& getDefaultRandGen()
	{
    #ifdef MCL_DONT_USE_CSPRNG
		static RandGen wrg;
    #elif defined(MCL_USE_WEB_CRYPTO_API)
		static mcl::RandomGeneratorJS rg;
		static RandGen wrg(rg);
    #else
		static cybozu::RandomGenerator rg;
		static RandGen wrg(rg);
    #endif
		return wrg;
	}

## GetMasterSecretKey

For loop around `SetByCSPRNG`. It's just this function we're trying to duplicate.
```
func (sec *SecretKey) GetMasterSecretKey(k int) (msk []SecretKey) {
	msk = make([]SecretKey, k)
	msk[0] = *sec
	for i := 1; i < k; i++ {
		msk[i].SetByCSPRNG()
	}
	return msk
}
```

## GetPublicKey

Function defined here:

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L519-L524>
  func (sec SecretKey) GetPublicKey() (pub PublicKey)

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L175>
  // Remember, this returns public key (0th arg as pub).
  ```
  void blsGetPublicKey(blsPublicKey *pub, const blsSecretKey *sec)
  {
    Gmul(*cast(&pub->v), getBasePoint(), *cast(&sec->v));
  }
  ```

Gmul is just a multimethod. There are 2 Gmul defined. C++ chooses the right function
based on the types passed to the function. See them defined here:
  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L16>

  inline void Gmul(G1& z, const G1& x, const Fr& y) { G1::mul(z, x, y); }
  inline void Gmul(G2& z, const G2& x, const Fr& y) { G2::mul(z, x, y); }

Gmul is either G1mul or G2mul. The diff between G1 and G2 is a #ifdef on BLS_ETH <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L50>

getBasePoint returns a static G1/G2. The G1 structure is:

  <https://github.com/herumi/bls/blob/master/ffi/go/bls/mcl.go#L569>
  // G2 --
  type G2 struct {
    X Fp2
    Y Fp2
    Z Fp2
  }
  <https://github.com/herumi/bls/blob/master/ffi/go/bls/mcl.go#L476>
  // Fp2 -- x = D[0] + D[1] i where i^2 = -1
  type Fp2 struct {
    D [2]Fp
  }
  // See FP2.go, which has two FPs defined.

  <https://github.com/herumi/bls/blob/master/ffi/go/bls/mcl.go#L569>
  type G1 struct {
    X Fp
    Y Fp
    Z Fp
  }
  <https://github.com/herumi/bls/blob/master/ffi/go/bls/mcl.go#L306>
  type Fp struct {
    v C.mclBnFp
  }
  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L78>
  typedef struct {
    uint64_t d[MCLBN_FP_UNIT_SIZE];
  } mclBnFp;

The real interesting thing is that this G1 struct corresponds perfectly to
ECP struct.

  <https://github.com/miracl/core/blob/master/go/ECP.go#L24>
  type ECP struct {
    x *FP
    y *FP
    z *FP
  }

Moving on to Fr, that's really interesting, I think I know the exact corollary
in PAIR.go

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L89>
  /*
    G1 and G2 are isomorphism to Fr
  */
  typedef struct {
    uint64_t d[MCLBN_FR_UNIT_SIZE];
  } mclBnFr;

I believe the exactly corollary is in BIG64.go, see this:

  <https://github.com/miracl/core/blob/master/go/BIG64.go#L30>
  type BIG struct {
    w [NLEN]Chunk
  }

  <https://github.com/miracl/core/blob/master/go/BIG64.go#L34>
  type DBIG struct {
    w [2 * NLEN]Chunk
  }

  // Chunk is int64
  <https://github.com/miracl/core/blob/master/go/ARCH64.go#L24>

Now that Fr corresponds exactly to what Shigeo said about Fr being just Fp
with a modulo, see the definition here:

  <https://github.com/miracl/core/blob/master/go/FP.go#L26>
  type FP struct {
    x   *BIG
    XES int32
  }

## FrSub

From Shigeo:

  The difference between the Fr and Fp class of mcl is only the modulo parameter.
  If MIRACL has the Fp class, then you may be able to use it as the Fr by setting the parameter.

Luckily, MIRACL *does* have FP class (FP.go), so FrSub is just:

  https://github.com/miracl/core/blob/master/go/FP.go#L381-L385

The original FrSub takes 3 args. The 0th arg is the return variable. The 1th
and 2nd args are where the subtraction should take place.

## var sk bls.SecretKey; sk.Set(polynomial, &id)

sk.Set is defined here:

  https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L305-L312
  Calls into C.blsSecretKeyShare

  https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L543
  Calls into mcl::evaluatePolynomial

  evaluatePolynomial uses mul and add, see Herumi's email.

Herumi's email:

  Secret Key "mul" and "add" is defined...

  https://github.com/herumi/mcl/blob/master/include/mcl/fp.hpp#L499
  https://github.com/herumi/mcl/blob/master/include/mcl/fp.hpp#L502
  The naive algorithm is (x + y) % r and (x * y) % r.

## GetLittleEndian, Serialize.

Defined here:

  // https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L180
  // GetLittleEndian -- alias of Serialize
  func (sec SecretKey) GetLittleEndian() []byte { // k
    return sec.Serialize()
  }

  // https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L186
  // SetLittleEndian --
  func (sec SecretKey) SetLittleEndian(buf []byte) error {
    // #nosec
    err := C.blsSecretKeySetLittleEndian(&sec.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf)))
    if err != 0 {
      return fmt.Errorf("err blsSecretKeySetLittleEndian %x", err)
    }
    return nil
  }

Herumi:
The class of Secret Key is Fr, then https://github.com/herumi/mcl/blob/master/include/mcl/fp.hpp#L308-L322
