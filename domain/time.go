// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

import "time"

// Bogota is Colombia's time zone: fixed UTC-5, no daylight saving. time.FixedZone is used
// instead of time.LoadLocation("America/Bogota") so this doesn't depend on the system (or the
// binary) having the IANA time zone database loaded.
var Bogota = time.FixedZone("America/Bogota", -5*60*60)
