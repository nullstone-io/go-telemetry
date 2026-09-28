// Package workflows is Nullstone's vocabulary for the outcome of a workflow or activity: who caused a
// failure, the customer (their code, configuration, or cloud account) or Nullstone (a bug, infrastructure,
// an upstream dependency).
//
// It is deliberately independent of any execution engine. Producers classify their errors
// (Classifier, WithFailure, Match); the engine that observes an execution's outcome calls
// Classify, Annotate and the Record* functions so dashboards and alerts can separate a
// customer's failing build from a regression of ours.
//
// The default is internal: a failure is only attributed to the user when a producer positively
// identified it as such. An unrecognised error is most likely a Nullstone regression, and it
// must page rather than blend in with customers' own failing builds and terraform runs.
package workflows

import (
	"context"
	"errors"
	"fmt"
)

// Class says who (or what) caused a failure
type Class string

const (
	// ClassUser is a failure caused by the customer's own code, configuration, or cloud account
	ClassUser Class = "user"
	// ClassInternal is a Nullstone bug, an infrastructure fault, or an upstream dependency Nullstone relies on
	ClassInternal Class = "internal"
	// ClassCancelled means the execution was cancelled (by a user or the system); it is not a failure
	ClassCancelled Class = "cancelled"
	// ClassTimeout means the execution exceeded its timeout
	ClassTimeout Class = "timeout"
)

// IsFailure reports whether this class counts towards failure metrics; cancellations and timeouts do not
func (c Class) IsFailure() bool {
	return c == ClassUser || c == ClassInternal
}

// Category is a free-form, low-cardinality label a service assigns to a failure (e.g. a build step or
// the subsystem that broke). Each service owns its own bounded list of categories; this package only
// defines the default.
const (
	// CategoryUnknown is the default for anything not positively classified
	CategoryUnknown = "unknown"
)

// MaxCodeLen bounds Info.Code so a terraform diagnostic or stderr excerpt never bloats a span or metric attribute
const MaxCodeLen = 128

// Info classifies a failure. Code is an optional short detail (a terraform diagnostic summary,
// an exit code, an eviction reason), bounded to MaxCodeLen characters.
type Info struct {
	Class    Class  `json:"class"`
	Category string `json:"category"`
	Code     string `json:"code,omitempty"`
}

var (
	// Unknown is the default classification: an internal failure nobody identified
	Unknown = Info{Class: ClassInternal, Category: CategoryUnknown}
	// Cancelled is the classification of a cancelled execution
	Cancelled = Info{Class: ClassCancelled}
	// Timeout is the classification of an execution that exceeded its timeout
	Timeout = Info{Class: ClassTimeout}
)

// User builds a user classification
func User(category string) Info { return Info{Class: ClassUser, Category: category} }

// Internal builds an internal classification
func Internal(category string) Info { return Info{Class: ClassInternal, Category: category} }

// IsZero reports whether no classification was set at all
func (i Info) IsZero() bool {
	return i.Class == "" && i.Category == "" && i.Code == ""
}

// Normalized fills blanks with the internal/unknown default and bounds Code
func (i Info) Normalized() Info {
	if i.Class == "" {
		i.Class = ClassInternal
	}
	if i.Category == "" && i.Class.IsFailure() {
		i.Category = CategoryUnknown
	}
	if len(i.Code) > MaxCodeLen {
		i.Code = i.Code[:MaxCodeLen]
	}
	return i
}

// Classifier is implemented by error types that know what caused them.
// A type that crosses a serialization boundary must carry its Info in a serialized field.
type Classifier interface {
	FailureInfo() Info
}

// ClassifiedError attaches an Info to an arbitrary error (see WithFailure).
// Only Message and Info are serialized; the wrapped Err is not.
type ClassifiedError struct {
	Message string `json:"message"`
	Info    Info   `json:"info"`
	Err     error  `json:"-"`
}

// WithFailure classifies err without defining an error type for it.
//
// Use it for plain errors (fmt.Errorf, sentinel errors, third-party errors). An error type that is
// serialized across an execution boundary with its own details should carry an Info field and
// implement Classifier instead, so those details are not lost to the wrapper.
func WithFailure(err error, info Info) error {
	if err == nil {
		return nil
	}
	return &ClassifiedError{Message: err.Error(), Info: info.Normalized(), Err: err}
}

func (e *ClassifiedError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *ClassifiedError) Unwrap() error { return e.Err }

func (e *ClassifiedError) FailureInfo() Info { return e.Info }

// Matcher classifies errors of a type that cannot implement Classifier itself (e.g. one owned by another module)
type Matcher func(err error) (Info, bool)

var matchers []Matcher

// RegisterMatcher adds a Matcher consulted by Classify, in registration order, after Classifier
func RegisterMatcher(m Matcher) {
	matchers = append(matchers, m)
}

// Match registers a Matcher that classifies every error of type T (anywhere in the chain) as info
func Match[T error](info Info) {
	RegisterMatcher(func(err error) (Info, bool) {
		var t T
		if errors.As(err, &t) {
			return info, true
		}
		return Info{}, false
	})
}

// Classify determines who caused err. Resolution order:
//  1. a cancelled or expired context in the chain: Cancelled / Timeout
//  2. the error (or any error in its chain) implements Classifier
//  3. a registered Matcher matches
//  4. otherwise Unknown
//
// A nil error returns the zero Info. Engine-specific wrappers (e.g. Temporal's) must be
// unwrapped by the caller first.
func Classify(err error) Info {
	if err == nil {
		return Info{}
	}
	if errors.Is(err, context.Canceled) {
		return Cancelled
	} else if errors.Is(err, context.DeadlineExceeded) {
		return Timeout
	}
	var classifier Classifier
	if errors.As(err, &classifier) {
		return classifier.FailureInfo().Normalized()
	}
	for _, m := range matchers {
		if info, ok := m(err); ok {
			return info.Normalized()
		}
	}
	return Unknown
}

// ErrorType is the Go type of err, for the error.type attribute
func ErrorType(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%T", err)
}
