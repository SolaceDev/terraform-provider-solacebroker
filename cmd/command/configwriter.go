// terraform-provider-solacebroker
//
// Copyright 2023 Solace Corporation. All rights reserved.
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
package terraform

import (
	"bytes"
	"embed"
	"os"
	"text/template"
)

var (
	//go:embed templates
	templatefiles embed.FS
)
var terraformTemplate *template.Template
var terraformVariableTemplate *template.Template

func init() {
	var err error
	terraformTemplateString, _ := templatefiles.ReadFile("templates/terraform.template")
	terraformVarsTemplateString, _ := templatefiles.ReadFile("templates/variables.template")
	terraformTemplate, err = template.New("Object Template").Parse(string(terraformTemplateString))
	terraformVariableTemplate, err = template.New("Variable Template").Parse(string(terraformVarsTemplateString))
	if err != nil {
		panic(err)
	}
}

func GenerateTerraformFile(terraformObjectInfo *ObjectInfo) error {
	var codeStream bytes.Buffer
	err := terraformTemplate.Execute(&codeStream, terraformObjectInfo)
	if err != nil {
		LogCLIError("\nError: Templating error : " + err.Error() + "\n\n")
		os.Exit(1)
	}
	GenerateTerraformVariableFile(terraformObjectInfo)
	return os.WriteFile(terraformObjectInfo.FileName, codeStream.Bytes(), 0664)
}

func GenerateTerraformVariableFile(terraformObjectInfo *ObjectInfo) error {
	var codeStreamVariables bytes.Buffer
	err := terraformVariableTemplate.Execute(&codeStreamVariables, terraformObjectInfo)
	if err != nil {
		LogCLIError("\nError: Templating error : " + err.Error() + "\n\n")
		os.Exit(1)
	}
	return os.WriteFile("variables"+terraformObjectInfo.FileName, codeStreamVariables.Bytes(), 0664)
}
