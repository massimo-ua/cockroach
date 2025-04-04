// Copyright 2023 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

// Code generated by "stringer"; DO NOT EDIT.

package sql

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[implicitTxn-0]
	_ = x[explicitTxn-1]
	_ = x[upgradedExplicitTxn-2]
}

func (i txnType) String() string {
	switch i {
	case implicitTxn:
		return "implicitTxn"
	case explicitTxn:
		return "explicitTxn"
	case upgradedExplicitTxn:
		return "upgradedExplicitTxn"
	default:
		return "txnType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
