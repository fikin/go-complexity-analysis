# go-complexity-analysis

go-complexity-analysis calculates:
* the Cyclomatic complexities
* the Halstead complexities (difficulty and volume)
* the Maintainability indices
* lines of code
  of golang functions.

Additionally it counts package imports from same application. This is indicator of how interconnected application packages are with each other. Higher number suggests more interconnections and lower code refactoring capability.

# Install

```sh
$ go get github.com/fikin/go-complexity-analysis/cmd/complexity
```

# Usage

```sh
$ go vet -vettool=$(which complexity) [flags] [directory/file]
```

## Flags

`--cycloover`: show functions with the Cyclomatic complexity > N (default: 10)

`--maintunder`: show functions with the Maintainability index < N (default: 20)

`--selfimpdepth`: how many path levels must be common between package and its import to be considered a self-import (default same as package)

`--csvstats`: show functions stats in csv format. other flags are still valid.

`--csvtotals`: show package totals in csv format. other flags are still valid.

## Output

```
<filename>:<line>:<column>: func <funcname> seems to be complex (cyclomatic complexity=<cyclomatic complexity>)
<filename>:<line>:<column>: func <funcname> seems to have low maintainability (maintainability index=<maintainability index>)
<filename>,<line>,<column>,<funcname>,<cyclomatic complexity>,<maintainability index>,<halstead difficulty>,<halstead volume>,<loc>,<imports count>,<same app imports>
<pkgname>,<functions count>,-1,total,<cyclomatic complexity>,<maintainability index>,<halstead difficulty>,<halstead volume>,<loc>,<imports count>,<self imports count>
```

## Examples

```go
$ go vet -vettool=$(which complexity) --cycloover 10 ./...
$ go vet -vettool=$(which complexity) --maintunder 20 main.go
$ go vet -vettool=$(which complexity) --cycloover 5 --maintunder 30 ./src
$ go vet -vettool=$(which complexity) --csvstats ./src
$ go vet -vettool=$(which complexity) --cycloover 5 --maintunder 30 --csvstats ./...
$ go vet -vettool=$(which complexity) --csvtotals ./src
$ go vet -vettool=$(which complexity) --cycloover 5 --maintunder 30 --csvtotals ./...
$ go vet -vettool=$(which complexity) --cycloover 5 --maintunder 30 --csvtotals --selfimpdepth 4 ./...
```

## Github Actions

You can use the Github Actions to execute the complexity command on Github pull requests with [reviewdog](https://github.com/reviewdog/reviewdog).

See [fikin/go-complexity-analysis-action](https://github.com/fikin/go-complexity-analysis-action) for the details.


# Metrics

## Cyclomatic Complexity

The Cyclomatic complexity indicates the complexity of a program.

This program calculates the complexities of each function by counting idependent paths with the following rules.
```
Initial value: 1
+1: if, for, case, ||, &&
```

## Halstead Metrics

Calculation of each Halstead metrics can be found [here](https://www.verifysoft.com/en_halstead_metrics.html).

### Rules

1. Comments are not considered in Halstead Metrics
2. Operands and Operators are divided as follows:

#### Operands

- [Identifiers](!https://golang.org/ref/spec#Identifiers)
- [Constants](!https://golang.org/ref/spec#Constants)
- [Variables](!https://golang.org/ref/spec#Variables)

#### Operators
- [Operators](!https://golang.org/ref/spec#Operators_and_punctuation)
    - Parenthesis, such as "()", is counted as one operator
- [Keywords](!https://golang.org/ref/spec#Keywords)

## Maintainability Index

The Maintainability index represents maintainability of a program.

The value is calculated with the Cyclomatic complexity and the Halstead volume by using the following formula.
```
Maintainability Index = 171 - 5.2 * ln(Halstead Volume) - 0.23 * (Cyclomatic Complexity) - 16.2 * ln(Lines of Code)
```

This program shows normalized values instead of the original ones [introduced by Microsoft](https://docs.microsoft.com/en-us/archive/blogs/codeanalysis/maintainability-index-range-and-meaning).
```
Normalized Maintainability Index = MAX(0,(171 - 5.2 * ln(Halstead Volume) - 0.23 * (Cyclomatic Complexity) - 16.2 * ln(Lines of Code))*100 / 171)
```

The thresholds are as follows:
```
0-9 = Red
10-19 = Yellow
20-100 = Green
```

## Same application interconnection index

Same application package imports count is metric about inter-connectedness. 

Package imports as total count are not very meaningful, but imports of packages from same application are indication of problematic design.

In go, a good practice is every package to declare its externally needed interfaces which results in no extra package import. Otherwise it is likely the programming is happening against structures which is anti-pattern. 

In same application, all reusable and common data structures can be abstracted in single own package. For one application, this would result in single same application import count.

Criteria to determine same application import is based on package and import having same base path, up to indicated self import depth level value (--selfimpdepth).

This analyzer is printing import counts per package (if --csvtotoal is used) and repeats same values in each function from same package (if --csvstats is used).

By applying post processing on package statistics, one can count distinct packages per application (N) and how many of all imports are to same application packages (M, max of same application values per application). Then:
* for minimum complexity, M = N - 1. This is the absolute minimum complexity possible, where each package is coded independently of each other and only one package (main) integrates them all together.
* for practical purposes M = N. This is the optimal value, where besides one package used to integrated the rest, there is one package containing reusable application data model.
* anything where M > N would likely need code refactoring.

## CSV export

The analyzer can print data in csv format in order to offer easy import into other tools.
