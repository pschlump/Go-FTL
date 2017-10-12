package gosrp

import (
	// "crypto/hmac"
	// cryptorand "crypto/rand"
	// "crypto/sha256"
	// "crypto/subtle"
	// "fmt"
	// "io"
	"./big" // "math/big"
	"fmt"
	// mathrand "math/rand"
)

// Phase 1 - just make it work in test code
// Phase 2 - make tests that connect to Redis
// Phase 3 - test it with multiple calls  - Simulate Client
// Phase 4 - test it with real client

/*

From: http://srp.stanford.edu/ndss.html

(With additional notes by me)

The SRP Protocol
================

What follows is a complete description of the entire SRP authentication process from beginning to end, starting with the password setup steps.

			Table 3: Mathematical Notation for SRP
			---------------------------------------

		Var		Description

		n	   	A large prime number. All computations are performed modulo n.
		g	   	A primitive root modulo n (often called a generator)		- this is 2 or 5 - from the table based on bitsize
		N		Modulo number												- big number from table - use based on standard (js/rfc5054-2048-sha256.js)
																			- this is really the same as 'n'
		s	   	A random string used as the user's salt
		P	   	The user's password
		x	   	A private key derived from the password and salt
		v	   	The host's password verifier
		u	   	Random scrambling parameter, publicly revealed 				- (hash of A.B)
		a,b	   	Ephemeral private keys, generated randomly and not publicly revealed
		A,B	   	Corresponding public keys
				Client: A = g^a, compute and send A to server
				Server: B = v + g^b
		H()	   	One-way hash function. 										-  In this H() will be Sha256
		m,n	   	The two quantities (strings) m and n concatenated
		K	   	Session key

		C		Carol's Username (carol@example.com, also referred to as I in some cases)
		D.B.	Database
		t		Random UUID used as salt for generating session ID in steps 9,10
		r		Random UUID with timeout for keeping data during authorization

Table 3 shows the notation used in this section. The values n and g are well-known values, agreed to beforehand.

To establish a password P with Steve, Carol picks a random salt s, and computes

		x = H(s, P)
		v = g^x

Steve stores v and s as Carol's password verifier and salt. Remember that the computation of v is implicitly reduced modulo n. x is
discarded because it is equivalent to the plaintext password P.

The AKE protocol also allows Steve to have a password z with a corresponding public key held by Carol; in SRP, we set z = 0 so
that it drops out of the equations. Since this private key is 0, the corresponding public key is 1. Consequently, instead of
safeguarding its own password z, Steve needs only to keep Carol's verifier v secret to assure mutual authentication. This frees Carol
from having to remember Steve's public key and simplifies the protocol.

To authenticate, Carol and Steve engage in the protocol described in Table 4. A description of each step follows:

			Table 4: The Secure Remote Password Protocol
			--------------------------------------------

		Step	Carol				Communication				Steve
		1.								C -->					(lookup s, v from D.B. by username(C)) - send back s		/api/srp_login
		2.		x = H(s, P)				<-- s, t, r				t is 2nd salt sent back to client
																r is 3rd random - temporary for saving data
																In D.B. create the key(r){C,t,s,v} + Timeout
																In D.B. create the key(t){C,r} Session Key

		3.		A = g^a					A,r -->																				/api/srp_confirm
																Fetch from d.b. using(r){C,t,s,v}
		4.								<-- B, u				B = v + g^b			Lookup D.B. (C), get s,v - gen b      ?? u
		5.		S = (B - g^x)^(a + ux)							S = (A · v^u)^b		Both sides can now compute S
		6.		K = H(S)										K = H(S)			Both sides compute the same K - Update D.B. with A,B,b,K
																In d.b. Update(r){C,t,s,v,B,S,K,M,K}

		7.		M[1] = H(A, B, K)		M[1],r -->				(verify M[1])		Lookup D.B. getting s, A, B, b, K
		8.		(verify M[2])			<-- M[2]				M[2] = H(A, M[1], K)

		9.		U = H(t,K)										U = H(t,K)			Generate session ID, store K, C, s, info in D.B. with key U
		10.		Use U as key for communication
				(Encrypt with K)								(Decrypt with K) (Encrypt Responses with K)
				(Decrypt with K)
																Update(t) with logged in


1. Carol sends Steve her username, (e.g. carol@example.com).
	Example #8
2. Steve looks up Carol's password entry and fetches her password verifier v and her salt s. He sends s to Carol.
	Carol computes her long-term private key x using s and her real password P.
	Example s=#13, v=??
		var v = xxx.generateVerifier(s,identity,password);
		The verifier is computed as v = g^x (mod N).
		g=#2, N=#1000001
		x=??
3. Carol generates a random number a, 1 < a < n, computes her ephemeral public key A = g^a, and sends it to Steve.
4. Steve generates his own random number b, 1 < b < n, computes his ephemeral public key B = v + g^b, and sends
	it back to Carol, along with the randomly generated parameter u.
5. Carol and Steve compute the common exponential value S = g^(ab + bux) using the values available to each of them.
	If Carol's password P entered in Step 2 matches the one she originally used to generate v, then both values of
	S will match.
6. Both sides hash the exponential S into a cryptographically strong session key.
7. Carol sends Steve M[1] as evidence that she has the correct session key. Steve computes M[1] himself and verifies
	that it matches what Carol sent him.
8. Steve sends Carol M[2] as evidence that he also has the correct session key. Carol also verifies M[2] herself,
	accepting only if it matches Steve's value.

This protocol is mostly the result of substituting the equations of Section 3.2.1 into the generic AKE protocol, adding explicit
flows to exchange information like the user's identity and the salt s. Both sides will agree on the session key S = g^(ab + bux) if all
steps are executed correctly. SRP also adds the two flows at the end to verify session key agreement using a one-way hash function.
Once the protocol run completes successfully, both parties may use K to encrypt subsequent session traffic.

Version 0.0.1

https://github.com/RuslanZavacky/srp-6a-demo/blob/master/srp/Server/Srp.php
https://github.com/RuslanZavacky/srp-6a-demo
https://en.wikipedia.org/wiki/Secure_Remote_Password_protocol
http://srp.stanford.edu/analysis.html
https://en.wikipedia.org/wiki/Secure_Remote_Password_protocol -- Shows how to gen the B, A key

*/

