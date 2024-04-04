package tibgs

import (
	"crypto/sha256"
	// "fmt"
	// "encoding/binary"
	// "errors"
	// "math/big"

	"github.com/Nik-U/pbc"
)

// parameters of the curve

const str = `type a
	q 8780710799663312522437781984754049815806883199414208211028653399266475630880222957078625179422662221423155858769582317459277713367317481324925129998224791
	h 12016012264891146079388821366740534204802954401251311822919615131047207289359704531102844802183906537786776
	r 730750818665451621361119245571504901405976559617
	exp2 159
	exp1 107
	sign1 1
	sign0 1`

var pairing, _ = pbc.NewPairingFromString(str)   //等同于c中 pairing_init_set_str(my->pairing, (const char *)params);

var GSUSK *TIBGSUserSecretKey
var GSMPK *TIBGSMasterPublicKey

var M, N uint32
var Index, Level uint32

const GSSigLen = 533
//
//
//

//TIBGS
type TIBGSMasterPublicKey struct{
	g, g2, h1, u0, u1, u2, u3, u4, n *pbc.Element //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
}

type TIBGSMasterPublicKeyBytes struct{ //bytes
	G, G2, H1, U0, U1, U2, U3, U4 []byte //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
	Bn []byte
}

type TIBGSMasterSecretKeyi struct{
	h2i *pbc.Element //G2上的一个点
}

type Sharealphar struct{
	alphai,ri *pbc.Element //zr上的点
}

type SharealpharBytes struct{ //bytes
	Alphai []byte
	Ri []byte
}

type TIBGSGroupSecretKeyi struct{
	a0i, a2i, a3i, a4i, a5i *pbc.Element //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点 
}

type TIBGSGroupVerifyKeyi struct{
	gai *pbc.Element //G1上的点
}

type TIBGSGroupVerifyKeyiBytes struct{ //bytes
	Gai []byte //G1上的点
}

type TIBGSUserSecretKey struct{
	b0, b3, b4, b5 *pbc.Element //b0, b3, b4是G2上的点 b5是G1上的点
}

type TIBGSUserSecretKeyBytes struct{ //bytes
	B0, B3, B4, B5 []byte //b0, b3, b4是G2上的点 b5是G1上的点
}

type TIBGSOK struct{
	ok1, ok2 *pbc.Element //GT上的点
}

type TIBGSOKBytes struct{//bytes
	Ok1, Ok2 []byte //GT上的点
}

type TIBGSPOK struct{
	c, s1, s2, s3 *pbc.Element //Zr上的数
}

type TIBGSPOKBytes struct{
	C, S1, S2, S3 []byte //Zr上的数
}

type TIBGSSIG struct{
	c0, c5, c6, e1, e2, e3 *pbc.Element //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	pok *TIBGSPOK
}

type TIBGSSIGBytes struct{
	C0, C5, C6, E1, E2, E3 []byte //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	Pok *TIBGSPOKBytes
}



//TIBGS



