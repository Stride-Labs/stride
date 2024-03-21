package keeper_test

import (
	math "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v20/x/stakedym/types"
)

// Helper function to create and set 5 slashRecord objects with various attributes
func (s *KeeperTestSuite) createAndSetSlashRecords() []types.SlashRecord {
	SlashRecords := []types.SlashRecord{}
	valAddresses := []string{"valA", "valB", "valC", "valD", "valE"}
	offsets := []math.Int{
		math.NewInt(-1),
		math.NewInt(0),
		math.NewInt(1),
		math.NewInt(2),
		math.NewInt(10),
	}
	for i := 0; i < 5; i++ {
		slashRecord := types.SlashRecord{
			Id:               uint64(i),
			Time:             uint64(s.Ctx.BlockTime().Unix()),
			NativeAmount:     offsets[i],
			ValidatorAddress: valAddresses[i],
		}
		SlashRecords = append(SlashRecords, slashRecord)
		s.App.StakedymKeeper.SetSlashRecord(s.Ctx, slashRecord)
	}
	return SlashRecords
}

func (s *KeeperTestSuite) TestGetAllSlashRecords() {
	expectedSlashRecords := s.createAndSetSlashRecords()
	actualSlashRecords := s.App.StakedymKeeper.GetAllSlashRecords(s.Ctx)
	s.Require().Len(actualSlashRecords, len(expectedSlashRecords), "number of SlashRecords")
	s.Require().ElementsMatch(expectedSlashRecords, actualSlashRecords, "contents of SlashRecords")
}

func (s *KeeperTestSuite) TestSetSlashRecord() {
	expectedSlashRecords := s.createAndSetSlashRecords()
	// make a slash record with a NEW ID and set it, then make sure a new record was added
	newSlashRecord := types.SlashRecord{
		Id:               uint64(5),
		Time:             uint64(s.Ctx.BlockTime().Unix()),
		NativeAmount:     math.NewInt(1),
		ValidatorAddress: "valZ",
	}
	s.App.StakedymKeeper.SetSlashRecord(s.Ctx, newSlashRecord)
	actualSlashRecords := s.App.StakedymKeeper.GetAllSlashRecords(s.Ctx)
	s.Require().Len(actualSlashRecords, len(expectedSlashRecords)+1, "number of SlashRecords with new slashRecord added")
	s.Require().Equal(newSlashRecord, actualSlashRecords[5], "contents of newly added SlashRecord")

	// make a slash record with an existing ID and set it, then make sure no new record was added (just existing modified)
	overwriteSlashRecord := types.SlashRecord{
		Id:               uint64(0),
		Time:             uint64(s.Ctx.BlockTime().Unix()),
		NativeAmount:     math.NewInt(1),
		ValidatorAddress: "valZ",
	}
	s.App.StakedymKeeper.SetSlashRecord(s.Ctx, overwriteSlashRecord)
	actualSlashRecords = s.App.StakedymKeeper.GetAllSlashRecords(s.Ctx)
	s.Require().Len(actualSlashRecords, len(expectedSlashRecords)+1, "number of SlashRecords same as before overwriting")
	s.Require().Equal(overwriteSlashRecord, actualSlashRecords[0], "contents of newly added SlashRecord")
}

func (s *KeeperTestSuite) TestIncrementSlashRecordId() {
	prevSlashRecordId := s.App.StakedymKeeper.IncrementSlashRecordId(s.Ctx)
	currSlashRecordId := s.App.StakedymKeeper.IncrementSlashRecordId(s.Ctx)
	nextSlashRecordId := s.App.StakedymKeeper.IncrementSlashRecordId(s.Ctx)
	s.Require().Equal(prevSlashRecordId+1, currSlashRecordId, "incremented slash record id (tests incrementing)")
	s.Require().Equal(currSlashRecordId+1, nextSlashRecordId, "incremented slash record id again (test storing incremented val)")
}
