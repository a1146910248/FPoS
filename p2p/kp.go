package p2p

//import (
//	"fmt"
//	"github.com/consensys/gnark-crypto/backend/plonk"
//	"github.com/consensys/gnark-crypto/ecc"
//	"github.com/consensys/gnark-crypto/frontend"
//	"github.com/consensys/gnark-crypto/frontend/cs/r1cs"
//)
//
//// 转账电路结构定义
//type TransferCircuit struct {
//	OldBalance frontend.Variable // 私密输入：原余额
//	Amount     frontend.Variable // 私密输入：转账金额
//	NewBalance frontend.Variable `gnark:",public"` // 公开输出：新余额
//}
//
//// 定义电路约束
//func (c *TransferCircuit) Define(api frontend.API) error {
//	// 约束1: 新余额 = 原余额 - 转账金额
//	calculated := api.Sub(c.OldBalance, c.Amount)
//	api.AssertIsEqual(c.NewBalance, calculated)
//
//	// 约束2: 转账金额必须为正数
//	api.AssertIsLessOrEqual(1, c.Amount)
//
//	// 约束3: 原余额必须 >= 转账金额
//	api.AssertIsLessOrEqual(c.Amount, c.OldBalance)
//	return nil
//}
//
//// 生成证明和验证密钥
//func GenerateKeys() (plonk.ProvingKey, plonk.VerifyingKey, error) {
//	var circuit TransferCircuit
//
//	// 编译电路
//	cc, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	// 创建SRS（结构化参考字符串）
//	srs, err := plonk.NewKZGSetup(cc, ecc.BN254)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	// 生成证明密钥和验证密钥
//	pk, vk, err := plonk.Setup(cc, srs)
//	return pk, vk, err
//}
//
//// 生成转账证明
//func GenerateProof(pk plonk.ProvingKey, oldBal, amount int) (plonk.Proof, frontend.Witness, error) {
//	// 创建赋值对象
//	assignment := &TransferCircuit{
//		OldBalance: oldBal,
//		Amount:     amount,
//		NewBalance: oldBal - amount,
//	}
//
//	// 创建完整witness
//	fullWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
//	if err != nil {
//		return nil, nil, err
//	}
//
//	// 生成证明
//	proof, err := plonk.Prove(cc, pk, fullWitness)
//	return proof, fullWitness.Public(), err
//}
//
//// 验证转账证明
//func VerifyProof(vk plonk.VerifyingKey, proof plonk.Proof, publicWitness frontend.Witness) error {
//	return plonk.Verify(proof, vk, publicWitness)
//}
//
//func main() {
//	// 示例调用
//	pk, vk, err := GenerateKeys()
//	if err != nil {
//		panic(err)
//	}
//
//	// 创建交易参数
//	oldBalance := 1000
//	transferAmount := 300
//	newBalance := oldBalance - transferAmount
//
//	// 生成证明
//	proof, publicWitness, err := GenerateProof(pk, oldBalance, transferAmount)
//	if err != nil {
//		panic(err)
//	}
//
//	// 验证证明
//	err = VerifyProof(vk, proof, publicWitness)
//	if err != nil {
//		fmt.Println("验证失败:", err)
//	} else {
//		fmt.Println("验证成功!")
//	}
//}
