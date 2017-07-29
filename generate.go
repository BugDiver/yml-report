package main

import (
	"encoding/base64"

	"github.com/bugdiver/yml-report/gauge_messages"
)

type tokenKind string
type status string
type errorType string

const (
	pass                  status    = "pass"
	fail                  status    = "fail"
	skip                  status    = "skip"
	notExecuted           status    = "not executed"
	stepKind              tokenKind = "step"
	conceptKind           tokenKind = "concept"
	assertionErrorType    errorType = "assertion"
	verificationErrorType errorType = "verification"
)

type item interface {
	kind() tokenKind
}

type suiteResult struct {
	ProjectName            string       `yaml:"projectName"`
	Timestamp              string       `yaml:"timestamp"`
	SuccessRate            int          `yaml:"successRate"`
	Environment            string       `yaml:"environment"`
	Tags                   string       `yaml:"tags"`
	ExecutionTime          int64        `yaml:"executionTime"`
	ExecutionStatus        status       `yaml:"executionStatus"`
	SpecResults            []spec       `yaml:"specResults"`
	BeforeSuiteHookFailure *hookFailure `yaml:"beforeSuiteHookFailure"`
	AfterSuiteHookFailure  *hookFailure `yaml:"afterSuiteHookFailure"`
	PassedSpecsCount       int          `yaml:"passedSpecsCount"`
	FailedSpecsCount       int          `yaml:"failedSpecsCount"`
	SkippedSpecsCount      int          `yaml:"skippedSpecsCount"`
}

type spec struct {
	SpecHeading           string       `yaml:"specHeading"`
	FileName              string       `yaml:"fileName"`
	Tags                  []string     `yaml:"tags"`
	ExecutionTime         int64        `yaml:"executionTime"`
	ExecutionStatus       status       `yaml:"executionStatus"`
	Scenarios             []scenario   `yaml:"scenarios"`
	IsTableDriven         bool         `yaml:"isTableDriven"`
	Datatable             *table       `yaml:"datatable"`
	BeforeSpecHookFailure *hookFailure `yaml:"beforeSpecHookFailure"`
	AfterSpecHookFailure  *hookFailure `yaml:"afterSpecHookFailure"`
	PassedScenarioCount   int          `yaml:"passedScenarioCount"`
	FailedScenarioCount   int          `yaml:"failedScenarioCount"`
	SkippedScenarioCount  int          `yaml:"skippedScenarioCount"`
}

type scenario struct {
	Heading                   string       `yaml:"scenarioHeading"`
	Tags                      []string     `yaml:"tags"`
	ExecutionTime             int64        `yaml:"executionTime"`
	ExecutionStatus           status       `yaml:"executionStatus"`
	Contexts                  []item       `yaml:"contexts"`
	Teardowns                 []item       `yaml:"teardowns"`
	Items                     []item       `yaml:"items"`
	BeforeScenarioHookFailure *hookFailure `yaml:"beforeScenarioHookFailure"`
	AfterScenarioHookFailure  *hookFailure `yaml:"afterScenarioHookFailure"`
	SkipErrors                []string     `yaml:"skipErrors"`
	TableRowIndex             int          `yaml:"tableRowIndex"`
}

type step struct {
	ItemType              tokenKind    `yaml:"itemType"`
	StepText              string       `yaml:"stepText"`
	Table                 *table       `yaml:"table"`
	BeforeStepHookFailure *hookFailure `yaml:"beforeStepHookFailure"`
	AfterStepHookFailure  *hookFailure `yaml:"afterStepHookFailure"`
	Result                *result      `yaml:"result"`
}

func (s *step) kind() tokenKind {
	return stepKind
}

type result struct {
	Status        status    `yaml:"status"`
	StackTrace    string    `yaml:"stackTrace"`
	Screenshot    string    `yaml:"screenshot"`
	ErrorMessage  string    `yaml:"errorMessage"`
	ExecutionTime int64     `yaml:"executionTime"`
	SkippedReason string    `yaml:"skippedReason"`
	Messages      []string  `yaml:"messages"`
	ErrorType     errorType `yaml:"errorType"`
}

type hookFailure struct {
	ErrMsg     string `yaml:"errorMessage"`
	Screenshot string `yaml:"screenshot"`
	StackTrace string `yaml:"stackTrace"`
}

type concept struct {
	ItemType    tokenKind `yaml:"itemType"`
	ConceptStep *step     `yaml:"conceptStep"`
	Items       []item    `yaml:"items"`
	Result      result    `yaml:"result"`
}

