//
// Copyright © 2016 Samsung CNCT
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
//
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
)

const (
	// GCI Metadata Server Default Values
	gciMetadataFlavorHeader      = "Metadata-Flavor"
	gciMetadataFlavorHeaderValue = "Google"
	gciDefaultScheme             = "http"
	gciDefaultMetadataHost       = "metadata.google.internal"

	// Our primary attribute collection of interest.
	gciDefaultURI = gciDefaultScheme + "://" + gciDefaultMetadataHost

	kubednsSvcName   = "kube-dns"
	kubednsNamespace = ".kube-system"
	kubernetesSvc    = ".svc"
	kubernetesDomain = ".cluster.local."
	kubednsQuestion  = kubednsSvcName + kubednsNamespace + kubernetesSvc + kubernetesDomain

	etcResolvConf = "/etc/resolv.conf"
)

const (
	cmdDnsmasq    = "/usr/sbin/dnsmasq"
	argForeground = "--keep-in-foreground"
	argNoDaemon   = "--no-daemon"
	argNoResolv   = "--no-resolv"
	argServer     = "--server=/<domain>/<dns server ip>"
)

var (
	dnsClusterIP   string
	metaResolverIP string
)

func init() {
	log.SetFlags(0)
}

func getDNSSingleIP(question string) string {
	addrs, err := net.LookupHost(question)
	if err != nil || len(addrs) == 0 {
		return ""
	}
	return addrs[0]
}

// ValidateResolvConf checks to ensure that the current metadata
// servers IP address is present as a nameserver entry.
func ValidateResolvConf(metaIP string) bool {
	// Validations:
	// 0 - /etc/resolv.conf is present and not empty
	// 1 - at least one 'nameserver' is present for the metadata server's IP
	// 2 - TODO: return a list of other nameservers found that don't match

	fileData, err := ioutil.ReadFile(etcResolvConf)
	if err != nil {
		log.Print(fmt.Sprintf("gci-dnsmas: ValidateResolveConf error: %v", err))
		return false
	}
	for _, fileLine := range strings.Split(string(fileData), "\n") {
		if strings.Contains(fileLine, metaIP) {
			return true
		}
	}
	return false
}

func main() {
	// Query DNS for the current kube-dns service cluster ip
	// Yes, we need DNS to, configure DNS... shhhh....
	dnsClusterIP = getDNSSingleIP(kubednsQuestion)
	if len(dnsClusterIP) == 0 {
		log.Fatal("Can't continue without a valid value for kube-dns service ClusterIP Range CIDR")
	}
	log.Print(fmt.Sprintf("gci-dnsmasq: kube-dns service cluster ip: %s\n", dnsClusterIP))

	// Query DNS for the current kube-dns service cluster ip
	// Yes, we need DNS to, configure DNS... shhhh....
	metaResolverIP = getDNSSingleIP(gciDefaultMetadataHost)
	if len(dnsClusterIP) == 0 {
		log.Fatal("Can't continue without a valid value for kube-dns service ClusterIP Range CIDR")
	}
	log.Print(fmt.Sprintf("gci-dnsmasq: metadata server as resolver ip: %s\n", metaResolverIP))

	cmd := exec.Command(cmdDnsmasq, argForeground, argNoDaemon, argServer)
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	if err := cmd.Run(); err != nil {
		log.Fatal(fmt.Sprintf("gci-dnsmasq: error return: %v, stderr: %s, stdout: %s", err, stderrBuf, stdoutBuf))
	}
	log.Print(fmt.Sprintf("gci-dnsmasq: dnsmasq: %s", stdoutBuf))
}
