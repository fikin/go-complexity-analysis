package complexity

import (
	"flag"
	"fmt"
	"math"

	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const docComp = "complexity is cyclomatic complexity and maintanability index analyzer"

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "complexity",
	Doc:  docComp,
	Run:  runComp,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

type statsType struct {
	filename       string
	line           int
	col            int
	funcname       string
	loc            int
	constLoc       int
	cyclo          int
	maintenability int
	halsbreadDiff  float64
	halsbreadVol   float64
	timeToCode     float64
	tooComplex     bool
	notMaintenable bool
}

var (
	cycloover  int
	maintunder int
	asCsv      bool
)

func init() {
	flag.IntVar(&cycloover, "cycloover", 10, "print functions with the Cyclomatic complexity > N")
	flag.IntVar(&maintunder, "maintunder", 20, "print functions with the Maintainability index < N")
	flag.BoolVar(&asCsv, "csv", false, "print stats in csv")
}

func runComp(pass *analysis.Pass) (interface{}, error) {
	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("internal error, wrong inspector.Inspector type")
	}
	someCheckFailed := false
	inspector.Preorder([]ast.Node{(*ast.File)(nil)}, func(n ast.Node) {
		astVisitFunctions(n, func(nn *ast.FuncDecl) {
			stats := calcFuncStats(pass, nn)
			printFuncStats(stats)
			if flag.Lookup("test.v") != nil {
				// Only when `go test`
				pass.Reportf(nn.Pos(), "Cyclomatic complexity: %d, Halstead difficulty: %0.3f, volume: %0.3f", stats.cyclo, stats.halsbreadDiff, stats.halsbreadVol)
			} else if stats.tooComplex || stats.notMaintenable {
				someCheckFailed = true
			}
		})
	})
	if someCheckFailed {
		return nil, fmt.Errorf("some functions are too complex or big")
	}
	return nil, nil
}

type branchVisitor func(n ast.Node) (w ast.Visitor)

// Visit is callback from ast to visit the node
func (v branchVisitor) Visit(n ast.Node) (w ast.Visitor) {
	return v(n)
}

func calcFuncStats(pass *analysis.Pass, n *ast.FuncDecl) statsType {
	nPos := n.Pos()
	pos := pass.Fset.File(nPos).Position(nPos)

	stats := statsType{
		filename: pos.Filename,
		line:     pos.Line,
		col:      pos.Column,
		funcname: n.Name.Name,
		loc:      countLOC(pass.Fset, n),
		constLoc: countVarsLOC(pass.Fset, n),
		cyclo:    calcCycloComp(n),
	}
	stats.halsbreadDiff, stats.halsbreadVol = calcHalstComp(n)
	stats.maintenability = calcMaintIndex(stats.halsbreadVol, stats.cyclo, stats.loc)
	stats.tooComplex = stats.cyclo > cycloover
	stats.notMaintenable = stats.maintenability < maintunder
	stats.timeToCode = stats.halsbreadDiff * stats.halsbreadVol / (18 * 3600)

	return stats
}

func astVisitFunctions(n ast.Node, cb func(*ast.FuncDecl)) {
	var v ast.Visitor
	v = branchVisitor(func(nn ast.Node) ast.Visitor {
		switch nnn := nn.(type) {
		case *ast.FuncDecl:
			cb(nnn)
		}
		return v
	})
	ast.Walk(v, n)
}

func calcHalstComp(fd *ast.FuncDecl) (difficulty float64, volume float64) {
	operators, operands := map[string]int{}, map[string]int{}

	walkDecl(fd, operators, operands)

	distOpt := len(operators) // distinct operators
	distOpd := len(operands)  // distinct operands
	var sumOpt, sumOpd int
	for _, val := range operators {
		sumOpt += val
	}

	for _, val := range operands {
		sumOpd += val
	}

	nVocab := distOpt + distOpd
	length := sumOpt + sumOpd
	volume = float64(length) * log2Of(float64(nVocab))
	divisor := float64(2 * distOpd)
	if distOpd == 0 {
		divisor = 0.0000000000001
	}
	difficulty = float64(distOpt*sumOpd) / divisor

	return
}

