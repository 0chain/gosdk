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

// New ones from 0chain repo.
bls.GetMasterPublicKey
`(*bls.Sign) Recover`

## GT

So GT is just FP12.

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L107>
  ```
  typedef struct {
    mclBnFp d[12];
  } mclBnGT;
  ```

## bls.PublicKey -> Equals

This is how Equals is implemented for projected and affine.

  <include/mcl/ec.hpp>

  ```
	bool operator==(const EcT& rhs) const
	{
		switch (mode_) {
		case ec::Jacobi:
			return ec::isEqualJacobi(*this, rhs);
		case ec::Proj:
			return ec::isEqualProj(*this, rhs);
		case ec::Affine:
		default:
			return x == rhs.x && y == rhs.y && z == rhs.z;
		}
	}
  ```

  ```
  // (x/z, y/z)
  template<class E>
  bool isEqualProj(const E& P1, const E& P2)
  {
    typedef typename E::Fp F;
    bool zero1 = P1.isZero();
    bool zero2 = P2.isZero();
    if (zero1) {
      return zero2;
    }
    if (zero2) return false;
    F t1, t2;
    F::mul(t1, P1.x, P2.z);
    F::mul(t2, P2.x, P1.z);
    if (t1 != t2) return false;
    F::mul(t1, P1.y, P2.z);
    F::mul(t2, P2.y, P1.z);
    return t1 == t2;
  }
  ```

## bls.PublicKey -> Set

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L457>
  ```
  // Set --
  func (pub *PublicKey) Set(mpk []PublicKey, id *ID) error {
    // #nosec
    ret := C.blsPublicKeyShare(&pub.v, &mpk[0].v, (C.mclSize)(len(mpk)), &id.v)
    if ret != 0 {
      return fmt.Errorf("err blsPublicKeyShare")
    }
    return nil
  }
  ```

  <https://github.com/herumi/bls/blob/4ae022a6bb71dc518d81f22141d71d2a1f767ab3/src/bls_c_impl.hpp#L567>
  ```
  int blsPublicKeyShare(blsPublicKey *pub, const blsPublicKey *mpk, mclSize k, const blsId *id)
  {
    bool b;
    mcl::evaluatePolynomial(&b, *cast(&pub->v), cast(&mpk->v), k, *cast(&id->v));
    return b ? 0 : -1;
  }
  ```

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/lagrange.hpp#L64>
  ```
  /*
    out = f(x) = c[0] + c[1] * x + c[2] * x^2 + ... + c[cSize - 1] * x^(cSize - 1)
    @retval 0 if succeed else -1 (if cSize == 0)
  */
  template<class G, class T>
  void evaluatePolynomial(bool *pb, G& out, const G *c, size_t cSize, const T& x)
  {
    if (cSize == 0) {
      *pb = false;
      return;
    }
    if (cSize == 1) {
      out = c[0];
      *pb = true;
      return;
    }
    G y = c[cSize - 1];
    for (int i = (int)cSize - 2; i >= 0; i--) {
      G::mul(y, y, x);
      G::add(y, y, c[i]);
    }
    out = y;
    *pb = true;
  }
  ```

