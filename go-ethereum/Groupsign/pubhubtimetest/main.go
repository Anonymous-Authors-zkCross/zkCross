package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/Nik-U/pbc"
	"strings"
	"time"

	//"unsafe"
	//"time"
	"os"
	//"bufio" //缓存IO
	//"fmt"
	//"io"
	//"io/ioutil" //io 工具包

	"encoding/hex"
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


var pairing, _ = pbc.NewPairingFromString(str)

type TIBGSMasterPublicKey struct{
	g, g2, h1, u0, u1, u2, u3, u4, n *pbc.Element //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
}

type TIBGSMasterSecretKeyi struct{
	h2i *pbc.Element //G2上的一个点
}

type Sharealphar struct{
	alphai,ri *pbc.Element
}

type TIBGSGroupSecretKeyi struct{
	a0i, a2i, a3i, a4i, a5i *pbc.Element //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点
}

type TIBGSGroupVerifyKeyi struct{
	gai *pbc.Element //G1上的点
}

type TIBGSUserSecretKey struct{
	b0, b3, b4, b5 *pbc.Element //b0, b3, b4是G2上的点 b5是G1上的点
}

type TIBGSOK struct{
	ok1, ok2 *pbc.Element //GT上的点
}

type TIBGSPOK struct{
	c, s1, s2, s3 *pbc.Element //Zr上的数
}

type TIBGSSIG struct{
	c0, c5, c6, e1, e2, e3 *pbc.Element //c5, e1是G1上的点  c0, c6, e2是G2上的点  e3是GT上的点
	pok TIBGSPOK
}

var gstatic = []byte{86,128,202,65,129,43,121,169,37,55,68,20,250,195,252,225,59,245,228,150,17,53,79,113,124,91,189,249,157,250,124,10,195,133,15,124,165,45,119,11,104,244,204,204,106,189,172,17,84,198,17,53,172,21,89,170,56,164,225,154,195,127,29,73,142,57,28,134,143,128,214,189,149,72,108,147,24,194,17,242,231,21,233,253,119,7,247,91,166,162,172,232,29,242,138,187,86,227,209,146,88,36,204,155,247,34,156,167,98,171,98,208,52,25,48,160,119,57,0,3,164,67,89,130,85,246,255,244}
var g2static = []byte{60,153,28,98,67,154,113,16,117,120,181,29,234,86,207,225,171,23,134,192,22,236,107,212,207,93,43,26,192,208,27,169,114,115,10,179,118,25,66,55,169,194,205,38,172,140,141,39,71,152,26,200,97,11,17,73,106,27,24,138,199,11,13,157,24,78,211,63,217,221,169,202,207,134,184,36,121,94,75,145,204,104,72,229,68,158,10,151,94,29,159,76,140,7,35,254,148,5,251,240,18,141,113,0,39,129,152,167,129,51,140,213,204,239,79,171,165,193,71,210,80,166,165,94,228,154,1,133}
// var h1static = []byte{146,57,123,198,171,171,126,151,125,8,13,215,238,187,71,141,249,7,64,116,83,56,91,173,170,130,165,218,172,39,209,219,250,166,42,140,38,207,89,212,48,211,13,245,124,58,172,235,91,156,173,68,96,113,115,180,235,196,254,17,138,16,113,113,66,72,203,192,12,28,214,217,129,114,38,255,84,212,1,187,85,13,188,67,4,239,136,123,40,20,201,41,113,11,157,90,29,205,123,217,166,51,172,186,37,13,150,206,138,175,40,107,61,224,116,43,146,232,81,120,254,147,7,169,115,166,164,198}
var u0static = []byte{33,198,191,118,110,0,171,144,76,112,162,224,84,35,127,76,146,190,225,228,173,50,207,50,120,113,235,235,192,248,154,89,196,228,107,53,205,219,52,202,35,232,53,74,89,82,17,12,108,107,147,62,1,17,201,175,199,165,119,121,169,76,212,205,84,111,199,136,65,224,130,60,220,165,77,130,73,80,191,109,220,233,7,250,16,27,228,63,123,153,66,119,43,58,103,38,157,4,90,134,49,179,239,169,78,203,153,167,120,194,212,208,236,159,160,57,69,1,247,87,204,40,3,191,195,98,197,127}
var u1static = []byte{134,101,209,13,93,59,143,183,33,192,71,144,93,157,215,69,197,98,22,43,32,228,37,169,69,117,239,123,38,116,239,255,232,129,96,84,109,131,221,34,135,186,250,68,199,225,56,181,32,238,251,89,64,58,66,184,152,163,240,35,185,78,29,255,32,90,21,17,195,116,184,70,26,196,160,218,50,119,19,110,156,104,214,222,37,31,238,43,124,162,210,176,189,53,236,222,163,189,3,37,85,18,42,26,134,118,85,236,155,55,154,139,242,159,156,245,46,68,210,6,35,204,88,128,127,97,217,26}
var u2static = []byte{132,123,240,68,4,50,183,140,79,162,119,200,178,44,170,12,214,31,15,207,162,144,159,147,111,192,215,127,42,1,218,125,56,201,81,114,243,53,109,70,36,161,56,160,203,21,135,252,245,108,244,95,94,79,76,106,100,149,1,237,58,71,253,70,164,190,33,1,243,227,229,198,55,19,81,66,246,66,53,40,192,199,192,4,29,219,149,90,5,19,239,165,109,166,171,88,37,220,9,160,236,149,35,118,80,220,153,113,207,155,62,134,148,34,202,143,79,208,23,173,4,142,92,187,175,105,132,253}
var u3static = []byte{92,248,171,53,226,107,27,134,195,216,221,30,133,227,150,106,38,155,71,183,235,65,132,59,180,39,181,35,159,197,135,197,228,162,31,49,169,135,121,51,234,187,99,197,88,111,222,242,120,152,56,56,230,125,126,15,102,152,29,214,65,5,138,105,67,217,43,156,241,192,12,245,103,190,53,29,103,11,92,167,69,137,122,186,110,217,190,2,235,167,31,117,37,131,222,104,236,102,170,190,187,226,107,63,109,45,115,202,2,140,19,173,8,96,42,46,46,68,133,69,171,164,181,224,9,67,54,173}
var u4static = []byte{37,130,102,158,66,112,163,71,3,23,88,31,182,102,95,23,156,134,56,37,25,13,131,215,149,156,90,101,63,221,149,171,92,189,132,122,193,229,186,179,4,180,202,90,229,43,223,179,138,138,238,222,162,24,231,172,185,3,93,11,96,207,8,209,83,57,29,157,19,111,215,133,130,195,212,144,101,215,24,115,131,208,109,93,37,194,198,111,11,207,14,2,158,162,88,19,212,175,20,118,243,25,235,177,151,213,147,242,32,70,202,163,112,237,175,219,77,235,147,220,38,39,121,113,219,2,203,26}
var nstatic = []byte{93,122,146,183,222,144,80,111,128,201,250,100,43,213,102,147,47,168,91,130,155,17,166,2,189,212,173,9,121,59,157,110,10,44,123,170,130,53,80,159,221,208,65,185,167,197,154,135,40,99,194,165,163,207,157,124,11,37,232,186,75,135,111,194,153,80,154,234,8,167,95,162,109,225,40,97,30,113,253,88,84,5,44,107,184,130,145,10,111,246,147,115,230,167,0,99,102,157,58,50,183,98,51,202,101,219,66,17,204,98,80,225,177,146,118,198,160,98,135,232,167,249,145,162,41,108,224,93}


