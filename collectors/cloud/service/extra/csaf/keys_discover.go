// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package csaf

import (
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/crypto/openpgp"
	"confirmate.io/collectors/cloud/internal/pointer"
	"confirmate.io/core/api/ontology"

	"github.com/gocsaf/csaf/v3/csaf"
	"github.com/lmittmann/tint"
)

// collectKeys collects the PGP keys and returns the respective keys in the ontology format as well as the keyring
// needed for collecting the Security Advisories (collectSecurityAdvisories)
func (d *csafCollector) collectKeys(pgpkeys []csaf.PGPKey, parentId string) (keys []ontology.IsResource, keyring openpgp.EntityList) {
	for _, pgpkey := range pgpkeys {
		key, openPGPEntity := d.handleKey(pgpkey, parentId)
		keys = append(keys, key)
		keyring = append(keyring, openPGPEntity)
	}
	return
}

// handleKey handles a [csaf.PGPKey]: First we try to fetch the actual key and provide a [openpgp.Entity]. Then we use
// this information as well as the information provided by [csaf.PGPKey] to create a Key ontology object.
func (d *csafCollector) handleKey(pgpkey csaf.PGPKey, parentId string) (key *ontology.Key, openPGPEntity *openpgp.Entity) {
	var (
		err error
		// isAccessible denotes that the key is accessible. We assume it is but set it to false if we could not fetch it
		isAccessible = true
	)

	// 1st: Try to fetch key for creating the OpenGPG entity
	openPGPEntity, err = d.fetchKey(pgpkey)
	if err != nil {
		// If we could not fetch the key we assume that the key exists but is not accessible
		isAccessible = false
		log.Warn("Could not fetch key", slog.String("key url", pointer.Deref(pgpkey.URL)), tint.Err(err))
	}

	// 2nd: Create the key in the ontology format
	key = &ontology.Key{
		Algorithm:                  "PGP",
		Id:                         pointer.Deref(pgpkey.URL),
		Raw:                        collector.Raw(pgpkey),
		ParentId:                   &parentId,
		InternetAccessibleEndpoint: isAccessible,
	}
	return
}

func (d *csafCollector) fetchKey(keyinfo csaf.PGPKey) (key *openpgp.Entity, err error) {
	var (
		res  *http.Response
		keys openpgp.EntityList
	)

	res, err = d.client.Get(pointer.Deref(keyinfo.URL))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	keys, err = openpgp.ReadArmoredKeyRing(res.Body)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, errors.New("no key in key file")
	} else if len(keys) > 1 {
		return nil, errors.New("too many keys in file")
	}

	key = keys[0]

	if !strings.EqualFold(hex.EncodeToString(key.PrimaryKey.Fingerprint), string(keyinfo.Fingerprint)) {
		return nil, errors.New("fingerprints do not match")
	}

	return
}
