/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gxprotocol

// Result is a RPC result
type Result interface {
	// SetError sets error.
	SetError(error)
	// Error gets error.
	Error() error
	// SetResult sets invoker result.
	SetResult(interface{})
	// Result gets invoker result.
	Result() interface{}
	// SetAttachments replaces the existing attachments with the specified param.
	SetAttachments(map[string]interface{})
	// Attachments gets all attachments
	Attachments() map[string]interface{}

	// AddAttachment adds the specified map to existing attachments in this instance.
	AddAttachment(string, interface{})
	// Attachment gets attachment by key with default value.
	Attachment(string, interface{}) interface{}
}
