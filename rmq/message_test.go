// Copyright © 2017 Meltwater
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

package rmq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageFromAttrs(t *testing.T) {
	var headers = make(map[string]string)
	headers["amqp.routingKey"] = "routingKey from tarball header"
	headers["myHeaderString"] = "myString"

	m := *NewMessageFromAttrs([]byte("Message"), headers)

	assert.Equal(t, "routingKey from tarball header", m.RoutingKey)
	assert.Equal(t, []byte("Message"), m.Body)
	assert.Equal(t, "myString", m.Headers["myHeaderString"])
}
