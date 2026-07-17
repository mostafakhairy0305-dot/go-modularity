package domain

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity/gomodularity"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Format selects a report encoding.
type Format string

const (
	// FormatText renders a human-readable report.
	FormatText Format = "text"
	// FormatJSON renders the versioned JSON schema.
	FormatJSON Format = "json"
	// FormatCSV renders one row per entity and metric.
	FormatCSV Format = "csv"
	// FormatWeb renders a self-contained interactive HTML report.
	FormatWeb Format = "web"
)

// ParseFormat validates a format name.
func ParseFormat(name string) (Format, bool) {
	switch Format(name) {
	case FormatText, FormatJSON, FormatCSV, FormatWeb:
		return Format(name), true
	}

	return "", false
}

// FormatValue renders a metric value deterministically: the shortest
// decimal representation that round-trips, identical on every platform.
func FormatValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

// TextOptions configures the text renderer.
type TextOptions struct {
	// Color wraps values in ANSI quality colors. Callers enable it only
	// when the destination understands escapes (a terminal).
	Color bool
	// Explain appends a notes section with the reasons behind n/a cells
	// and dropped metric components.
	Explain bool
}

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

// naCell marks a value that must not be read (metric not applicable).
const naCell = "–"

// metricQuality maps a metric's values to quality colors. Bounded metrics
// live in [0, 1] and use fixed thresholds; unbounded ones are judged
// relative to the other types in the same package. Metrics with no entry
// (abstractness, instability) have no inherent good/bad direction.
type metricQuality struct {
	lowerBetter bool
	bounded     bool
}

var qualityByMetric = map[string]metricQuality{
	metrics.MetricCAMC:        {bounded: true},
	metrics.MetricTCC:         {bounded: true},
	metrics.MetricReusability: {bounded: true},
	metrics.MetricLCOM96b:     {lowerBetter: true, bounded: true},
	metrics.MetricDistance:    {lowerBetter: true, bounded: true},
	metrics.MetricAMC:         {lowerBetter: true},
	metrics.MetricLCOM1:       {lowerBetter: true},
	metrics.MetricCBO:         {lowerBetter: true},
}

// columnAbbrev holds the short table headings. Single-word metric names get
// a short, readable abbreviation; standard metric initialisms (AMC, LCOM1,
// CAMC, TCC, CBO) fall through to their uppercase name.
var columnAbbrev = map[string]string{
	metrics.MetricAbstractness: "Abst",
	metrics.MetricInstability:  "Inst",
	metrics.MetricDistance:     "Dist",
	metrics.MetricLCOM96b:      "LCOM96b",
	metrics.MetricReusability:  "Reuse",
}

func abbrev(name string) string {
	if short, ok := columnAbbrev[name]; ok {
		return short
	}

	return strings.ToUpper(name)
}

// tableCell is one table cell: an optional unstyled prefix (tree glyphs),
// the visible text, and its ANSI style.
type tableCell struct {
	prefix string
	text   string
	style  string
}

func (c tableCell) width() int {
	return utf8.RuneCountInString(c.prefix) + utf8.RuneCountInString(c.text)
}

// treeNode is one node of the package tree: a path segment (or several,
// when single-child chains are compressed) and, for package nodes, the
// package's report. Aggregates summarize the node's whole subtree.
type treeNode struct {
	name     string
	pkg      *gomodularity.PackageReport
	children []*treeNode

	typesTotal int                     // types in the subtree, applicable or not
	typeAgg    map[string]*columnStats // subtree type metrics
	pkgAgg     map[string]*columnStats // subtree package metrics
}

func (n *treeNode) child(name string) *treeNode {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}

	c := &treeNode{name: name}
	n.children = append(n.children, c)

	return c
}

// aggregate fills the subtree summaries: the mean inputs and value ranges
// of every type metric below the node, and of every package metric, so
// package rows can carry their types' means and directories the means of
// everything they contain.
func (n *treeNode) aggregate() {
	n.typeAgg = make(map[string]*columnStats)

	n.pkgAgg = make(map[string]*columnStats)
	if n.pkg != nil {
		n.typesTotal = len(n.pkg.Types)
		collectPackageStats(n.pkg, n.typeAgg, n.pkgAgg)
	}

	for _, c := range n.children {
		c.aggregate()
		n.typesTotal += c.typesTotal
		mergeStats(n.typeAgg, c.typeAgg)
		mergeStats(n.pkgAgg, c.pkgAgg)
	}
}