By luck, found the ECP * FP mul code in herumi:

  <https://github.com/herumi/mcl/blob/7bfe60c5373767d674b9aefea5fbf737897397eb/include/mcl/ec.hpp#L933>
  ```
  template<class tag, size_t maxBitSize, template<class _tag, size_t _maxBitSize>class FpT>
  static inline void mul(EcT& z, const EcT& x, const FpT<tag, maxBitSize>& y)
  {
    fp::Block b;
    y.getBlock(b);
    mulArray(z, x, b.p, b.n, false);
  }
  ```

  <https://github.com/herumi/mcl/blob/7bfe60c5373767d674b9aefea5fbf737897397eb/include/mcl/fp.hpp#L379>
  ```
  void getBlock(fp::Block& b) const
  {
    b.n = op_.N;
    if (isMont()) {
      op_.fromMont(b.v_, v_);
      b.p = &b.v_[0];
    } else {
      b.p = &v_[0];
    }
  }
  ```

  <https://github.com/herumi/mcl/blob/7bfe60c5373767d674b9aefea5fbf737897397eb/include/mcl/ec.hpp#L1320>
  ```
	static inline void mulArray(EcT& z, const EcT& x, const fp::Unit *y, size_t yn, bool isNegative, bool constTime = false, bool useGLV = true)
	{
		if (!constTime) {
			if (yn == 0) {
				z.clear();
				return;
			}
			yn = fp::getNonZeroArraySize(y, yn);
			if (yn <= 1 && mulSmallInt(z, x, *y, isNegative)) return;
		}
		if (useGLV && mulArrayGLV && (yn * sizeof(fp::Unit) > 8)) {
			mulArrayGLV(z, x, y, yn, isNegative, constTime);
			return;
		}
		mulArrayBase(z, x, y, yn, isNegative, constTime);
	}
  ```

  <https://github.com/herumi/mcl/blob/7bfe60c5373767d674b9aefea5fbf737897397eb/include/mcl/ec.hpp#L1685>
  ```
	static void mulArrayGLV(Ec& z, const Ec& x, const mcl::fp::Unit *y, size_t yn, bool isNegative, bool constTime)
	{
		mpz_class s;
		bool b;
		mcl::gmp::setArray(&b, s, y, yn);
		assert(b);
		if (isNegative) s = -s;
		mul(z, x, s, constTime);
	}
  ```

Well PublicKey is of type ECP2.

I think this is just the exact same that I did for SecretKey.Set()! More or
less the same thing anyway.

## bls.Sign -> Recover

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L586>
  ```
  // Recover --
  func (sig *Sign) Recover(sigVec []Sign, idVec []ID) error {
    if len(sigVec) != len(idVec) {
      return fmt.Errorf("err Sign.Recover bad size")
    }
    // #nosec
    ret := C.blsSignatureRecover(&sig.v, &sigVec[0].v, (*C.blsId)(&idVec[0].v), (C.mclSize)(len(idVec)))
    if ret != 0 {
      return fmt.Errorf("err blsSignatureRecover")
    }
    return nil
  }
  ```

  <https://github.com/herumi/bls/blob/4ae022a6bb71dc518d81f22141d71d2a1f767ab3/src/bls_c_impl.hpp#L588>
  ```
  int blsSignatureRecover(blsSignature *sig, const blsSignature *sigVec, const blsId *idVec, mclSize n)
  {
    bool b;
    mcl::LagrangeInterpolation(&b, *cast(&sig->v), cast(&idVec->v), cast(&sigVec->v), n);
    return b ? 0 : -1;
  }
  ```

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/lagrange.hpp#L11>
  ```
  /*
    recover out = f(0) by { (x, y) | x = S[i], y = f(x) = vec[i] }
    @retval 0 if succeed else -1
  */
  template<class G, class F>
  void LagrangeInterpolation(bool *pb, G& out, const F *S, const G *vec, size_t k)
  {
    if (k == 0) {
      *pb = false;
      return;
    }
    if (k == 1) {
      out = vec[0];
      *pb = true;
      return;
    }
    /*
      delta_{i,S}(0) = prod_{j != i} S[j] / (S[j] - S[i]) = a / b
      where a = prod S[j], b = S[i] * prod_{j != i} (S[j] - S[i])
    */
    F a = S[0];
    for (size_t i = 1; i < k; i++) {
      a *= S[i];
    }
    if (a.isZero()) {
      *pb = false;
      return;
    }
    /*
      f(0) = sum_i f(S[i]) delta_{i,S}(0)
    */
    G r;
    r.clear();
    for (size_t i = 0; i < k; i++) {
      F b = S[i];
      for (size_t j = 0; j < k; j++) {
        if (j != i) {
          F v = S[j] - S[i];
          if (v.isZero()) {
            *pb = false;
            return;
          }
          b *= v;
        }
      }
      G t;
      G::mul(t, vec[i], a / b);
      r += t;
    }
    out = r;
    *pb = true;
  }
  ```

Very luckily, this just looks the exact same as SHARE.go -> Recover

  <https://github.com/miracl/core/blob/master/go/SHARE.go#L140>

Where the singular argument []SHARE corresponds to `sigVec []Sign, idVec []ID`.
Especially since sigVec and idVec need to be of the same length.

  <https://github.com/miracl/core/blob/master/go/SHARE.go#L64>
  ```
  type SHARE struct {
    ID   byte       // Unique Share ID
    NSR   byte		// Number of Shares required for recovery
    B    []byte		// Share
  }
  ```

