package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"time"

	"github.com/nabaz-io/nabaz/pkg/testrunner/models"
	"github.com/nabaz-io/nabaz/pkg/testrunner/scm/history/git"
	"github.com/nabaz-io/nabaz/pkg/testrunner/testengine"
)

func CreateNabazRun(testsToSkip map[string]models.SkippedTest, totalDuration float64, testEngine *testengine.TestEngine, history git.GitHistory, testResults []models.TestRun) *models.NabazRun {
	skippedTests := make([]models.SkippedTest, 0, len(testsToSkip))
	for _, v := range testsToSkip {
		skippedTests = append(skippedTests, v)
	}

	longestDuration := totalDuration
	if testEngine.LastNabazRun != nil {
		longestDuration = math.Max(totalDuration, testEngine.LastNabazRun.LongestDuration)
	}

	return &models.NabazRun{
		RunID:           time.Now().UnixNano(),
		CommitID:        history.HEAD(),
		TestsRan:        testResults,
		TestsSkipped:    skippedTests,
		RunDuration:     totalDuration,
		LongestDuration: longestDuration,
	}
}

func SendNabazStarted() error {
	t := models.ExecutionTelemtry{
		Os:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	return SendAnonymousTelemetry(t)
}

func NewAnnonymousTelemetry(nabazRun *models.NabazRun, hashedRepoName string) models.ResultTelemetry {
	return models.ResultTelemetry{
		RepoName:        hashedRepoName,
		Os:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		RunDuration:     nabazRun.RunDuration,
		LongestDuration: nabazRun.LongestDuration,
		TestsSkipped:    len(nabazRun.TestsSkipped),
		TestsRan:        len(nabazRun.TestsRan),
		TestsFailed:     len(nabazRun.FailedTests()),
	}
}

func SendAnonymousTelemetry(telemetry models.Telemetry) error {
	j, err := json.Marshal(telemetry)
	if err != nil {
		return err
	}

	res, err := http.Post("https://api.nabaz.io/stats", "application/json", bytes.NewBuffer(j))

	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("bad status code %d", res.StatusCode)
	}

	return nil
}