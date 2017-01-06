#
# Copyright © 2016 Samsung CNCT
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License. 
#
# Dockerfile - GCI DNS Redirector
#
# - Example Commands:
# docker build --rm -t sostheim/gci-dnsmasq .
#
FROM alpine:3.4
MAINTAINER Rick Sostheim
LABEL vendor="Samsung CNCT"

# Add dnsmasq (verison is currently 2.76 for Alpine:3.4)
RUN apk --update --no-cache add dnsmasq

# Redirect Domain List
ENV DOMAIN_LIST=example.com

COPY gci_dnsmasq /
ENTRYPOINT ["/gci_dnsmasq"]