// collectPackageStats feeds one package's applicable metric values into the
// aggregation maps.
func collectPackageStats(pkg *gomodularity.PackageReport, typeAgg, pkgAgg map[string]*columnStats) {
	for i := range pkg.Types {
		for _, r := range pkg.Types[i].Metrics {
			if r.Applicable {
				addStat(typeAgg, r.Name, r.Value)
			}
		}
	}

	for _, r := range pkg.Metrics {
		if r.Applicable {
			addStat(pkgAgg, r.Name, r.Value)
		}
	}
}

func addStat(m map[string]*columnStats, name string, value float64) {
	st := m[name]
	if st == nil {
		st = &columnStats{min: value, max: value}
		m[name] = st
	}

	st.sum += value
	st.count++
	st.min = math.Min(st.min, value)
	st.max = math.Max(st.max, value)
}

func mergeStats(dst, src map[string]*columnStats) {
	for name, s := range src {
		d := dst[name]
		if d == nil {
			c := *s
			dst[name] = &c

			continue
		}

		d.sum += s.sum
		d.count += s.count
		d.min = math.Min(d.min, s.min)
		d.max = math.Max(d.max, s.max)
	}
}

// compress merges directory chains with a single child and no package of
// their own, so "internal" → "features" renders as "internal/features".
func (n *treeNode) compress() {
	for n.pkg == nil && len(n.children) == 1 {
		c := n.children[0]
		n.name = n.name + "/" + c.name
		n.pkg = c.pkg
		n.children = c.children
	}

	for _, c := range n.children {
		c.compress()
	}
}

// buildTree arranges the report's packages by their module-relative paths.
// The returned root carries the module-root package (if analyzed) and one
// child per top-level path segment.
func buildTree(report gomodularity.Report) *treeNode {
	root := &treeNode{}

	for i := range report.Packages {
		pkg := &report.Packages[i]

		rel := relPath(pkg.Path, report.Module)
		if rel == "." {
			root.pkg = pkg

			continue
		}

		node := root
		for seg := range strings.SplitSeq(rel, "/") {
			node = node.child(seg)
		}

		node.pkg = pkg
	}

	for _, c := range root.children {
		c.compress()
	}

	return root
}

func relPath(path, module string) string {
	if module == "" {
		return path
	}

	if path == module {
		return "."
	}

	if strings.HasPrefix(path, module+"/") {
		return path[len(module)+1:]
	}

	return path
}

// Text renders the whole report as one tree table: package paths branch
// from the module root, each package row carries its package metrics plus
// the means of its types' metrics, directory rows carry the means of
// everything they contain, and types follow as branches with their metric
// columns. Means average applicable values only. With Explain, the
// reasons behind n/a cells follow as a notes section.
func Text(report gomodularity.Report, opts TextOptions) string {
	var b strings.Builder
	b.WriteString(report.Tool.Name)
	b.WriteString(" ")
	b.WriteString(report.Tool.Version)
	b.WriteString(" — schema ")
	b.WriteString(report.SchemaVersion)
	b.WriteString("\n")

	if report.Module != "" {
		b.WriteString("module ")
		b.WriteString(paint(report.Module, ansiBold, opts.Color))
		b.WriteString("\n")
	}

	if len(report.Packages) == 0 {
		return b.String()
	}

	b.WriteString("\n")

	table := &textTable{
		pkgCols:  packageColumns(report),
		typeCols: reportColumns(report),
	}

	header := make([]tableCell, 0, 1+len(table.pkgCols)+len(table.typeCols))

	header = append(header, tableCell{text: "PATH / TYPE", style: ansiDim})
	for _, name := range append(append([]string{}, table.pkgCols...), table.typeCols...) {
		header = append(header, tableCell{text: abbrev(name), style: ansiDim})
	}

	table.rows = append(table.rows, header)

	root := buildTree(report)

	sections := make([]*treeNode, 0, len(root.children)+1)
	if root.pkg != nil {
		sections = append(sections, &treeNode{name: ".", pkg: root.pkg})
	}

	sections = append(sections, root.children...)
	for i, section := range sections {
		section.aggregate()

		if i > 0 {
			table.rows = append(table.rows, nil)
		}

		table.emitNode(section, "", "")
	}

	// Global column widths keep every branch aligned as one table.
	sawNA := false

	widths := make([]int, 1+len(table.pkgCols)+len(table.typeCols))
	for _, row := range table.rows {
		for c, cell := range row {
			widths[c] = max(widths[c], cell.width())
			if cell.text == naCell {
				sawNA = true
			}
		}
	}

	for _, row := range table.rows {
		if len(row) == 0 {
			b.WriteString("\n")

			continue
		}

		last := len(row) - 1
		for last > 0 && row[last].text == "" && row[last].prefix == "" {
			last--
		}

		for c, cell := range row[:last+1] {
			b.WriteString(cell.prefix)
			b.WriteString(paint(cell.text, cell.style, opts.Color))

			if c < last {
				b.WriteString(strings.Repeat(" ", widths[c]-cell.width()+2))
			}
		}

		b.WriteString("\n")
	}

	if sawNA {
		b.WriteString("\n")
		b.WriteString(paint(naCell+" = not applicable", ansiDim, opts.Color))
		b.WriteString("\n")
	}

	if opts.Explain {
		writeNotes(&b, report, opts.Color)
	}

	return b.String()
}

