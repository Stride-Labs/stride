package types

import fmt "fmt"

// Confirm there are no duplicate delegation record Ids and that the amounts are not nil
func ValidateDelegationRecordGenesis(delegationRecords []DelegationRecord) error {
	ids := map[uint64]bool{}
	for _, delegationRecord := range delegationRecords {
		if delegationRecord.NativeAmount.IsNil() {
			return ErrInvalidGenesisRecords.Wrapf("uninitialized native amount in delegation record %d", delegationRecord.Id)
		}
		if _, ok := ids[delegationRecord.Id]; ok {
			return ErrInvalidGenesisRecords.Wrapf("duplicate delegation record %d", delegationRecord.Id)
		}
		ids[delegationRecord.Id] = true
	}
	return nil
}

// Confirm there are no duplicate unbonding record Ids and that the amounts are not nil
func ValidateUnbondingRecordGenesis(unbondingRecords []UnbondingRecord) error {
	ids := map[uint64]bool{}
	for _, unbondingRecord := range unbondingRecords {
		if unbondingRecord.NativeAmount.IsNil() {
			return ErrInvalidGenesisRecords.Wrapf("uninitialized native amount in unbonding record %d", unbondingRecord.Id)
		}
		if unbondingRecord.StTokenAmount.IsNil() {
			return ErrInvalidGenesisRecords.Wrapf("uninitialized sttoken amount in unbonding record %d", unbondingRecord.Id)
		}
		if _, ok := ids[unbondingRecord.Id]; ok {
			return ErrInvalidGenesisRecords.Wrapf("duplicate unbonding record %d", unbondingRecord.Id)
		}
		ids[unbondingRecord.Id] = true
	}
	return nil
}

// Confirm there are no duplicate slash record Ids and that the amounts are not nil
func ValidateRedemptionRecordGenesis(redemptionRecords []RedemptionRecord) error {
	ids := map[string]bool{}
	for _, redemptionRecord := range redemptionRecords {
		idKey := fmt.Sprintf("%d-%s", redemptionRecord.UnbondingRecordId, redemptionRecord.Redeemer)

		if redemptionRecord.NativeAmount.IsNil() {
			return ErrInvalidGenesisRecords.Wrapf("uninitialized native amount in redemption record %s", idKey)
		}
		if redemptionRecord.StTokenAmount.IsNil() {
			return ErrInvalidGenesisRecords.Wrapf("uninitialized sttoken amount in redemption record %s", idKey)
		}

		if _, ok := ids[idKey]; ok {
			return ErrInvalidGenesisRecords.Wrapf("duplicate redemption record %s", idKey)
		}
		ids[idKey] = true
	}

	return nil
}

// Confirm there are no duplicate slash record Ids and that the amounts are not nil
func ValidateSlashRecordGenesis(slashRecords []SlashRecord) error {
	ids := map[uint64]bool{}
	for _, slashRecord := range slashRecords {
		if slashRecord.NativeAmount.IsNil() {
			return ErrInvalidGenesisRecords.Wrapf("uninitialized native amount in slash record %d", slashRecord.Id)
		}
		if _, ok := ids[slashRecord.Id]; ok {
			return ErrInvalidGenesisRecords.Wrapf("duplicate slash record %d", slashRecord.Id)
		}
		ids[slashRecord.Id] = true
	}
	return nil
}