type GoSrp struct {
	State  int
	Salt_s []byte
	Xv_s   []byte
	Salt   *big.Int
	Xv     *big.Int
	XN     *big.Int
	Xg     *big.Int
	Xk     *big.Int
	Key_s  string
	Xb     *big.Int
	Xb_s   string
	XB     *big.Int
	XB_s   string
	Xa     *big.Int
	Xa_s   string
	XA     *big.Int
	XA_s   string
	XI_s   []byte
	XS     *big.Int
	XS_s   string
	XM     string
	XHAMK  string
	Xu     *big.Int
	Xu_s   string
	Xavu   *big.Int
}

func GoSrpNew(C []byte) (rv *GoSrp) {
	N, _ := big.NewInt(0).SetString("AC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73", 16)
	return &GoSrp{
		State: 0,
		XI_s:  C,
		Xg:    big.NewInt(2),
		XN:    N,
	}
}

var bits = 2048

var g_debug1 = true
var g_db_pos = 0
var g_ranv = []string{
	"a5e998caa34be4c8843ceaa3897d4812da588c518af40216e685a8325736b12e7ddd5d2d905b01fbe7cdc7d11dbd25ac59a7f0ec51c1f10efe3d6f91b1550418",
	"706a423a9b390a79a21a53b5ebb02bcf55be72fa4f9b151f03630558cf0309f9a6e5fe876ae82bd1e1e822ed46d08d353c9aaff3fbc5aa77f1d921e2150c6751",
	"2a0c1260b9726fc4860feb7f3f0e500eaf330394c1344bd0f584c74924be5fb1c6415911ff21274b10f43bab43a0697fc0e554b8d4809014303c58c545fe49f6",
	"37686caac6d5f387301ec38565af22024f95893cf8fcd51306e2ba049bd966c182b6a1dda8155540bdf918bfb46b8f610c2dbebe8862e431a30a605a65bb1c77",
	"5327effd2a7f5ff1c37313c1b75d09501c9a9edc4efcf18690f58bb1f87aee5a880c4a4bd470bd425365e40d1c6a05fb26d0f553339b0faf12fbca1f82511bf9",
	"fa8606fe621cd0ae9a16b7920357c83f56a2f9243c266f1f34a2222e97060cc6ad9cd6116ec7787fd406b75a3eb0ae5ad2fa0b0fc6aa84ce26947946e69f3bce",
	"a7afc01c620ef349e4c95fda9a59359abe08bdff3b4143018bcdb9d703a5bcae404e944581bfda2b7ad2c36d7c00a3bae2d0954638b16ee1e8b13691d809ade6",
	"951d8d52017cfa5260206e7db0bed8535b6435a1aa3ba7d2310fc03b96a1538cee798fda92b433ae577411a7147c8f186c26ff582a3bbcbe93e48e0ade190b7d",
	"4d72896c1ef128209e72a44a4d5c30ef22b6cc0aba3aa00a481028f8e58d640ba531aa6c09fb2b874478edf1752ae3b165bfeb99128d8246e68c09271000aed1",
	"3f2d07a6ddc2894d631a452abb60233ab8857586a77bd59897621fb8212ff730444850aee4fd3217971b70d8fc1fc8e18d66aac6c884a652815c4523e4e55ac3",
}