var gstatic = []byte{86,128,202,65,129,43,121,169,37,55,68,20,250,195,252,225,59,245,228,150,17,53,79,113,124,91,189,249,157,250,124,10,195,133,15,124,165,45,119,11,104,244,204,204,106,189,172,17,84,198,17,53,172,21,89,170,56,164,225,154,195,127,29,73,142,57,28,134,143,128,214,189,149,72,108,147,24,194,17,242,231,21,233,253,119,7,247,91,166,162,172,232,29,242,138,187,86,227,209,146,88,36,204,155,247,34,156,167,98,171,98,208,52,25,48,160,119,57,0,3,164,67,89,130,85,246,255,244}
var g2static = []byte{60,153,28,98,67,154,113,16,117,120,181,29,234,86,207,225,171,23,134,192,22,236,107,212,207,93,43,26,192,208,27,169,114,115,10,179,118,25,66,55,169,194,205,38,172,140,141,39,71,152,26,200,97,11,17,73,106,27,24,138,199,11,13,157,24,78,211,63,217,221,169,202,207,134,184,36,121,94,75,145,204,104,72,229,68,158,10,151,94,29,159,76,140,7,35,254,148,5,251,240,18,141,113,0,39,129,152,167,129,51,140,213,204,239,79,171,165,193,71,210,80,166,165,94,228,154,1,133}
// var h1static = []byte{146,57,123,198,171,171,126,151,125,8,13,215,238,187,71,141,249,7,64,116,83,56,91,173,170,130,165,218,172,39,209,219,250,166,42,140,38,207,89,212,48,211,13,245,124,58,172,235,91,156,173,68,96,113,115,180,235,196,254,17,138,16,113,113,66,72,203,192,12,28,214,217,129,114,38,255,84,212,1,187,85,13,188,67,4,239,136,123,40,20,201,41,113,11,157,90,29,205,123,217,166,51,172,186,37,13,150,206,138,175,40,107,61,224,116,43,146,232,81,120,254,147,7,169,115,166,164,198}
var u0static = []byte{33,198,191,118,110,0,171,144,76,112,162,224,84,35,127,76,146,190,225,228,173,50,207,50,120,113,235,235,192,248,154,89,196,228,107,53,205,219,52,202,35,232,53,74,89,82,17,12,108,107,147,62,1,17,201,175,199,165,119,121,169,76,212,205,84,111,199,136,65,224,130,60,220,165,77,130,73,80,191,109,220,233,7,250,16,27,228,63,123,153,66,119,43,58,103,38,157,4,90,134,49,179,239,169,78,203,153,167,120,194,212,208,236,159,160,57,69,1,247,87,204,40,3,191,195,98,197,127}
var u1static = []byte{134,101,209,13,93,59,143,183,33,192,71,144,93,157,215,69,197,98,22,43,32,228,37,169,69,117,239,123,38,116,239,255,232,129,96,84,109,131,221,34,135,186,250,68,199,225,56,181,32,238,251,89,64,58,66,184,152,163,240,35,185,78,29,255,32,90,21,17,195,116,184,70,26,196,160,218,50,119,19,110,156,104,214,222,37,31,238,43,124,162,210,176,189,53,236,222,163,189,3,37,85,18,42,26,134,118,85,236,155,55,154,139,242,159,156,245,46,68,210,6,35,204,88,128,127,97,217,26}
var u2static = []byte{132,123,240,68,4,50,183,140,79,162,119,200,178,44,170,12,214,31,15,207,162,144,159,147,111,192,215,127,42,1,218,125,56,201,81,114,243,53,109,70,36,161,56,160,203,21,135,252,245,108,244,95,94,79,76,106,100,149,1,237,58,71,253,70,164,190,33,1,243,227,229,198,55,19,81,66,246,66,53,40,192,199,192,4,29,219,149,90,5,19,239,165,109,166,171,88,37,220,9,160,236,149,35,118,80,220,153,113,207,155,62,134,148,34,202,143,79,208,23,173,4,142,92,187,175,105,132,253}
var u3static = []byte{92,248,171,53,226,107,27,134,195,216,221,30,133,227,150,106,38,155,71,183,235,65,132,59,180,39,181,35,159,197,135,197,228,162,31,49,169,135,121,51,234,187,99,197,88,111,222,242,120,152,56,56,230,125,126,15,102,152,29,214,65,5,138,105,67,217,43,156,241,192,12,245,103,190,53,29,103,11,92,167,69,137,122,186,110,217,190,2,235,167,31,117,37,131,222,104,236,102,170,190,187,226,107,63,109,45,115,202,2,140,19,173,8,96,42,46,46,68,133,69,171,164,181,224,9,67,54,173}
var u4static = []byte{37,130,102,158,66,112,163,71,3,23,88,31,182,102,95,23,156,134,56,37,25,13,131,215,149,156,90,101,63,221,149,171,92,189,132,122,193,229,186,179,4,180,202,90,229,43,223,179,138,138,238,222,162,24,231,172,185,3,93,11,96,207,8,209,83,57,29,157,19,111,215,133,130,195,212,144,101,215,24,115,131,208,109,93,37,194,198,111,11,207,14,2,158,162,88,19,212,175,20,118,243,25,235,177,151,213,147,242,32,70,202,163,112,237,175,219,77,235,147,220,38,39,121,113,219,2,203,26}
var nstatic = []byte{93,122,146,183,222,144,80,111,128,201,250,100,43,213,102,147,47,168,91,130,155,17,166,2,189,212,173,9,121,59,157,110,10,44,123,170,130,53,80,159,221,208,65,185,167,197,154,135,40,99,194,165,163,207,157,124,11,37,232,186,75,135,111,194,153,80,154,234,8,167,95,162,109,225,40,97,30,113,253,88,84,5,44,107,184,130,145,10,111,246,147,115,230,167,0,99,102,157,58,50,183,98,51,202,101,219,66,17,204,98,80,225,177,146,118,198,160,98,135,232,167,249,145,162,41,108,224,93}

// C returns the result of f(x) = C1*x+C2*x^2+...+Ctk-1*x^(tk-1)
func C(Cs []*pbc.Element, index int) *pbc.Element {
	x := pairing.NewZr().SetInt32(int32(index))
	ret := pairing.NewZr().Set0()
	for _, coef := range Cs {
		ret.Add(ret, coef)
		ret.MulZn(ret, x)
	}
	return ret
}


// GenCoef generates the coefficients used by func C
func NewGenCoef(t int) []*pbc.Element {
	Cs := make([]*pbc.Element, t-1)
	for i := 0; i < t-1; i++ {
		Cs[i] = pairing.NewZr().Rand()
	}
	return Cs
}

// L generates the Lagrange coefficient of an index   L(i)  i=indexOfArray+1<=k
func L(K, indexOfArray int) *pbc.Element { // indexOfArray represents index of SelectedNodes
	if indexOfArray>K-1{
		return nil
	}
	L := pairing.NewZr().Set1()
	I := pairing.NewZr().SetInt32(int32(indexOfArray +1))  //I <= K  indexofarray从0开始 I从1开始
	for j:=0;j<K;j++{
		J := pairing.NewZr().SetInt32(int32(j+1)) //J从1开始
		if(J.Equals(I)){
			continue
		}else{
			// fmt.Println("J,I",J,I)
			temp1 := pairing.NewZr().Sub(J, I) // 2-1  3-1
			temp2 := pairing.NewZr().Div(J, temp1) // 2/2-1 3/3-1
			// fmt.Println("Li",L,"temp2",temp2)
			L.Mul(L, temp2) 
			// fmt.Println("L", L)
		}
	}
	// fmt.Println("L is",L)
	return L
}

