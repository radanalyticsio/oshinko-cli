/*
Copyright 2017 The Kubernetes Authors.

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

package controller

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func TestBuildOriginatingIdentity(t *testing.T) {
	userInfo := v1beta1.UserInfo{
		Username: "person@place.com",
		UID:      "abcd-1234",
		Groups:   []string{"stuff-dev", "main-eng"},
		Extra:    map[string]v1beta1.ExtraValue{"foo": {"bar", "baz"}},
	}

	e := osb.OriginatingIdentity{
		Platform: "kubernetes",
		Value:    `{"foo":["bar","baz"],"groups":["stuff-dev","main-eng"],"uid":"abcd-1234","username":"person@place.com"}`,
	}

	g, err := buildOriginatingIdentity(&userInfo)

	if err != nil {
		t.Fatalf("Unexpected Error, %+v", err)
	}

	if e.Platform != g.Platform {
		t.Fatalf("Unexpected Platform, %s", expectedGot(e.Platform, g.Platform))
	}

	if e.Value != g.Value {
		t.Fatalf("Unexpected Value, %s", expectedGot(e.Value, g.Value))
	}
}
