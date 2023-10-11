/*   
   Copyright 2023 Rodolfo González González for the modifications in
   this fork.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package gormcache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	bolt "go.etcd.io/bbolt"
)

// BboltClient is a wrapper for bbolt client
type BboltClient struct {
	client *bolt.DB
}

// NewBboltClient returns a new RedisClient instance
func NewBboltClient(client *bolt.DB) *BboltClient {
	return &BboltClient{
		client: client,
	}
}

// Get gets value from bbolt by key using json encoding/decoding
func (r *BboltClient) Get(ctx context.Context, key string) (interface{}, error) {
	var data []byte
	if r.client == nil {
		log.Fatalln("NO CLIENT")
	}
	err := r.client.View(func(tx *bolt.Tx) error {
		data = tx.Bucket([]byte("DB")).Get([]byte(key))
		return nil
	})

	if err != nil || len(data) == 0 {
		return nil, nil
	}
	//log.Printf("get cache %v, key: %v", data, key)
	return data, nil
}

// Set sets value to bbolt by key with ttl using json encoding/decoding
func (r *BboltClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value) // encode value to json bytes using json encoding/decoding
	if err != nil {
		return err
	}
	err = r.client.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte("DB")).Put([]byte(key), data)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
