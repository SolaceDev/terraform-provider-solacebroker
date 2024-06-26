// terraform-provider-solacebroker
//
// Copyright 2024 Solace Corporation. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package generator

import "testing"

func TestGenerateTerraformFile(t *testing.T) {
	type args struct {
		terraformObjectInfo *ObjectInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"CanGenerateFile",
			args{terraformObjectInfo: &ObjectInfo{
				BasicAuthentication: true,
				FileName:            "/tmp/somefile.tf",
				BrokerResources:     []map[string]string{}},
			},
			false,
		},
		{
			"FailToGenerateFile",
			args{terraformObjectInfo: &ObjectInfo{}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenerateTerraformFile(tt.args.terraformObjectInfo); (err != nil) != tt.wantErr {
				t.Errorf("GenerateTerraformFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
