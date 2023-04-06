package types

func (d *LSMTokenDeposit) GetKey() []byte {
	return LSMTokenDepositKey(d.ChainId, d.ValidatorAddress, d.Denom)
}
