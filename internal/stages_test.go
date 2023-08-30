package internal

import (
	"os"
	"regexp"
	"testing"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func TestStages(t *testing.T) {
	os.Setenv("CODECRAFTERS_RANDOM_SEED", "1234567890")

	testCases := map[string]tester_utils.TesterOutputTestCase{
		"bencoded_string_failure": {
			StageName:           "bencode-string",
			CodePath:            "./test_helpers/scenarios/init/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/init/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bencoded_string_success": {
			StageName:           "bencode-string",
			CodePath:            "./test_helpers/scenarios/init/success",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/init/success",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"pass_all": {
			StageName:           "dl-file",
			CodePath:            "./test_helpers/scenarios/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/pass_all",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	re := regexp.MustCompile("Running ./your_bittorrent.sh .*")
	testerOutput = re.ReplaceAll(testerOutput, []byte("Running ./your_bittorrent.sh <truncated>"))

	return testerOutput
}