// PIz returns the product of Zr inputs
func PIz(vals []*pbc.Element) *pbc.Element {
	accum := pairing.NewZr().Set1()
	for _, v := range vals {
		if v != nil {
			accum.MulZn(accum, v)
		}
	}
	return accum
}
func PIg(vals []*pbc.Element) *pbc.Element {
	accum := pairing.NewG1().Set1()
	for _, v := range vals {
		if v != nil {
			accum.Mul(accum, v)
		}
	}
	return accum
}

func (key *Sharealphar) GSShadowToBytes() *SharealpharBytes { //1
	bytes := &SharealpharBytes{}
	// bytes.alphai = key.alphai.Bytes()
	// bytes.ri = key.ri.Bytes()	
	bytes.Sset(key.alphai.Bytes(), key.ri.Bytes())

	return bytes
}

func (bytes *SharealpharBytes) Sset(alphai, ri []byte){
	bytes.Alphai = alphai
	bytes.Ri = ri
}

func (key *TIBGSMasterPublicKey) GSmpkToBytes() *TIBGSMasterPublicKeyBytes { //1
	bytes := &TIBGSMasterPublicKeyBytes{}
	// copy(bytes.g, key.g.CompressedBytes()) //bu copu
	// copy(bytes.g2 , key.g2.CompressedBytes())
	// copy(bytes.h1 , key.h1.CompressedBytes())
	// copy(bytes.u0 , key.u0.CompressedBytes())
	// copy(bytes.u1 , key.u1.CompressedBytes())
	// copy(bytes.u2 , key.u2.CompressedBytes())
	// copy(bytes.u3 , key.u3.CompressedBytes())
	// copy(bytes.u4 , key.u4.CompressedBytes())
	// copy(bytes.bn , key.n.Bytes())

	// bytes.g = key.g.CompressedBytes()
	// bytes.g2 = key.g2.CompressedBytes()
	// bytes.h1 = key.h1.CompressedBytes()
	// bytes.u0 = key.u0.CompressedBytes()
	// bytes.u1 = key.u1.CompressedBytes()
	// bytes.u2 = key.u2.CompressedBytes()
	// bytes.u3 = key.u3.CompressedBytes()
	// bytes.u4 = key.u4.CompressedBytes()
	// bytes.bn = key.n.Bytes()
	bytes.Sset( key.g.CompressedBytes(),key.g2.CompressedBytes(),key.h1.CompressedBytes(),key.u0.CompressedBytes(),key.u1.CompressedBytes(),key.u2.CompressedBytes(),key.u3.CompressedBytes(),key.u4.CompressedBytes(),key.n.Bytes())
	return bytes
}

func (bytes *TIBGSMasterPublicKeyBytes) Sset(g, g2, h1, u0, u1, u2, u3, u4, bn []byte){
	bytes.G = g
	bytes.G2 = g2
	bytes.H1 = h1
	bytes.U0 = u0
	bytes.U1 = u1
	bytes.U2 = u2
	bytes.U3 = u3
	bytes.U4 = u4
	bytes.Bn = bn
}

func (key *TIBGSUserSecretKey) GSUSKToBytes() *TIBGSUserSecretKeyBytes { //2
	bytes := &TIBGSUserSecretKeyBytes{}
	bytes.B0 = key.b0.CompressedBytes()
	bytes.B3 = key.b3.CompressedBytes()
	bytes.B4 = key.b4.CompressedBytes()
	bytes.B5 = key.b5.CompressedBytes()
	return bytes
}

func (key *TIBGSGroupVerifyKeyi) GSGVKiToBytes() *TIBGSGroupVerifyKeyiBytes { //2
	bytes := &TIBGSGroupVerifyKeyiBytes{}
	bytes.Gai = key.gai.CompressedBytes()
	return bytes
}

func(key *TIBGSPOK) GSPOKTOBytes() *TIBGSPOKBytes{
	bytes := &TIBGSPOKBytes{}
	bytes.C = key.c.Bytes() //GT Zr上的元素不能CompressedBytes
	bytes.S1 = key.s1.Bytes()
	bytes.S2 = key.s2.Bytes()
	bytes.S3 = key.s3.Bytes()
	return bytes
}

func(key *TIBGSSIG) GSSIGTOBytes() *TIBGSSIGBytes{
	bytes := &TIBGSSIGBytes{}
	bytes.C0 = key.c0.CompressedBytes()
	bytes.C5 = key.c5.CompressedBytes()
	bytes.C6 = key.c6.CompressedBytes()
	bytes.E1 = key.e1.CompressedBytes()
	bytes.E2 = key.e2.CompressedBytes()
	bytes.E3 = key.e3.Bytes() //GT Zr上的元素不能CompressedBytesbytes.e3 = key.e3.CompressedBytes()
	bytes.Pok = key.pok.GSPOKTOBytes()
	return bytes
}

