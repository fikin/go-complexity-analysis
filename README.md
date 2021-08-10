# go-complexity-analysis

go-complexity-analysis calculates:
* the Cyclomatic complexities
* the Halstead complexities (difficulty and volume)
* the Maintainability indices
* lines of code
  of golang functions.

Additionally it counts imports of packages from same application. This is indicator of interconnection between packages akin self imports. Higher number suggests more interconnections and lower refactoring capability.

# Install

```sh
$ go get github.com/shoooooman/go-complexity-analysis/cmd/complexity
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
<filename>,<line>,<column>,<funcname>,<cyclomatic complexity>,<maintainability index>,<halstead difficulty>,<halstead volume>,<loc>,<imports count>,<self imports count>
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

See [shoooooman/go-complexity-analysis-action](https://github.com/shoooooman/go-complexity-analysis-action) for the details.


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

## Package interconnection index

Self-import metric is count of how many of all package imports, are of packages from same application. Same application import is assumed when package and the import paths are having same path, up to the specified self import depth value.

This counter is used for post analysis of the entire application, looking for how inter-connected packages are. 

In go, minimum package inter-connection is N-1, where N is application's packages count.

For practical purposes N is also a good value.

Anything >N is indication what packages are perhaps more interconnected than necessary.

Determining N is possible only via post processing tools. One can export --csvtotals and count packages from same application. Application self-import is max of all such packages.

## CSV export

The analyzer can print data in csv format in order to offer easy import into other tools.
