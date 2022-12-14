package helpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAmountToLotSize(t *testing.T) {
	type args struct {
		lot       float64
		precision int
		amount    float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test with lot of zero and invalid amount",
			args: args{
				lot:       0.00100000,
				precision: 8,
				amount:    0.00010000,
			},
			want: 0,
		},
		{
			name: "test with lot",
			args: args{
				lot:       0.00100000,
				precision: 3,
				amount:    1.39,
			},
			want: 1.389,
		},
		{
			name: "test with big decimal",
			args: args{
				lot:       0.00100000,
				precision: 8,
				amount:    11.31232419283240912834434,
			},
			want: 11.312,
		},
		{
			name: "test with big number",
			args: args{
				lot:       0.0010000,
				precision: 8,
				amount:    11232821093480213.31232419283240912834434,
			},
			want: 11232821093480213.3123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := AmountToLotSize(tt.args.lot, tt.args.precision, tt.args.amount)
			assert.Equal(t, tt.want, size)
			fmt.Println(size)
		})
	}
}

func TestValidData(t *testing.T) {
	// fmt.Println(AlignPrice(19750.11512, "0.01"))
	// fmt.Println(AlignQuantity(451.1222, "0.1"))
	// AlignPriceToString(702.00, "")
	fmt.Println(AlignQuantityToString(2.48941996515, "0.01"))
}
