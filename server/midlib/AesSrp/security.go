//
// Package aessrp implements encrypted authentication and encrypted REST.
// SRP-6a for login authentication, followed by AES 256 bit encrypted RESTful calls.
// A security model with roles is also implemented.
//
// Copyright (C) Philip Schlump, 2013-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 0001
//

package AesSrp

import "github.com/pschlump/godebug" //

type SecurityConfigType struct {
	Roles        []string
	AccessLevels map[string][]string
	Privilages   map[string][]string
	MayAccessApi map[string][]string
}

type RolesWithBitMask struct {
	Name    string `json:"title"`
	BitMask uint64 `json:"bitMask"`
}

// This is per-server
func SetupRoles(rolesName []string, accessLevels map[string][]string) ([]RolesWithBitMask, map[string]uint64, []RolesWithBitMask) {

	var bm uint64 = 1
	rn := make([]RolesWithBitMask, 0, len(rolesName))
	rn_h := make(map[string]uint64)
	for _, vv := range rolesName {
		rn = append(rn, RolesWithBitMask{Name: vv, BitMask: bm})
		rn_h[vv] = bm
		bm = bm << 1
	}

	godebug.Db2Printf(db_SetupRoles, "rn=%+v, rn_h=%+v, %s\n", rn, rn_h, godebug.LF())

	an := make([]RolesWithBitMask, 0, len(rolesName))
	for kk, vv := range accessLevels {
		t := uint64(0)
		for _, ww := range vv {
			t |= rn_h[ww]
		}
		an = append(an, RolesWithBitMask{Name: kk, BitMask: t})
	}

	godebug.Db2Printf(db_SetupRoles, "an=%+v\n", an)

	return rn, rn_h, an
}