func walkDecl(n ast.Node, opt map[string]int, opd map[string]int) {
	switch n := n.(type) {
	case *ast.GenDecl:
		appendValidSymb(n.Lparen.IsValid(), n.Rparen.IsValid(), opt, "()")

		if n.Tok.IsOperator() {
			opt[n.Tok.String()]++
		} else {
			opd[n.Tok.String()]++
		}
		for _, s := range n.Specs {
			walkSpec(s, opt, opd)
		}
	case *ast.FuncDecl:
		if n.Recv == nil {
			opt["func"]++
			opt[n.Name.Name]++
			opt["()"]++
		} else {
			opt["func"]++
			opt[n.Name.Name]++
			opt["()"] += 2
		}
		walkStmt(n.Body, opt, opd)
	}
}

func walkStmt(n ast.Node, opt map[string]int, opd map[string]int) {
	switch n := n.(type) {
	case *ast.DeclStmt:
		walkDecl(n.Decl, opt, opd)
	case *ast.ExprStmt:
		walkExpr(n.X, opt, opd)
	case *ast.SendStmt:
		walkExpr(n.Chan, opt, opd)
		if n.Arrow.IsValid() {
			opt["<-"]++
		}
		walkExpr(n.Value, opt, opd)
	case *ast.IncDecStmt:
		walkExpr(n.X, opt, opd)
		if n.Tok.IsOperator() {
			opt[n.Tok.String()]++
		}
	case *ast.AssignStmt:
		if n.Tok.IsOperator() {
			opt[n.Tok.String()]++
		}
		for _, exp := range n.Lhs {
			walkExpr(exp, opt, opd)
		}
		for _, exp := range n.Rhs {
			walkExpr(exp, opt, opd)
		}
	case *ast.GoStmt:
		if n.Go.IsValid() {
			opt["go"]++
		}
		walkExpr(n.Call, opt, opd)
	case *ast.DeferStmt:
		if n.Defer.IsValid() {
			opt["defer"]++
		}
		walkExpr(n.Call, opt, opd)
	case *ast.ReturnStmt:
		if n.Return.IsValid() {
			opt["return"]++
		}
		for _, e := range n.Results {
			walkExpr(e, opt, opd)
		}
	case *ast.BranchStmt:
		if n.Tok.IsOperator() {
			opt[n.Tok.String()]++
		} else {
			opd[n.Tok.String()]++
		}
		if n.Label != nil {
			walkExpr(n.Label, opt, opd)
		}
	case *ast.BlockStmt:
		appendValidSymb(n.Lbrace.IsValid(), n.Rbrace.IsValid(), opt, "{}")
		for _, s := range n.List {
			walkStmt(s, opt, opd)
		}
	case *ast.IfStmt:
		if n.If.IsValid() {
			opt["if"]++
		}
		if n.Init != nil {
			walkStmt(n.Init, opt, opd)
		}
		walkExpr(n.Cond, opt, opd)
		walkStmt(n.Body, opt, opd)
		if n.Else != nil {
			opt["else"]++
			walkStmt(n.Else, opt, opd)
		}
	case *ast.SwitchStmt:
		if n.Switch.IsValid() {
			opt["switch"]++
		}
		if n.Init != nil {
			walkStmt(n.Init, opt, opd)
		}
		if n.Tag != nil {
			walkExpr(n.Tag, opt, opd)
		}
		walkStmt(n.Body, opt, opd)
	case *ast.SelectStmt:
		if n.Select.IsValid() {
			opt["select"]++
		}
		walkStmt(n.Body, opt, opd)
	case *ast.ForStmt:
		if n.For.IsValid() {
			opt["for"]++
		}
		if n.Init != nil {
			walkStmt(n.Init, opt, opd)
		}
		if n.Cond != nil {
			walkExpr(n.Cond, opt, opd)
		}
		if n.Post != nil {
			walkStmt(n.Post, opt, opd)
		}
		walkStmt(n.Body, opt, opd)
	case *ast.RangeStmt:
		if n.For.IsValid() {
			opt["for"]++
		}
		if n.Key != nil {
			walkExpr(n.Key, opt, opd)
			if n.Tok.IsOperator() {
				opt[n.Tok.String()]++
			} else {
				opd[n.Tok.String()]++
			}
		}
		if n.Value != nil {
			walkExpr(n.Value, opt, opd)
		}
		opt["range"]++
		walkExpr(n.X, opt, opd)
		walkStmt(n.Body, opt, opd)
	case *ast.CaseClause:
		if n.List == nil {
			opt["default"]++
		} else {
			for _, c := range n.List {
				walkExpr(c, opt, opd)
			}
		}
		if n.Colon.IsValid() {
			opt[":"]++
		}
		if n.Body != nil {
			for _, b := range n.Body {
				walkStmt(b, opt, opd)
			}
		}
	}
}

