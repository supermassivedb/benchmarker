// BSD 3-Clause License
//
// (C) Copyright 2025, Alex Gaetano Padula & SuperMassive authors
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//  1. Redistributions of source code must retain the above copyright notice, this
//     list of conditions and the following disclaimer.
//
//  2. Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//
//  3. Neither the name of the copyright holder nor the names of its
//     contributors may be used to endorse or promote products derived from
//     this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"sync"
	"time"
)

// BenchmarkConfig holds the benchmark configuration
type BenchmarkConfig struct {
	ClusterAddress string  // Address for cluster i.e. localhost:4000
	Username       string  // Username for authentication
	Password       string  // Password for authentication
	NumWorkers     int     // Number of workers, each worker will perform a fraction of the total operations( is a goroutine)
	NumOperations  int     // Total number of operations to perform
	KeySize        int     // Size of the key
	ValueSize      int     // Size of the value
	ReadRatio      float64 // Read operation ratio
	WriteRatio     float64 // Write operation ratio
	DeleteRatio    float64 // Delete operation ratio
	// TLS configuration options
	UseTLS        bool   // Whether to use TLS
	TLSCertFile   string // Path to TLS cert file
	TLSKeyFile    string // Path to TLS key file
	TLSCACertFile string // Path to CA cert file for verifying server
	TLSSkipVerify bool   // Skip TLS verification (not recommended for production)
	TLSServerName string // Server name for TLS verification
}

// OperationStats tracks the operation counts
type OperationStats struct {
	Gets    int        // Number of GET operations
	Puts    int        // Number of PUT operations
	Deletes int        // Number of DELETE operations
	mutex   sync.Mutex // Mutex for stats
}

// Benchmarker entry point
func main() {
	config := BenchmarkConfig{
		ClusterAddress: "localhost:4000",
		Username:       "username", // Just some defaults y'all
		Password:       "password",
		NumWorkers:     8,
		NumOperations:  1000000,
		KeySize:        16,
		ValueSize:      16,
		ReadRatio:      0.8,
		WriteRatio:     0.15,
		DeleteRatio:    0.05,
		// Default TLS settings
		UseTLS:        false,
		TLSSkipVerify: false,
	}

	// Parse flags
	flag.StringVar(&config.ClusterAddress, "cluster", config.ClusterAddress, "Cluster address")
	flag.StringVar(&config.Username, "user", config.Username, "Username for authentication")
	flag.StringVar(&config.Password, "pass", config.Password, "Password for authentication")
	flag.IntVar(&config.NumWorkers, "workers", config.NumWorkers, "Number of workers")
	flag.IntVar(&config.NumOperations, "operations", config.NumOperations, "Total number of operations")
	flag.IntVar(&config.KeySize, "keysize", config.KeySize, "Key size in bytes")
	flag.IntVar(&config.ValueSize, "valuesize", config.ValueSize, "Value size in bytes")
	flag.Float64Var(&config.ReadRatio, "readratio", config.ReadRatio, "Read operation ratio")
	flag.Float64Var(&config.WriteRatio, "writeratio", config.WriteRatio, "Write operation ratio")
	flag.Float64Var(&config.DeleteRatio, "deleteratio", config.DeleteRatio, "Delete operation ratio")

	// TLS flags
	flag.BoolVar(&config.UseTLS, "tls", config.UseTLS, "Use TLS for connections")
	flag.StringVar(&config.TLSCertFile, "tls-cert", config.TLSCertFile, "Path to TLS certificate file")
	flag.StringVar(&config.TLSKeyFile, "tls-key", config.TLSKeyFile, "Path to TLS key file")
	flag.StringVar(&config.TLSCACertFile, "tls-ca-cert", config.TLSCACertFile, "Path to CA certificate file for server verification")
	flag.BoolVar(&config.TLSSkipVerify, "tls-skip-verify", config.TLSSkipVerify, "Skip TLS certificate verification (not recommended)")
	flag.StringVar(&config.TLSServerName, "tls-server-name", config.TLSServerName, "Server name for TLS verification")

	flag.Parse()

	// Run the benchmark
	RunBenchmark(config)
}

