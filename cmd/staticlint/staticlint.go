// Analyzer staticlint consisting of:
// 1. Standard static analyzers of the golang.org/x/tools/go/analysis/passes package;
// 2. All SA class analyzers of the staticcheck.io package;
// 3. Two analyzers for the remaining classes of the staticcheck.io package;
//
// Usage:
// $ staticlint            # checking files in the current directory
// $ staticlint main.go    # checking main.go file
// $ staticlint ./...      # checking all files in the current directory and subdirectories
// $ staticlint ./foo ./internal/...  # checking all files in the foo subdirectory and in all internal subdirectories
package main

import (
	"strings"

	"staticlint/customlint"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		appends.Analyzer,             // Package appends defines an Analyzer that detects if there is only one variable in append.
		asmdecl.Analyzer,             // Package asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
		assign.Analyzer,              // Package assign defines an Analyzer that detects useless assignments.
		atomic.Analyzer,              // Package atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
		atomicalign.Analyzer,         // Package atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
		bools.Analyzer,               // Package bools defines an Analyzer that detects common mistakes involving boolean operators.
		buildssa.Analyzer,            // Package buildssa defines an Analyzer that constructs the SSA representation of an error-free package and returns the set of all functions within it.
		buildtag.Analyzer,            // Package buildtag defines an Analyzer that checks build tags.
		cgocall.Analyzer,             // Package cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
		composite.Analyzer,           // Package composite defines an Analyzer that checks for unkeyed composite literals.
		copylock.Analyzer,            // Package copylock defines an Analyzer that checks for locks erroneously passed by value.
		ctrlflow.Analyzer,            // Package ctrlflow is an analysis that provides a syntactic control-flow graph (CFG) for the body of a function.
		deepequalerrors.Analyzer,     // Package deepequalerrors defines an Analyzer that checks for the use of reflect.DeepEqual with error values.
		defers.Analyzer,              // Package defers defines an Analyzer that checks for common mistakes in defer statements.
		directive.Analyzer,           // Package directive defines an Analyzer that checks known Go toolchain directives.
		errorsas.Analyzer,            // The errorsas package defines an Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
		fieldalignment.Analyzer,      // Package fieldalignment defines an Analyzer that detects structs that would use less memory if their fields were sorted.
		findcall.Analyzer,            // Package findcall defines an Analyzer that serves as a trivial example and test of the Analysis API.
		framepointer.Analyzer,        // Package framepointer defines an Analyzer that reports assembly code that clobbers the frame pointer before saving it.
		httpmux.Analyzer,             // The httpmux command runs the httpmux analyzer.
		httpresponse.Analyzer,        // Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
		ifaceassert.Analyzer,         // Package ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
		inspect.Analyzer,             // Package inspect defines an Analyzer that provides an AST inspector (golang.org/x/tools/go/ast/inspector.Inspector) for the syntax trees of a package.
		loopclosure.Analyzer,         // Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
		lostcancel.Analyzer,          // Package lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
		nilfunc.Analyzer,             // Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
		nilness.Analyzer,             // Package nilness inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons.
		pkgfact.Analyzer,             // The pkgfact package is a demonstration and test of the package fact mechanism.
		printf.Analyzer,              // Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
		reflectvaluecompare.Analyzer, // Package reflectvaluecompare defines an Analyzer that checks for accidentally using == or reflect.DeepEqual to compare reflect.Value values.
		shadow.Analyzer,              // Package shadow defines an Analyzer that checks for shadowed variables.
		shift.Analyzer,               // Package shift defines an Analyzer that checks for shifts that exceed the width of an integer.
		sigchanyzer.Analyzer,         // Package sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal.Notify.
		slog.Analyzer,                // Package slog defines an Analyzer that checks for mismatched key-value pairs in log/slog calls.
		sortslice.Analyzer,           // Package sortslice defines an Analyzer that checks for calls to sort.Slice that do not use a slice type as first argument.
		stdmethods.Analyzer,          // Package stdmethods defines an Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
		stdversion.Analyzer,          // Package stdversion reports uses of standard library symbols that are "too new" for the Go version in force in the referring file.
		stringintconv.Analyzer,       // Package stringintconv defines an Analyzer that flags type conversions from integers to strings.
		structtag.Analyzer,           // Package structtag defines an Analyzer that checks struct field tags are well formed.
		testinggoroutine.Analyzer,    // Package testinggoroutine defines an Analyzerfor detecting calls to Fatal from a test goroutine.
		tests.Analyzer,               // Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
		timeformat.Analyzer,          // Package timeformat defines an Analyzer that checks for the use of time.Format or time.Parse calls with a bad format.
		unmarshal.Analyzer,           // The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
		unreachable.Analyzer,         // Package unreachable defines an Analyzer that checks for unreachable code.
		unsafeptr.Analyzer,           // Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
		unusedresult.Analyzer,        // Package unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
		unusedwrite.Analyzer,         // Package unusedwrite checks for unused writes to the elements of a struct or array object.
		usesgenerics.Analyzer,        // Package usesgenerics defines an Analyzer that checks for usage of generic features added in Go 1.18.

		customlint.Analyzer, // Package customlint defines an custom Analyzer that disables for os.Exit calls.
	)

	// Source: https://staticcheck.io/docs/checks/#SA
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	for _, a := range [][]*lint.Analyzer{
		quickfix.Analyzers,   // Source: https://staticcheck.io/docs/checks/#QF
		simple.Analyzers,     // Source: https://staticcheck.io/docs/checks/#S
		stylecheck.Analyzers, // Source: https://staticcheck.io/docs/checks/#ST
	} {
		for i := 0; i < 2; i++ {
			analyzers = append(analyzers, a[i].Analyzer)
		}
	}

	multichecker.Main(analyzers...)
}