func (s *concept) kind() tokenKind {
	return conceptKind
}

type table struct {
	Headers []string `yaml:"headers"`
	Rows    []row    `yaml:"rows"`
}

type row struct {
	Cells []string `yaml:"cells"`
}

func toSuiteResult(psr *gauge_messages.ProtoSuiteResult) suiteResult {
	suiteResult := suiteResult{
		ProjectName:            psr.GetProjectName(),
		Environment:            psr.GetEnvironment(),
		Tags:                   psr.GetTags(),
		ExecutionTime:          psr.GetExecutionTime(),
		PassedSpecsCount:       len(psr.GetSpecResults()) - int(psr.GetSpecsFailedCount()) - int(psr.GetSpecsSkippedCount()),
		FailedSpecsCount:       int(psr.GetSpecsFailedCount()),
		SkippedSpecsCount:      int(psr.GetSpecsSkippedCount()),
		BeforeSuiteHookFailure: toHookFailure(psr.GetPreHookFailure()),
		AfterSuiteHookFailure:  toHookFailure(psr.GetPostHookFailure()),
		SuccessRate:            int(psr.GetSuccessRate()),
		Timestamp:              psr.GetTimestamp(),
		ExecutionStatus:        pass,
	}
	if psr.GetFailed() {
		suiteResult.ExecutionStatus = fail
	}
	suiteResult.SpecResults = make([]spec, 0)
	for _, protoSpecRes := range psr.GetSpecResults() {
		suiteResult.SpecResults = append(suiteResult.SpecResults, toSpec(protoSpecRes))
	}
	return suiteResult
}

func toSpec(psr *gauge_messages.ProtoSpecResult) spec {
	spec := spec{
		SpecHeading:           psr.GetProtoSpec().GetSpecHeading(),
		IsTableDriven:         psr.GetProtoSpec().GetIsTableDriven(),
		FileName:              psr.GetProtoSpec().GetFileName(),
		Tags:                  make([]string, 0),
		FailedScenarioCount:   int(psr.GetScenarioFailedCount()),
		SkippedScenarioCount:  int(psr.GetScenarioSkippedCount()),
		PassedScenarioCount:   int(psr.GetScenarioCount() - psr.GetScenarioFailedCount() - psr.GetScenarioSkippedCount()),
		ExecutionTime:         psr.GetExecutionTime(),
		ExecutionStatus:       getStatus(psr.GetFailed(), psr.GetSkipped()),
		BeforeSpecHookFailure: toHookFailure(psr.GetProtoSpec().GetPreHookFailure()),
		AfterSpecHookFailure:  toHookFailure(psr.GetProtoSpec().GetPostHookFailure()),
	}
	if psr.GetProtoSpec().GetTags() != nil {
		spec.Tags = psr.GetProtoSpec().GetTags()
	}
	spec.Scenarios = make([]scenario, 0)
	for _, item := range psr.GetProtoSpec().GetItems() {
		switch item.GetItemType() {
		case gauge_messages.ProtoItem_Scenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetScenario(), -1))
		case gauge_messages.ProtoItem_TableDrivenScenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetTableDrivenScenario().GetScenario(), int(item.GetTableDrivenScenario().GetTableRowIndex())))
		case gauge_messages.ProtoItem_Table:
			spec.Datatable = toTable(item.GetTable())
		}
	}
	return spec
}

func toScenario(protoSce *gauge_messages.ProtoScenario, tableRowIndex int) scenario {
	sce := scenario{
		Heading:                   protoSce.GetScenarioHeading(),
		ExecutionTime:             protoSce.GetExecutionTime(),
		Tags:                      make([]string, 0),
		ExecutionStatus:           getScenarioStatus(protoSce),
		Contexts:                  make([]item, 0),
		Items:                     make([]item, 0),
		Teardowns:                 make([]item, 0),
		BeforeScenarioHookFailure: toHookFailure(protoSce.GetPreHookFailure()),
		AfterScenarioHookFailure:  toHookFailure(protoSce.GetPostHookFailure()),
		TableRowIndex:             tableRowIndex,
		SkipErrors:                make([]string, 0),
	}
	if protoSce.GetSkipErrors() != nil {
		sce.SkipErrors = protoSce.GetSkipErrors()
	}
	if protoSce.GetTags() != nil {
		sce.Tags = protoSce.GetTags()
	}
	if protoSce.GetContexts() != nil {
		sce.Contexts = toItems(protoSce.GetContexts())
	}
	if protoSce.GetScenarioItems() != nil {
		sce.Items = toItems(protoSce.GetScenarioItems())
	}
	if protoSce.GetTearDownSteps() != nil {
		sce.Teardowns = toItems(protoSce.GetTearDownSteps())
	}
	return sce
}