// textTable accumulates the tree table's rows before column widths are
// known. A nil row renders as a blank separator line.
type textTable struct {
	pkgCols  []string
	typeCols []string
	rows     [][]tableCell
}

// naTableCell renders a value that must not be read.
func naTableCell() tableCell {
	return tableCell{text: naCell, style: ansiDim}
}

// emitNode appends one package or directory node and its subtree: the node
// row, its type branches, then child nodes, drawing the tree glyphs.
func (t *textTable) emitNode(node *treeNode, prefix, connector string) {
	t.nodeRow(node, prefix+connector)

	childPrefix := childIndent(prefix, connector)
	typeCount := nodeTypeCount(node, t.typeCols)
	total := typeCount + len(node.children)

	for i := range typeCount {
		t.typeRow(node, i, childPrefix, branchGlyph(i, total))
	}

	for i, child := range node.children {
		if typeCount+i > 0 {
			t.rows = append(t.rows, []tableCell{{prefix: childPrefix + "│"}})
		}

		t.emitNode(child, childPrefix, branchGlyph(typeCount+i, total))
	}
}

// nodeTypeCount is the number of type rows a node contributes: its package's
// types when any type column is displayed, otherwise none.
func nodeTypeCount(node *treeNode, typeCols []string) int {
	if node.pkg != nil && len(typeCols) > 0 {
		return len(node.pkg.Types)
	}

	return 0
}

// branchGlyph is the tree connector for the child at index of total: the
// corner glyph for the last child, a tee for the rest.
func branchGlyph(index, total int) string {
	if index == total-1 {
		return "└── "
	}

	return "├── "
}

// childIndent extends a node's prefix for its children, continuing the
// vertical guide under a tee and leaving blank space under a corner.
func childIndent(prefix, connector string) string {
	switch connector {
	case "├── ":
		return prefix + "│   "
	case "└── ":
		return prefix + "    "
	}

	return prefix
}

// nodeRow appends the spanning row of one package or directory node: its
// own package metrics (or, for directories, the means of the contained
// packages) followed by the means over all types in its subtree.
func (t *textTable) nodeRow(node *treeNode, label string) {
	row := make([]tableCell, 0, 1+len(t.pkgCols)+len(t.typeCols))
	row = append(row, tableCell{prefix: label, text: node.name, style: ansiBold})
	row = append(row, nodePkgCells(node, t.pkgCols)...)
	row = append(row, nodeTypeAggCells(node, t.typeCols)...)

	t.rows = append(t.rows, row)
}

// nodePkgCells renders a node's package columns: the package's own metric
// values, or the mean over the contained packages for a directory node.
func nodePkgCells(node *treeNode, pkgCols []string) []tableCell {
	if node.pkg != nil {
		return packageMetricCells(node.pkg, pkgCols)
	}

	cells := make([]tableCell, 0, len(pkgCols))
	for _, name := range pkgCols {
		cells = append(cells, meanCell(node.pkgAgg[name], boundedColorFor(name)))
	}

	return cells
}

// nodeTypeAggCells renders a node's type columns as the means over all types
// in its subtree; empty when the subtree holds no types.
func nodeTypeAggCells(node *treeNode, typeCols []string) []tableCell {
	if node.typesTotal == 0 {
		return nil
	}

	cells := make([]tableCell, 0, len(typeCols))
	for _, name := range typeCols {
		st := node.typeAgg[name]
		cells = append(
			cells,
			meanCell(st, func(v float64) string { return valueColor(name, v, st) }),
		)
	}

	return cells
}

