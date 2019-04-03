/*
Copyright 2019 The NSQ-Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"bytes"
	"testing"
)

func TestHash(t *testing.T) {
	first := map[string]string{
		"a": "x",
		"b": "y",
		"c": "z",
	}
	second := map[string]string{
		"b": "y",
		"c": "z",
		"a": "x",
	}

	third := map[string]string{
		"c": "z",
		"a": "x",
		"b": "y",
	}

	firstHash, err := Hash(first)
	if err != nil {
		t.Fatalf("Hash first content error: %v", err)
	}
	secondHash, err := Hash(second)
	if err != nil {
		t.Fatalf("Hash second content error: %v", err)
	}
	thirdHash, err := Hash(third)
	if err != nil {
		t.Fatalf("Hash third content error: %v", err)
	}

	if !(bytes.Equal(firstHash, secondHash) && bytes.Equal(secondHash, thirdHash)) {
		t.Fatalf("hash different sequence map with same key/value pairs error.\n"+
			"First: %v, Hash: %v.\n"+
			"Second: %v, Hash: %v.\n"+
			"Third: %v, Hash: %v.\n", first, firstHash, second, secondHash, third, thirdHash)
	}
}
