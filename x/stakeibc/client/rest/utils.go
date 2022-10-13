package rest

type AddValidatorProposalReq struct {
	Title            string `protobuf:"bytes,1,opt,name=Title,proto3" json:"Title,omitempty"`
	Description      string `protobuf:"bytes,2,opt,name=Description,proto3" json:"Description,omitempty"`
	HostZone         string `protobuf:"bytes,3,opt,name=host_zone,json=hostZone,proto3" json:"host_zone,omitempty"`
	ValidatorName    string `protobuf:"bytes,4,opt,name=validator_name,json=validatorName,proto3" json:"validator_name,omitempty"`
	ValidatorAddress string `protobuf:"bytes,5,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
}