// typeRow appends one type's branch row with its metric values, colored
// against the subtree's column ranges.
func (t *textTable) typeRow(node *treeNode, index int, prefix, connector string) {
	typ := &node.pkg.Types[index]
	byName := metricsByName(typ.Metrics)

	row := make([]tableCell, 0, 1+len(t.pkgCols)+len(t.typeCols))
	row = append(row, tableCell{prefix: prefix + connector, text: typ.Name})

	for range t.pkgCols {
		row = append(row, tableCell{})
	}

	row = append(row, typeMetricCells(byName, t.typeCols, node)...)

	t.rows = append(t.rows, row)
}

// typeMetricCells renders one type's metric columns, coloring each applicable
// value against the subtree's column range and marking the rest n/a.
func typeMetricCells(
	byName map[string]metrics.MetricResult,
	typeCols []string,
	node *treeNode,
) []tableCell {
	cells := make([]tableCell, 0, len(typeCols))
	for _, name := range typeCols {
		r, ok := byName[name]
		if !ok || !r.Applicable {
			cells = append(cells, naTableCell())

			continue
		}

		cells = append(cells, tableCell{
			text:  formatCell(r.Value),
			style: valueColor(name, r.Value, node.typeAgg[name]),
		})
	}

	return cells
}

// metricsByName indexes metric results by their metric name.
func metricsByName(results []metrics.MetricResult) map[string]metrics.MetricResult {
	byName := make(map[string]metrics.MetricResult, len(results))
	for _, r := range results {
		byName[r.Name] = r
	}

	return byName
}

// packageMetricCells renders a package row's own metric values in column
// order; blank cells fill metrics absent from the display set.
func packageMetricCells(pkg *gomodularity.PackageReport, cols []string) []tableCell {
	byName := metricsByName(pkg.Metrics)

	cells := make([]tableCell, 0, len(cols))
	for _, name := range cols {
		r, ok := byName[name]
		switch {
		case !ok:
			cells = append(cells, tableCell{})
		case !r.Applicable:
			cells = append(cells, naTableCell())
		default:
			cells = append(
				cells,
				tableCell{text: formatCell(r.Value), style: ansiBold + boundedColor(name, r.Value)},
			)
		}
	}

	return cells
}

// meanCell renders one aggregated column: the mean of the applicable
// values, or the n/a marker when none exist.
func meanCell(st *columnStats, color func(float64) string) tableCell {
	if st == nil || st.count == 0 {
		return naTableCell()
	}

	value := st.sum / float64(st.count)

	return tableCell{text: formatCell(value), style: ansiBold + color(value)}
}

// boundedColorFor adapts boundedColor to the meanCell color callback.
func boundedColorFor(name string) func(float64) string {
	return func(value float64) string { return boundedColor(name, value) }
}

// columnStats aggregates one metric column: the mean input and the value
// range used for relative coloring of unbounded metrics.
type columnStats struct {
	sum   float64
	count int
	min   float64
	max   float64
}

// packageColumns lists the package-level metrics present anywhere in the
// report, in the fixed metric order.
func packageColumns(report gomodularity.Report) []string {
	present := make(map[string]bool)

	for i := range report.Packages {
		for _, r := range report.Packages[i].Metrics {
			present[r.Name] = true
		}
	}

	var cols []string

	for _, name := range metrics.PackageMetricOrder() {
		if present[name] {
			cols = append(cols, name)
		}
	}

	return cols
}

// reportColumns lists the type-level metrics present anywhere in the
// report, in the fixed metric order.
func reportColumns(report gomodularity.Report) []string {
	present := make(map[string]bool)

	for i := range report.Packages {
		for j := range report.Packages[i].Types {
			for _, r := range report.Packages[i].Types[j].Metrics {
				present[r.Name] = true
			}
		}
	}

	var cols []string

	for _, name := range metrics.TypeMetricOrder() {
		if present[name] {
			cols = append(cols, name)
		}
	}

	return cols
}

