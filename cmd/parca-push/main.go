// Copyright (c) 2023 The Parca Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/google/pprof/profile"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grun "github.com/oklog/run"
	profilestorepb "github.com/parca-dev/parca/gen/proto/go/parca/profilestore/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type flags struct {
	Path              string            `kong:"arg,help='Path to the profile data.'"`
	Labels            map[string]string `kong:"help='Labels to attach to the profile data. For example --labels=__name__=process_cpu --labels=node=foo',short='l'"`
	Normalized        bool              `kong:"help='Whether the profile sample addresses are already normalized by the mapping offset.',default='false'"`
	OverrideTimestamp bool              `kong:"help='Update the timestamp in the pprof profile to be the current time.'"`

	RemoteStore FlagsRemoteStore `embed:"" prefix:"remote-store-"`
}

// FlagsRemoteStore provides remote store configuration flags.
type FlagsRemoteStore struct {
	Address            string `kong:"help='gRPC address to send profiles and symbols to.'"`
	BearerToken        string `kong:"help='Bearer token to authenticate with store.'"`
	BearerTokenFile    string `kong:"help='File to read bearer token from to authenticate with store.'"`
	Insecure           bool   `kong:"help='Send gRPC requests via plaintext instead of TLS.'"`
	InsecureSkipVerify bool   `kong:"help='Skip TLS certificate verification.'"`
}

func main() {
	flags := flags{}
	kong.Parse(&flags)
	if err := run(flags); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(flags flags) error {
	var g grun.Group
	ctx, cancel := context.WithCancel(context.Background())
	g.Add(func() error {
		labels := []*profilestorepb.Label{}
		for name, value := range flags.Labels {
			labels = append(labels, &profilestorepb.Label{
				Name:  name,
				Value: value,
			})
		}

		conn, err := grpcConn(prometheus.NewRegistry(), flags.RemoteStore)
		if err != nil {
			return fmt.Errorf("create gRPC connection: %w", err)
		}
		defer conn.Close()

		var profileContent []byte
		if flags.Path == "-" {
			profileContent, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read profile from stdin: %w", err)
			}
		} else {
			profileContent, err = os.ReadFile(flags.Path)
			if err != nil {
				return fmt.Errorf("read profile file: %w", err)
			}
		}

		p, err := profile.ParseData(profileContent)
		if err != nil {
			return fmt.Errorf("parse pprof profile: %w", err)
		}

		if flags.OverrideTimestamp {
			now := time.Now()
			p.TimeNanos = now.UnixNano()
			buf := bytes.NewBuffer(nil)
			if err := p.Write(buf); err != nil {
				return fmt.Errorf("serialize pprof profile: %w", err)
			}
			profileContent = buf.Bytes()
		}

		profilestoreClient := profilestorepb.NewProfileStoreServiceClient(conn)
		_, err = profilestoreClient.WriteRaw(ctx, &profilestorepb.WriteRawRequest{
			Series: []*profilestorepb.RawProfileSeries{{
				Labels: &profilestorepb.LabelSet{
					Labels: labels,
				},
				Samples: []*profilestorepb.RawSample{{
					RawProfile: profileContent,
				}},
			}},
			Normalized: flags.Normalized,
		})
		if err != nil {
			return fmt.Errorf("write profile: %w", err)
		}

		return nil
	}, func(error) {
		cancel()
	})

	g.Add(grun.SignalHandler(ctx, os.Interrupt, os.Kill))
	return g.Run()
}

func grpcConn(reg prometheus.Registerer, flags FlagsRemoteStore) (*grpc.ClientConn, error) {
	met := grpc_prometheus.NewClientMetrics()
	met.EnableClientHandlingTimeHistogram()
	reg.MustRegister(met)

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(
			met.UnaryClientInterceptor(),
		),
	}
	if flags.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		config := &tls.Config{
			//nolint:gosec
			InsecureSkipVerify: flags.InsecureSkipVerify,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	}

	if flags.BearerToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&perRequestBearerToken{
			token:    flags.BearerToken,
			insecure: flags.Insecure,
		}))
	}

	if flags.BearerTokenFile != "" {
		b, err := os.ReadFile(flags.BearerTokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read bearer token from file: %w", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(&perRequestBearerToken{
			token:    string(b),
			insecure: flags.Insecure,
		}))
	}

	return grpc.Dial(flags.Address, opts...)
}

type perRequestBearerToken struct {
	token    string
	insecure bool
}

func (t *perRequestBearerToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t *perRequestBearerToken) RequireTransportSecurity() bool {
	return !t.insecure
}
