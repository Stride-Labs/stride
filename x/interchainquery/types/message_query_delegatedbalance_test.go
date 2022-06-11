package types

// func TestMsgQueryDelegatedbalance_ValidateBasic(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		msg  MsgQueryDelegatedbalance
// 		err  error
// 	}{
// 		{
// 			name: "invalid address",
// 			msg: MsgQueryDelegatedbalance{
// 				Creator: "invalid_address",
// 			},
// 			err: sdkerrors.ErrInvalidAddress,
// 		}, {
// 			name: "valid address",
// 			msg: MsgQueryDelegatedbalance{
// 				Creator: sample.AccAddress(),
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := tt.msg.ValidateBasic()
// 			if tt.err != nil {
// 				require.ErrorIs(t, err, tt.err)
// 				return
// 			}
// 			require.NoError(t, err)
// 		})
// 	}
// }
