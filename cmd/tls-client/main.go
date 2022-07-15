/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

const (
	CONFIG_PATH = "../../configs"
	CONFIG_NAME = "app"
)

func main() {
	flag.Parse()

	log.Println("Client running ...")

	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	rpcCreds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: "client-x-id"})
	trnCreds, err := credentials.NewClientTLSFromFile("../../certs/server.pem", "localhost")
	if err != nil {
		log.Fatalln(err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(trnCreds),
		grpc.WithPerRPCCredentials(rpcCreds),
	}
	// opts = append(opts, grpc.WithBlock())

	// Set up a connection to the server.
	conn, err := grpc.Dial(cfg.Grpc.Addr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := cc.NewCloudControlServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := cc.APICommand{
		Service:  "EC2",
		Resource: "SSHKeypair",
		Action:   "List",
		Params:   []string{"4608f97b-d5d4-11ec-a835-0242ac110002", "admin"},
	}

	resp, err := c.UnaryCall(ctx, &cc.APIRequest{JobID: uuid.NewString(), Cmd: &cmd})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Server response: %s", resp)
}
