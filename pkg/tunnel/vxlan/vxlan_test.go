// Copyright 2023 FabEdge Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vxlan

import (
	"github.com/fabedge/fabedge/pkg/tunnel"
	"github.com/vishvananda/netlink"
	"net"
	"testing"
)

func TestCreateMulticastVxlan(t *testing.T) {
	config := VxlanMulticastConfig{
		Name:         "vxlantest0",
		VtepAddress:  "192.168.222.10/24",
		GroupAddress: net.IPv4(239, 0, 0, 1),
	}
	m := VxlanManager{
		MTU:         1450,
		VNI:         10,
		Port:        4789,
		VtepDevName: GetFirstInterface(),
	}
	err := m.createMulticastVxlan(config)
	if err != nil {
		t.Error(err)
	}

	link, err := netlink.LinkByName(config.Name)
	if link == nil || err != nil {
		t.Error(err)
	}

	err = m.deleteVxlan(config.Name)
	if err != nil {
		t.Error(err)
	}

	link, err = netlink.LinkByName(config.Name)
	if link != nil {
		t.Error("error deleting vxlan")
	}
}

func TestCreateUnicastVxlan(t *testing.T) {
	config := VxlanUnicastConfig{
		Name:          "vxlantest0",
		VtepAddress:   "192.168.222.10/24",
		LocalAddress:  "192.168.2.50",
		RemoteAddress: "192.168.2.61",
	}
	m := VxlanManager{
		MTU:         1450,
		VNI:         10,
		Port:        4789,
		VtepDevName: GetFirstInterface(),
	}
	err := m.createUnicastVxlan(config)
	if err != nil {
		t.Error(err)
	}

	link, err := netlink.LinkByName(config.Name)
	if link == nil || err != nil {
		t.Error(err)
	}

	err = m.deleteVxlan(config.Name)
	if err != nil {
		t.Error(err)
	}

	link, err = netlink.LinkByName(config.Name)
	if link != nil {
		t.Error("error deleting vxlan")
	}
}

func TestVxlanManager(t *testing.T) {
	m := CreateVxlanManager(1450, 10, 4789, GetFirstInterface())
	conn := tunnel.ConnConfig{
		Name:            "vxlantest0",
		LocalAddress:    []string{"192.168.2.61"},
		RemoteAddress:   []string{"192.168.2.50"},
		EndpointAddress: "192.168.222.10/24",
	}
	err := m.LoadConn(conn)
	if err != nil {
		t.Error(err)
	}
	err = m.InitiateConn(conn.Name)
	if err != nil {
		t.Error(err)
	}

	link, err := netlink.LinkByName(conn.Name)
	if link == nil || err != nil {
		t.Error(err)
	}

	err = m.UnloadConn(conn.Name)
	if err != nil {
		t.Error(err)
	}

	link, err = netlink.LinkByName(conn.Name)
	if link != nil {
		t.Error("error deleting vxlan")
	}
}
