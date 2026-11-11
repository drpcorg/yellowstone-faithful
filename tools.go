package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// exists checks whether a file or directory exists.
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		// file does not exist
		return false, nil
	}
	// other error
	return false, err
}

// isFile checks whether a path is a file.
func isFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// isJSONFile checks whether a path is a JSON file.
func isJSONFile(filepath string) bool {
	return filepath[len(filepath)-5:] == ".json"
}

// isYAMLFile checks whether a path is a YAML file.
func isYAMLFile(filepath string) bool {
	return filepath[len(filepath)-5:] == ".yaml" || filepath[len(filepath)-4:] == ".yml"
}

// loadFromJSON loads a JSON file into dst (which must be a pointer).
func loadFromJSON(configFilepath string, dst any) error {
	file, err := os.Open(configFilepath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	return fasterJson.NewDecoder(file).Decode(dst)
}

// loadFromYAML loads a YAML file into dst (which must be a pointer).
func loadFromYAML(configFilepath string, dst any) error {
	file, err := os.Open(configFilepath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return yaml.NewDecoder(file).Decode(dst)
}

type timer struct {
	reqID string
	start time.Time
	prev  time.Time
}

func newTimer(reqID string) *timer {
	now := time.Now()
	return &timer{
		reqID: reqID,
		start: now,
		prev:  now,
	}
}

func (t *timer) time(name string) {
	klog.V(4).Infof("[%s]: %q: %s (overall %s)", t.reqID, name, time.Since(t.prev), time.Since(t.start))
	t.prev = time.Now()
}

//	pub enum RewardType {
//	    Fee,
//	    Rent,
//	    Staking,
//	    Voting,
//	}
func rewardTypeToString(typ int) string {
	switch typ {
	case 1:
		return "Fee"
	case 2:
		return "Rent"
	case 3:
		return "Staking"
	case 4:
		return "Voting"
	default:
		return "Unknown"
	}
}

func rewardTypeStringToInt(typ string) int {
	switch typ {
	case "Fee":
		return 1
	case "Rent":
		return 2
	case "Staking":
		return 3
	case "Voting":
		return 4
	default:
		return 0
	}
}

const CodeNotFound = -32009
const CodeSkipped = -32009
const CodeNotAvailable = -32004
const CodeUnsupportedTransactionVersion = -32015

func checkTransactionVersion(tx *solana.Transaction, maxSupported *uint64) *jsonrpc2.Error {
	var version uint64
	switch tx.Message.GetVersion() {
	case solana.MessageVersionLegacy:
		return nil
	case solana.MessageVersionV1:
		version = 1
	}
	if maxSupported != nil && version <= *maxSupported {
		return nil
	}
	return &jsonrpc2.Error{
		Code:    CodeUnsupportedTransactionVersion,
		Message: fmt.Sprintf("Transaction version (%d) is not supported by the requesting client. Please try the request again with the following configuration parameter: \"maxSupportedTransactionVersion\": %d", version, version),
	}
}