type GSCompressedSIGBytes struct {
	SIG []byte
}
func(sig *TIBGSSIG) GSSigToCompressedBytes() *GSCompressedSIGBytes{
	sigBytes := sig.GSSIGTOBytes()
	CsigBytes := &GSCompressedSIGBytes{}
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.C0...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.C5...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.C6...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.E1...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.E2...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.E3...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.Pok.C...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.Pok.S1...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.Pok.S2...)
	CsigBytes.SIG = append(CsigBytes.SIG, sigBytes.Pok.S3...)
	return CsigBytes
}
func(key *TIBGSOK) GSOKTObytes() *TIBGSOKBytes{
	bytes := &TIBGSOKBytes{}
	bytes.Ok1 = key.ok1.Bytes() //GT 上的元素不能CompressedBytes
	bytes.Ok2 = key.ok2.Bytes()
	return bytes
}
///////////////////////////////////bytes->
func (bytes *TIBGSMasterPublicKeyBytes) BytesToGSmpk() *TIBGSMasterPublicKey {
	key := &TIBGSMasterPublicKey{} //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
	key.g, key.g2, key.h1, key.u0, key.u1, key.u2, key.u3, key.u4, key.n = pairing.NewG1(), pairing.NewG2(), pairing.NewG1(), pairing.NewG2(), pairing.NewG2(), pairing.NewG2(), pairing.NewG2(), pairing.NewG2(), pairing.NewGT()
	key.g.SetCompressedBytes(bytes.G)
	key.g2.SetCompressedBytes(bytes.G2)
	key.h1.SetCompressedBytes(bytes.H1)
	key.u0.SetCompressedBytes(bytes.U0)
	key.u1.SetCompressedBytes(bytes.U1)
	key.u2.SetCompressedBytes(bytes.U2)
	key.u3.SetCompressedBytes(bytes.U3)
	key.u4.SetCompressedBytes(bytes.U4)
	key.n.SetBytes(bytes.Bn)
	return key
}

func (bytes *SharealpharBytes) BytesToGSShadow() *Sharealphar {
	key := &Sharealphar{}
	key.alphai, key.ri = pairing.NewZr(), pairing.NewZr()
	key.alphai.SetBytes(bytes.Alphai)
	key.ri.SetBytes(bytes.Ri)
	return key
}

func (bytes *TIBGSUserSecretKeyBytes) BytesToGSUSK() *TIBGSUserSecretKey {
	key := &TIBGSUserSecretKey{}
	key.b0, key.b3, key.b4, key.b5 = pairing.NewG2(), pairing.NewG2(), pairing.NewG2(), pairing.NewG1()
	key.b0.SetCompressedBytes(bytes.B0)
	key.b3.SetCompressedBytes(bytes.B3)
	key.b4.SetCompressedBytes(bytes.B4)
	key.b5.SetCompressedBytes(bytes.B5)
	return key
}

func (bytes *TIBGSGroupVerifyKeyiBytes) BytesToGSGVKi() *TIBGSGroupVerifyKeyi {
	key := &TIBGSGroupVerifyKeyi{}
	key.gai = pairing.NewG1()
	key.gai.SetCompressedBytes(bytes.Gai)
	return key
}

func (bytes *TIBGSPOKBytes) BytesToGSPOK() *TIBGSPOK{
	key := &TIBGSPOK{}
	key.c, key.s1, key.s2, key.s3 = pairing.NewZr(), pairing.NewZr(), pairing.NewZr(), pairing.NewZr()
	key.c.SetBytes(bytes.C)
	key.s1.SetBytes(bytes.S1)
	key.s2.SetBytes(bytes.S2)
	key.s3.SetBytes(bytes.S3)
	return key
}

func (bytes *TIBGSSIGBytes) BytesToGSSig() *TIBGSSIG{
	key := &TIBGSSIG{}
	key.c0, key.c5, key.c6, key.e1, key.e2, key.e3 = pairing.NewG2(), pairing.NewG1(), pairing.NewG2(), pairing.NewG1(), pairing.NewG2(), pairing.NewGT()
	key.c0.SetCompressedBytes(bytes.C0)
	key.c5.SetCompressedBytes(bytes.C5)
	key.c6.SetCompressedBytes(bytes.C6)
	key.e1.SetCompressedBytes(bytes.E1)
	key.e2.SetCompressedBytes(bytes.E2)
	key.e3.SetBytes(bytes.E3)
	key.pok = bytes.Pok.BytesToGSPOK()
	return key
}

func (csig *GSCompressedSIGBytes) GSCompressedBytesToSig() *TIBGSSIG{
	c0b, c5b, c6b, e1b, e2b, e3b, pokcb, poks1b, poks2b, poks3b := csig.SIG[0:65], csig.SIG[65:130], csig.SIG[130:195], csig.SIG[195:260], csig.SIG[260:325], csig.SIG[325:453], csig.SIG[453:473], csig.SIG[473:493], csig.SIG[493:513], csig.SIG[513:533]
	c0, c5, c6, e1, e2, e3 := pairing.NewG2(), pairing.NewG1(), pairing.NewG2(), pairing.NewG1(), pairing.NewG2(), pairing.NewGT()
	c, s1, s2, s3 := pairing.NewZr(), pairing.NewZr(), pairing.NewZr(), pairing.NewZr()
	c0.SetCompressedBytes(c0b)
	c5.SetCompressedBytes(c5b)
	c6.SetCompressedBytes(c6b)
	e1.SetCompressedBytes(e1b)
	e2.SetCompressedBytes(e2b)
	e3.SetBytes(e3b)
	c.SetBytes(pokcb)
	s1.SetBytes(poks1b)
	s2.SetBytes(poks2b)
	s3.SetBytes(poks3b)
	pok := &TIBGSPOK{c,s1,s2,s3}
	ssig := &TIBGSSIG{c0, c5, c6, e1, e2, e3, pok}
	return ssig
}

