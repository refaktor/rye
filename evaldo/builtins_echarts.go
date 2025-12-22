//go:build !no_echarts
// +build !no_echarts

package evaldo

import (
	"io"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/refaktor/rye/env"
)

// ---- helpers ----

// blockToLineData converts a Rye Block of integers/decimals into []opts.LineData
func blockToLineData(ps *env.ProgramState, blk env.Block) ([]opts.LineData, *env.Error) {
	items := make([]opts.LineData, 0, blk.Series.Len())
	for i := 0; i < blk.Series.Len(); i++ {
		obj := blk.Series.Get(i)
		switch v := obj.(type) {
		case env.Integer:
			items = append(items, opts.LineData{Value: v.Value})
		case env.Decimal:
			items = append(items, opts.LineData{Value: v.Value})
		default:
			return nil, MakeBuiltinError(ps, "line-data block must contain integers or decimals", "line-data")
		}
	}
	return items, nil
}

// blockToBarData converts a Rye Block of integers/decimals into []opts.BarData
func blockToBarData(ps *env.ProgramState, blk env.Block) ([]opts.BarData, *env.Error) {
	items := make([]opts.BarData, 0, blk.Series.Len())
	for i := 0; i < blk.Series.Len(); i++ {
		obj := blk.Series.Get(i)
		switch v := obj.(type) {
		case env.Integer:
			items = append(items, opts.BarData{Value: v.Value})
		case env.Decimal:
			items = append(items, opts.BarData{Value: v.Value})
		default:
			return nil, MakeBuiltinError(ps, "bar-data block must contain integers or decimals", "bar-data")
		}
	}
	return items, nil
}

// blockToPieData converts a Rye Block (alternating name, value) into []opts.PieData
func blockToPieData(ps *env.ProgramState, blk env.Block) ([]opts.PieData, *env.Error) {
	items := make([]opts.PieData, 0)
	for i := 0; i < blk.Series.Len(); i += 2 {
		if i+1 >= blk.Series.Len() {
			return nil, MakeBuiltinError(ps, "pie-data needs pairs of name and value", "pie-data")
		}
		nameObj := blk.Series.Get(i)
		valueObj := blk.Series.Get(i + 1)

		name, ok := nameObj.(env.String)
		if !ok {
			return nil, MakeBuiltinError(ps, "pie-data name must be string", "pie-data")
		}

		var value interface{}
		switch v := valueObj.(type) {
		case env.Integer:
			value = v.Value
		case env.Decimal:
			value = v.Value
		default:
			return nil, MakeBuiltinError(ps, "pie-data value must be integer or decimal", "pie-data")
		}

		items = append(items, opts.PieData{Name: name.Value, Value: value})
	}
	return items, nil
}

// blockToScatterData converts a Rye Block of 2-element blocks into []opts.ScatterData
func blockToScatterData(ps *env.ProgramState, blk env.Block) ([]opts.ScatterData, *env.Error) {
	items := make([]opts.ScatterData, 0)
	for i := 0; i < blk.Series.Len(); i++ {
		pointObj := blk.Series.Get(i)
		pointBlk, ok := pointObj.(env.Block)
		if !ok || pointBlk.Series.Len() != 2 {
			return nil, MakeBuiltinError(ps, "scatter-data needs blocks of 2 values [x y]", "scatter-data")
		}

		xObj := pointBlk.Series.Get(0)
		yObj := pointBlk.Series.Get(1)

		var x, y interface{}
		switch v := xObj.(type) {
		case env.Integer:
			x = v.Value
		case env.Decimal:
			x = v.Value
		default:
			return nil, MakeBuiltinError(ps, "scatter x value must be integer or decimal", "scatter-data")
		}

		switch v := yObj.(type) {
		case env.Integer:
			y = v.Value
		case env.Decimal:
			y = v.Value
		default:
			return nil, MakeBuiltinError(ps, "scatter y value must be integer or decimal", "scatter-data")
		}

		items = append(items, opts.ScatterData{Value: []interface{}{x, y}})
	}
	return items, nil
}