probably NSR is just len(sigVec).
To get ID, just convert bls.ID back to a number (originally it's just an int).
Once you have ID and the signature, create a bunch of NewSHARE as shown in
the miracl/core unit test TestMPIN.go

  <https://github.com/miracl/core/blob/master/go/TestMPIN.go#L107>
  ```
  Sh1:=core.NewSHARE(1,3,TOKEN[:],R)  // indicate 3 shares required for recovery
  Sh2:=core.NewSHARE(2,3,TOKEN[:],R)
  // Sh3:=core.NewSHARE(3,3,TOKEN[:],R)	// must comment out or Go throws an error (Grrr..)
  Sh4:=core.NewSHARE(4,3,TOKEN[:],R)

  var Shares [3]*core.SHARE
  ```

Or actually create a new share with `sh := new(core.SHARE)` and then set the
fields.

## bls.PublicKey -> Recover

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L467>
  ```
  // Recover --
  func (pub *PublicKey) Recover(pubVec []PublicKey, idVec []ID) error {
    if len(pubVec) != len(idVec) {
      return fmt.Errorf("err PublicKey.Recover bad size")
    }
    // #nosec
    ret := C.blsPublicKeyRecover(&pub.v, &pubVec[0].v, (*C.blsId)(&idVec[0].v), (C.mclSize)(len(idVec)))
    if ret != 0 {
      return fmt.Errorf("err blsPublicKeyRecover")
    }
    return nil
  }
  ```

  <https://github.com/herumi/bls/blob/4ae022a6bb71dc518d81f22141d71d2a1f767ab3/src/bls_c_impl.hpp#L581>
  ```
  int blsPublicKeyRecover(blsPublicKey *pub, const blsPublicKey *pubVec, const blsId *idVec, mclSize n)
  {
    bool b;
    mcl::LagrangeInterpolation(&b, *cast(&pub->v), cast(&idVec->v), cast(&pubVec->v), n);
    return b ? 0 : -1;
  }
  ```

## bls.PublicKey -> Add

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L452>
  ```
  // Add --
  func (pub *PublicKey) Add(rhs *PublicKey) {
    C.blsPublicKeyAdd(&pub.v, &rhs.v)
  }
  ```

  <https://github.com/herumi/bls/blob/4ae022a6bb71dc518d81f22141d71d2a1f767ab3/src/bls_c_impl.hpp#L600>
  ```
  void blsPublicKeyAdd(blsPublicKey *pub, const blsPublicKey *rhs)
  {
    *cast(&pub->v) += *cast(&rhs->v);
  }
  ```

## bls.GetMasterPublicKey

We already have GetPublicKey method of SecretKey, so this is a simple port.

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L320>
  ```
  // GetMasterPublicKey --
  func GetMasterPublicKey(msk []SecretKey) (mpk []PublicKey) {
    n := len(msk)
    mpk = make([]PublicKey, n)
    for i := 0; i < n; i++ {
      mpk[i] = *msk[i].GetPublicKey()
    }
    return mpk
  }
  ```

## bls.ID

bls.ID is just Fr

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L47>
  // ID --
  type ID struct {
    v C.blsId
  }

  <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L52>
  typedef struct {
    mclBnFr v;
  } blsId;

## bls.SecretKey.SetHexString

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L259>
  ```
  // SetHexString --
  func (sec *SecretKey) SetHexString(s string) error {
    buf := []byte(s)
    // #nosec
    err := C.blsSecretKeySetHexStr(&sec.v, (*C.char)(unsafe.Pointer(&buf[0])), C.mclSize(len(buf)))
    if err != 0 {
      return fmt.Errorf("err blsSecretKeySetHexStr %s", s)
    }
    return nil
  }
  ```

  <https://github.com/herumi/bls/blob/4ae022a6bb71dc518d81f22141d71d2a1f767ab3/src/bls_c_impl.hpp#L986>
  ```
  int blsSecretKeySetHexStr(blsSecretKey *sec, const char *buf, mclSize bufSize)
  {
    return cast(&sec->v)->deserialize(buf, bufSize, 16) > 0 ? 0 : -1;
  }
  ```

## bls.ID.SetDecString

Overall plan for SecDecString is to just ditch it and use SetHexString instead.
Where it is used in the library is simple enough anyway to just use hexstring
instead...

So this is just FP.serialize.

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L135>
  ```
  // SetDecString --
  func (id *ID) SetDecString(s string) error {
    buf := []byte(s)
    // #nosec
    err := C.blsIdSetDecStr(&id.v, (*C.char)(unsafe.Pointer(&buf[0])), C.mclSize(len(buf)))
    if err != 0 {
      return fmt.Errorf("err blsIdSetDecStr %s", s)
    }
    return nil
  }
  ```

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L896>
  ```
  int blsIdSetDecStr(blsId *id, const char *buf, mclSize bufSize)
  {
    return cast(&id->v)->deserialize(buf, bufSize, 10) > 0 ? 0 : -1;
  }
  ```

## getBasePoint ##

Eureka! getBasePoint is just the same thing as ECP2_generator!

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

OK, by luck, found the deserialization here: https://github.com/herumi/mcl/blob/b32aecd583af99ee6a63c12e14224140fde442d9/include/mcl/ec.hpp#L1076

  ```
  cybozu::MemoryOutputStream mos(buf + adj, n);
  P.x.save(pb, mos, IoSerialize); if (!*pb) return;
  if (adj) {
    buf[0] = P.y.isOdd() ? 3 : 2;
  } else {
    if (P.y.isOdd()) {
      buf[n - 1] |= 0x80;
    }
  }
  ```

and pseudocode documented here: https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/api.md#serialization-format

  size = mclBn_getG1ByteSize() # resp. mclBn_getG1ByteSize() * 2
  if P is zero:
    return [0] * size
  else:
    P = P.normalize()
    s = P.x.serialize()
    # x in Fp2 is odd <=> x.a is odd
    if P.y is odd: # resp. P.y.d[0] is odd
      s[byte-length(s) - 1] |= 0x80
    return s

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

## SecretKey.Add

really just Fr.add

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L279>
  // Add --
  func (sec SecretKey) Add(rhs SecretKey) {
    C.blsSecretKeyAdd(&sec.v, &rhs.v)
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L578>
  void blsSecretKeyAdd(blsSecretKey sec, const blsSecretKey rhs)
  {
    *cast(&sec->v) += *cast(&rhs->v);
  }

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

And see this following analysis from other place in the document about ECP2:

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

Oh snaps correlates to!

  type ECP2 struct {
    x FP2
    y FP2
    z FP2
  }

Also see this in mcl lib:

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L96>
  ```
  typedef struct {
    mclBnFp x, y, z;
  } mclBnG1;

  typedef struct {
    mclBnFp2 x, y, z;
  } mclBnG2;
  ```

Oh snaps, found a bunch of MCL docs here: <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/api.md#mclbnfp>

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
  inline void precomputeG2(Fp6 Qcoeff, const G2& Q_)

But interestingly enough, that's a CONST on the Q_, so that doesn't actually do anything to the BasePoint.

BN254.Init initializes a global:

  // BLS.go
  var G2_TAB []FP4

And FP4 is just:

  // FP4.go
  type FP4 struct {
    a FP2
    b FP2
  }
  // FP2.go
  type FP2 struct {
    a FP
    b FP
  }

The original source precomputes a G2 as g_Q

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

Oh snaps correlates to!

  type ECP2 struct {
    x FP2
    y FP2
    z FP2
  }

The code we want to port is:

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L104>
	if (curve == MCL_BN254) {
		const char *Qx_BN254 = "11ccb44e77ac2c5dc32a6009594dbe331ec85a61290d6bbac8cc7ebb2dceb128 f204a14bbdac4a05be9a25176de827f2e60085668becdd4fc5fa914c9ee0d9a";
		const char *Qy_BN254 = "7c13d8487903ee3c1c5ea327a3a52b6cc74796b1760d5ba20ed802624ed19c8 8f9642bbaacb73d8c89492528f58932f2de9ac3e80c7b0e41f1a84f1c40182";
		g_Q.x.setStr(&b, Qx_BN254, 16);
		g_Q.y.setStr(&b, Qy_BN254, 16);
		g_Q.z = 1;
	} else {

And if I'm not mistaken, that's just the same as ECP_generator? This is pretty
damn good!

## SecretKey.Sign ##

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L526>
  // Sign -- Constant Time version
  func (sec SecretKey) Sign(m string) (sig Sign) {
    sig = new(Sign)
    buf := []byte(m)
    // #nosec
    C.blsSign(&sig.v, &sec.v, unsafe.Pointer(&buf[0]), C.mclSize(len(buf)))
    return sig
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L186>
  void blsSign(blsSignature sig, const blsSecretKey sec, const void m, mclSize size)
  {
    blsHashToSignature(sig, m, size);
    Fr s = *cast(&sec->v);
    GmulCT(*cast(&sig->v), *cast(&sig->v), s);
  }

LOL how lucky I am to find Core_Sign! It fits perfectly!

  // BLS.go
  func Core_Sign(SIG []byte, M []byte, S []byte) int {}

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

As can be seen from the following snippets, just need to convert string to ECP,
then perform Add.

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L535>
  // Add --
  func (sig Sign) Add(rhs Sign) {
    C.blsSignatureAdd(&sig.v, &rhs.v)
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L588>
  void blsSignatureAdd(blsSignature *sig, const blsSignature *rhs)
  {
    *cast(&sig->v) += *cast(&rhs->v);
  }

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

  https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L284
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
  // GetPublicKey --
  func (sec SecretKey) GetPublicKey() (pub PublicKey) {
    pub = new(PublicKey)
    C.blsGetPublicKey(&pub.v, &sec.v)
    return pub
  }

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L175>
  // Remember, this returns public key (0th arg as pub).
  ```
  void blsGetPublicKey(blsPublicKey *pub, const blsSecretKey *sec)
  {
    Gmul(*cast(&pub->v), getBasePoint(), *cast(&sec->v));
  }
  ```

Now if I'm not mistaken, that just looks a lot like MPIN_GET_SERVER_SECRET,
which is just horribly named because it doesn't have the word 'public' in it?

  MPIN.go-153-func MPIN_GET_SERVER_SECRET(S []byte, SST []byte) int {
  MPIN.go:154:    Q := ECP2_generator()
  MPIN.go-155-    s := FromBytes(S) // S rly just FR, see: <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L56>
  MPIN.go-156-    Q = G2mul(Q, s)
  MPIN.go-157-    Q.ToBytes(SST,false)
  MPIN.go-158-    return 0
  MPIN.go-159-}

The reason why I think it could possibly work:
+ turns out ECP2_generator is just the same thing as getBasePoint
+ It does G2mul, which is the same thing as Gmul! Because getBasePoint is a G2

Gmul is just a multimethod. There are 2 Gmul defined. C++ chooses the right function
based on the types passed to the function. See them defined here:
  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L16>

  inline void Gmul(G1& z, const G1& x, const Fr& y) { G1::mul(z, x, y); }
  inline void Gmul(G2& z, const G2& x, const Fr& y) { G2::mul(z, x, y); }

Gmul is either G1mul or G2mul. The diff between G1 and G2 is a #ifdef on BLS_ETH <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L50>

getBasePoint returns a static G1/G2. The G2 structure is:

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
  // And MCLBN_FR_UNIT_SIZE is 4 or 6. It's 6 if it's bls384.

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

## SecretKey SerializeToHexStr

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L205>
  func (sec SecretKey) SerializeToHexStr() string {
    return hex.EncodeToString(sec.Serialize())
  }

  <https://github.com/herumi/bls-go-binary/blob/master/bls/bls.go#L159>
  // Serialize --
  func (sec SecretKey) Serialize() []byte {
    buf := make([]byte, 2048)
    // #nosec
    n := C.blsSecretKeySerialize(unsafe.Pointer(&buf[0]), C.mclSize(len(buf)), &sec.v)
    if n == 0 {
      panic("err blsSecretKeySerialize")
    }
    return buf[:n]
  }

## PublicKey SerializeToHexStr

## FrSub

From Shigeo:

  The difference between the Fr and Fp class of mcl is only the modulo parameter.
  If MIRACL has the Fp class, then you may be able to use it as the Fr by setting the parameter.

Luckily, MIRACL *does* have FP class (FP.go), so FrSub is just:

  <https://github.com/miracl/core/blob/master/go/FP.go#L381-L385>
  func (F FP) sub(b FP) {
    n := NewFPcopy(b)
    n.neg()
    F.add(n)
  }

The original FrSub takes 3 args. The 0th arg is the return variable. The 1th
and 2nd args are where the subtraction should take place.

## var sk bls.SecretKey; sk.Set(polynomial, &id)

sk.Set is defined here:

  <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L305-L312>
  // Calls into C.blsSecretKeyShare
  ```
  // Set --
  func (sec *SecretKey) Set(msk []SecretKey, id *ID) error {
    // #nosec
    ret := C.blsSecretKeyShare(&sec.v, &msk[0].v, (C.mclSize)(len(msk)), &id.v)
    if ret != 0 {
      return fmt.Errorf("err blsSecretKeyShare")
    }
    return nil
  }
  ```

  <https://github.com/herumi/bls/blob/3005a32a97ebdcb426d59caaa9868a074fe7b35a/src/bls_c_impl.hpp#L543>
  // Calls into mcl::evaluatePolynomial
  ```
  int blsSecretKeyShare(blsSecretKey *sec, const blsSecretKey* msk, mclSize k, const blsId *id)
  {
    bool b;
    mcl::evaluatePolynomial(&b, *cast(&sec->v), cast(&msk->v), k, *cast(&id->v));
    return b ? 0 : -1;
  }
  ```

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/lagrange.hpp#L64>
  ```
  /*
    out = f(x) = c[0] + c[1] * x + c[2] * x^2 + ... + c[cSize - 1] * x^(cSize - 1)
    @retval 0 if succeed else -1 (if cSize == 0)
  */
  template<class G, class T>
  void evaluatePolynomial(bool *pb, G& out, const G *c, size_t cSize, const T& x)
  {
    if (cSize == 0) {
      *pb = false;
      return;
    }
    if (cSize == 1) {
      out = c[0];
      *pb = true;
      return;
    }
    G y = c[cSize - 1];
    for (int i = (int)cSize - 2; i >= 0; i--) {
      G::mul(y, y, x);
      G::add(y, y, c[i]);
    }
    out = y;
    *pb = true;
  }
  ```

Herumi's email:

  // Secret Key "mul" and "add" is defined...
  <https://github.com/herumi/mcl/blob/master/include/mcl/fp.hpp#L499>
	`static inline void add(FpT& z, const FpT& x, const FpT& y) { op_.fp_add(z.v_, x.v_, y.v_, op_.p); }`

  The naive algorithm is (x + y) % r and (x * y) % r.

Notice how out is 'z', while 'x' and 'y' are const.

Remember that SecretKey and bls.ID are both just Fr

Notice how these are all consts except for the first two vars.

  <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/impl/bn_c_impl.hpp#L669>
  `int mclBn_FrEvaluatePolynomial(mclBnFr *out, const mclBnFr *cVec, mclSize cSize, const mclBnFr *x)`

  evaluatePolynomial uses mul and add, see Herumi's email.

Now if I'm not mistaken, is this just BIG.go "func smul"?

  // BIG.go
  ```
  func smul(a *BIG, b *BIG) *BIG {
    carry := Chunk(0)
    c := NewBIG()
    for i := 0; i < NLEN; i++ {
      carry = 0
      for j := 0; j < NLEN; j++ {
        if i+j < NLEN {
          carry, c.w[i+j] = muladd(a.w[i], b.w[j], carry, c.w[i+j])
        }
      }
    }
    return c
  }
  ```

Nah, muladd doesn't look promising:

  ```
  /* set this[i]+=x*y+c, and return high part */
  func muladd(a Chunk, b Chunk, c Chunk, r Chunk) (Chunk, Chunk) {

    tp,bt := bits.Mul64(uint64(a),uint64(b))  // use math/bits intrinsic
    bot := Chunk(bt&uint64(BMASK))
    top := Chunk((tp << (64-BASEBITS)) | (bt >> BASEBITS))
    bot += c; bot += r
    carry := bot>>BASEBITS
    bot &= BMASK
    top+=carry
    return top, bot

  }
  ```

Modmul and Modadd look more promising!


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
