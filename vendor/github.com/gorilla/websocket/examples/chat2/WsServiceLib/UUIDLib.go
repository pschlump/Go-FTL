package WsServiceLib

//
// Copyright (C) Philip Schlump, 2015-2017.
//

import (
	"encoding/json"
	"fmt"
	"os"

	newUuid "github.com/pborman/uuid" // Modified pool to have NewAuth for authorized connections
	"github.com/pschlump/uuid"
	"github.com/taskcluster/slugid-go/slugid"
)

/*
-------------------------------------------------------
UUID Notes
-------------------------------------------------------

import (
	newUuid "github.com/pborman/uuid"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/taskcluster/slugid-go/slugid"
)

	id0 := lib.GetUUIDAsString()
	id0_slug := UUIDToSlug(id0)
*/

func UUIDToSlug(uuid string) (slug string) {
	// slug = id
	uuidType := newUuid.Parse(uuid)
	if uuidType != nil {
		slug = slugid.Encode(uuidType)
		return
	}
	fmt.Fprintf(os.Stderr, "slug: ERROR: Cannot encode invalid uuid '%v' into a slug\n", uuid) // Xyzzy - logrus
	return
}

func SlugToUUID(slug string) (uuid string) {
	// uuid = slug
	uuidType := slugid.Decode(slug)
	if uuidType != nil {
		uuid = uuidType.String()
		return
	}
	fmt.Fprintf(os.Stderr, "slug: ERROR: Cannot decode invalid slug '%v' into a UUID\n", slug) // Xyzzy - logrus
	return
}

func StructToJson(mm interface{}) (rv string) {
	trv, err := json.Marshal(mm)
	if err != nil {
		rv = "{}"
		return
	}
	rv = string(trv)
	return
}

const ReturnPacked = true

func UUIDAsStr() (s_id string) {
	id, _ := uuid.NewV4()
	s_id = id.String()
	return
}

func UUIDAsStrPacked() (s_id string) {
	if ReturnPacked {
		s := UUIDAsStr()
		return UUIDToSlug(s)
	} else {
		return UUIDAsStr()
	}
}