func walkSpec(spec ast.Spec, opt map[string]int, opd map[string]int) {
	switch spec := spec.(type) {
	case *ast.ValueSpec:
		for _, n := range spec.Names {
			walkExpr(n, opt, opd)
			if spec.Type != nil {
				walkExpr(spec.Type, opt, opd)
			}
			if spec.Values != nil {
				for _, v := range spec.Values {
					walkExpr(v, opt, opd)
				}
			}
		}
	}
}

func walkExpr(exp ast.Expr, opt map[string]int, opd map[string]int) {
	switch exp := exp.(type) {
	case *ast.ParenExpr:
		appendValidSymb(exp.Lparen.IsValid(), exp.Rparen.IsValid(), opt, "()")
		walkExpr(exp.X, opt, opd)
	case *ast.SelectorExpr:
		walkExpr(exp.X, opt, opd)
		walkExpr(exp.Sel, opt, opd)
	case *ast.IndexExpr:
		walkExpr(exp.X, opt, opd)
		appendValidSymb(exp.Lbrack.IsValid(), exp.Rbrack.IsValid(), opt, "{}")
		walkExpr(exp.Index, opt, opd)
	case *ast.SliceExpr:
		walkExpr(exp.X, opt, opd)
		appendValidSymb(exp.Lbrack.IsValid(), exp.Rbrack.IsValid(), opt, "[]")
		if exp.Low != nil {
			walkExpr(exp.Low, opt, opd)
		}
		if exp.High != nil {
			walkExpr(exp.High, opt, opd)
		}
		if exp.Max != nil {
			walkExpr(exp.Max, opt, opd)
		}
	case *ast.TypeAssertExpr:
		walkExpr(exp.X, opt, opd)
		appendValidSymb(exp.Lparen.IsValid(), exp.Rparen.IsValid(), opt, "()")
		if exp.Type != nil {
			walkExpr(exp.Type, opt, opd)
		}
	case *ast.CallExpr:
		walkExpr(exp.Fun, opt, opd)
		appendValidSymb(exp.Lparen.IsValid(), exp.Rparen.IsValid(), opt, "()")
		if exp.Ellipsis != 0 {
			opt["..."]++
		}
		for _, a := range exp.Args {
			walkExpr(a, opt, opd)
		}
	case *ast.StarExpr:
		if exp.Star.IsValid() {
			opt["*"]++
		}
		walkExpr(exp.X, opt, opd)
	case *ast.UnaryExpr:
		if exp.Op.IsOperator() {
			opt[exp.Op.String()]++
		} else {
			opd[exp.Op.String()]++
		}
		walkExpr(exp.X, opt, opd)
	case *ast.BinaryExpr:
		walkExpr(exp.X, opt, opd)
		opt[exp.Op.String()]++
		walkExpr(exp.Y, opt, opd)
	case *ast.KeyValueExpr:
		walkExpr(exp.Key, opt, opd)
		if exp.Colon.IsValid() {
			opt[":"]++
		}
		walkExpr(exp.Value, opt, opd)
	case *ast.BasicLit:
		if exp.Kind.IsLiteral() {
			opd[exp.Value]++
		} else {
			opt[exp.Value]++
		}
	case *ast.FuncLit:
		walkExpr(exp.Type, opt, opd)
		walkStmt(exp.Body, opt, opd)
	case *ast.CompositeLit:
		appendValidSymb(exp.Lbrace.IsValid(), exp.Rbrace.IsValid(), opt, "{}")
		if exp.Type != nil {
			walkExpr(exp.Type, opt, opd)
		}
		for _, e := range exp.Elts {
			walkExpr(e, opt, opd)
		}
	case *ast.Ident:
		if exp.Obj == nil {
			opt[exp.Name]++
		} else {
			opd[exp.Name]++
		}
	case *ast.Ellipsis:
		if exp.Ellipsis.IsValid() {
			opt["..."]++
		}
		if exp.Elt != nil {
			walkExpr(exp.Elt, opt, opd)
		}
	case *ast.FuncType:
		if exp.Func.IsValid() {
			opt["func"]++
		}
		appendValidSymb(true, true, opt, "()")
		if exp.Params.List != nil {
			for _, f := range exp.Params.List {
				walkExpr(f.Type, opt, opd)
			}
		}
	case *ast.ChanType:
		if exp.Begin.IsValid() {
			opt["chan"]++
		}
		if exp.Arrow.IsValid() {
			opt["<-"]++
		}
		walkExpr(exp.Value, opt, opd)
	}
}

