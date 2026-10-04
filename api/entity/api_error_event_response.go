// SPDX-License-Identifier: Apache-2.0
package entity

// PostErrorEventResponse confirms durable event acceptance.
type PostErrorEventResponse struct {
	Fingerprint string `json:"fingerprint"`
	Duplicate   bool   `json:"duplicate"`
}
