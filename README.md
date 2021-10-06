# go-complexity-analysis

go-complexity-analysis calculates:
* the Cyclomatic complexities
* the Halstead complexities (difficulty, volume, time to code)
* the Maintainability index
* lines of code
* lines of code of (only) variable and constant declarations

of golang functions.

# Install and usage as go-vet tool

## Build

```sh
$ go get github.com/fikin/go-complexity-analysis
$ go build -o path_to_binary cmd/complexity/main.go
```

## Use

```sh
$ go vet -vettool=<path_to_binary> [flags] [directory/file]
```

# Install and use as golangcli-lint plugin

## Build

```sh
$ go build -buildmode=plugin -o path_to_plugin_dir github.com/fikin/complexity/plugin/complexity
```

## Use

Add to your `.golangci.yml` file

```yaml
linters-settings:
  custom:
    complexity:
      path: <path to plugin>.so
      description: Complexity checks cyclomatic complexity and maintainability index
      original-url: github.com/fikin/complexity
```

# Flags

`--cycloover`: show functions with the Cyclomatic complexity > N (default: 10)

`--maintunder`: show functions with the Maintainability index < N (default: 20)

`--csv`: print (all) function stats in csv format

## Output

```
<filename>:<line>:<column>: func <funcname> seems to be complex (cyclomatic complexity=<cyclomatic complexity>)
<filename>:<line>:<column>: func <funcname> seems to have low maintainability (maintainability index=<maintainability index>)
```

If csv flag is used, following are the columns printed:

```
<file name>,<line>,<column>,<function name>,<cyclomatic complexity>,<maintainability index>,<halstead difficulty>,<halstead volume>,<time to code>,<loc>,<varDeclarationLoc>,<tooComplex>,<notMaintainable>```

## Examples

```go
$ go vet -vettool=$(which complexity) --cycloover 10 ./...
$ go vet -vettool=$(which complexity) --maintunder 20 main.go
$ go vet -vettool=$(which complexity) --cycloover 5 --maintunder 30 ./src
$ go vet -vettool=$(which complexity) --maxlines 30 ./src
$ go vet -vettool=$(which complexity) --csv ./src
```

# Github Actions

You can use the Github Actions to execute the complexity command on Github pull requests with [reviewdog](https://github.com/reviewdog/reviewdog).

See [fikin/go-complexity-analysis-action](https://github.com/fikin/go-complexity-analysis-action) for the details.


# Metrics

## Cyclomatic Complexity

The Cyclomatic complexity indicates the complexity of a program.

For background reference see explanations in [wikipedia](https://en.wikipedia.org/wiki/Cyclomatic_complexity).

This program calculates the complexities of each function by counting independent paths with the following rules.
```
Initial value: 1
+1: if, for, range, select, switch, final-else, chan read, chan write, ||, &&
+2: go subroutine
```

The thresholds are as follows:
```
0-10 = Green
11-... = Red
```

### Differences related to Go-lang nature

Else (final) in if-(else-if-)else construct is considered own branch as Go coding practice discourages such constructs.

Channel read and writes are considered complex operations. Channel operations are used for inter-process communication in Go and require proper channel handling awareness, they are not plain assignments.

Go subroutines spawning are considered extra complex. Subroutines life cycle design and tracking is requiring extra caution from developers, especially if there is use of up-values (enclosing function variables).

Go select and switch constructs are considered single complexity i.e. different case statements are not counted as individual execution paths.
In Go, case statements are used in places where typically inheritance or polymorphism would otherwise have been used. These situations are not tracked by cyclomatic complexity analysis.
Additionally, while in some situations it would be possible to split cases into multiple functions, this would not lead to reduced complexity (aka. function extraction), nor to improved code readability.
Since the focus of this analyzer is to be of more practical value, it was decided to not count individual case statements.

## Halstead Metrics

Calculation of each Halstead metrics can be found [here](https://www.verifysoft.com/en_halstead_metrics.html) and [wikipedia](https://en.wikipedia.org/wiki/Halstead_complexity_measures).

This analyzer is calculating halstead difficulty, volume and time-to-code metrics. They are provided in csv output format.

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

# Maintainability Index

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

# Lines of code

In csv output format, the analyzer is outputting function's total lines of code.

Additionally it calculates the function's total lines of codes for all constant and variable declarations.
This metrics can be used to reveal if some function is having large halstead volume (and thus low maintainability index) due to too much configuration data. This is applicable specifically for table-driven test case coding practice.

# CSV export

The analyzer can print data in csv format in order to offer easy import into other tools.

For example, here is how one can gather statistics in csv format from current directory:

```
( echo "file name,line,column,function name,cyclomatic complexity,maintainability index,halstead difficulty,halstead volume,\
time to code,loc,declLoc,too complex,not maintainable" && \
  ( go vet -vettool=[path to binary] --csv ./... 2>&1 ) \
) | sed '/#/d' | sed '/complexity: /d' > [out file].csv
```