func appendValidSymb(lvalid bool, rvalid bool, opt map[string]int, symb string) {
	if lvalid && rvalid {
		opt[symb]++
	}
}

// calcMaintComp calculates the maintainability index
// source: https://docs.microsoft.com/en-us/archive/blogs/codeanalysis/maintainability-index-range-and-meaning
func calcMaintIndex(halstComp float64, cycloComp, loc int) int {
	origVal := 171.0 - 5.2*logOf(halstComp) - 0.23*float64(cycloComp) - 16.2*logOf(float64(loc))
	normVal := int(math.Max(0.0, origVal*100.0/171.0))
	return normVal
}

func logOf(val float64) float64 {
	switch val {
	case 0:
		return 0
	default:
		return math.Log(val)
	}
}

func log2Of(val float64) float64 {
	switch val {
	case 0:
		return 0
	default:
		return math.Log2(val)
	}
}

// calcCycloComp calculates the Cyclomatic complexity
func calcCycloComp(fd *ast.FuncDecl) int {
	comp := 1
	var v ast.Visitor
	v = branchVisitor(func(n ast.Node) (w ast.Visitor) {
		switch n := n.(type) {
		case *ast.GoStmt: // subroutines are double complexity
			comp += 2
		case *ast.SendStmt: // writing to channels
			comp++
		case *ast.UnaryExpr:
			if n.Op == token.ARROW { // channel reading
				comp++
			}
		case *ast.IfStmt:
			comp++
			if _, ok := n.Else.(*ast.BlockStmt); ok { // include final else
				comp++
			}
		case *ast.ForStmt, *ast.RangeStmt, *ast.SelectStmt, *ast.SwitchStmt:
			comp++
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				comp++
			}
		}
		return v
	})
	ast.Walk(v, fd)

	return comp
}

func countVarsLOC(fs *token.FileSet, n *ast.FuncDecl) int {
	loc := 0
	var v ast.Visitor
	v = branchVisitor(func(nn ast.Node) ast.Visitor {
		switch nnn := nn.(type) {
		case *ast.ValueSpec:
			loc += countLOC(fs, nn)
		case *ast.AssignStmt:
			if nnn.Tok == token.DEFINE { // variable declaration & assignment
				loc += countLOC(fs, nn)
			}
		}
		return v
	})
	ast.Walk(v, n)
	return loc
}

// counts lines of a function
func countLOC(fs *token.FileSet, n ast.Node) int {
	f := fs.File(n.Pos())
	startLine := f.Line(n.Pos())
	endLine := f.Line(n.End())
	return endLine - startLine + 1
}

func printFuncStats(stats statsType) {
	if asCsv {
		fmt.Printf("%s,%d,%d,%s,%d,%d,%0.3f,%0.3f,%0.3f,%d,%d,%t,%t\n",
			stats.filename, stats.line, stats.col, stats.funcname,
			stats.cyclo, stats.maintenability, stats.halsbreadDiff,
			stats.halsbreadVol, stats.timeToCode,
			stats.loc, stats.constLoc,
			stats.tooComplex, stats.notMaintenable)
		return
	}
	if stats.tooComplex {
		msg := fmt.Sprintf("func %s seems to be complex (cyclomatic complexity=%d)", stats.funcname, stats.cyclo)
		fmt.Printf("%s:%d:%d: %s\n", stats.filename, stats.line, stats.col, msg)
	}
	if stats.notMaintenable {
		msg := fmt.Sprintf("func %s seems to have low maintainability (maintainability index=%d)", stats.funcname, stats.maintenability)
		fmt.Printf("%s:%d:%d: %s\n", stats.filename, stats.line, stats.col, msg)
	}
}