// writeNotes appends the notes section: per package, the reasons dropped
// from table cells.
func writeNotes(b *strings.Builder, report gomodularity.Report, color bool) {
	wrote := false

	for i := range report.Packages {
		pkg := &report.Packages[i]

		notes := packageNotes(pkg)
		if len(notes) == 0 {
			continue
		}

		if !wrote {
			b.WriteString("\n")
			b.WriteString(paint("notes", ansiDim, color))
			b.WriteString("\n")

			wrote = true
		}

		b.WriteString("  ")
		b.WriteString(paint(pkg.Path, ansiDim, color))
		b.WriteString("\n")

		for _, note := range notes {
			b.WriteString("    ")
			b.WriteString(paint(note, ansiDim, color))
			b.WriteString("\n")
		}
	}
}

// packageNotes collects the reasons dropped from table cells: n/a
// explanations and component notes. Identical reasons are aggregated into
// one line per metric, listing the affected types — otherwise packages
// full of method-less types bury the table under repeated boilerplate.
func packageNotes(pkg *gomodularity.PackageReport) []string {
	var notes []string

	for _, r := range pkg.Metrics {
		if r.Reason != "" {
			notes = append(notes, r.Name+": "+r.Reason)
		}
	}

	for _, name := range metrics.TypeMetricOrder() {
		type entry struct {
			reason string
			types  []string
		}

		var entries []entry

		index := make(map[string]int)

		for i := range pkg.Types {
			for _, r := range pkg.Types[i].Metrics {
				if r.Name != name || r.Reason == "" {
					continue
				}

				j, ok := index[r.Reason]
				if !ok {
					j = len(entries)
					index[r.Reason] = j
					entries = append(entries, entry{reason: r.Reason})
				}

				entries[j].types = append(entries[j].types, pkg.Types[i].Name)
			}
		}

		for _, e := range entries {
			who := strings.Join(e.types, ", ")
			if len(e.types) == len(pkg.Types) && len(pkg.Types) > 1 {
				who = "all types"
			}

			notes = append(notes, name+": "+e.reason+" ("+who+")")
		}
	}

	return notes
}

// formatCell renders a value for a table cell with two decimals, keeping
// columns uniform. Machine formats keep FormatValue.
func formatCell(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

// valueColor picks the quality color for one value of the named metric.
// Unbounded metrics are normalized against their package column's range;
// a column with no spread (or a lone type) stays plain.
func valueColor(name string, value float64, st *columnStats) string {
	q, ok := qualityByMetric[name]
	if !ok {
		return ""
	}

	if !q.bounded {
		if st == nil || st.max == st.min {
			return ""
		}

		value = (value - st.min) / (st.max - st.min)
	}

	return thresholdColor(q.lowerBetter, value)
}

// boundedColor colors package-row values, where no column context exists;
// only metrics with fixed 0–1 thresholds get a color.
func boundedColor(name string, value float64) string {
	q, ok := qualityByMetric[name]
	if !ok || !q.bounded {
		return ""
	}

	return thresholdColor(q.lowerBetter, value)
}

func thresholdColor(lowerBetter bool, score float64) string {
	if lowerBetter {
		score = 1 - score
	}

	switch {
	case score >= 0.66:
		return ansiGreen
	case score >= 0.33:
		return ansiYellow
	default:
		return ansiRed
	}
}

// paint wraps text in an ANSI style when coloring is enabled.
func paint(text, style string, enabled bool) string {
	if !enabled || style == "" {
		return text
	}

	return style + text + ansiReset
}

// CSVHeader is the fixed CSV column set.
func CSVHeader() []string {
	return []string{
		"package",
		"type",
		"metric",
		"scope",
		"value",
		"applicable",
		"reason",
		"definition",
	}
}

// CSVRecords flattens the report into one record per entity and metric, in
// report order.
func CSVRecords(report gomodularity.Report) [][]string {
	var records [][]string

	appendRecords := func(pkgPath, typeName string, results []metrics.MetricResult) {
		for _, r := range results {
			value := ""
			if r.Applicable {
				value = FormatValue(r.Value)
			}

			records = append(records, []string{
				pkgPath, typeName, r.Name, string(r.Scope), value,
				strconv.FormatBool(r.Applicable), r.Reason, r.Definition,
			})
		}
	}

	for i := range report.Packages {
		pkg := &report.Packages[i]
		appendRecords(pkg.Path, "", pkg.Metrics)

		for j := range pkg.Types {
			appendRecords(pkg.Path, pkg.Types[j].Name, pkg.Types[j].Metrics)
		}
	}

	return records
}
