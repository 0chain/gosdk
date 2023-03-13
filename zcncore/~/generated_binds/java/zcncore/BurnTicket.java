// Code generated by gobind. DO NOT EDIT.

// Java class zcncore.BurnTicket is a proxy for talking to a Go program.
//
//   autogenerated by gobind -lang=java command-line-arguments
package zcncore;

import go.Seq;

/**
 * BurnTicket model used for deserialization of the response received from sharders
 */
public final class BurnTicket implements Seq.Proxy {
	static { Zcncore.touch(); }
	
	private final int refnum;
	
	@Override public final int incRefnum() {
	      Seq.incGoRef(refnum, this);
	      return refnum;
	}
	
	public BurnTicket(String hash, long nonce) {
		this.refnum = __NewBurnTicket(hash, nonce);
		Seq.trackGoRef(refnum, this);
	}
	
	private static native int __NewBurnTicket(String hash, long nonce);
	
	BurnTicket(int refnum) { this.refnum = refnum; Seq.trackGoRef(refnum, this); }
	
	public final native String getHash();
	public final native void setHash(String v);
	
	public final native long getNonce();
	public final native void setNonce(long v);
	
	@Override public boolean equals(Object o) {
		if (o == null || !(o instanceof BurnTicket)) {
		    return false;
		}
		BurnTicket that = (BurnTicket)o;
		String thisHash = getHash();
		String thatHash = that.getHash();
		if (thisHash == null) {
			if (thatHash != null) {
			    return false;
			}
		} else if (!thisHash.equals(thatHash)) {
		    return false;
		}
		long thisNonce = getNonce();
		long thatNonce = that.getNonce();
		if (thisNonce != thatNonce) {
		    return false;
		}
		return true;
	}
	
	@Override public int hashCode() {
	    return java.util.Arrays.hashCode(new Object[] {getHash(), getNonce()});
	}
	
	@Override public String toString() {
		StringBuilder b = new StringBuilder();
		b.append("BurnTicket").append("{");
		b.append("Hash:").append(getHash()).append(",");
		b.append("Nonce:").append(getNonce()).append(",");
		return b.append("}").toString();
	}
}

