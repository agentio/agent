//
//  Copyright 2014 Radtastical Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
package agent

// Ubuntu-specific configuration
const ControlPath = "/home/control"
const PathForNginx = "/usr/sbin/nginx"
const PathForNginxConfiguration = "/etc/nginx/nginx.conf"
const PathForUpstartConfiguration = "/etc/init"

// Installation-specific configuration
const HostName = "alpha.agent.io"

// Internal constants
const PasswordSalt = "agent.io"