func (bytes *TIBGSOKBytes) BytesToGSOK() *TIBGSOK{
	key := &TIBGSOK{}
	key.ok1, key.ok2 = pairing.NewGT(),pairing.NewGT()
	key.ok1.SetBytes(bytes.Ok1)
	key.ok2.SetBytes(bytes.Ok2)
	return key
}
//实现方便，单节点生成
func NewSetup(numOfLevel0, thresholdOfLevel0 int, grpID string) (*TIBGSMasterPublicKey, []*Sharealphar, error){
	var mpk TIBGSMasterPublicKey //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
	mpk.g = pairing.NewG1().SetBytes(gstatic)
	mpk.g2 = pairing.NewG2().SetBytes(g2static)
	// mpk.h1 = pairing.NewG1().SetXBytes(h1static)
	mpk.u0 = pairing.NewG2().SetBytes(u0static)
	mpk.u1 = pairing.NewG2().SetBytes(u1static)
	mpk.u2 = pairing.NewG2().SetBytes(u2static)
	mpk.u3 = pairing.NewG2().SetBytes(u3static)
	mpk.u4 = pairing.NewG2().SetBytes(u4static)
	mpk.n = pairing.NewGT().SetBytes(nstatic)
	//随机生成alpha和r
	alpha := pairing.NewZr().Rand()
	// alpha := pairing.NewZr().SetInt32(int32(10))
	r := pairing.NewZr().Rand()

	mpk.h1 = pairing.NewG1().PowZn(mpk.g, alpha)

	poly1 := NewGenCoef(thresholdOfLevel0) //产生多项式系数 k-1个？？
	nalpha := make([]*pbc.Element, numOfLevel0) // s存储n个分享 1-n
	for index := 1; index <= numOfLevel0; index++ { // C returns the result of f(x) = C1*x+C2*x^2+...+Ctk-1*x^(tk-1) 
		nalpha[index-1] = pairing.NewZr().Add(C(poly1, index), alpha) //s[i-1] = ploy(i) + alpha
	}
	nr := make([]*pbc.Element, numOfLevel0) // s存储n个分享 1-n
	poly2 := NewGenCoef(thresholdOfLevel0) //产生多项式系数 k-1个？？
	for index := 1; index <= numOfLevel0; index++ { // C returns the result of f(x) = C1*x+C2*x^2+...+Ctk-1*x^(tk-1) 
		nr[index-1] = pairing.NewZr().Add(C(poly2, index), r) //s[i-1] = ploy(i) + alpha
	}
	nss := make([]*Sharealphar, numOfLevel0)
	for i := 0; i < numOfLevel0; i++ {
		var tmpSs Sharealphar
		tmpSs.alphai = nalpha[i] //g2^s[i]
		tmpSs.ri = nr[i] //1
		nss[i] = &tmpSs
	}
	return &mpk, nss, nil
}
//为其他n-1节点生成mski,gski,gvki
func Gen3key(mpk *TIBGSMasterPublicKey, ar *Sharealphar, grpID string) (*TIBGSMasterSecretKeyi, *TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi){
	falphai := pairing.NewZr().Set(ar.alphai)
	fri := pairing.NewZr().Set(ar.ri)

	//门限 计算mski
	var mski TIBGSMasterSecretKeyi  //G2上的一个点
	//通过门限求出alphai n
	mski.h2i = pairing.NewG2().PowZn(mpk.g2, falphai) //g2^alphai

	I := G(grpID)
	// fmt.Println("G(grpID)",G(grpID))
	var gski TIBGSGroupSecretKeyi //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点 
	//秘密分享ri n
	a0 := pairing.NewG2().PowZn(mpk.u1, I) //u1^G(grpID)
	a0.Mul(mpk.u0, a0)  //u0 * u1^G(grpID)
	a0.PowZn(a0, fri) //(u0 * u1^G(grpID))^ri
	gski.a0i = pairing.NewG2().Mul(mski.h2i, a0) // h2i*(u0 * u1^G(grpID))^
	// fmt.Println("gski.a0i",gski.a0i)
	gski.a2i = pairing.NewG2().PowZn(mpk.u2, fri)
	gski.a3i = pairing.NewG2().PowZn(mpk.u3, fri)
	gski.a4i = pairing.NewG2().PowZn(mpk.u4, fri)
	gski.a5i = pairing.NewG1().PowZn(mpk.g, fri)

	var gvki TIBGSGroupVerifyKeyi //G1上的点
	gvki.gai = pairing.NewG1().PowZn(mpk.g, falphai)
	return &mski, &gski, &gvki
}



//G returns the hash of IDs in an ID string
func G(ID string) *pbc.Element {
	I :=pairing.NewZr().SetFromStringHash(ID,sha256.New())
	return I
}

