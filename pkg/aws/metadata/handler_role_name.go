// Copyright 2017 uSwitch
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
package metadata

import (
	"context"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/uswitch/kiam/pkg/k8s"
	"net/http"
	"time"
)

type roleHandler struct {
	roleFinder  k8s.RoleFinder
	clientIP    clientIPFunc
	defaultRole string
}

func (h *roleHandler) Handle(ctx context.Context, w http.ResponseWriter, req *http.Request) (int, error) {
	roleNameTimings := metrics.GetOrRegisterTimer("roleNameHandler", metrics.DefaultRegistry)
	startTime := time.Now()
	defer roleNameTimings.UpdateSince(startTime)

	err := req.ParseForm()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ip, err := h.clientIP(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	role, err := findRole(ctx, h.roleFinder, ip)
	if err != nil {
		metrics.GetOrRegisterMeter("roleNameHandler.findRoleError", metrics.DefaultRegistry).Mark(1)
		return http.StatusInternalServerError, err
	}

	if role == "" && h.defaultRole != "" {
		metrics.GetOrRegisterMeter("roleNameHandler.defaultRole", metrics.DefaultRegistry).Mark(1)
		fmt.Fprint(w, h.defaultRole)
		return http.StatusOK, nil
	}

	if role == "" {
		metrics.GetOrRegisterMeter("credentialsHandler.emptyRole", metrics.DefaultRegistry).Mark(1)
		return http.StatusNotFound, EmptyRoleError
	}

	fmt.Fprint(w, role)
	metrics.GetOrRegisterMeter("roleNameHandler.success", metrics.DefaultRegistry).Mark(1)

	return http.StatusOK, nil
}