func getScenarioStatus(scn *gauge_messages.ProtoScenario) status {
	switch scn.GetExecutionStatus() {
	case gauge_messages.ExecutionStatus_FAILED:
		return fail
	case gauge_messages.ExecutionStatus_PASSED:
		return pass
	case gauge_messages.ExecutionStatus_SKIPPED:
		return skip
	default:
		return notExecuted
	}
}

func toTable(protoTable *gauge_messages.ProtoTable) *table {
	rows := make([]row, len(protoTable.GetRows()))
	for i, r := range protoTable.GetRows() {
		rows[i] = row{
			Cells: r.GetCells(),
		}
	}
	headers := make([]string, 0)
	if protoTable.GetHeaders().GetCells() != nil {
		headers = protoTable.GetHeaders().GetCells()
	}
	return &table{Headers: headers, Rows: rows}
}

func toItems(protoItems []*gauge_messages.ProtoItem) []item {
	items := make([]item, 0)
	for _, i := range protoItems {
		switch i.GetItemType() {
		case gauge_messages.ProtoItem_Step:
			items = append(items, toStep(i.GetStep()))
		case gauge_messages.ProtoItem_Concept:
			items = append(items, toConcept(i.GetConcept()))
		}
	}
	return items
}

func toStep(protoStep *gauge_messages.ProtoStep) *step {
	res := protoStep.GetStepExecutionResult().GetExecutionResult()
	result := &result{
		Status:        getStepStatus(protoStep.GetStepExecutionResult()),
		Screenshot:    base64.StdEncoding.EncodeToString(res.GetScreenShot()),
		StackTrace:    res.GetStackTrace(),
		ErrorMessage:  res.GetErrorMessage(),
		ExecutionTime: res.GetExecutionTime(),
		Messages:      make([]string, 0),
		ErrorType:     getErrorType(res.GetErrorType()),
	}
	if protoStep.GetStepExecutionResult().GetSkipped() {
		result.SkippedReason = protoStep.GetStepExecutionResult().GetSkippedReason()
	}
	if res.GetMessage() != nil {
		result.Messages = res.GetMessage()
	}
	var tableParam *table
	if protoStep.GetFragments() != nil {
		for _, f := range protoStep.GetFragments() {
			if f.GetParameter().GetParameterType() == gauge_messages.Parameter_Table || f.GetParameter().GetParameterType() == gauge_messages.Parameter_Special_Table {
				tableParam = toTable(f.GetParameter().GetTable())
			}
		}
	}
	return &step{
		ItemType:              stepKind,
		StepText:              protoStep.GetActualText(),
		Result:                result,
		Table:                 tableParam,
		BeforeStepHookFailure: toHookFailure(protoStep.GetStepExecutionResult().GetPreHookFailure()),
		AfterStepHookFailure:  toHookFailure(protoStep.GetStepExecutionResult().GetPostHookFailure()),
	}
}

func toConcept(protoConcept *gauge_messages.ProtoConcept) *concept {
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.GetConceptExecutionResult()
	return &concept{
		ItemType:    conceptKind,
		ConceptStep: toStep(protoConcept.GetConceptStep()),
		Items:       toItems(protoConcept.GetSteps()),
	}
}

func toHookFailure(failure *gauge_messages.ProtoHookFailure) *hookFailure {
	if failure == nil {
		return nil
	}
	return &hookFailure{
		ErrMsg:     failure.GetErrorMessage(),
		Screenshot: base64.StdEncoding.EncodeToString(failure.GetScreenShot()),
		StackTrace: failure.GetStackTrace(),
	}
}
func getErrorType(protoErrType gauge_messages.ProtoExecutionResult_ErrorType) errorType {
	if protoErrType == gauge_messages.ProtoExecutionResult_VERIFICATION {
		return verificationErrorType
	}
	return assertionErrorType
}

func getStatus(failed, skipped bool) status {
	if failed {
		return fail
	}
	if skipped {
		return skip
	}
	return pass
}

func getStepStatus(res *gauge_messages.ProtoStepExecutionResult) status {
	if res.GetSkipped() {
		return skip
	}
	if res.GetExecutionResult() == nil {
		return notExecuted
	}
	if res.GetExecutionResult().GetFailed() {
		return fail
	}
	return pass
}
