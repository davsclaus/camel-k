/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trait

import (
	"sort"

	"github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/apache/camel-k/pkg/builder"
	"github.com/apache/camel-k/pkg/builder/springboot"
	"github.com/apache/camel-k/pkg/util"
)

type springBootTrait struct {
	BaseTrait `property:",squash"`
}

func newSpringBootTrait() *springBootTrait {
	return &springBootTrait{
		BaseTrait: newBaseTrait("springboot"),
	}
}

// IsAuto determines if we should apply automatic configuration
func (trait *springBootTrait) IsAuto() bool {
	return false
}

// IsEnabled is used to determine if the trait needs to be executed
func (trait *springBootTrait) IsEnabled() bool {
	if trait.Enabled == nil {
		return false
	}
	return *trait.Enabled
}

func (trait *springBootTrait) appliesTo(e *Environment) bool {
	if e.Context != nil && e.Context.Status.Phase == v1alpha1.IntegrationContextPhaseBuilding {
		return true
	}
	if e.Integration != nil && e.Integration.Status.Phase == v1alpha1.IntegrationPhaseDeploying {
		return true
	}
	if e.Integration != nil && e.Integration.Status.Phase == "" {
		return true
	}

	return false
}

func (trait *springBootTrait) apply(e *Environment) error {

	//
	// Integration
	//

	if e.Integration != nil && e.Integration.Status.Phase == "" {
		util.StringSliceUniqueAdd(&e.Integration.Spec.Dependencies, "runtime:spring-boot")

		// sort the dependencies to get always the same list if they don't change
		sort.Strings(e.Integration.Spec.Dependencies)
	}

	if e.Integration != nil && e.Integration.Status.Phase == v1alpha1.IntegrationPhaseDeploying {
		// Override env vars
		e.EnvVars["JAVA_MAIN_CLASS"] = "org.springframework.boot.loader.PropertiesLauncher"
		e.EnvVars["LOADER_PATH"] = "/deployments/dependencies/"
	}

	//
	// Integration Context
	//

	if e.Context != nil && e.Context.Status.Phase == v1alpha1.IntegrationContextPhaseBuilding {
		// add custom initialization logic
		e.Steps = append(e.Steps, builder.NewStep("initialize/spring-boot", builder.IntiPhase, springboot.Initialize))
		e.Steps = append(e.Steps, builder.NewStep("build/compute-boot-dependencies", builder.ProjectBuildPhase+1, springboot.ComputeDependencies))

		// replace project generator
		for i := 0; i < len(e.Steps); i++ {
			if e.Steps[i].Phase() == builder.ProjectGenerationPhase {
				e.Steps[i] = builder.NewStep("generate/spring-boot", builder.ProjectGenerationPhase, springboot.GenerateProject)
			}
		}
	}

	return nil
}