// match with __construct
func (gs *GoSrp) Setup(verifier []byte, salt []byte) {
	zero := big.NewInt(0)
	ok := false

	gs.State = 1
	gs.Salt_s = salt
	gs.Salt, ok = big.NewInt(0).SetString(string(gs.Salt_s), 16)
	gs.Xv_s = verifier
	gs.Xv, ok = big.NewInt(0).SetString(string(gs.Xv_s), 16)
	if !ok {
		fmt.Printf("Error on convert of string to big, variable v \n")
		panic("")
	}
	fmt.Printf("Hash=%s\n", Hashstring(gs.XN.HexString()+gs.Xg.HexString()))
	gs.Xk, ok = big.NewInt(0).SetString(Hashstring(gs.XN.HexString()+gs.Xg.HexString()), 16)
	if !ok {
		fmt.Printf("Error on convert of string to big, variable k\n")
		panic("")
	}
	gs.Key_s = ""
	for {
		gs.Xb = randlong(bits)
		gs.Xb_s = gs.Xb.HexString()

		gPowed := big.NewInt(0).Exp(gs.Xg, gs.Xb, gs.XN)
		t1 := big.NewInt(0).Mul(gs.Xk, gs.Xv)
		t2 := big.NewInt(0).Add(t1, gPowed)
		gs.XB = big.NewInt(0).Mod(t2, gs.XN)

		tf := big.NewInt(0).Mod(gs.XB, gs.XN)
		if tf.Cmp(zero) != 0 {
			break
		}
	}
	gs.XB_s = gs.XB.HexString()

	// Xyzzy - conver to hext values - and test
}

func (gs *GoSrp) CalculateA() string {
	zero := big.NewInt(0)

	for {
		gs.Xa = randlong(bits)
		gs.Xa_s = gs.Xa.HexString()

		gPowed := big.NewInt(0).Exp(gs.Xg, gs.Xa, gs.XN)
		// t1 := big.NewInt(0).Mul(gs.Xk, gs.Xv)
		// t2 := big.NewInt(0).Add(t1, gPowed)
		// gs.XB = big.NewInt(0).Mod(t2, gs.XN)
		gs.XA = gPowed

		tf := big.NewInt(0).Mod(gs.XA, gs.XN)
		if tf.Cmp(zero) != 0 {
			break
		}
	}

	gs.XA_s = gs.XA.HexString()
	gs.State = 2

	return gs.XA_s
}

func (gs *GoSrp) IssueChallenge(A_s string) (B_s []byte) {
	zero := big.NewInt(0)
	ok := true

	gs.XA, ok = big.NewInt(0).SetString(A_s, 16)
	if !ok {
		fmt.Printf("A did not parse - error, [%s]\n", A_s)
		panic("")
	}
	gs.XA_s = gs.XA.HexString()
	if gs.XA_s != A_s {
		fmt.Printf("Bad conversion\n")
		panic("")
	}

	tf := big.NewInt(0).Mod(gs.XA, gs.XN)
	if tf.Cmp(zero) == 0 {
		fmt.Printf("Bad A Value - error, %s\n", gs.XA.HexString())
		panic("")
	}
	fmt.Printf("db A [%s] B [%s], hash [%s]\n", gs.XA_s, gs.XB_s, Hashstring(gs.XA_s+gs.XB_s))
	u, ok := big.NewInt(0).SetString(Hashstring(gs.XA_s+gs.XB_s), 16)
	t1 := big.NewInt(0).Set(gs.XA)
	t2 := big.NewInt(0).Exp(gs.Xv, u, gs.XN)
	avu := big.NewInt(0).Mul(t1, t2)
	gs.XS = big.NewInt(0).Exp(avu, gs.Xb, gs.XN)
	gs.XS_s = gs.XS.HexString()
	gs.Key_s = Hashstring(gs.XS_s)

	// Compute M, HAMK
	gs.XM = Hashstring(gs.XA_s + gs.XB_s + gs.XS_s)
	gs.XHAMK = Hashstring(gs.XA_s + gs.XM + gs.XS_s)

	gs.Xu = u
	gs.Xu_s = u.HexString()
	gs.Xavu = avu
	gs.State = 3

	_ = ok

	return []byte(gs.XB.HexString())
}

