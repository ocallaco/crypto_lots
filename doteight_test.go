package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotEightOperations(t *testing.T) {
	de1 := ToDotEight("10.101")
	de2 := ToDotEight("5.0505")

	assert.Equal(t, "15.15150000", (de1 + de2).ToString())
	assert.Equal(t, "5.05050000", (de1 - de2).ToString())
	assert.Equal(t, "2.00000000", (de1.Div(de2)).ToString())
	assert.Equal(t, "51.01510050", (de1.Mul(de2)).ToString())

	de3 := ToDotEight("1")
	de4 := ToDotEight("5")
	de5 := ToDotEight("3")

	assert.Equal(t, "0.20000000", de3.Div(de4).ToString())
	assert.Equal(t, "0.33333333", de3.Div(de5).ToString())

	de6 := DotEight(9223372036854775807) // int64 max
	assert.Equal(t, "92233720368.54775807", de6.ToString())
	assert.Equal(t, de6.ToString(), de6.Mul(de3).ToString())
}