// func GrpSetUp(mpk *TIBGSMasterPublicKey, mski *TIBGSMasterSecretKeyi,  grpID string, ri, alphai *pbc.Element)(*TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi){
// 	I := G(grpID)
// 	// fmt.Println("G(grpID)",G(grpID))
// 	var gski TIBGSGroupSecretKeyi //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点 
// 	//秘密分享ri n
// 	a0 := pairing.NewG2().PowZn(mpk.u1, I) //u1^G(grpID)
// 	a0.Mul(mpk.u0, a0)  //u0 * u1^G(grpID)
// 	a0.PowZn(a0, ri) //(u0 * u1^G(grpID))^ri
// 	gski.a0i = pairing.NewG2().Mul(mski.h2i, a0) // h2i*(u0 * u1^G(grpID))^
// 	// fmt.Println("gski.a0i",gski.a0i)
// 	gski.a2i = pairing.NewG2().PowZn(mpk.u2, ri)
// 	gski.a3i = pairing.NewG2().PowZn(mpk.u3, ri)
// 	gski.a4i = pairing.NewG2().PowZn(mpk.u4, ri)
// 	gski.a5i = pairing.NewG1().PowZn(mpk.g, ri)

// 	var gvki TIBGSGroupVerifyKeyi //G1上的点
// 	gvki.gai = pairing.NewG1().PowZn(mpk.g, alphai)
// 	return &gski, &gvki
// }

func ExtShare(gski *TIBGSGroupSecretKeyi,userID string) *TIBGSUserSecretKey{
	I := G(userID)
	var uski TIBGSUserSecretKey //b0, b3, b4是G2上的点 b5是G1上的点
	mid := pairing.NewG2().PowZn(gski.a2i, I) //a2i^G(userID)
	uski.b0 = pairing.NewG2().Mul(gski.a0i, mid) //a0i * a2i^G(userID)
	uski.b3 = gski.a3i
	uski.b4 = gski.a4i
	uski.b5 = gski.a5i
	return &uski
}

func VerifyShare(uski *TIBGSUserSecretKey, gvki *TIBGSGroupVerifyKeyi, mpk *TIBGSMasterPublicKey, grpID string, userID string) bool{
	IG := G(grpID)
	IU := G(userID)
	
	left1 := pairing.NewGT().Pair(mpk.g , uski.b0) //e(g , b0)
	right11 := pairing.NewGT().Pair(gvki.gai, mpk.g2) //e(gvk , g2)
	right12 := pairing.NewGT().Pair(uski.b5, mpk.u0) //e(b5, u0)
	u1grpID := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	right13 := pairing.NewGT().Pair(uski.b5, u1grpID) //e(b5,u1^G(grpID))
	u2userID := pairing.NewG2().PowZn(mpk.u2, IU) //u2^G(userID)
	right14 := pairing.NewGT().Pair(uski.b5, u2userID) //e(b5, u2^G(userID))
	right1 := pairing.NewGT().Mul(right11, right12) // e(gvk , g2) e(b5, u0)
	right1.Mul(right1, right13) //e(gvk , g2) e(b5, u0) e(b5,u1^G(grpID))
	right1.Mul(right1, right14) //e(gvk , g2) e(b5, u0) e(b5,u1^G(grpID)) e(b5, u2^G(userID))
	left2 := pairing.NewGT().Pair(mpk.g , uski.b3)  //e(g, b3)
	right2 := pairing.NewGT().Pair(uski.b5, mpk.u3) //e(b5, u3)
	left3 := pairing.NewGT().Pair(mpk.g, uski.b4) //e(g, b4)
	right3 := pairing.NewGT().Pair(uski.b5, mpk.u4) //e(b5, u4)
	return left1.Equals(right1) && left2.Equals(right2) && left3.Equals(right3)
}

func ReconstKey(uskis []*TIBGSUserSecretKey,mpk *TIBGSMasterPublicKey, K int, grpID string, userID string) *TIBGSUserSecretKey{
	IG := G(grpID)
	IU := G(userID)
	var usk TIBGSUserSecretKey //b0, b3, b4是G2上的点 b5是G1上的点
	// KL := make([]*pbc.Element, K)
	var b01,b31,b41,b51 *pbc.Element
	mb0,mb3,mb4,mb5:=pairing.NewG2().Set1(),pairing.NewG2().Set1(),pairing.NewG2().Set1(),pairing.NewG1().Set1()
	
	for i:=0;i<K;i++{
		KL:=L(K,i) //i从0开始  在L里会+1
		b01=pairing.NewG2().PowZn(uskis[i].b0, KL)
		mb0.Mul(mb0,b01)
		b31=pairing.NewG2().PowZn(uskis[i].b3, KL)
		mb3.Mul(mb3,b31)
		b41=pairing.NewG2().PowZn(uskis[i].b4, KL)
		mb4.Mul(mb4,b41)
		b51=pairing.NewG1().PowZn(uskis[i].b5, KL)
		mb5.Mul(mb5,b51)
	}
	b02 := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	b02.Mul(mpk.u0, b02)  //u0 * u1^G(grpID)
	b03 := pairing.NewG2().PowZn(mpk.u2, IU)//u2^G(UserID)
	b02.Mul(b02,b03) //u0 * u1^G(grpID)*u2^G(UserID)
	r2:= pairing.NewZr().Rand() //随机
	b02.PowZn(b02, r2) //(u0 * u1^G(grpID)*u2^G(UserID))^r2
	usk.b0 = pairing.NewG2().Mul(mb0,b02)
	usk.b3 = pairing.NewG2().PowZn(mpk.u3, r2) //u3^r2
	usk.b3.Mul(mb3, usk.b3)
	usk.b4 = pairing.NewG2().PowZn(mpk.u4, r2)
	usk.b4.Mul(mb4, usk.b4)
	usk.b5 = pairing.NewG1().PowZn(mpk.g, r2)
	usk.b5.Mul(mb5, usk.b5)
	return &usk
}

