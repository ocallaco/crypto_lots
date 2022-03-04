package main

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type DotEight int64

func (de DotEight) ToString() string {
	left := de / 1e8
	right := de % 1e8
	if right < 0 {
		right = -right
	}
	return fmt.Sprintf("%d.%08d", left, right)
}

func (de DotEight) ToFloat64() float64 {
	return float64(de) / 1e8
}

func (de DotEight) Mul(de2 DotEight) DotEight {
	b1 := big.NewInt(int64(de))
	b2 := big.NewInt(int64(de2))

	res := &big.Int{}
	res.Mul(b1, b2)
	res.Div(res, big.NewInt(1e8))
	return DotEight(res.Int64())
}

func (de DotEight) Div(de2 DotEight) DotEight {
	b1 := big.NewInt(int64(de))
	b2 := big.NewInt(int64(de2))

	res := &big.Int{}
	rem := &big.Int{}
	res.DivMod(b1, b2, rem)

	s := big.NewInt(1e8)
	res.Mul(res, s)
	rem.Mul(rem, s)
	rem.Div(rem, b2)
	res.Add(res, rem)
	return DotEight(res.Int64())
}

func ToDotEight(s string) DotEight {
	comps := strings.Split(s, ".")
	if len(comps) > 2 {
		panic(fmt.Sprintf("invalid string for DotEight: %s\n", s))
	}

	left, err := strconv.Atoi(comps[0])
	if err != nil {
		panic(fmt.Sprintf("invalid string for DotEight: %s\n", s))
	}
	right := 0
	if len(comps) == 2 {
		rightStr := (comps[1] + "00000000")[0:8]
		right, err = strconv.Atoi(rightStr)
		if err != nil {
			panic(fmt.Sprintf("invalid string for DotEight: %s\n", s))
		}
	}
	dotEight := int64(left) * int64(1e8)
	dotEight = dotEight + int64(right)
	return DotEight(dotEight)
}
