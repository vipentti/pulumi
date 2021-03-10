// Copyright 2016-2021, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package encoding

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ValidateFileEquality compares maps of files for equality.
func ValidateFileEquality(t *testing.T, actual, expected map[string][]byte) {
	for name, file := range expected {
		assert.Contains(t, actual, name)
		assert.Equal(t, string(file), string(actual[name]), name)
	}
}

func mustLoadFile(dir, file string) []byte {
	fileBytes, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		panic(err)
	}

	return fileBytes
}

func TestRoundTrip(t *testing.T) {
	files := []string{
		"Pulumi.trivial-dev.yaml",
		"Pulumi.simple-dev.yaml", // Currently failing due to a marshaling bug (extra newline)
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			yamlBytes := mustLoadFile("internal", file)
			fileAST, err := NewFileAST(yamlBytes)
			assert.NoError(t, err)
			out := fileAST.Marshal()
			assert.Equal(t, string(yamlBytes), string(out), file)
		})
	}
}
