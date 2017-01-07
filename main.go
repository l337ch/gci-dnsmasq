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
	"os"
	"os/exec"
)

const (
	// DNS names we resolve
	gciDefaultMetadataHost = "metadata.google.internal"
	kubednsSvcName         = "kube-dns"
	kubednsNamespace       = ".kube-system"
	kubernetesSvc          = ".svc"
	kubernetesDomain       = ".cluster.local"
	kubednsHostname        = kubednsSvcName + kubednsNamespace + kubernetesSvc + kubernetesDomain

	etcResolvConf = "/etc/resolv.conf"
)

const (
	cmdDnsmasq    = "/usr/sbin/dnsmasq"
	argForeground = "--keep-in-foreground"
	argNoDaemon   = "--no-daemon"
	argNoResolv   = "--no-resolv"
	argServer     = "--server=/redacted.com/10.10.10.10"
)

var (
	dnsClusterIP   string
	metaResolverIP string
	envArgs        string
)

func init() {
	log.SetFlags(0)

	envArgs := os.Getenv("DNSMASQ_CMD_ARGS")
	log.Print(fmt.Sprintf("gci-dnsmasq: DNSMASQ_CMD_ARGS: %s", envArgs))
}

func getDNSSingleIP(question string) string {
	addrs, err := net.LookupHost(question)
	if err != nil || len(addrs) == 0 {
		return ""
	}
	return addrs[0]
}

// ValidateResolvConf checks to ensure that a desired nameserver entry exists
func ValidateResolvConf(nameserverIP string) bool {
	// Validations:
	// 0 - /etc/resolv.conf is present and not empty
	// 1 - at least one 'nameserver' is present for the nameserver IP argument
	// 2 - TODO: return a list of other nameservers found that don't match
	fileData, err := ioutil.ReadFile(etcResolvConf)
	if err != nil {
		log.Print(fmt.Sprintf("gci-dnsmasq: ValidateResolveConf error: %v", err))
		return false
	}

	// -- debug cruft
	// log.Print(">>>>>>>>>>>>>>>>>>>>>>>>> /etc/resolv.conf >>>>>>>>>>>>>>>>>>>>>>>>>\n")
	// log.Print(fmt.Sprintf("%s", string(fileData)))
	// log.Print(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")

	matchIP := []byte(nameserverIP)
	for _, fileLine := range bytes.Split(fileData, []byte("\n")) {
		if bytes.Contains(fileLine, matchIP) {
			return true
		}
	}
	return false
}

func main() {
	// Query DNS for the current kube-dns service cluster ip
	// Yes, we need DNS to, configure DNS... shhhh....
	metaResolverIP = getDNSSingleIP(gciDefaultMetadataHost)
	if len(metaResolverIP) == 0 {
		log.Fatal(fmt.Sprintf("gci-dnsmasq: failed to resolve: %s, can't continue without a valid value for DNS resolver for forwarding",
			gciDefaultMetadataHost))
	}
	log.Print(fmt.Sprintf("gci-dnsmasq: metadata server as resolver ip: %s\n", metaResolverIP))

	dnsClusterIP = getDNSSingleIP(kubednsHostname)
	if len(dnsClusterIP) == 0 {
		log.Fatal(fmt.Sprintf("gci-dnsmasq: failed to resolve: %s, can't continue without a valid value for kube-dns service ClusterIP",
			kubednsHostname))
	}
	log.Print(fmt.Sprintf("gci-dnsmasq: kube-dns service cluster ip (vip): %s\n", dnsClusterIP))

	isPresent := ValidateResolvConf(dnsClusterIP)
	log.Print(fmt.Sprintf("gci-dnsmasq: kube-dns nameserver present in /etc/resolv.conf: %t\n", isPresent))

	cmd := exec.Command(cmdDnsmasq, argForeground, argServer)
	if len(envArgs) != 0 {
		cmd = exec.Command(cmdDnsmasq, envArgs)
	}

	log.Print(fmt.Sprintf("gci-dnsmasq: starting dnsmasq: cmd: %s argv: %v", cmdDnsmasq, cmd.Args))

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	if err := cmd.Run(); err != nil {
		log.Fatal(fmt.Sprintf("gci-dnsmasq: error return: %v, stderr: %s, stdout: %s", err, stderrBuf, stdoutBuf))
	}
	log.Print(fmt.Sprintf("gci-dnsmasq: dnsmasq: %s", stdoutBuf))
}
