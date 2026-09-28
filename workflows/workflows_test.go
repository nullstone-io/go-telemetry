package workflows

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type selfClassified struct {
	Failure Info `json:"failure"`
}

func (e *selfClassified) Error() string     { return "self classified" }
func (e *selfClassified) FailureInfo() Info { return e.Failure }

type external struct{ Field string }

func (e external) Error() string { return "invalid field " + e.Field }

func init() {
	Match[external](User(CategoryConfig))
}

func TestClassify(t *testing.T) {
	tfPlan := Info{Class: ClassUser, Category: CategoryTerraformPlan, Code: "Unsupported argument"}
	tests := map[string]struct {
		err  error
		want Info
	}{
		"nil":                                   {nil, Info{}},
		"plain error is internal/unknown":       {errors.New("pg: can't find column"), Unknown},
		"wrapped plain error is unknown":        {fmt.Errorf("saving: %w", errors.New("boom")), Unknown},
		"cancelled context":                     {fmt.Errorf("docker build cancelled: %w", context.Canceled), Cancelled},
		"expired context":                       {context.DeadlineExceeded, Timeout},
		"classifier":                            {&selfClassified{Failure: tfPlan}, tfPlan},
		"classifier wrapped in context":         {fmt.Errorf("running plan: %w", &selfClassified{Failure: tfPlan}), tfPlan},
		"classifier with blank info is unknown": {&selfClassified{}, Unknown},
		"classifier with class only":            {&selfClassified{Failure: Info{Class: ClassUser}}, Info{Class: ClassUser, Category: CategoryUnknown}},
		"matcher":                               {external{Field: "vars"}, User(CategoryConfig)},
		"matcher through wrapping":              {fmt.Errorf("parsing: %w", external{Field: "vars"}), User(CategoryConfig)},
		"WithFailure":                           {WithFailure(errors.New("deployment failed"), User(CategoryDeployRollout)), User(CategoryDeployRollout)},
		"WithFailure bounds the code": {
			WithFailure(errors.New("x"), Info{Class: ClassInternal, Category: CategoryDatabase, Code: strings.Repeat("a", 500)}),
			Info{Class: ClassInternal, Category: CategoryDatabase, Code: strings.Repeat("a", MaxCodeLen)},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Classify(test.err); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Classify() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestWithFailure(t *testing.T) {
	if WithFailure(nil, User(CategoryConfig)) != nil {
		t.Error("WithFailure(nil) should be nil")
	}
	sentinel := errors.New("deployment failed")
	err := WithFailure(fmt.Errorf("watching rollout: %w", sentinel), User(CategoryDeployRollout))
	if !errors.Is(err, sentinel) {
		t.Error("the wrapped chain should be kept")
	}
	if err.Error() != "watching rollout: deployment failed" {
		t.Errorf("Error() = %q", err.Error())
	}
	// A deserialized ClassifiedError has no Err; its Message stands in
	rehydrated := &ClassifiedError{Message: "deployment failed", Info: User(CategoryDeployRollout)}
	if rehydrated.Error() != "deployment failed" {
		t.Errorf("Error() = %q", rehydrated.Error())
	}
}

func TestInfo_Attributes(t *testing.T) {
	attrs := Info{Class: ClassUser, Category: CategoryDockerBuild, Code: "exit 1"}.Attributes(ErrorType(errors.New("x")))
	got := map[string]string{}
	for _, kv := range attrs {
		got[string(kv.Key)] = kv.Value.AsString()
	}
	want := map[string]string{
		AttrClass: "user", AttrCategory: "docker-build", AttrCode: "exit 1", AttrErrorType: "*errors.errorString",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Attributes() = %v, want %v", got, want)
	}
	if n := len(Cancelled.Attributes("")); n != 1 {
		t.Errorf("a cancellation carries only the class attribute, got %d", n)
	}
}

func TestCompletionStatus(t *testing.T) {
	for info, want := range map[Info]string{Cancelled: CompletionCancelled, Timeout: CompletionTimeout, Unknown: CompletionFailed, User(CategoryConfig): CompletionFailed} {
		if got := CompletionStatus(info); got != want {
			t.Errorf("CompletionStatus(%+v) = %q, want %q", info, got, want)
		}
	}
}
