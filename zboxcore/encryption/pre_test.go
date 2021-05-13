package encryption

import (
	"fmt"
	"testing"

	"go.dedis.ch/kyber/v3/group/edwards25519"
)

func TestMnemonic(t *testing.T) {
	suite := edwards25519.NewBlakeSHA256Ed25519()

	rand := suite.XOF([]byte("travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"))
	key := suite.Scalar().Pick(rand)

	Tag = []byte("filetype:audio")

	pre.T = pre.SuiteObj.Point().Pick(suite.RandomStream()) // Pick a random point T from the curve
	pre.Ht = pre.hash1(suite, Tag, "abcd")      // Ht  = H1(tagA,skA)
	var g kyber.Group = pre.SuiteObj
	pre.EncryptedKey = g.Point().Add(pre.Ht, pre.T) // C1  = T + Ht

}