/*
	echo "ST/bits=  [{$this->ST}/{$this->bits}]\n\n";
	echo "verifier= [{$this->verifier}]\n\n";
	echo "salt=     [{$this->salt}]\n\n";
	echo "Nhex=     [{$this->Nhex}]\n\n";
	echo "g=        [{$this->g}]\n\n";
	echo "khex=     [{$this->khex}]\n\n";
	echo "vhex=     [{$this->vhex}]\n\n";
	echo "key=      [{$this->key}]\n\n";
	echo "bhex=     [{$this->bhex}]\n\n";
	echo "Bhex=     [{$this->Bhex}]\n\n";
*/
func (gs *GoSrp) TestDump1() {
	// Salt   *big.Int
	// Xv     *big.Int
	// XN     *big.Int
	// Xg     *big.Int
	// Xk     *big.Int
	// Xb     *big.Int
	// XB     *big.Int
	fmt.Printf("ST/bits=  [%d/%d]\n\n", gs.State, bits)
	// echo "verifier= [{$this->verifier}]\n\n";
	// echo "salt=     [{$this->salt}]\n\n";
	fmt.Printf("salt=     [%s]\n", gs.Salt.HexString())
	// echo "Nhex=     [{$this->Nhex}]\n\n";
	// echo "g=        [{$this->g}]\n\n";
	fmt.Printf("khex=     [%s]\n", gs.Xk.HexString())
	fmt.Printf("vhex=     [%s]\n", gs.Xv.HexString())
	fmt.Printf("key=      [%s]\n", gs.Key_s)
	fmt.Printf("bhex=     [%s]\n", gs.Xb.HexString())
	fmt.Printf("Bhex=     [%s]\n", gs.XB.HexString())
}

/*
	echo "Simulated Client\n\n";
	echo "ST/bits=  [{$this->ST}/{$this->bits}]\n\n";
	echo "ahex=     [{$this->ahex}]\n\n";
	echo "Ahex=     [{$this->Ahex}]\n\n";
*/
func (gs *GoSrp) TestDump2() {
	fmt.Printf("Simulated Client\n\n")
	fmt.Printf("ST/bits=  [%d/%d]\n\n", gs.State, bits)
	fmt.Printf("ahex=     [%s]\n", gs.Xa.HexString())
	fmt.Printf("Ahex=     [%s]\n", gs.XA.HexString())
}

/*
	public function dumpVars3() {
		echo "ST/bits=  [{$this->ST}/{$this->bits}]\n\n";
		echo "Ahex=     [{$this->Ahex}]\n\n";
		echo "Shex=     [{$this->Shex}]\n\n";
		echo "M=        [{$this->M}]\n\n";
		echo "HAMK=     [{$this->HAMK}]\n\n";
		echo "key=      [{$this->key}]\n\n";
	}
*/
func (gs *GoSrp) TestDump3() {
	fmt.Printf("ST/bits=  [%d/%d]\n\n", gs.State, bits)
	fmt.Printf("Ahex=     [%s]\n", gs.XA.HexString())
	fmt.Printf("Shex=     [%s]\n", gs.XS.HexString())
	fmt.Printf("M=        [%s]\n", gs.XM)
	fmt.Printf("HAMK=     [%s]\n", gs.XHAMK)
	fmt.Printf("key=      [%s]\n", gs.Key_s)
	fmt.Printf("--\n")
	fmt.Printf("uhex=     [%s]\n", gs.Xu.HexString())
	fmt.Printf("vhex=     [%s]\n", gs.Xv.HexString())
	fmt.Printf("avuhex=   [%s]\n", gs.Xavu.HexString())
}