func Sign(mpk *TIBGSMasterPublicKey,usk *TIBGSUserSecretKey, message string, grpID string, userID string ) *TIBGSSIG{
	msg := G(message)
	IG := G(grpID)
	IU := G(userID)
	var ssig TIBGSSIG //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	var POK TIBGSPOK //c, s1, s2, s3 Zr上的数  
	r3, f, rID := pairing.NewZr().Rand(), pairing.NewZr().Rand(), pairing.NewZr().Rand()
	c02 := pairing.NewG2().PowZn(usk.b3, msg) //b3^m
	c03 := pairing.NewG2().PowZn(usk.b4, rID) //b4^G(rID)
	c04 := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	c04.Mul(mpk.u0, c04) //u0 * u1^G(grpID)
	c043 := pairing.NewG2().PowZn(mpk.u2, IU) //u2^G(userID) 
	c04.Mul(c04, c043) //u0 * u1^G(grpID)*u2^G(userID)
	c044 := pairing.NewG2().PowZn(mpk.u3, msg) //u3^m
	c04.Mul(c04, c044) //u0 * u1^G(grpID)*u2^G(userID) * u3^m
	c045 := pairing.NewG2().PowZn(mpk.u4, rID) // u4^G(rID)
	c04.Mul(c04, c045) //u0 * u1^G(grpID)*u2^G(userID) * u3^m * u4^G(rID)
	c04.PowZn(c04, r3) //(u0 * u1^G(grpID) * u2^G(userID) * u3^m * u4^G(rID))^r3	
	ssig.c0 = pairing.NewG2().Mul(usk.b0, c02) // b0 * b3^m 
	ssig.c0.Mul(ssig.c0, c03) //b0 * b3^m * b4^G(rID)
	ssig.c0.Mul(ssig.c0, c04) // b0 * b3^m * b4^G(rID) * (u0 * u1^G(grpID) * u2^G(userID) * u3^m * u4^G(rID))^r3
	ssig.c5 = pairing.NewG1().PowZn(mpk.g, r3) //g^r3
	ssig.c5.Mul(usk.b5 , ssig.c5) //b5 * g^r3
	c61 := pairing.NewG2().PowZn(mpk.u2, IU) //u2^G(userID)
	c62 := pairing.NewG2().PowZn(mpk.u4, rID) //u4^G(rID)
	ssig.c6 = pairing.NewG2().Mul(c61 , c62) // u2^G(userID) * u4^G(rID)
	ssig.e1 = pairing.NewG1().PowZn(mpk.g, f) //g^f
	ssig.e2 = pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	ssig.e2.Mul(mpk.u0, ssig.e2) //u0 * u1^G(grpID)
	ssig.e2.PowZn(ssig.e2, f) //(u0 * u1^G(grpID))^f
	e31 := pairing.NewGT().PowZn(mpk.n, IU) //n^G(userID)
	e32 := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1, g2)
	e32.PowZn(e32, f) // e(h1, g2)^f
	ssig.e3 = pairing.NewGT().Mul(e31, e32) //n^G(userID) * e(h1, g2)^f
	//生成POK
	k1, k2, k3 := pairing.NewZr().Rand(), pairing.NewZr().Rand(), pairing.NewZr().Rand()
	hatf := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	hatf.Mul(mpk.u0, hatf) //u0*u1^G(grpID)
	bigg := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1,g2)
	pr11 := pairing.NewG2().PowZn(mpk.u2, k1) //u2^k1
	pr12 := pairing.NewG2().PowZn(mpk.u4, k2) //u4^k2
	pr1 := pairing.NewG2().Mul(pr11, pr12) //u2^k1 * u4^k2
	pr2 := pairing.NewG1().PowZn(mpk.g,k3) //g^k3
	pr3 := pairing.NewG2().PowZn(hatf, k3) //hatf^k3
	pt41 := pairing.NewGT().PowZn(mpk.n, k1) //n^k1
	pt42 := pairing.NewGT().PowZn(bigg, k3) //bigg^k3
	pr4 := pairing.NewGT().Mul(pt41,pt42) //n^k1* bigg^k3
	
	POK.c = pairing.NewZr().SetFromStringHash(pr1.String()+pr2.String()+pr3.String()+pr4.String(), sha256.New())
	cx := pairing.NewZr().Mul(POK.c, IU) //c*IU
	POK.s1 = pairing.NewZr().Add(k1, cx) //k1 + c*IU
	cy := pairing.NewZr().Mul(POK.c, rID) //c*rID
	POK.s2 = pairing.NewZr().Add(k2, cy) //k2+c*rID
	cz := pairing.NewZr().Mul(POK.c, f)
	POK.s3 = pairing.NewZr().Add(k3, cz)
	ssig.pok = &POK
	return &ssig
}

func NewSign(mpk *TIBGSMasterPublicKey,usk *TIBGSUserSecretKey, message []byte, grpID string, userID string ) *GSCompressedSIGBytes{
	msg := string(message)
	sig := Sign(mpk ,usk , msg , grpID , userID)
	ssig := sig.GSSigToCompressedBytes()
	return ssig
}

