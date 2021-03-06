// Copyright (c) 2013,2014 SmugMug, Inc. All rights reserved.
// 
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//     * Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
//       copyright notice, this list of conditions and the following
//       disclaimer in the documentation and/or other materials provided
//       with the distribution.
// 
// THIS SOFTWARE IS PROVIDED BY SMUGMUG, INC. ``AS IS'' AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
// PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL SMUGMUG, INC. BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE
// GOODS OR SERVICES;LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
// IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
// OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
// ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"net/http"
	ep "github.com/smugmug/godynamo/endpoint"
	batch_write_item "github.com/smugmug/godynamo/endpoints/batch_write_item"
	conf_iam "github.com/smugmug/godynamo/conf_iam"
	"github.com/smugmug/godynamo/conf"
	"github.com/smugmug/godynamo/conf_file"
)

// this tests "RetryBatchWrite", which does NOT do intelligent splitting and re-assembling
// of requests and responses
func Test1() {
	conf_file.Read()
	if conf.Vals.Initialized == false {
		panic("the conf.Vals global conf struct has not been initialized")
	}

	iam_ready_chan := make(chan bool)
	go conf_iam.GoIAM(iam_ready_chan)
	iam_ready := <- iam_ready_chan
	if iam_ready {
		fmt.Printf("using iam\n")
	} else {
		fmt.Printf("not using iam\n")
	}

	tn := "test-godynamo-livetest"
	b := batch_write_item.NewBatchWriteItem()
	b.RequestItems[tn] = make([]batch_write_item.RequestInstance,0)
	for i := 1; i <= 300; i++ {
		var p batch_write_item.PutRequest
		p.Item = make(ep.Item)
		k := fmt.Sprintf("AHashKey%d",i)
		v := fmt.Sprintf("%d",i)
		p.Item["TheHashKey"] = ep.AttributeValue{S:k}
		p.Item["TheRangeKey"] = ep.AttributeValue{N:v}
		p.Item["SomeValue"] = ep.AttributeValue{N:v}
		b.RequestItems[tn] =
			append(b.RequestItems[tn],
			batch_write_item.RequestInstance{PutRequest:&p})
	}
	bs,_ := batch_write_item.Split(*b)
	for _,bsi := range bs {
		body,code,err := bsi.RetryBatchWrite(0)
		if err != nil || code != http.StatusOK {
			fmt.Printf("error: %v\n%v\n%v\n",body,code,err)
		} else {
			fmt.Printf("worked!: %v\n%v\n%v\n",body,code,err)
		}
	}
}

// this tests "DoBatchGet", which breaks up requests that are larger than the limit
// and re-assembles responses
func Test2() {
	b := batch_write_item.NewBatchWriteItem()
	tn := "test-godynamo-livetest"
	b.RequestItems[tn] = make([]batch_write_item.RequestInstance,0)
	for i := 201; i <= 300; i++ {
		var p batch_write_item.PutRequest
		p.Item = make(ep.Item)
		k := fmt.Sprintf("AHashKey%d",i)
		v := fmt.Sprintf("%d",i)
		p.Item["TheHashKey"] = ep.AttributeValue{S:k}
		p.Item["TheRangeKey"] = ep.AttributeValue{N:v}
		p.Item["SomeValue"] = ep.AttributeValue{N:v}
		b.RequestItems[tn] =
			append(b.RequestItems[tn],
			batch_write_item.RequestInstance{PutRequest:&p})
	}
	body,code,err := b.DoBatchWrite()
	fmt.Printf("%v\n%v\n%v\n",body,code,err)
}

func main() {
	Test1()
	Test2()
}
