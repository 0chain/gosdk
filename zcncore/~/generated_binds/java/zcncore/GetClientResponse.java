// Code generated by gobind. DO NOT EDIT.

// Java class zcncore.GetClientResponse is a proxy for talking to a Go program.
//
//   autogenerated by gobind -lang=java command-line-arguments
package zcncore;

import go.Seq;

public final class GetClientResponse implements Seq.Proxy {
	static { Zcncore.touch(); }
	
	private final int refnum;
	
	@Override public final int incRefnum() {
	      Seq.incGoRef(refnum, this);
	      return refnum;
	}
	
	GetClientResponse(int refnum) { this.refnum = refnum; Seq.trackGoRef(refnum, this); }
	
	public GetClientResponse() { this.refnum = __New(); Seq.trackGoRef(refnum, this); }
	
	private static native int __New();
	
	public final native String getID();
	public final native void setID(String v);
	
	public final native String getVersion();
	public final native void setVersion(String v);
	
	public final native long getCreationDate();
	public final native void setCreationDate(long v);
	
	public final native String getPublicKey();
	public final native void setPublicKey(String v);
	
	@Override public boolean equals(Object o) {
		if (o == null || !(o instanceof GetClientResponse)) {
		    return false;
		}
		GetClientResponse that = (GetClientResponse)o;
		String thisID = getID();
		String thatID = that.getID();
		if (thisID == null) {
			if (thatID != null) {
			    return false;
			}
		} else if (!thisID.equals(thatID)) {
		    return false;
		}
		String thisVersion = getVersion();
		String thatVersion = that.getVersion();
		if (thisVersion == null) {
			if (thatVersion != null) {
			    return false;
			}
		} else if (!thisVersion.equals(thatVersion)) {
		    return false;
		}
		long thisCreationDate = getCreationDate();
		long thatCreationDate = that.getCreationDate();
		if (thisCreationDate != thatCreationDate) {
		    return false;
		}
		String thisPublicKey = getPublicKey();
		String thatPublicKey = that.getPublicKey();
		if (thisPublicKey == null) {
			if (thatPublicKey != null) {
			    return false;
			}
		} else if (!thisPublicKey.equals(thatPublicKey)) {
		    return false;
		}
		return true;
	}
	
	@Override public int hashCode() {
	    return java.util.Arrays.hashCode(new Object[] {getID(), getVersion(), getCreationDate(), getPublicKey()});
	}
	
	@Override public String toString() {
		StringBuilder b = new StringBuilder();
		b.append("GetClientResponse").append("{");
		b.append("ID:").append(getID()).append(",");
		b.append("Version:").append(getVersion()).append(",");
		b.append("CreationDate:").append(getCreationDate()).append(",");
		b.append("PublicKey:").append(getPublicKey()).append(",");
		return b.append("}").toString();
	}
}

