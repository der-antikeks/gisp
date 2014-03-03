package ecs

import (
// "testing"
)

const (
	TestAType ComponentType = iota
	TestBType
	TestCType
)

type TestAComponent struct{ Value float64 }

func (c TestAComponent) Type() ComponentType { return TestAType }

type TestBComponent struct{ Value float64 }

func (c TestBComponent) Type() ComponentType { return TestBType }

type TestCComponent struct{ Value float64 }

func (c TestCComponent) Type() ComponentType { return TestCType }