func convert(array interface{}) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

func G(ID string) *pbc.Element {
	I :=pairing.NewZr().SetFromStringHash(ID,sha256.New())
	return I
}

func CN(Cs []*pbc.Element, N int) []*pbc.Element {
	NN := make([]*pbc.Element, N)
	// fmt.Println("NmanagerEXLEL", NN)

	for i := 0; i < N; i++{

		NN[i] = CNi(Cs,N,i)
	}
	// fmt.Println("Nmanager", NN)
	return NN
}

func CNi(Cs []*pbc.Element, N, i int) *pbc.Element{
	sum := pairing.NewZr().Set0()
	pre := int32(1)
	for _, coef := range Cs {
		temp1 := pairing.NewZr().MulInt32(coef, pre)
		sum.Add(sum, temp1)
		pre = pre*int32((i+1))
	}
	// fmt.Println("sum",sum, i)
	return sum
}


// GenCoef generates the coefficients used by func C
func GenCoef(szero *pbc.Element, t int) []*pbc.Element {
	Cs := make([]*pbc.Element, t)
	Cs[0] = szero
	for i := 1; i < t; i++ {//生成K-1个系数
		Cs[i] = pairing.NewZr().Rand()
	}
	// fmt.Println("Cs",Cs)
	return Cs
}
func SharesGen(thresholdOfLevel0 int, numOfLevel0 int, s *pbc.Element)([]*pbc.Element, []*pbc.Element){
	//t1:=time.Now()
	//每个manager生成 s,t
	// s := pairing.NewZr().Rand()
	// t := pairing.NewZr().Rand()
	//产生多项式系数  k-1次多项式 t个系数，第一个是要分享的
	sk := GenCoef(s, thresholdOfLevel0)
	// tk := GenCoef(t, thresholdOfLevel0)
	//产生N个管理员的分享

	SCN := CN(sk, numOfLevel0)
	// TCN := CN(tk, numOfLevel0)
	//t2:=time.Now()
	//f, _ := os.OpenFile("SharesGen.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return SCN, nil
}
//通过n个分享生成 alphai 和 ri
func ReconManager(ss []*pbc.Element, tt []*pbc.Element, N int)(*pbc.Element, *pbc.Element){
	//t1:=time.Now()
	if len(ss) == N && len(tt)==N{
		alphai := pairing.NewZr().Set0()
		ri := pairing.NewZr().Set0()
		for i := 0; i < N; i++{
			alphai.Add(alphai, ss[i])
			ri.Add(ri, tt[i])
		}
		//t2:=time.Now()
		//fmt.Println("time:ReconManager",t2.Sub(t1))
		//f, _ := os.OpenFile("ReconManager.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
		//fmt.Fprintln(f,t2.Sub(t1))
		return alphai,ri
	}
	return nil,nil
}

