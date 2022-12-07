package types

func (p *Path) IsNative() bool {
	return p.Denom == "ustrd" // TODO: Replace with param
}

func (p *Path) GetIBCDenomHash() {

}