// blockToStringSlice converts a Rye Block of strings into []string
func blockToStringSlice(ps *env.ProgramState, blk env.Block) ([]string, *env.Error) {
	items := make([]string, 0, blk.Series.Len())
	for i := 0; i < blk.Series.Len(); i++ {
		obj := blk.Series.Get(i)
		switch v := obj.(type) {
		case env.String:
			items = append(items, v.Value)
		default:
			return nil, MakeBuiltinError(ps, "block must contain strings", "set-x-axis")
		}
	}
	return items, nil
}

// ---- builtins ----

var Builtins_echarts = map[string]*env.Builtin{

	//
	// ##### ECharts — Chart Creation #####
	//
	// bar-chart creates a new bar chart instance.
	// Tests:
	// equal { bar-chart |type? } 'native
	// Args: none
	// Returns: native echarts-bar
	"bar-chart": {
		Argsn: 0,
		Doc:   "Creates a new ECharts bar chart instance.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			bar := charts.NewBar()
			return *env.NewNative(ps.Idx, bar, "echarts-bar")
		},
	},

	// bar-data converts a block of numbers into bar data items.
	// Tests:
	// equal { bar-data { 10 20 30 } |type? } 'native
	// Args:
	// * values: Block of integers or decimals
	// Returns: native echarts-bar-data (slice of opts.BarData)
	"bar-data": {
		Argsn: 1,
		Doc:   "Converts a block of numbers into ECharts bar data items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:
				items, err := blockToBarData(ps, blk)
				if err != nil {
					return err
				}
				return *env.NewNative(ps.Idx, items, "echarts-bar-data")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "bar-data")
			}
		},
	},

	// title-opts creates title options for a chart.
	// Tests:
	// equal { title-opts "My Title" "subtitle" |type? } 'native
	// Args:
	// * title: String - main title
	// * subtitle: String - subtitle (optional, can be "")
	// Returns: native echarts-title-opts
	"title-opts": {
		Argsn: 2,
		Doc:   "Creates ECharts title options with title and subtitle.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			title, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "title-opts")
			}
			subtitle, ok := arg1.(env.String)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "title-opts")
			}
			titleOpts := opts.Title{
				Title:    title.Value,
				Subtitle: subtitle.Value,
			}
			return *env.NewNative(ps.Idx, titleOpts, "echarts-title-opts")
		},
	},

	// with-title-opts wraps title options into a chart option function.
	// Tests:
	// equal { title-opts "T" "S" |with-title-opts |type? } 'native
	// Args:
	// * title-opts: native echarts-title-opts
	// Returns: native echarts-global-opts (a charts.GlobalOpts functional option)
	"with-title-opts": {
		Argsn: 1,
		Doc:   "Wraps title options into a chart global option.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch nat := arg0.(type) {
			case env.Native:
				titleOpts, ok := nat.Value.(opts.Title)
				if !ok {
					return MakeNativeArgError(ps, 1, []string{"echarts-title-opts"}, "with-title-opts")
				}
				opt := charts.WithTitleOpts(titleOpts)
				return *env.NewNative(ps.Idx, opt, "echarts-global-opts")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "with-title-opts")
			}
		},
	},

	// echarts-bar//Global-options! sets global options on a bar chart.
	// Tests:
	// equal { bar-chart |Global-options! title-opts "T" "S" |with-title-opts |type? } 'native
	// Args:
	// * bar: native echarts-bar
	// * opts: one or more native echarts-global-opts (variadic via block)
	// Returns: native echarts-bar (same instance for chaining)
	"echarts-bar//Global-options!": {
		Argsn: 2,
		Doc:   "Sets global options on an ECharts bar chart. Accepts a single option or a block of options.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			barNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-bar//Global-options!")
			}
			bar, ok := barNat.Value.(*charts.Bar)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-bar"}, "echarts-bar//Global-options!")
			}

			// Collect options
			var gopts []charts.GlobalOpts
			switch opt := arg1.(type) {
			case env.Native:
				gopt, ok := opt.Value.(charts.GlobalOpts)
				if !ok {
					return MakeNativeArgError(ps, 2, []string{"echarts-global-opts"}, "echarts-bar//Global-options!")
				}
				gopts = append(gopts, gopt)
			case env.Block:
				for i := 0; i < opt.Series.Len(); i++ {
					obj := opt.Series.Get(i)
					nat, ok := obj.(env.Native)
					if !ok {
						return MakeBuiltinError(ps, "block must contain native echarts-global-opts", "echarts-bar//Global-options!")
					}
					gopt, ok := nat.Value.(charts.GlobalOpts)
					if !ok {
						return MakeBuiltinError(ps, "block must contain echarts-global-opts natives", "echarts-bar//Global-options!")
					}
					gopts = append(gopts, gopt)
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.NativeType, env.BlockType}, "echarts-bar//Global-options!")
			}

			bar.SetGlobalOptions(gopts...)
			return barNat
		},
	},

	// echarts-bar//X-axis! sets the x-axis labels on a bar chart.
	// Tests:
	// equal { bar-chart |X-axis! { "Mon" "Tue" } |type? } 'native
	// Args:
	// * bar: native echarts-bar
	// * labels: Block of strings
	// Returns: native echarts-bar (same instance for chaining)
	"echarts-bar//X-axis!": {
		Argsn: 2,
		Doc:   "Sets the x-axis labels on an ECharts bar chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			barNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-bar//X-axis!")
			}
			bar, ok := barNat.Value.(*charts.Bar)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-bar"}, "echarts-bar//X-axis!")
			}
			switch blk := arg1.(type) {
			case env.Block:
				labels, err := blockToStringSlice(ps, blk)
				if err != nil {
					return err
				}
				bar.SetXAxis(labels)
				return barNat
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "echarts-bar//X-axis!")
			}
		},
	},

	// echarts-bar//Add-series adds a data series to a bar chart.
	// Tests:
	// equal { bar-chart |Add-series "Cat A" bar-data { 10 20 30 } |type? } 'native
	// Args:
	// * bar: native echarts-bar
	// * name: String - series name
	// * data: native echarts-bar-data
	// Returns: native echarts-bar (same instance for chaining)
	"echarts-bar//Add-series": {
		Argsn: 3,
		Doc:   "Adds a data series to an ECharts bar chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			barNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-bar//Add-series")
			}
			bar, ok := barNat.Value.(*charts.Bar)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-bar"}, "echarts-bar//Add-series")
			}
			name, ok := arg1.(env.String)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "echarts-bar//Add-series")
			}
			dataNat, ok := arg2.(env.Native)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.NativeType}, "echarts-bar//Add-series")
			}
			data, ok := dataNat.Value.([]opts.BarData)
			if !ok {
				return MakeNativeArgError(ps, 3, []string{"echarts-bar-data"}, "echarts-bar//Add-series")
			}
			bar.AddSeries(name.Value, data)
			return barNat
		},
	},

	// echarts-bar//Render renders the bar chart to a writer (e.g., a file).
	// Tests:
	// ; This would need file I/O to test properly
	// Args:
	// * bar: native echarts-bar
	// * writer: native (io.Writer, e.g., from open//write)
	// Returns: native echarts-bar on success
	"echarts-bar//Render": {
		Argsn: 2,
		Doc:   "Renders an ECharts bar chart to a writer (file, etc.).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			barNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-bar//Render")
			}
			bar, ok := barNat.Value.(*charts.Bar)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-bar"}, "echarts-bar//Render")
			}
			writerNat, ok := arg1.(env.Native)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.NativeType}, "echarts-bar//Render")
			}
			writer, ok := writerNat.Value.(io.Writer)
			if !ok {
				return MakeNativeArgError(ps, 2, []string{"io.Writer (e.g., file)"}, "echarts-bar//Render")
			}
			err := bar.Render(writer)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "echarts-bar//Render")
			}
			return barNat
		},
	},

	//
	// ##### ECharts — Line Chart #####
	//

	"line-chart": {
		Argsn: 0,
		Doc:   "Creates a new ECharts line chart instance.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			line := charts.NewLine()
			return *env.NewNative(ps.Idx, line, "echarts-line")
		},
	},

	"line-data": {
		Argsn: 1,
		Doc:   "Converts a block of numbers into ECharts line data items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:
				items, err := blockToLineData(ps, blk)
				if err != nil {
					return err
				}
				return *env.NewNative(ps.Idx, items, "echarts-line-data")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "line-data")
			}
		},
	},

	"echarts-line//Global-options!": {
		Argsn: 2,
		Doc:   "Sets global options on an ECharts line chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			lineNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-line//Global-options!")
			}
			line, ok := lineNat.Value.(*charts.Line)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-line"}, "echarts-line//Global-options!")
			}

			var gopts []charts.GlobalOpts
			switch opt := arg1.(type) {
			case env.Native:
				gopt, ok := opt.Value.(charts.GlobalOpts)
				if !ok {
					return MakeNativeArgError(ps, 2, []string{"echarts-global-opts"}, "echarts-line//Global-options!")
				}
				gopts = append(gopts, gopt)
			case env.Block:
				for i := 0; i < opt.Series.Len(); i++ {
					obj := opt.Series.Get(i)
					nat, ok := obj.(env.Native)
					if !ok {
						return MakeBuiltinError(ps, "block must contain native echarts-global-opts", "echarts-line//Global-options!")
					}
					gopt, ok := nat.Value.(charts.GlobalOpts)
					if !ok {
						return MakeBuiltinError(ps, "block must contain echarts-global-opts natives", "echarts-line//Global-options!")
					}
					gopts = append(gopts, gopt)
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.NativeType, env.BlockType}, "echarts-line//Global-options!")
			}

			line.SetGlobalOptions(gopts...)
			return lineNat
		},
	},

	"echarts-line//X-axis!": {
		Argsn: 2,
		Doc:   "Sets the x-axis labels on an ECharts line chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			lineNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-line//X-axis!")
			}
			line, ok := lineNat.Value.(*charts.Line)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-line"}, "echarts-line//X-axis!")
			}
			switch blk := arg1.(type) {
			case env.Block:
				labels, err := blockToStringSlice(ps, blk)
				if err != nil {
					return err
				}
				line.SetXAxis(labels)
				return lineNat
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "echarts-line//X-axis!")
			}
		},
	},

	"echarts-line//Add-series": {
		Argsn: 3,
		Doc:   "Adds a data series to an ECharts line chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			lineNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-line//Add-series")
			}
			line, ok := lineNat.Value.(*charts.Line)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-line"}, "echarts-line//Add-series")
			}
			name, ok := arg1.(env.String)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "echarts-line//Add-series")
			}
			dataNat, ok := arg2.(env.Native)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.NativeType}, "echarts-line//Add-series")
			}
			data, ok := dataNat.Value.([]opts.LineData)
			if !ok {
				return MakeNativeArgError(ps, 3, []string{"echarts-line-data"}, "echarts-line//Add-series")
			}
			line.AddSeries(name.Value, data)
			return lineNat
		},
	},

	"echarts-line//Render": {
		Argsn: 2,
		Doc:   "Renders an ECharts line chart to a writer (file, etc.).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			lineNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-line//Render")
			}
			line, ok := lineNat.Value.(*charts.Line)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-line"}, "echarts-line//Render")
			}
			writerNat, ok := arg1.(env.Native)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.NativeType}, "echarts-line//Render")
			}
			writer, ok := writerNat.Value.(io.Writer)
			if !ok {
				return MakeNativeArgError(ps, 2, []string{"io.Writer"}, "echarts-line//Render")
			}
			err := line.Render(writer)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "echarts-line//Render")
			}
			return lineNat
		},
	},

	//
	// ##### ECharts — Pie Chart #####
	//

	"pie-chart": {
		Argsn: 0,
		Doc:   "Creates a new ECharts pie chart instance.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pie := charts.NewPie()
			return *env.NewNative(ps.Idx, pie, "echarts-pie")
		},
	},

	"pie-data": {
		Argsn: 1,
		Doc:   "Converts a block of name-value pairs into ECharts pie data items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:
				items, err := blockToPieData(ps, blk)
				if err != nil {
					return err
				}
				return *env.NewNative(ps.Idx, items, "echarts-pie-data")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pie-data")
			}
		},
	},

	"echarts-pie//Global-options!": {
		Argsn: 2,
		Doc:   "Sets global options on an ECharts pie chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pieNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-pie//Global-options!")
			}
			pie, ok := pieNat.Value.(*charts.Pie)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-pie"}, "echarts-pie//Global-options!")
			}

			var gopts []charts.GlobalOpts
			switch opt := arg1.(type) {
			case env.Native:
				gopt, ok := opt.Value.(charts.GlobalOpts)
				if !ok {
					return MakeNativeArgError(ps, 2, []string{"echarts-global-opts"}, "echarts-pie//Global-options!")
				}
				gopts = append(gopts, gopt)
			case env.Block:
				for i := 0; i < opt.Series.Len(); i++ {
					obj := opt.Series.Get(i)
					nat, ok := obj.(env.Native)
					if !ok {
						return MakeBuiltinError(ps, "block must contain native echarts-global-opts", "echarts-pie//Global-options!")
					}
					gopt, ok := nat.Value.(charts.GlobalOpts)
					if !ok {
						return MakeBuiltinError(ps, "block must contain echarts-global-opts natives", "echarts-pie//Global-options!")
					}
					gopts = append(gopts, gopt)
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.NativeType, env.BlockType}, "echarts-pie//Global-options!")
			}

			pie.SetGlobalOptions(gopts...)
			return pieNat
		},
	},

	"echarts-pie//Add-series": {
		Argsn: 2,
		Doc:   "Adds a data series to an ECharts pie chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pieNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-pie//Add-series")
			}
			pie, ok := pieNat.Value.(*charts.Pie)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-pie"}, "echarts-pie//Add-series")
			}
			dataNat, ok := arg1.(env.Native)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.NativeType}, "echarts-pie//Add-series")
			}
			data, ok := dataNat.Value.([]opts.PieData)
			if !ok {
				return MakeNativeArgError(ps, 2, []string{"echarts-pie-data"}, "echarts-pie//Add-series")
			}
			pie.AddSeries("pie", data)
			return pieNat
		},
	},

	"echarts-pie//Render": {
		Argsn: 2,
		Doc:   "Renders an ECharts pie chart to a writer (file, etc.).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pieNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-pie//Render")
			}
			pie, ok := pieNat.Value.(*charts.Pie)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-pie"}, "echarts-pie//Render")
			}
			writerNat, ok := arg1.(env.Native)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.NativeType}, "echarts-pie//Render")
			}
			writer, ok := writerNat.Value.(io.Writer)
			if !ok {
				return MakeNativeArgError(ps, 2, []string{"io.Writer"}, "echarts-pie//Render")
			}
			err := pie.Render(writer)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "echarts-pie//Render")
			}
			return pieNat
		},
	},

	//
	// ##### ECharts — Scatter Chart #####
	//

	"scatter-chart": {
		Argsn: 0,
		Doc:   "Creates a new ECharts scatter chart instance.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			scatter := charts.NewScatter()
			return *env.NewNative(ps.Idx, scatter, "echarts-scatter")
		},
	},

	"scatter-data": {
		Argsn: 1,
		Doc:   "Converts a block of [x y] pairs into ECharts scatter data items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:
				items, err := blockToScatterData(ps, blk)
				if err != nil {
					return err
				}
				return *env.NewNative(ps.Idx, items, "echarts-scatter-data")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "scatter-data")
			}
		},
	},

	"echarts-scatter//Global-options!": {
		Argsn: 2,
		Doc:   "Sets global options on an ECharts scatter chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			scatterNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-scatter//Global-options!")
			}
			scatter, ok := scatterNat.Value.(*charts.Scatter)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-scatter"}, "echarts-scatter//Global-options!")
			}

			var gopts []charts.GlobalOpts
			switch opt := arg1.(type) {
			case env.Native:
				gopt, ok := opt.Value.(charts.GlobalOpts)
				if !ok {
					return MakeNativeArgError(ps, 2, []string{"echarts-global-opts"}, "echarts-scatter//Global-options!")
				}
				gopts = append(gopts, gopt)
			case env.Block:
				for i := 0; i < opt.Series.Len(); i++ {
					obj := opt.Series.Get(i)
					nat, ok := obj.(env.Native)
					if !ok {
						return MakeBuiltinError(ps, "block must contain native echarts-global-opts", "echarts-scatter//Global-options!")
					}
					gopt, ok := nat.Value.(charts.GlobalOpts)
					if !ok {
						return MakeBuiltinError(ps, "block must contain echarts-global-opts natives", "echarts-scatter//Global-options!")
					}
					gopts = append(gopts, gopt)
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.NativeType, env.BlockType}, "echarts-scatter//Global-options!")
			}

			scatter.SetGlobalOptions(gopts...)
			return scatterNat
		},
	},

	"echarts-scatter//X-axis!": {
		Argsn: 2,
		Doc:   "Sets the x-axis labels on an ECharts scatter chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			scatterNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-scatter//X-axis!")
			}
			scatter, ok := scatterNat.Value.(*charts.Scatter)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-scatter"}, "echarts-scatter//X-axis!")
			}
			switch blk := arg1.(type) {
			case env.Block:
				labels, err := blockToStringSlice(ps, blk)
				if err != nil {
					return err
				}
				scatter.SetXAxis(labels)
				return scatterNat
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "echarts-scatter//X-axis!")
			}
		},
	},

	"echarts-scatter//Add-series": {
		Argsn: 3,
		Doc:   "Adds a data series to an ECharts scatter chart.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			scatterNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-scatter//Add-series")
			}
			scatter, ok := scatterNat.Value.(*charts.Scatter)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-scatter"}, "echarts-scatter//Add-series")
			}
			name, ok := arg1.(env.String)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "echarts-scatter//Add-series")
			}
			dataNat, ok := arg2.(env.Native)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.NativeType}, "echarts-scatter//Add-series")
			}
			data, ok := dataNat.Value.([]opts.ScatterData)
			if !ok {
				return MakeNativeArgError(ps, 3, []string{"echarts-scatter-data"}, "echarts-scatter//Add-series")
			}
			scatter.AddSeries(name.Value, data)
			return scatterNat
		},
	},

	"echarts-scatter//Render": {
		Argsn: 2,
		Doc:   "Renders an ECharts scatter chart to a writer (file, etc.).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			scatterNat, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "echarts-scatter//Render")
			}
			scatter, ok := scatterNat.Value.(*charts.Scatter)
			if !ok {
				return MakeNativeArgError(ps, 1, []string{"echarts-scatter"}, "echarts-scatter//Render")
			}
			writerNat, ok := arg1.(env.Native)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.NativeType}, "echarts-scatter//Render")
			}
			writer, ok := writerNat.Value.(io.Writer)
			if !ok {
				return MakeNativeArgError(ps, 2, []string{"io.Writer"}, "echarts-scatter//Render")
			}
			err := scatter.Render(writer)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "echarts-scatter//Render")
			}
			return scatterNat
		},
	},

	//
	// ##### ECharts — Additional Global Options #####
	//

	"legend-opts": {
		Argsn: 0,
		Doc:   "Creates ECharts legend options.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			legendOpts := opts.Legend{}
			return *env.NewNative(ps.Idx, legendOpts, "echarts-legend-opts")
		},
	},

	"with-legend-opts": {
		Argsn: 1,
		Doc:   "Wraps legend options into a chart global option.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch nat := arg0.(type) {
			case env.Native:
				legendOpts, ok := nat.Value.(opts.Legend)
				if !ok {
					return MakeNativeArgError(ps, 1, []string{"echarts-legend-opts"}, "with-legend-opts")
				}
				opt := charts.WithLegendOpts(legendOpts)
				return *env.NewNative(ps.Idx, opt, "echarts-global-opts")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "with-legend-opts")
			}
		},
	},

	"tooltip-opts": {
		Argsn: 0,
		Doc:   "Creates ECharts tooltip options.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tooltipOpts := opts.Tooltip{}
			return *env.NewNative(ps.Idx, tooltipOpts, "echarts-tooltip-opts")
		},
	},

	"with-tooltip-opts": {
		Argsn: 1,
		Doc:   "Wraps tooltip options into a chart global option.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch nat := arg0.(type) {
			case env.Native:
				tooltipOpts, ok := nat.Value.(opts.Tooltip)
				if !ok {
					return MakeNativeArgError(ps, 1, []string{"echarts-tooltip-opts"}, "with-tooltip-opts")
				}
				opt := charts.WithTooltipOpts(tooltipOpts)
				return *env.NewNative(ps.Idx, opt, "echarts-global-opts")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "with-tooltip-opts")
			}
		},
	},
}