// L generates the Lagrange coefficient of an index   L(i)  i=indexOfArray+1<=k
func L(K, indexOfArray int) *pbc.Element { // indexOfArray represents index of SelectedNodes
	if indexOfArray>K-1{
		return nil
	}
	L := pairing.NewZr().Set1()
	I := pairing.NewZr().SetInt32(int32(indexOfArray +1))  //I <= K
	for j:=0;j<K;j++{
		J := pairing.NewZr().SetInt32(int32(j+1))
		if(J.Equals(I)){
			continue
		}else{
			temp1 := pairing.NewZr().Sub(J, I)
			temp2 := pairing.NewZr().Div(J, temp1)
			L.Mul(L, temp2)
			// fmt.Println("L", L)
		}
	}
	return L
}
//k个g^alphai 还原个g^alpha
func Galpha(SelectedNodes []*pbc.Element, K int ) *pbc.Element{
	if len(SelectedNodes)< K{
		fmt.Println("galphai not enough")
		return nil
	}
	h1 := pairing.NewG1().Set1()
	for i:=0;i<K;i++{
		temp := pairing.NewG1().PowZn(SelectedNodes[i], L(K,i))
		h1.Mul(h1, temp)
	}
	return h1
}
/*
type TestStructTobytes struct {
  data int64
}
type SliceMock struct {
  addr uintptr
  len int
  cap int
}

func main() {

  var testStruct = &TestStructTobytes{100}
  Len := unsafe.Sizeof(*testStruct)
  testBytes := &SliceMock{
    addr: uintptr(unsafe.Pointer(testStruct)),
    cap: int(Len),
    len: int(Len),
  }
  data := *(*[]byte)(unsafe.Pointer(testBytes))
  fmt.Println("[]byte is : ", data)
}

*/

func Setup(numOfLevel0, thresholdOfLevel0 int, alphai *pbc.Element, kgalphai []*pbc.Element)(*TIBGSMasterPublicKey, *TIBGSMasterSecretKeyi){
	t1:=time.Now()
	var mpk TIBGSMasterPublicKey //g，h1是G1上的点 g2,u0,u1,u2,u3,u4是G2上的点  n是GT上的点
	mpk.g = pairing.NewG1().SetBytes(gstatic)
	//fmt.Println("mpk.gori=",mpk.g)
	//gyh
	//fmt.Println("byte_mpk_g:", hex.EncodeToString(mpk.g.Bytes()))
	//b_g, err := hex.DecodeString("5680ca41812b79a925374414fac3fce13bf5e49611354f717c5bbdf99dfa7c0ac3850f7ca52d770b68f4cccc6abdac1154c61135ac1559aa38a4e19ac37f1d498e391c868f80d6bd95486c9318c211f2e715e9fd7707f75ba6a2ace81df28abb56e3d1925824cc9bf7229ca762ab62d0341930a077390003a443598255f6fff4")
	b_g, err := hex.DecodeString(hex.EncodeToString(mpk.g.Bytes()))
	if err!=nil{
		panic(err)
	}
	mpk.g.SetBytes(b_g)
	//fmt.Println("mpk.g=",mpk.g)

	mpk.g2 = pairing.NewG2().SetBytes(g2static)
	//fmt.Println("byte_mpk_g2:", hex.EncodeToString(mpk.g2.Bytes()))
	/*
	b_g2, err := hex.DecodeString("3c991c62439a71107578b51dea56cfe1ab1786c016ec6bd4cf5d2b1ac0d01ba972730ab376194237a9c2cd26ac8c8d2747981ac8610b11496a1b188ac70b0d9d184ed33fd9dda9cacf86b824795e4b91cc6848e5449e0a975e1d9f4c8c0723fe9405fbf0128d7100278198a781338cd5ccef4faba5c147d250a6a55ee49a0185")
	if err!=nil{
		panic(err)
	}
	mpk.g2.SetBytes(b_g2)

	 */
	//fmt.Println("mpk.g2=",mpk.g2)

	// mpk.h1 = pairing.NewG1().SetXBytes(h1static)
	mpk.u0 = pairing.NewG2().SetBytes(u0static)
	//fmt.Println("byte_mpk_u0:", hex.EncodeToString(mpk.u0.Bytes()))
	//fmt.Println("mpk.u0=",mpk.u0)  21c6bf766e00ab904c70a2e054237f4c92bee1e4ad32cf327871ebebc0f89a59c4e46b35cddb34ca23e8354a5952110c6c6b933e0111c9afc7a57779a94cd4cd546fc78841e0823cdca54d824950bf6ddce907fa101be43f7b9942772b3a67269d045a8631b3efa94ecb99a778c2d4d0ec9fa0394501f757cc2803bfc362c57f
	mpk.u1 = pairing.NewG2().SetBytes(u1static)
	//fmt.Println("byte_mpk_u1:", hex.EncodeToString(mpk.u1.Bytes()))
	mpk.u2 = pairing.NewG2().SetBytes(u2static)
	//fmt.Println("byte_mpk_u2:", hex.EncodeToString(mpk.u2.Bytes()))
	mpk.u3 = pairing.NewG2().SetBytes(u3static)
	//fmt.Println("byte_mpk_u3:", hex.EncodeToString(mpk.u3.Bytes()))
	mpk.u4 = pairing.NewG2().SetBytes(u4static)
	//fmt.Println("byte_mpk_u4:", hex.EncodeToString(mpk.u4.Bytes()))
	mpk.n = pairing.NewGT().SetBytes(nstatic)
	//fmt.Println("byte_mpk_n:", hex.EncodeToString(mpk.n.Bytes()))
	//拉格朗日插值计算h1
	mpk.h1 = Galpha(kgalphai, thresholdOfLevel0) //通过k个g^alphai求出g^alpha
	//fmt.Println("byte_mpk_h1:", hex.EncodeToString(mpk.h1.Bytes()))
	//门限 计算mski
	var mski TIBGSMasterSecretKeyi  //G2上的一个点
	//通过门限求出alphai n
	mski.h2i = pairing.NewG2().PowZn(mpk.g2, alphai) //g2^alphai
	t2:=time.Now()
	fmt.Println("time:Setup",t2.Sub(t1))
	//f, _ := os.OpenFile("Setup.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &mpk, &mski
}