func Verify(ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey, message string, grpID string) bool{
	msg := G(message)
	IG := G(grpID)	
	
	t := pairing.NewZr().Rand()
	M := pairing.NewGT().Rand()

	d1 := pairing.NewG1().PowZn(mpk.g, t) //g^t
	d22 := pairing.NewG2().PowZn(mpk.u1, IG) // u1^G(grpID)
	d23 := pairing.NewG2().PowZn(mpk.u3, msg) //u3^m
	d2 := pairing.NewG2().Mul(mpk.u0, d22) //u0 * u1^G(grpID)
	d2.Mul(d2, d23) //u0 * u1^G(grpID) * u3^m
	d2.Mul(d2, ssig.c6) //u0 * u1^G(grpID) * u3^m *c6
	d2.PowZn(d2, t) //(u0 * u1^G(grpID) * u3^m *c6)^t
	zeta := pairing.NewGT().Pair(mpk.h1, mpk.g2) //e(h1, g2)
	zeta.PowZn(zeta, t) //e(h1, g2)^t
	zeta.Mul(M, zeta) //M * e(h1, g2)^t
	temp1 := pairing.NewGT().Pair(ssig.c5, d2) //e(c5, d2)
	temp2 := pairing.NewGT().Pair(d1, ssig.c0) //e(d1, c0)
	right := pairing.NewGT().Div(temp1, temp2) // e(c5, d2) / e(d1, c0)
	right.Mul(zeta, right) // zeta * (e(c5, d2) / e(d1, c0))
	//验证POK
	rr11 := pairing.NewG2().PowZn(mpk.u2, ssig.pok.s1) //u2^s1
	rr12 := pairing.NewG2().PowZn(mpk.u4, ssig.pok.s2) //u4^s2
	negc := pairing.NewZr().Neg(ssig.pok.c) //-c
	rr13 := pairing.NewG2().PowZn(ssig.c6, negc) //c6^-c
	rr1:= pairing.NewG2().Mul(rr11,rr12) //u2^s1 * u4^s2
	rr1.Mul(rr1,rr13) //u2^s1 * u4^s2* c6^-c
	rr21 := pairing.NewG2().PowZn(mpk.g, ssig.pok.s3) //g^s3
	rr22 := pairing.NewG2().PowZn(ssig.e1, negc) //e1^-c
	rr2 := pairing.NewG2().Mul(rr21,rr22) //g^s3 * e1^-c
	hatf := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	hatf.Mul(mpk.u0, hatf) //u0*u1^G(grpID)
	bigg := pairing.NewGT().Pair(mpk.h1, mpk.g2)
	rr31 := pairing.NewG2().PowZn(hatf, ssig.pok.s3) //hatf^s3
	rr32 := pairing.NewG2().PowZn(ssig.e2, negc)//e2^-c
	rr3 := pairing.NewG2().Mul(rr31,rr32) //hatf^s3 * e2^-c
	tt41 := pairing.NewGT().PowZn(mpk.n, ssig.pok.s1)//n^s1
	tt42 := pairing.NewGT().PowZn(bigg, ssig.pok.s3) //bigg^s3
	tt43 := pairing.NewGT().PowZn(ssig.e3, negc) //e3^-c
	tt4:= pairing.NewGT().Mul(tt41,tt42)//n^s1* bigg^s3
	tt4.Mul(tt4, tt43) //n^s1* bigg^s3 *e3^-c
	cc := pairing.NewZr().SetFromStringHash(rr1.String()+rr2.String()+rr3.String()+tt4.String(), sha256.New())
	return M.Equals(right) && cc.Equals(ssig.pok.c)
}

func NewVerify(sig *GSCompressedSIGBytes, mpk *TIBGSMasterPublicKey, message []byte) bool{
	msg := string(message)
	ssig := sig.GSCompressedBytesToSig()
	return Verify(ssig , mpk , msg , "computer" )
}

func OpenPart(gski *TIBGSGroupSecretKeyi, ssig *TIBGSSIG) *TIBGSOK{
	var OKi TIBGSOK //ok1, ok2 GT上的点
	OKi.ok1 = pairing.NewGT().Pair(ssig.e1, gski.a0i)
	OKi.ok2 = pairing.NewGT().Pair(gski.a5i, ssig.e2)
	return &OKi
}

func NewOpenPart(gski *TIBGSGroupSecretKeyi, sig *GSCompressedSIGBytes) *TIBGSOKBytes{
	ssig:=sig.GSCompressedBytesToSig()
	oki := OpenPart(gski,ssig)
	return oki.GSOKTObytes()
}

func Open(OKK []*TIBGSOK, K int) *pbc.Element{
	temp1,temp2:= pairing.NewGT().Set1(),pairing.NewGT().Set1()
	var m1,m2 *pbc.Element
	for i:=0;i<K;i++{
		KL:=L(K,i) //i从0开始  在L里会+1
		m1=pairing.NewGT().PowZn(OKK[i].ok1, KL)
		temp1.Mul(temp1,m1)
		m2=pairing.NewGT().PowZn(OKK[i].ok2, KL)
		temp2.Mul(temp2,m2)
	}
	gama := pairing.NewGT().Div(temp1,temp2)
	return gama
}

func FindUser(UIDS []string, gama *pbc.Element,ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey)string{
	for _,ID := range UIDS{
		GU := G(ID)
		right:=pairing.NewGT().PowZn(mpk.n, GU)
		right.Mul(right,gama)
		if ssig.e3.Equals(right){
			return ID
		}
	}
	return ""
}
//TIBGS


