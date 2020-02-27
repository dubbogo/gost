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

package gxpage

// DefaultPage is the default implementation of Page interface
type DefaultPage struct {
	requestOffset int
	pageSize      int
	totalSize     int
	data          []interface{}
	totalPages    int
	hasNext       bool
}

// GetOffSet will return the offset
func (d *DefaultPage) GetOffset() int {
	return d.requestOffset
}

// GetPageSize will return the page size
func (d *DefaultPage) GetPageSize() int {
	return d.pageSize
}

// GetTotalPages will return the number of total pages
func (d *DefaultPage) GetTotalPages() int {
	return d.totalPages
}

// GetData will return the data
func (d *DefaultPage) GetData() []interface{} {
	return d.data
}

// GetDataSize will return the size of data.
// it's len(GetData())
func (d *DefaultPage) GetDataSize() int {
	return len(d.GetData())
}

// HasNext will return whether has next page
func (d *DefaultPage) HasNext() bool {
	return d.hasNext
}

// HasData will return whether this page has data.
func (d *DefaultPage) HasData() bool {
	return d.GetDataSize() > 0
}

func NewDefaultPage(requestOffset int, pageSize int,
	data []interface{}, totalSize int) *DefaultPage {

	remain := totalSize % pageSize
	totalPages := totalSize / pageSize
	if remain > 0 {
		totalPages++
	}

	hasNext := totalSize-requestOffset-pageSize > 0

	return &DefaultPage{
		requestOffset: requestOffset,
		pageSize:      pageSize,
		data:          data,
		totalSize:     totalSize,
		totalPages:    totalPages,
		hasNext:       hasNext,
	}
}