//实现方便，单节点生成
func NewSetup(numOfLevel0, thresholdOfLevel0 int, grpID string) (*TIBGSMasterPublicKey, *TIBGSMasterSecretKeyi, *TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi, []*Sharealphar, error){
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
	r := pairing.NewZr().Rand()

	mpk.h1 = pairing.NewG1().PowZn(mpk.g, alpha)

	//生成n个节点的alphai和ri
	sk := GenCoef(alpha, thresholdOfLevel0)
	tk := GenCoef(r, thresholdOfLevel0)
	//产生N个管理员的分享
	nalpha := CN(sk, numOfLevel0)
	nr := CN(tk, numOfLevel0)

	nss := make([]*Sharealphar, numOfLevel0)
	for i := 0; i < numOfLevel0; i++ {
		var tmpSs Sharealphar
		tmpSs.alphai = nalpha[i] //g2^s[i]
		tmpSs.ri = nr[i] //1
		nss[i] = &tmpSs
	}

	falphai := nalpha[0]
	fri := nr[0]
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
	return &mpk, &mski, &gski, &gvki, nss, nil
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

func GrpSetUp(mpk *TIBGSMasterPublicKey, mski *TIBGSMasterSecretKeyi,  grpID string, ri, alphai *pbc.Element)(*TIBGSGroupSecretKeyi, *TIBGSGroupVerifyKeyi){
	t1:=time.Now()
	I := G(grpID)
	// fmt.Println("G(grpID)",G(grpID))
	var gski TIBGSGroupSecretKeyi //a0i, a2i, a3i, a4i是G2上的点  a5i是G1上的点
	//秘密分享ri n
	a0 := pairing.NewG2().PowZn(mpk.u1, I) //u1^G(grpID)
	a0.Mul(mpk.u0, a0)  //u0 * u1^G(grpID)
	a0.PowZn(a0, ri) //(u0 * u1^G(grpID))^ri
	gski.a0i = pairing.NewG2().Mul(mski.h2i, a0) // h2i*(u0 * u1^G(grpID))^
	// fmt.Println("gski.a0i",gski.a0i)
	gski.a2i = pairing.NewG2().PowZn(mpk.u2, ri)
	gski.a3i = pairing.NewG2().PowZn(mpk.u3, ri)
	gski.a4i = pairing.NewG2().PowZn(mpk.u4, ri)
	gski.a5i = pairing.NewG1().PowZn(mpk.g, ri)

	var gvki TIBGSGroupVerifyKeyi //G1上的点
	gvki.gai = pairing.NewG1().PowZn(mpk.g, alphai)
	t2:=time.Now()
	fmt.Println("time:GrpSetUp",t2.Sub(t1))
	//f, _ := os.OpenFile("GrpSetUp.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &gski, &gvki
}

func ExtShare(gski *TIBGSGroupSecretKeyi,userID string) *TIBGSUserSecretKey{
	t1:=time.Now()
	I := G(userID)
	var uski TIBGSUserSecretKey //b0, b3, b4是G2上的点 b5是G1上的点
	mid := pairing.NewG2().PowZn(gski.a2i, I) //a2i^G(userID)
	uski.b0 = pairing.NewG2().Mul(gski.a0i, mid) //a0i * a2i^G(userID)
	uski.b3 = gski.a3i
	uski.b4 = gski.a4i
	uski.b5 = gski.a5i
	t2:=time.Now()
	fmt.Println("time:ExtShare",t2.Sub(t1))
	//f, _ := os.OpenFile("ExtShare.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &uski
}

func VerifyShare(uski *TIBGSUserSecretKey, gvki *TIBGSGroupVerifyKeyi, mpk *TIBGSMasterPublicKey, grpID string, userID string) bool{
	//t1:=time.Now()
	IG := G(grpID)
	IU := G(userID)

	left1 := pairing.NewGT().Pair(mpk.g , uski.b0) //e(g , b0)
	right11 := pairing.NewGT().Pair(gvki.gai, mpk.g2) //e(gvk , g2)
	right12 := pairing.NewGT().Pair(uski.b5, mpk.u0) //e(b5, u0)
	u1grpID := pairing.NewG2().PowZn(mpk.u1, IG) //u1^G(grpID)
	right13 := pairing.NewGT().Pair(uski.b5,u1grpID) //e(b5,u1^G(grpID))
	u2userID := pairing.NewG2().PowZn(mpk.u2, IU) //u2^G(userID)
	right14 := pairing.NewGT().Pair(uski.b5, u2userID) //e(b5, u2^G(userID))
	right1 := pairing.NewGT().Mul(right11, right12) // e(gvk , g2) e(b5, u0)
	right1.Mul(right1, right13) //e(gvk , g2) e(b5, u0) e(b5,u1^G(grpID))
	right1.Mul(right1, right14) //e(gvk , g2) e(b5, u0) e(b5,u1^G(grpID)) e(b5, u2^G(userID))
	left2 := pairing.NewGT().Pair(mpk.g , uski.b3)  //e(g, b3)
	right2 := pairing.NewGT().Pair(uski.b5, mpk.u3) //e(b5, u3)
	left3 := pairing.NewGT().Pair(mpk.g, uski.b4) //e(g, b4)
	right3 := pairing.NewGT().Pair(uski.b5, mpk.u4) //e(b5, u4)
	tf:=left1.Equals(right1) && left2.Equals(right2) && left3.Equals(right3)
	//t2:=time.Now()
	// fmt.Println("time:VerifyShare",t2.Sub(t1))
	//f, _ := os.OpenFile("VerifyShare.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return tf
}

func ReconstKey(uskis []*TIBGSUserSecretKey, K int,mpk *TIBGSMasterPublicKey, grpID string, userID string) *TIBGSUserSecretKey{
	//t1:=time.Now()
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
	// fmt.Println("mb0,mb3,mb4,mb5",mb0,mb3,mb4,mb5)
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
	//t2:=time.Now()
	// fmt.Println("time:ReconstKey",t2.Sub(t1))
	//f, _ := os.OpenFile("ReconstKey.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &usk
}

func Sign(mpk *TIBGSMasterPublicKey,usk *TIBGSUserSecretKey, message string, grpID string, userID string ) *TIBGSSIG{
	t1:=time.Now()
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
	ssig.c0.Mul(ssig.c0, c04) // b0 * b3^m * b4^G(rID) * (u0 * u1^G(grpID) * u2^G(userID) * u3^m * u4^G(rID))^r3\

	/*
	fmt.Println("b_c0=", ssig.c0)
	b_c0, err := hex.DecodeString(hex.EncodeToString(ssig.c0.Bytes()))
	if err!=nil{
		panic(err)
	}


	ssig.c0.SetBytes(b_c0)
	fmt.Println("b_c0_1=", ssig.c0)
	*/
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
	ssig.pok = POK

	t2:=time.Now()
	fmt.Println("time:Sign=",t2.Sub(t1))
	//file, _ := os.OpenFile("Sign.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(file,t2.Sub(t1))
	return &ssig
}

func Verify(ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey, message string, grpID string) bool{
	t1:=time.Now()
	msg := G(message)
	IG := G(grpID)

	t := pairing.NewZr().Rand()
	M := pairing.NewGT().Rand()
	// fmt.Println("M",M)
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
	// fmt.Println("cc",cc)
	// fmt.Println("pok.c",ssig.pok.c)
	tf:= M.Equals(right) && cc.Equals(ssig.pok.c)
	t2:=time.Now()
	fmt.Println("time:Verify=",t2.Sub(t1))
	//f, _ := os.OpenFile("Verify.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return tf
}

func OpenPart(gski *TIBGSGroupSecretKeyi, ssig *TIBGSSIG) *TIBGSOK{
	//t1:=time.Now()
	var OKi TIBGSOK //ok1, ok2 GT上的点
	OKi.ok1 = pairing.NewGT().Pair(ssig.e1, gski.a0i)
	OKi.ok2 = pairing.NewGT().Pair(gski.a5i, ssig.e2)
	//t2:=time.Now()
	// fmt.Println("time:OpenPart",t2.Sub(t1))
	//f, _ := os.OpenFile("OpenPart.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return &OKi
}

func Open(OKK []*TIBGSOK, K int) *pbc.Element{
	//t1:=time.Now()
	//lag
	if len(OKK) < K{
		fmt.Println("ok is not enough")
		return nil
	}
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
	//t2:=time.Now()
	// fmt.Println("time:Open",t2.Sub(t1))
	//f, _ := os.OpenFile("Open.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//fmt.Fprintln(f,t2.Sub(t1))
	return gama
}

func FindUser(UIDS []string, gama *pbc.Element,ssig *TIBGSSIG, mpk *TIBGSMasterPublicKey)string{
	//t1:=time.Now()
	for _,ID := range UIDS{
		GU := G(ID)
		right:=pairing.NewGT().PowZn(mpk.n, GU)
		right.Mul(right,gama)

		if ssig.e3.Equals(right){
			//t2:=time.Now()
			// fmt.Println("time:FindUser",t2.Sub(t1))
			//f, _ := os.OpenFile("FindUser.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
			//fmt.Fprintln(f,t2.Sub(t1))
			return ID
		}


	}
	return "no user here"
}

//gyh:合并byte
func BytesCombine1(pBytes ...[]byte) []byte {
	length := len(pBytes)
	s := make([][]byte, length)
	for index := 0; index < length; index++ {
		s[index] = pBytes[index]
	}
	sep := []byte("")
	return bytes.Join(s, sep)
}
//pairing.NewG1().Rand()  G1上一个点
//pairing.NewG2().Rand()  G2上一个点
//pairing.NewGT().Rand()  GT上一个点
//pairing.NewZr().SetBytes() 从byte还原出点 可用于固定点
//For some generator g in G1, generator h in G2, and x and y in Zr: e(gˣ, hʸ) = e(g,h)ˣʸ
//门限秘密分享的  s,t 都是Zr上的
//g1 := pairing.NewG1().PowZn(g, alpha) g1=g^alpha  g和g1都是G1上的点，alpha是Zr上的数
//el := pairing.NewZr().Add(x, y) el=x+y
//el := pairing.NewG1().Mul(x, y) el=x*y
//Set1 用于乘法  Set0 用于加法
//MulZn(x,i)  el = i*x
// el := pairing.NewGT().Pair(x , y) 双线行对运算 等同于c的pairing_apply
//el.Equals(x) 判断el与x是否相等  等同于c element_cmp
//pairing.NewZr().String()  转换为string
//
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func GetSHA256HashCode(message []byte)string{
	//方法一：
	//创建一个基于SHA256算法的hash.Hash接口的对象
	hash := sha256.New()
	//输入数据
	hash.Write(message)
	//计算哈希值
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	hashCode := hex.EncodeToString(bytes)
	//返回哈希值
	return hashCode

	//方法二：
	//bytes2:=sha256.Sum256(message)//计算哈希值，返回一个长度为32的数组
	//hashcode2:=hex.EncodeToString(bytes2[:])//将数组转换成切片，转换成16进制，返回字符串
	//return hashcode2
}

func main(){
	N,K:=1,1 //(11,15) (21,30) (31,45) (41,60) (51,75) (61,90)
	//管理员之间秘密分享
	//以2为例

	managerN := make([][]*pbc.Element, N)
	for i := 0; i < N; i++ {
        managerN[i] = make([]*pbc.Element, N)
    }
	//1. 生成 si秘密share
	for i:=0;i<N;i++{
		for j:=0;j<N;j++{
			s := pairing.NewZr().SetInt32(int32(i+1))//1...N  固定
			ssi,_ := SharesGen(K,N,s)
			// fmt.Println("ssi",ssi)
			managerN[i][j] = ssi[j]
		}
	}
	// fmt.Println(manager1,manager2,manager3,manager4,manager5,manager6)
	alphaN := make([]*pbc.Element, N)
	for i := 0; i < N; i++ {
        alphaN[i],_ = ReconManager(managerN[i],managerN[i],N)
    }
	//秘密分享结束

	managerk := make([]*pbc.Element, N) //存放N个g^alphai
	var gstatic = []byte{86,128,202,65,129,43,121,169,37,55,68,20,250,195,252,225,59,245,228,150,17,53,79,113,124,91,189,249,157,250,124,10,195,133,15,124,165,45,119,11,104,244,204,204,106,189,172,17,84,198,17,53,172,21,89,170,56,164,225,154,195,127,29,73,142,57,28,134,143,128,214,189,149,72,108,147,24,194,17,242,231,21,233,253,119,7,247,91,166,162,172,232,29,242,138,187,86,227,209,146,88,36,204,155,247,34,156,167,98,171,98,208,52,25,48,160,119,57,0,3,164,67,89,130,85,246,255,244}

	g := pairing.NewG1().SetBytes(gstatic)
	//fmt.Println("g",g)
	for i := 0; i < N; i++ {
        managerk[i] = pairing.NewG1().PowZn(g, alphaN[i])
	}

	//h1 := Galpha(managerk,K)
	//fmt.Println("h1",h1) //tibgs算出的

	//alpha := pairing.NewZr().SetInt32(int32(1+2+3+4+5))
	//h2 := pairing.NewG1().PowZn(g, alpha)
	//fmt.Println("h2",h2) //1+2+3+4+5算出的
	//1. setup
	mpkK := make([]*TIBGSMasterPublicKey, K)
	mskK := make([]*TIBGSMasterSecretKeyi, K)
	for i := 0; i < K; i++ {
		mpkK[i],mskK[i] = Setup(N,K,alphaN[i],managerk)
	}

	//2. Grpsetup　　gyh：管理员初始化整个群签名
	gskK := make([]*TIBGSGroupSecretKeyi, K)
	gvkK := make([]*TIBGSGroupVerifyKeyi, K)
	for i := 0; i < K; i++ {
		gskK[i],gvkK[i] = GrpSetUp(mpkK[i],mskK[i],"computer",alphaN[i],alphaN[i])
	}
	//3. ExtShare　　　　gyh：初始化用户的私钥
	uskK := make([]*TIBGSUserSecretKey, K)
	for i := 0; i < K; i++ {
		uskK[i] = ExtShare(gskK[i], "zhou")
	}

	//4. Reconstkey　　gyh：验证用户与管理员的秘钥相关性是否正确
	for i := 0; i < K; i++ {
		tf := VerifyShare(uskK[i], gvkK[i], mpkK[i], "computer","zhou")
		fmt.Println("tf=",tf)
	}

	//4. Reconstkey
	usk:=ReconstKey(uskK,K,mpkK[0],"computer","zhou")
	//fmt.Println("usk",usk)
	//5.sign
	mess := "19"
	mess_byte := []byte(mess)
	hash_mess := GetSHA256HashCode(mess_byte)

	ssig:=Sign(mpkK[0],usk,hash_mess,"computer","zhou")

	//将mpk的g,g2,u0,u1,u2,u3,u4,n,h1转化为byte类型
	g_byte := mpkK[0].g.Bytes()
	//g_16 := mpkK[0].g.String()
	//  fmt.Println(hex.DecodeString(string([]byte(hex.EncodeToString(g_byte)))))      !!!!!!!!!!!!!
	g_16:= hex.EncodeToString(g_byte)	//此处将byte存储变为16进制存储
	//fmt.Println("g_16",g_16)
	g2_byte := mpkK[0].g2.Bytes()
	g2_16:=hex.EncodeToString(g2_byte)
	//fmt.Println("g_byte",g_byte)
	u0_byte := mpkK[0].u0.Bytes()
	u0_16:=hex.EncodeToString(u0_byte)
	fmt.Println("u0_byte",u0_byte)
	u1_byte := mpkK[0].u1.Bytes()
	u1_16:=hex.EncodeToString(u1_byte)
	fmt.Println("u1_byte",u1_byte)
	u2_byte := mpkK[0].u2.Bytes()
	u2_16:=hex.EncodeToString(u2_byte)
	fmt.Println("u2_byte",u2_byte)
	u3_byte := mpkK[0].u3.Bytes()
	u3_16:=hex.EncodeToString(u3_byte)
	fmt.Println("u3_byte",u3_byte)
	u4_byte := mpkK[0].u4.Bytes()
	u4_16:=hex.EncodeToString(u4_byte)
	fmt.Println("u4_byte",u4_byte)
	n_byte := mpkK[0].n.Bytes()
	n_16:=hex.EncodeToString(n_byte)
	fmt.Println("n_byte",n_byte)
	h1_byte := mpkK[0].h1.Bytes()
	h1_16:=hex.EncodeToString(h1_byte)
	fmt.Println("h1_byte",h1_byte)

	//gyh：分别将c0,c5,c6,e1,e2,e3,pok转化为byte类型
	c0_byte := ssig.c0.Bytes()
	c0_16:=hex.EncodeToString(c0_byte)
	fmt.Println("c0_byte",c0_byte)
	c5_byte := ssig.c5.Bytes()
	c5_16:=hex.EncodeToString(c5_byte)
	fmt.Println("c5_byte",c5_byte)
	c6_byte := ssig.c6.Bytes()
	c6_16:=hex.EncodeToString(c6_byte)
	fmt.Println("c6_byte",c6_byte)
	e1_byte := ssig.e1.Bytes()
	e1_16:=hex.EncodeToString(e1_byte)
	fmt.Println("e1_byte",e1_byte)
	e2_byte := ssig.e2.Bytes()
	e2_16:=hex.EncodeToString(e2_byte)
	fmt.Println("e2_byte",e2_byte)
	e3_byte := ssig.e3.Bytes()
	e3_16:=hex.EncodeToString(e3_byte)
	fmt.Println("e3_byte",e3_byte)
	c_byte := ssig.pok.c.Bytes()
	c_16:=hex.EncodeToString(c_byte)
	fmt.Println("c_byte",c_byte)
	s1_byte := ssig.pok.s1.Bytes()
	s1_16:=hex.EncodeToString(s1_byte)
	fmt.Println("s1_byte",s1_byte)
	s2_byte := ssig.pok.s2.Bytes()
	s2_16:=hex.EncodeToString(s2_byte)
	fmt.Println("s2_byte",s2_byte)
	s3_byte := ssig.pok.s3.Bytes()
	s3_16:=hex.EncodeToString(s3_byte)
	fmt.Println("s3_byte",s3_byte)


	//计算mess的hash值

	fmt.Println("mses_byte",mess_byte)
	fmt.Println("hash_test需要的=",hash_mess)
	hash_mess_byte := []byte(hash_mess)
	fmt.Println("hash_mess_byte=",hash_mess_byte)
	fmt.Println("hash_mess_byte_encode=",hex.EncodeToString(hash_mess_byte))
	mess_send := hex.EncodeToString(hash_mess_byte)
	fmt.Println("hash_mess_byte_encode_byte=",[]byte(hex.EncodeToString(hash_mess_byte)))
	fmt.Println("hash_mess_byte_encode_byte_string=",string([]byte(hex.EncodeToString(hash_mess_byte))))
	test,_:= hex.DecodeString(string([]byte(hex.EncodeToString(hash_mess_byte))))
	fmt.Println("hash_mess_byte_encode_byte_string_Decode=",test)
	fmt.Println("hash_mess_byte_encode_byte_string_Decode_string=",string(test))
	//把mess的hash值传入
	//gyh: res_byte　为最后合并的传输结果
	f, _ := os.OpenFile("/home/u0/goproject/src/github.com/ethereum/go-ethereum/Groupsign/pubhubtimetest/output2.txt",os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
	fmt.Fprintln(f,g_16+"#"+g2_16+"#"+u0_16+"#"+u1_16+"#"+u2_16+"#"+u3_16+"#"+u4_16+"#"+n_16+"#"+h1_16+"#"+c0_16+"#"+c5_16+"#"+c6_16+"#"+e1_16+"#"+e2_16+"#"+e3_16+"#"+c_16+"#"+s1_16+"#"+s2_16+"#"+s3_16+"#"+mess_send)
	//res_byte := BytesCombine1(g_byte,g2_byte,u0_byte,u1_byte,u2_byte,u3_byte,u4_byte,n_byte,h1_byte,c0_byte,c5_byte,c6_byte,e1_byte,e2_byte,e3_byte,c_byte,s1_byte,s2_byte,s3_byte,hash_mess_byte)
	//fmt.Println("res_byte",res_byte)
	//f, _ := os.OpenFile("output2.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//f, _ := os.OpenFile("/home/u0/goproject/src/github.com/ethereum/go-ethereum/Groupsign/pubhubtimetest/output2.txt",os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
	//fmt.Fprintln(f,res_byte)
	//err2 := ioutil.WriteFile("/home/u0/goproject/src/github.com/yeongchingtarn/geth-pbft/hibe/testnewtibgs/pubhubtimetest/output2.txt", res_byte) //写入文件(字节数组)
	//check(err2)
	//res_byte_string:= string(res_byte)
	//fmt.Println("res_byte_STRING",res_byte_string)
	//6.verify
	fmt.Println( Verify(ssig,mpkK[0],hash_mess,"computer"))
	//7.openpart
	/*
	var res_len = len(res_byte) - 2000
	 var mpk_get_g [128]byte
	var mpk_get_g2 [128]byte
	var mpk_get_u0 [128]byte
	var mpk_get_u1 [128]byte
	var mpk_get_u2 [128]byte
	var mpk_get_u3 [128]byte
	var mpk_get_u4 [128]byte
	var mpk_get_n [128]byte
	var mpk_get_h1 [128]byte
	var ssig_get_c0 [128]byte
	var ssig_get_c5 [128]byte
	var ssig_get_c6 [128]byte
	var ssig_get_e1 [128]byte
	var ssig_get_e2 [128]byte
	var ssig_get_e3 [128]byte
	var pok_get_c [20]byte
	var pok_get_s1 [20]byte
	var pok_get_s2 [20]byte
	var pok_get_s3 [20]byte
	var mess_get [3000]byte  //
	for i := 0; i <= 127; i++ {
		//fmt.Println(res_byte[i])
		mpk_get_g[i] = res_byte[i]
		mpk_get_g2[i] = res_byte[i+128]
		mpk_get_u0[i] = res_byte[i+128*2]
		mpk_get_u1[i] = res_byte[i+128*3]
		mpk_get_u2[i] = res_byte[i+128*4]
		mpk_get_u3[i] = res_byte[i+128*5]
		mpk_get_u4[i] = res_byte[i+128*6]
		mpk_get_n[i] = res_byte[i+128*7]
		mpk_get_h1[i] = res_byte[i+128*8]

		ssig_get_c0[i] = res_byte[i+128*9]
		ssig_get_c5[i] = res_byte[i+128*10]
		ssig_get_c6[i] = res_byte[i+128*11]
		ssig_get_e1[i] = res_byte[i+128*12]
		ssig_get_e2[i] = res_byte[i+128*13]
		ssig_get_e3[i] = res_byte[i+128*14]
	}
	//mpk_get_h1[i] = res_byte[i]
	fmt.Println("h1=",mpk_get_h1)
	//ssig_get_c0[i] = res_byte[i+128]
	fmt.Println("c0=",ssig_get_c0)
	//ssig_get_c5[i] = res_byte[i+128*2]
	fmt.Println("c5=",ssig_get_c5)
	//ssig_get_c6[i] = res_byte[i+128*3]
	fmt.Println("c6=",ssig_get_c6)
	//ssig_get_e1[i] = res_byte[i+128*4]
	fmt.Println("e1=",ssig_get_e1)
	//ssig_get_e2[i] = res_byte[i+128*5]
	fmt.Println("e2=",ssig_get_e2)
	//ssig_get_e3[i] = res_byte[i+128*6]
	fmt.Println("e3=",ssig_get_e3)

	for i := 0; i <= 19; i++ {
		pok_get_c[i] = res_byte[i+128*15]
		//fmt.Println("c=",pok_get_c)
		pok_get_s1[i] = res_byte[i+128*15+20]
		//fmt.Println("s1=",pok_get_s1)
		pok_get_s2[i] = res_byte[i+128*15+20*2]
		//fmt.Println("s2=",pok_get_s2)
		pok_get_s3[i] = res_byte[i+128*15+20*3]
		//fmt.Println("s3=",pok_get_s3)
	}

	for i := 0; i <= res_len-1; i++ {

		//mess_get[i] = res_byte[i+128*7+20*4]
		mess_get[i] = res_byte[i+128*15+20*4]

	}
	mess_str := string(mess_get[:res_len])

	fmt.Println("res_len=",res_len)
	//pok_get_c[i] = res_byte[i+128*6+20]
	fmt.Println("c=",pok_get_c)
	//pok_get_s1[i] = res_byte[i+128*6+20*2]
	fmt.Println("s1=",pok_get_s1)
	//pok_get_s2[i] = res_byte[i+128*6+20*3]
	fmt.Println("s2=",pok_get_s2)
	//pok_get_s3[i] = res_byte[i+128*6+20*4]
	fmt.Println("s3=",pok_get_s3)
	fmt.Println("mess_get",mess_get[:res_len])
	fmt.Println("mess=",mess_str)



	 */

	time_start:=time.Now()

	okK := make([]*TIBGSOK, K)
	for i := 0; i < K; i++ {
		okK[i] = OpenPart(gskK[i],ssig)
	}
	//8.open
	gama:=Open(okK,K)
	suser:=make([]string, K)
	suser[0]="zhou"
	fmt.Println( FindUser(suser,gama,ssig,mpkK[0]))
	fmt.Println("公开所用的时间是：",time.Now().Sub(time_start))
}