// createConnection establishes either a TCP or TLS connection based on configuration
func createConnection(config BenchmarkConfig) (net.Conn, error) {
	if config.UseTLS {
		// Set up TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.TLSSkipVerify,
			ServerName:         config.TLSServerName,
		}

		// Load client certificate if provided
		if config.TLSCertFile != "" && config.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(config.TLSCertFile, config.TLSKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate for server verification if provided
		if config.TLSCACertFile != "" {
			caCert, err := ioutil.ReadFile(config.TLSCACertFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA certificate: %v", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		// Connect with TLS
		conn, err := tls.Dial("tcp", config.ClusterAddress, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("TLS connection failed: %v", err)
		}
		return conn, nil
	}

	// Regular TCP connection
	conn, err := net.Dial("tcp", config.ClusterAddress)
	if err != nil {
		return nil, fmt.Errorf("TCP connection failed: %v", err)
	}
	return conn, nil
}

// RunBenchmark runs the benchmark
func RunBenchmark(config BenchmarkConfig) {
	var wg sync.WaitGroup
	startTime := time.Now()
	stats := OperationStats{}

	// Generate auth string
	authString := fmt.Sprintf("AUTH %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s\x00%s", config.Username, config.Password))))

	// Pre-populate with some data
	populateData(config, authString)

	// Create workers
	operationsPerWorker := config.NumOperations / config.NumWorkers
	for i := 0; i < config.NumWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Create connection (TCP or TLS)
			conn, err := createConnection(config)
			if err != nil {
				fmt.Printf("Worker %d failed to connect: %v\n", workerID, err)
				return
			}
			defer conn.Close()

			// Authenticate with cluster protocol
			fmt.Fprintf(conn, "%s\r\n", authString)

			localStats := OperationStats{}

			for j := 0; j < operationsPerWorker; j++ {
				randomValue := rand.Float64()
				if randomValue < config.ReadRatio {
					key := generateRandomKey(workerID, j, config.KeySize)
					fmt.Fprintf(conn, "GET %s\r\n", key)
					localStats.Gets++
				} else if randomValue < config.ReadRatio+config.WriteRatio {
					key := generateRandomKey(workerID, j, config.KeySize)
					value := generateRandomValue(config.ValueSize)
					fmt.Fprintf(conn, "PUT %s %s\r\n", key, value)
					localStats.Puts++
				} else {
					key := generateRandomKey(workerID, j, config.KeySize)
					fmt.Fprintf(conn, "DEL %s\r\n", key)
					localStats.Deletes++
				}

				buffer := make([]byte, 1024)
				conn.Read(buffer)
			}

			stats.mutex.Lock()
			stats.Gets += localStats.Gets
			stats.Puts += localStats.Puts
			stats.Deletes += localStats.Deletes
			stats.mutex.Unlock()
		}(i)
	}

	wg.Wait()
	elapsedTime := time.Since(startTime)

	// Connection type for reporting
	connectionType := "TCP"
	if config.UseTLS {
		connectionType = "TLS"
	}

	// We report the results..
	fmt.Printf("Benchmark Results (%s):\n", connectionType)
	fmt.Printf("Total Operations: %d\n", config.NumOperations)
	fmt.Printf("  - GET operations: %d\n", stats.Gets)
	fmt.Printf("  - PUT operations: %d\n", stats.Puts)
	fmt.Printf("  - DEL operations: %d\n", stats.Deletes)
	fmt.Printf("Time Elapsed: %v\n", elapsedTime)
	fmt.Printf("Operations/sec: %.2f\n", float64(config.NumOperations)/elapsedTime.Seconds())
	fmt.Printf("  - GETs/sec: %.2f\n", float64(stats.Gets)/elapsedTime.Seconds())
	fmt.Printf("  - PUTs/sec: %.2f\n", float64(stats.Puts)/elapsedTime.Seconds())
	fmt.Printf("  - DELs/sec: %.2f\n", float64(stats.Deletes)/elapsedTime.Seconds())
}

// populateData populates the cluster with initial data
func populateData(config BenchmarkConfig, authString string) {
	conn, err := createConnection(config)
	if err != nil {
		fmt.Printf("Failed to connect for data population: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\r\n", authString)

	initialDataSize := 10000
	for i := 0; i < initialDataSize; i++ {
		key := fmt.Sprintf("benchmark_key_%d", i)
		value := generateRandomValue(config.ValueSize)
		fmt.Fprintf(conn, "PUT %s %s\r\n", key, value)
		buffer := make([]byte, 1024)
		conn.Read(buffer)
	}
}

// generateRandomKey generates a random key
func generateRandomKey(workerID, operationID, size int) string {
	return fmt.Sprintf("key_%d_%d_%s", workerID, operationID, randomString(size))
}

// generateRandomValue generates a random value
func generateRandomValue(size int) string {
	return randomString(size)
}

// randomString generates a random string
func randomString(size int) string {
	buffer := make([]byte, size)
	rand.Read(buffer)
	return base64.StdEncoding.EncodeToString(buffer)[:size]
}
