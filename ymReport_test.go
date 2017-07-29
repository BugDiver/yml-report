package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

)

var now = time.Now()

type testNameGenerator struct {
}

func (T testNameGenerator) randomName() string {
	return now.Format(timeFormat)
}

func TestCreatingReport(t *testing.T) {
	reportDir := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(reportDir)

	finalReportDir, err := createYMLReport(reportDir, make([]byte, 0), nil)
	if err != nil {
		t.Errorf("Obtained: %s \nExpected: %s ", err , nil)
	}


	expectedFinalReportDir := filepath.Join(reportDir, ymlReport)
	if finalReportDir != expectedFinalReportDir {
		t.Errorf("Obtained: %s \nExpected: %s ", finalReportDir , expectedFinalReportDir)
	}

	verifyYMLReportFileIsCopied(expectedFinalReportDir, t)
}

func TestCreatingReportWithNoOverWrite(t *testing.T) {
	reportDir := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(reportDir)

	nameGen := testNameGenerator{}
	finalReportDir, err := createYMLReport(reportDir, make([]byte, 0), nameGen)
	if err != nil {
		t.Errorf("Obtained: %s \nExpected: %s ", err , nil)
	}

	expectedFinalReportDir := filepath.Join(reportDir, ymlReport, nameGen.randomName())
	if finalReportDir != expectedFinalReportDir {
		t.Errorf("Obtained: %s \nExpected: %s ", finalReportDir , expectedFinalReportDir)
	}
	verifyYMLReportFileIsCopied(expectedFinalReportDir, t)
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func verifyYMLReportFileIsCopied(dest string, t *testing.T) {
	exists := fileExists(filepath.Join(dest, jsonReportFile))
	if  !exists {
		t.Errorf("Obtained: %s \nExpected: %s ", exists, true)
	}

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}


func TestCreatingReportShouldOverwriteReportsBasedOnEnv(t *testing.T) {
	os.Setenv(overwriteReportsEnvProperty, "true")
	nameGen := getNameGen()
	if nameGen != nil {
		t.Errorf("Obtained: %s\nExpected: %s",nameGen, nil)
	}

	os.Setenv(overwriteReportsEnvProperty, "false")
	nameGen = getNameGen()
	if nameGen != nameGenerator(timeStampedNameGenerator{}) {
		t.Errorf("Obtained: %s \nExpected: %s ",nameGen, timeStampedNameGenerator{})
	}
}