//====================================================================================================
// Copyright (C) 2016-present ShIroRRen <http://shiror.ren>.                                         =
//                                                                                                   =
// Licensed under the F2DLPR License.                                                                =
//                                                                                                   =
// YOU MAY NOT USE THIS FILE EXCEPT IN COMPLIANCE WITH THE LICENSE.                                  =
// Provided "AS IS", WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,                                   =
// unless required by applicable law or agreed to in writing.                                        =
//                                                                                                   =
// For the F2DLPR License terms and conditions, visit: <http://license.fileto.download>.             =
//====================================================================================================

package fvv

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var _fwv_type = reflect.TypeOf(FVVV{})

var _escape_table = [math.MaxUint8 + 1]byte{
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'\\': '\\',
}

type FormatOpt uint64

const (
	FmtOptCommon FormatOpt = 0

	FmtOptUseWrapper FormatOpt = 1 << 0
	FmtOptMinify     FormatOpt = 1 << 1

	FmtOptUseCRLF FormatOpt = 1 << 2
	FmtOptUseCR   FormatOpt = 1 << 3

	FmtOptUseSpace2 FormatOpt = 1 << 4
	FmtOptUseSpace4 FormatOpt = 1 << 5

	FmtOptIntBinary FormatOpt = 1 << 6
	FmtOptIntOctal  FormatOpt = 1 << 7
	FmtOptIntHex    FormatOpt = 1 << 8

	FmtOptDigitSep3 FormatOpt = 1 << 9
	FmtOptDigitSep4 FormatOpt = 1 << 10

	FmtOptUseColon  FormatOpt = 1 << 11
	FmtOptFullWidth FormatOpt = 1 << 12

	FmtOptKeepListSingle     FormatOpt = 1 << 13
	FmtOptForceUseSeparator  FormatOpt = 1 << 14
	FmtOptRawMultilineString FormatOpt = 1 << 15

	FmtOptNoDescs      FormatOpt = 1 << 16
	FmtOptNoLinks      FormatOpt = 1 << 17
	FmtOptFlattenPaths FormatOpt = 1 << 18
	FmtOptFWWStyle     FormatOpt = 1 << 19
)

type FVVV struct {
	Value      any
	Nodes      map[string]*FVVV
	Desc, Link string
}

func NewFVVV(value ...any) (fwv *FVVV) {
	fwv = &FVVV{
		Value: nil,
		Nodes: make(map[string]*FVVV),
	}
	if len(value) >= 1 {
		fwv.Value = value[0]
	}
	return
}

func (_fwv *FVVV) SubNode(ensure bool, raw_paths ...string) (ret *FVVV) {
	ret = _fwv
	if ret == nil {
		return
	}
	if len(raw_paths) == 0 {
		return
	}
	paths := make([]string, 0, len(raw_paths))
	for _, raw_path := range raw_paths {
		if strings.ContainsRune(raw_path, '.') {
			paths = append(paths, strings.Split(raw_path, ".")...)
		} else {
			paths = append(paths, raw_path)
		}
	}
	for _, path := range paths {
		if ret.Nodes == nil {
			if !ensure {
				return nil
			}
			ret.Nodes = make(map[string]*FVVV)
		}
		if ret.Nodes[path] == nil {
			if !ensure {
				return nil
			}
			ret.Nodes[path] = NewFVVV()
		}
		ret = ret.Nodes[path]
	}
	return
}

func (_fwv *FVVV) IsEmpty() bool {
	switch val := _fwv.Value.(type) {
	case bool, int64, int, float64:
		return false
	case string:
		return val == ""
	case []bool:
		return len(val) == 0
	case []int64:
		return len(val) == 0
	case []int:
		return len(val) == 0
	case []float64:
		return len(val) == 0
	case []string:
		return len(val) == 0
	case []*FVVV:
		return len(val) == 0
	default:
		return _fwv.Value == nil
	}
}

func (_fwv *FVVV) IsNotEmpty() bool {
	return !_fwv.IsEmpty()
}

func (_fwv *FVVV) IsNodesEmpty() bool {
	return len(_fwv.Nodes) == 0
}

func (_fwv *FVVV) IsNodesNotEmpty() bool {
	return !_fwv.IsNodesEmpty()
}

func Is[tgt_type any](fwv *FVVV) (ok bool) {
	if fwv == nil || fwv.Value == nil {
		return false
	}
	_, ok = fwv.Value.(tgt_type)
	return
}

func IsList[tgt_type any](fwv *FVVV) (ok bool) {
	if fwv == nil || fwv.Value == nil {
		return false
	}
	_, ok = fwv.Value.([]tgt_type)
	return
}

func (_fwv *FVVV) IsBool() bool {
	return Is[bool](_fwv)
}

func (_fwv *FVVV) IsInt() bool {
	return Is[int64](_fwv) || Is[int](_fwv)
}

func (_fwv *FVVV) IsFloat() bool {
	return Is[float64](_fwv)
}

func (_fwv *FVVV) IsString() bool {
	return Is[string](_fwv)
}

func (_fwv *FVVV) IsBoolList() bool {
	return IsList[bool](_fwv)
}

func (_fwv *FVVV) IsIntList() bool {
	return IsList[int64](_fwv) || IsList[int](_fwv)
}

func (_fwv *FVVV) IsFloatList() bool {
	return IsList[float64](_fwv)
}

func (_fwv *FVVV) IsStringList() bool {
	return IsList[string](_fwv)
}

func (_fwv *FVVV) IsFVVVList() bool {
	return IsList[*FVVV](_fwv)
}

func (_fwv *FVVV) IsValue() bool {
	return _fwv.IsNodesEmpty() &&
		(_fwv.IsBool() || _fwv.IsInt() || _fwv.IsFloat() || _fwv.IsString())
}

func (_fwv *FVVV) IsList() bool {
	return _fwv.IsNodesEmpty() &&
		(_fwv.IsBoolList() || _fwv.IsIntList() || _fwv.IsFloatList() || _fwv.IsStringList() || _fwv.IsFVVVList())
}

func Value[tgt_type any](fwv *FVVV, defaultValue ...tgt_type) (ret tgt_type) {
	if Is[tgt_type](fwv) {
		return fwv.Value.(tgt_type)
	}
	if len(defaultValue) >= 1 {
		return defaultValue[0]
	}
	return
}

func List[tgt_type any](fwv *FVVV, defaultValue ...[]tgt_type) (ret []tgt_type) {
	if Is[[]tgt_type](fwv) {
		return fwv.Value.([]tgt_type)
	}
	if len(defaultValue) >= 1 {
		return defaultValue[0]
	}
	return
}

func (_fwv *FVVV) Bool(defaultValue ...bool) bool {
	return Value(_fwv, defaultValue...)
}

func (_fwv *FVVV) Int(defaultValue ...int64) (ret int64) {
	switch val := _fwv.Value.(type) {
	case int64:
		ret = val
	case int:
		ret = int64(val)
	default:
		if len(defaultValue) >= 1 {
			ret = defaultValue[0]
		}
	}
	return
}

func (_fwv *FVVV) Float(defaultValue ...float64) float64 {
	return Value(_fwv, defaultValue...)
}

func (_fwv *FVVV) String(defaultValue ...string) string {
	return Value(_fwv, defaultValue...)
}

func (_fwv *FVVV) BoolList(defaultValue ...[]bool) []bool {
	return List(_fwv, defaultValue...)
}

func (_fwv *FVVV) IntList(defaultValue ...[]int64) (ret []int64) {
	switch val := _fwv.Value.(type) {
	case []int64:
		ret = val
	case []int:
		if val == nil {
			return nil
		}
		ret = make([]int64, len(val))
		for idx, int_val := range val {
			ret[idx] = int64(int_val)
		}
	default:
		if len(defaultValue) >= 1 {
			ret = defaultValue[0]
		}
	}
	return
}

func (_fwv *FVVV) FloatList(defaultValue ...[]float64) []float64 {
	return List(_fwv, defaultValue...)
}

func (_fwv *FVVV) StringList(defaultValue ...[]string) []string {
	return List(_fwv, defaultValue...)
}

func (_fwv *FVVV) FVVVList(defaultValue ...[]*FVVV) []*FVVV {
	return List(_fwv, defaultValue...)
}

func (_fwv *FVVV) Unlink() {
	_fwv.Desc = ""
	for _, value := range _fwv.Nodes {
		value.Unlink()
	}
}

func (_fwv *FVVV) ParseString(text string, targets ...any) (err error) {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	ctx := newTextCtx(text)
	scope_stack := make([]*FVVV, 0, 1)

	ctx.skip_blanks()
	has_wrapper := ctx.match_any('{', '｛')
	if err = _fwv._parse_main(ctx, scope_stack); err != nil {
		return
	}

	ctx.skip_blanks()
	if has_wrapper && (!ctx.match_any('}', '｝')) {
		return ctx.ErrStrNotFound("wrapper")
	}
	ctx.skip_blanks()
	if !ctx.is_eof() {
		return ctx.ErrWhyNotEOF()
	}

	for _, target := range targets {
		if err := _fwv.Unmarshal(target); err != nil {
			return err
		}
	}

	return nil
}

func (_fwv *FVVV) ToString(opts ...FormatOpt) string {
	var flags FormatOpt
	for _, opt := range opts {
		flags |= opt
	}
	ctx := NewFormatCtx(flags)
	var ret strings.Builder

	if ctx.use_wrapper {
		ret.WriteRune(ctx.fwv_begin)
		if !ctx.minify {
			ret.WriteString(ctx.newline)
		}
	}

	var level int
	if ctx.use_wrapper {
		level = 1
	}
	_fwv._to_string_root(ctx, &ret, level)
	if ctx.use_wrapper {
		if !ctx.minify {
			ret.WriteString(ctx.newline)
		}
		ret.WriteRune(ctx.fwv_end)
	}

	return ret.String()
}

func (_fwv *FVVV) Unmarshal(val any) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("expects a non-nil pointer")
	}
	return _to(_fwv, rv.Elem())
}

func (_fwv *FVVV) Marshal(val any) error {
	_fwv.Unlink()
	return _from(_fwv, reflect.ValueOf(val))
}

type textCtx struct {
	input       string
	index       int
	lines_start []int
}

func newTextCtx(input string) *textCtx {
	ctx := &textCtx{
		input:       input,
		index:       0,
		lines_start: []int{0},
	}
	const bom = "\xEF\xBB\xBF"
	if strings.HasPrefix(input, bom) {
		ctx.index += len(bom)
	}
	return ctx
}

func (_ctx *textCtx) prematch(tgts ...rune) bool {
	if _ctx.is_eof() {
		return false
	}
	idx_rune := rune(_ctx.input[_ctx.index])
	if idx_rune >= utf8.RuneSelf {
		idx_rune, _ = utf8.DecodeRuneInString(_ctx.input[_ctx.index:])
	}
	for _, tgt := range tgts {
		if tgt == idx_rune {
			return true
		}
	}
	return false
}

func (_ctx *textCtx) next() rune {
	if _ctx.is_eof() {
		return 0
	}
	byte := _ctx.input[_ctx.index]
	if byte < utf8.RuneSelf {
		_ctx.index++
		switch byte {
		case '\r':
			_ctx.lines_start = append(_ctx.lines_start, _ctx.index)
		case '\n':
			if _ctx.index >= 2 && _ctx.input[_ctx.index-2] == '\r' {
				_ctx.lines_start[len(_ctx.lines_start)-1] = _ctx.index
			} else {
				_ctx.lines_start = append(_ctx.lines_start, _ctx.index)
			}
		}
		return rune(byte)
	}
	rune, size := utf8.DecodeRuneInString(_ctx.input[_ctx.index:])
	_ctx.index += size
	return rune
}

func (_ctx *textCtx) match(tgt rune, bool_args ...bool) bool {
	skip_blanks, same_line := true, false
	if len(bool_args) >= 1 {
		skip_blanks = bool_args[0]
	}
	if len(bool_args) >= 2 {
		same_line = bool_args[1]
	}
	if skip_blanks {
		_ctx.skip_blanks(same_line)
	}
	if _ctx.prematch(tgt) {
		_ctx.next()
		return true
	}
	return false
}

func (_ctx *textCtx) match_any(tgts ...rune) bool {
	_ctx.skip_blanks()
	if _ctx.prematch(tgts...) {
		_ctx.next()
		return true
	}
	return false
}

func (_ctx *textCtx) skip_blanks(same_line_arg ...bool) {
	same_line := len(same_line_arg) >= 1 && same_line_arg[0]
	for !_ctx.is_eof() {
		idx_byte := _ctx.input[_ctx.index]
		if idx_byte < utf8.RuneSelf {
			if same_line && (idx_byte == '\n' || idx_byte == '\r') {
				return
			}
			if unicode.IsSpace(rune(idx_byte)) {
				_ctx.next()
				continue
			}
			return
		}
		idx_rune, _ := utf8.DecodeRuneInString(_ctx.input[_ctx.index:])
		if !unicode.IsSpace(idx_rune) {
			return
		}
		_ctx.next()
	}
}

func (_ctx *textCtx) is_eof() bool {
	return _ctx.index >= len(_ctx.input)
}

func (_ctx *textCtx) is_same_line() bool {
	before := len(_ctx.lines_start)
	_ctx.skip_blanks()
	return before == len(_ctx.lines_start)
}

func (_ctx *textCtx) makeError(msg string, args ...any) error {
	return fmt.Errorf("%d:%d: %s",
		len(_ctx.lines_start),
		_ctx.index-_ctx.lines_start[len(_ctx.lines_start)-1]+1,
		fmt.Sprintf(msg, args...))
}

func (_ctx *textCtx) ErrUnknown() error {
	return _ctx.makeError("Why??? IDK!!!")
}

func (_ctx *textCtx) ErrWhyEOF() error {
	return _ctx.makeError("Why EOF???")
}

func (_ctx *textCtx) ErrWhyNotEOF() error {
	return _ctx.makeError("Why not EOF???")
}

func (_ctx *textCtx) ErrRuneNotFound(tgt rune) error {
	return _ctx.makeError("Where is the '%c'?", tgt)
}

func (_ctx *textCtx) ErrStrNotFound(tgt string) error {
	return _ctx.makeError("Where is the %s?", tgt)
}

func (_ctx *textCtx) ErrNoValue(tgt string) error {
	return _ctx.makeError("Cannot find the value of ‘%s’", tgt)
}

func (_ctx *textCtx) ErrPlusList() error {
	return _ctx.makeError("Why plus with list?")
}

func (_ctx *textCtx) ErrValuePlusFVVV() error {
	return _ctx.makeError("Why value plus with FVVV?")
}

type FormatCtx struct {
	newline              string
	indent_unit          string
	assign_op            string
	list_begin, list_end rune
	fwv_begin, fwv_end   rune
	item_sep, stmt_sep   rune

	int_base int

	digit_sep_step int
	digit_sep_char rune

	use_wrapper bool

	minify bool

	full_width bool

	list_single bool
	force_sep   bool
	raw_str     bool

	no_descs, no_links bool
	flatten_paths      bool
	fww_style          bool
}

func NewFormatCtx(flags FormatOpt) *FormatCtx {
	ctx := &FormatCtx{
		newline:     "\n",
		indent_unit: "\t",
		assign_op:   " = ",
		list_begin:  '[', list_end: ']',
		fwv_begin: '{', fwv_end: '}',
		item_sep: ',', stmt_sep: ';',

		int_base: 10,
	}

	ctx.use_wrapper = flags&FmtOptUseWrapper != 0

	if flags&FmtOptUseCRLF != 0 {
		ctx.newline = "\r\n"
	} else if flags&FmtOptUseCR != 0 {
		ctx.newline = "\r"
	}

	if flags&FmtOptUseSpace2 != 0 {
		ctx.indent_unit = "  "
	} else if flags&FmtOptUseSpace4 != 0 {
		ctx.indent_unit = "    "
	}

	if flags&FmtOptIntHex != 0 {
		ctx.int_base = 16
	} else if flags&FmtOptIntOctal != 0 {
		ctx.int_base = 8
	} else if flags&FmtOptIntBinary != 0 {
		ctx.int_base = 2
	}

	if flags&FmtOptDigitSep3 != 0 {
		ctx.digit_sep_step = 3
	} else if flags&FmtOptDigitSep4 != 0 {
		ctx.digit_sep_step = 4
	}

	if ctx.full_width = flags&FmtOptFullWidth != 0; ctx.full_width {
		if flags&FmtOptUseColon != 0 {
			ctx.assign_op = "："
		}
		ctx.list_begin, ctx.list_end = '［', '］'
		ctx.fwv_begin, ctx.fwv_end = '｛', '｝'
		ctx.item_sep, ctx.stmt_sep = '，', '；'
		if ctx.digit_sep_step > 0 {
			ctx.digit_sep_char = '’'
		}
	} else {
		if flags&FmtOptUseColon != 0 {
			ctx.assign_op = ": "
		}
		if ctx.digit_sep_step > 0 {
			ctx.digit_sep_char = '\''
		}
	}

	ctx.list_single = flags&FmtOptKeepListSingle != 0
	ctx.force_sep = flags&FmtOptForceUseSeparator != 0
	ctx.raw_str = flags&FmtOptRawMultilineString != 0

	ctx.no_descs = flags&FmtOptNoDescs != 0
	ctx.flatten_paths = flags&FmtOptFlattenPaths != 0
	ctx.fww_style = flags&FmtOptFWWStyle != 0

	if ctx.minify = flags&FmtOptMinify != 0; ctx.minify {
		ctx.newline, ctx.indent_unit = "", ""

		ctx.assign_op = strings.TrimSpace(ctx.assign_op)
	}

	return ctx
}

func (_fwv *FVVV) _parse_main(ctx *textCtx, scope_stack []*FVVV) (err error) {
	scope_stack = append(scope_stack, _fwv)

	for {
		var idx_desc strings.Builder
		if err = _parse_desc(ctx, &idx_desc, scope_stack, false); err != nil {
			return
		}
		if !ctx.is_same_line() {
			idx_desc.Reset()
		}

		if ctx.is_eof() || ctx.prematch('}', '｝') {
			break
		}

		name := _parse_name(ctx)
		if name == "" {
			return ctx.ErrStrNotFound("name")
		}
		if err = _parse_desc(ctx, &idx_desc, scope_stack); err != nil {
			return
		}
		if !ctx.match_any('=', ':', '：') {
			return ctx.ErrRuneNotFound('=')
		}
		if err = _parse_desc(ctx, &idx_desc, scope_stack); err != nil {
			return
		}

		tgt_key := _fwv.SubNode(true, name)
		if ctx.match_any('[', '［') {
			tgt_list := make([]any, 0, 1)
			var list_type any
			for {
				var value_desc strings.Builder
				if err = _parse_desc(ctx, &value_desc, scope_stack, false); err != nil {
					return
				}
				if !ctx.is_same_line() {
					value_desc.Reset()
				}

				if ctx.is_eof() {
					return ctx.ErrWhyEOF()
				}
				if ctx.match_any('{', '｛') {
					list_type = _fwv
					tmp_value := NewFVVV()
					if err = tmp_value._parse_main(ctx, scope_stack); err != nil {
						return
					}
					if !ctx.match_any('}', '｝') {
						return ctx.ErrRuneNotFound('}')
					}
					if err = _parse_desc(ctx, &value_desc, scope_stack, false, true); err != nil {
						return
					}
					tmp_value.Desc = value_desc.String()
					tgt_list = append(tgt_list, tmp_value)
					if ctx.is_same_line() && !ctx.match_any(',', '，') && !ctx.prematch(']', '］') {
						return ctx.ErrStrNotFound("EOL")
					}
				} else {
					tgt_fwv := NewFVVV()
					if err = _parse_value(ctx, scope_stack, tgt_fwv, &idx_desc, true); err != nil {
						return
					}
					switch tmp_list := tgt_fwv.Value.(type) {
					case []bool:
						for _, val := range tmp_list {
							tgt_list = append(tgt_list, val)
						}
					case []int64:
						for _, val := range tmp_list {
							tgt_list = append(tgt_list, val)
						}
					case []float64:
						for _, val := range tmp_list {
							tgt_list = append(tgt_list, val)
						}
					case []string:
						for _, val := range tmp_list {
							tgt_list = append(tgt_list, val)
						}
					case []*FVVV:
						for _, val := range tmp_list {
							tgt_list = append(tgt_list, val)
						}
					default:
						if tgt_fwv.Value != nil {
							tgt_list = append(tgt_list, tgt_fwv.Value)
						} else {
							tgt_list = append(tgt_list, tgt_fwv)
						}
					}
					if list_type == nil {
						list_type = _get_type_ptr(tgt_list[len(tgt_list)-1])
					} else if !_is_same_type(list_type, tgt_list[len(tgt_list)-1]) {
						if _is_same_type(list_type, _fwv) ||
							_is_same_type(tgt_list[len(tgt_list)-1], _fwv) {
							return ctx.ErrValuePlusFVVV()
						}
						switch tgt_type := _get_type_ptr(tgt_list[len(tgt_list)-1]); tgt_type.(type) {
						case *string:
							list_type = tgt_type
						case *float64:
							if _, ok := list_type.(*string); !ok {
								list_type = tgt_type
							}
						case *int64:
							if _, ok := list_type.(*string); !ok {
								list_type = tgt_type
							} else if _, ok := list_type.(*float64); !ok {
								list_type = tgt_type
							}
						}
					}
				}
				if ctx.match_any(']', '］') {
					break
				}
			}
			switch list_type.(type) {
			case *string:
				final_list := make([]string, len(tgt_list))
				for idx, val := range tgt_list {
					final_list[idx] = _value_to_string(val)
				}
				tgt_key.Value = final_list

			case *float64:
				final_list := make([]float64, len(tgt_list))
				for idx, val := range tgt_list {
					switch val := val.(type) {
					case float64:
						final_list[idx] = val
					case int64:
						final_list[idx] = float64(val)
					case bool:
						if val {
							final_list[idx] = 1.0
						} else {
							final_list[idx] = 0.0
						}
					}
				}
				tgt_key.Value = final_list

			case *int64:
				final_list := make([]int64, len(tgt_list))
				for idx, val := range tgt_list {
					switch val := val.(type) {
					case int64:
						final_list[idx] = val
					case bool:
						if val {
							final_list[idx] = 1
						} else {
							final_list[idx] = 0
						}
					}
				}
				tgt_key.Value = final_list
			case *bool:
				final_list := make([]bool, len(tgt_list))
				for idx, val := range tgt_list {
					final_list[idx] = val.(bool)
				}
				tgt_key.Value = final_list

			case *FVVV:
				final_list := make([]*FVVV, len(tgt_list))
				for idx, val := range tgt_list {
					final_list[idx] = val.(*FVVV)
				}
				tgt_key.Value = final_list
			}
		} else if ctx.match_any('{', '｛') {
			if err = tgt_key._parse_main(ctx, scope_stack); err != nil {
				return
			}
			if !ctx.match_any('}', '｝') {
				return ctx.ErrRuneNotFound('}')
			}
		} else {
			if err = _parse_value(ctx, scope_stack, tgt_key, &idx_desc); err != nil {
				return
			}
			goto set_desc
		}

		if err = _parse_desc(ctx, &idx_desc, scope_stack, false, true); err != nil {
			return
		}
		if ctx.is_same_line() && !ctx.is_eof() && !ctx.match_any(';', '；') && !ctx.prematch('}', '｝') {
			return ctx.ErrStrNotFound("EOL")
		}
	set_desc:
		tgt_key.Desc = idx_desc.String()
	}
	return nil
}

func _parse_name(ctx *textCtx) string {
	ctx.skip_blanks()
	var name strings.Builder
	for !ctx.is_eof() && !ctx.prematch('=', ':', '：', '<') {
		name.WriteRune(ctx.next())
	}
	if name.Len() == 0 {
		return ""
	}
	return strings.TrimSpace(name.String())
}

func _parse_value(ctx *textCtx, scope_stack []*FVVV, tgt_fwv *FVVV, idx_desc *strings.Builder,
	in_list_arg ...bool,
) (err error) {
	in_list := false
	if len(in_list_arg) >= 1 {
		in_list = in_list_arg[0]
	}
	for {
		if err = _parse_desc(ctx, idx_desc, scope_stack, in_list, !in_list); err != nil {
			return
		}
		if ctx.is_eof() || ((in_list && (ctx.match_any(',', '，') || ctx.prematch(']', '］'))) ||
			(!in_list && (!ctx.is_same_line() || ctx.match_any(';', '；') || ctx.prematch('}', '｝')))) {
			return ctx.ErrStrNotFound("value")
		}

		var tmp_sb strings.Builder
		if ctx.prematch('"', '“', '`') {
			if err = _parse_text(ctx, &tmp_sb); err != nil {
				return
			}
			if tgt_fwv.Value == nil {
				tgt_fwv.Value = tmp_sb.String()
			} else {
				tgt_fwv.Link = ""
				tgt_fwv.Value = _value_to_string(tgt_fwv.Value) + tmp_sb.String()
			}
		} else {
			for !ctx.is_eof() && !ctx.prematch('<', '+') && !ctx.prematch('\r', '\n') {
				if (in_list && ctx.prematch(',', '，', ']', '］')) ||
					(!in_list && ctx.prematch(';', '；', '}', '｝')) {
					break
				}
				tmp_sb.WriteRune(ctx.next())
			}
			tmp_str := strings.TrimSpace(tmp_sb.String())
			if tmp_str == "" {
				return ctx.ErrStrNotFound("value")
			}

			is_true := strings.EqualFold(tmp_str, "true")
			if is_true || strings.EqualFold(tmp_str, "false") {
				if tgt_fwv.Value == nil {
					tgt_fwv.Value = is_true
				} else {
					tgt_fwv.Link = ""
					tgt_fwv.Value = _value_to_string(tgt_fwv.Value) + tmp_str
				}
			} else {
				var tmp_value any
				if _try_parse_number(&tmp_value, tmp_str) {
					if tgt_fwv.Value == nil {
						tgt_fwv.Value = tmp_value
					} else {
						tgt_fwv.Link = ""
						tgt_fwv.Value = _value_to_string(tgt_fwv.Value) + tmp_str
					}
				} else {
					target := _find_key(tmp_str, scope_stack)
					if target != nil {
						if tgt_fwv.Value != nil {
							switch target.Value.(type) {
							case []bool, []int64, []int, []float64, []string, []*FVVV:
								return ctx.ErrPlusList()
							}
						}
						if tgt_fwv.Value == nil {
							tgt_fwv.Link = tmp_str
							tgt_fwv.Value = target.Value
						} else {
							tgt_fwv.Link = ""
							tgt_fwv.Value = _value_to_string(tgt_fwv.Value) + _value_to_string(target.Value)
						}
						tgt_fwv.Nodes = target.Nodes
					} else {
						return ctx.ErrNoValue(tmp_str)
					}
				}
			}
		}
		if err = _parse_desc(ctx, idx_desc, scope_stack, false, true); err != nil {
			return
		}
		if ctx.is_eof() || !ctx.is_same_line() ||
			((in_list && (ctx.match_any(',', '，') || ctx.prematch(']', '］'))) ||
				(!in_list && (ctx.match_any(';', '；') || ctx.prematch('}', '｝')))) {
			return nil
		}

		if ctx.match('+') {
			continue
		}
		return ctx.ErrRuneNotFound('+')
	}
}

func _parse_desc(ctx *textCtx, desc *strings.Builder, scope_stack []*FVVV, bool_args ...bool) error {
	skip_blanks, same_line := false, false
	if len(bool_args) >= 1 {
		skip_blanks = bool_args[0]
	}
	if len(bool_args) >= 2 {
		same_line = bool_args[1]
	}
	for {
		orig_idx, orig_line := ctx.index, len(ctx.lines_start)
		if !ctx.match('<', true, same_line) {
			if !skip_blanks {
				ctx.index = orig_idx
				if len(ctx.lines_start) > orig_line {
					ctx.lines_start = ctx.lines_start[:orig_line]
				}
			}
			break
		}

		desc.Reset()
		for {
			if ctx.is_eof() {
				return ctx.ErrWhyEOF()
			}
			if ctx.match('>', false) {
				if target := _find_key(desc.String(), scope_stack); target != nil && target.IsString() {
					desc.Reset()
					desc.WriteString(target.String())
				}
				break
			}
			if ctx.match('\\', false) {
				if ctx.is_eof() {
					return ctx.ErrWhyEOF()
				}
				if ctx.match('>', false) {
					desc.WriteByte('>')
				} else {
					ch, tgt := ctx.next(), byte(0)
					if ch < rune(len(_escape_table)) {
						tgt = _escape_table[uint8(ch)]
					}
					if tgt != 0 {
						desc.WriteByte(tgt)
					} else {
						desc.WriteByte('\\')
						desc.WriteRune(ch)
					}
				}
			} else {
				desc.WriteRune(ctx.next())
			}
		}
	}
	return nil
}

func _parse_text(ctx *textCtx, text *strings.Builder) error {
	if ctx.match('`') {
		for {
			if ctx.is_eof() {
				return ctx.ErrWhyEOF()
			}
			if ctx.match('`', false) {
				break
			}
			text.WriteRune(ctx.next())
		}
		tmp_str := strings.TrimSpace(_trim_indent(text.String()))
		text.Reset()
		text.WriteString(tmp_str)
		return nil
	}

	is_full_width := ctx.match('“') || !ctx.match('"')
	for {
		if ctx.is_eof() {
			return ctx.ErrWhyEOF()
		}
		if (is_full_width && ctx.match('”', false)) || (!is_full_width && ctx.match('"', false)) {
			return nil
		}
		if ctx.match('\\', false) {
			if ctx.is_eof() {
				return ctx.ErrWhyEOF()
			}
			if is_full_width && ctx.match('”', false) {
				text.WriteRune('”')
			} else if !is_full_width && ctx.match('"', false) {
				text.WriteByte('"')
			} else {
				ch, tgt := ctx.next(), byte(0)
				if ch < rune(len(_escape_table)) {
					tgt = _escape_table[uint8(ch)]
				}
				if tgt != 0 {
					text.WriteByte(tgt)
				} else {
					text.WriteByte('\\')
					text.WriteRune(ch)
				}
			}
		} else {
			text.WriteRune(ctx.next())
		}
	}
}

func _try_parse_number(tgt_value *any, tgt_str string) bool {
	if tgt_str == "" {
		return false
	}

	first := tgt_str[0]
	if !(first >= '0' && first <= '9') && first != '+' && first != '-' && first != '.' {
		return false
	}
	if (first == '+' || first == '-' || first == '.') && len(tgt_str) == 1 {
		return false
	}

	var tmp_sb strings.Builder
	tmp_sb.Grow(len(tgt_str))
	idx := 0
	if tgt_str[idx] == '+' || tgt_str[idx] == '-' {
		tmp_sb.WriteByte(tgt_str[idx])
		idx++
	}

	is_hex, is_oct, is_bin := false, false, false
	if idx+1 < len(tgt_str) && tgt_str[idx] == '0' {
		switch tgt_str[idx+1] {
		case 'x', 'X':
			is_hex = true
			tmp_sb.WriteString("0x")
			idx += 2
		case 'o', 'O':
			is_oct = true
			tmp_sb.WriteString("0o")
			idx += 2
		case 'b', 'B':
			is_bin = true
			tmp_sb.WriteString("0b")
			idx += 2
		case '0', '1', '2', '3', '4', '5', '6', '7':
			is_oct = true
			tmp_sb.WriteByte('0')
			idx++
		}
	}

	has_dot := false
	has_exp := false
	for ; idx < len(tgt_str); idx++ {
		ch := tgt_str[idx]
		if ch == '\'' {
			continue
		}
		if ch >= utf8.RuneSelf {
			if rune, size := utf8.DecodeRuneInString(tgt_str[idx:]); rune == '’' {
				idx += size - 1
				continue
			}
		}

		if is_hex {
			if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
				tmp_sb.WriteByte(ch)
			} else {
				return false
			}
		} else if is_oct {
			if ch >= '0' && ch <= '7' {
				tmp_sb.WriteByte(ch)
			} else {
				return false
			}
		} else if is_bin {
			if ch == '0' || ch == '1' {
				tmp_sb.WriteByte(ch)
			} else {
				return false
			}
		} else if ch >= '0' && ch <= '9' {
			tmp_sb.WriteByte(ch)
		} else if ch == '.' {
			if has_dot {
				return false
			}
			has_dot = true
			tmp_sb.WriteByte(ch)
		} else if ch == 'e' || ch == 'E' {
			if has_exp {
				return false
			}
			has_exp = true
			tmp_sb.WriteByte(ch)

			if idx+1 < len(tgt_str) && (tgt_str[idx+1] == '+' || tgt_str[idx+1] == '-') {
				idx++
				tmp_sb.WriteByte(tgt_str[idx])
			}
		} else {
			return false
		}
	}

	final_str := tmp_sb.String()
	if final_str == "" || final_str == "+" || final_str == "-" {
		return false
	}

	if !is_hex && !is_bin && (has_dot || has_exp) {
		final_value, err := strconv.ParseFloat(final_str, 64)
		if err != nil {
			return false
		}
		*tgt_value = final_value
		return true
	}

	final_value, err := strconv.ParseInt(final_str, 0, 64)
	if err != nil {
		return false
	}

	*tgt_value = final_value
	return true
}

func _find_key(path string, scope_stack []*FVVV) (target *FVVV) {
	paths := strings.Split(path, ".")
	for idx := len(scope_stack) - 1; idx >= 0; idx-- {
		target = scope_stack[idx]
		for _, idx_path := range paths {
			if target.Nodes[idx_path] != nil {
				target = target.Nodes[idx_path]
			} else {
				target = nil
				break
			}
		}
		if target != nil {
			return target
		}
	}
	return
}

func (_fwv *FVVV) _to_string_root(ctx *FormatCtx, ret *strings.Builder, level int) {
	if _fwv.IsNodesEmpty() {
		return
	}

	keys := make([]string, 0, len(_fwv.Nodes))
	for key := range _fwv.Nodes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for idx, key := range keys {
		_fwv.Nodes[key]._to_string_main(ctx, key, ret, level, idx == len(keys)-1)
	}
}

func (_fwv *FVVV) _to_string_main(ctx *FormatCtx, name string, ret *strings.Builder, level int, is_back bool) {
	if _, is_text := _fwv.Value.(string); name == "" || (!is_text && _fwv.IsEmpty() && _fwv.IsNodesEmpty()) {
		return
	}
	ret.Grow(len(_fwv.Nodes) * 6)

	tgt_node := _fwv
	if ctx.flatten_paths {
		var tmp_name strings.Builder
		tmp_name.Grow(len(name) + 6)
		tmp_name.WriteString(name)
		for len(tgt_node.Nodes) == 1 &&
			(ctx.no_descs || tgt_node.Desc == "") &&
			(ctx.no_links || tgt_node.Link == "") {
			for sub_name, sub_node := range tgt_node.Nodes {
				tmp_name.WriteByte('.')
				tmp_name.WriteString(sub_name)
				tgt_node = sub_node
			}
		}
		name = tmp_name.String()
	}

	indent := ""
	if !ctx.minify && level > 0 {
		indent = strings.Repeat(ctx.indent_unit, level)
		ret.WriteString(indent)
	}
	ret.WriteString(name)
	ret.WriteString(ctx.assign_op)

	if !ctx.no_links && tgt_node.Link != "" {
		ret.WriteString(tgt_node.Link)
	} else if tgt_node.IsNodesNotEmpty() {
		if ctx.fww_style && tgt_node.Desc != "" {
			ret.WriteString(_escape_string(tgt_node.Desc, true, ctx.full_width))
			if !ctx.minify {
				ret.WriteByte(' ')
			}
		}
		if ctx.full_width && _sb_last_empty(ret) {
			_sb_trim_last(ret)
		}
		_to_string_fwv(ctx, tgt_node, ret, indent, level)
	} else if tgt_node.IsValue() {
		_to_string_value(ctx, tgt_node.Value, ret, indent, 0)
	} else if tgt_node.IsList() {
		multiline := false
		if !ctx.minify && !ctx.list_single {
			if multiline = tgt_node.IsFVVVList(); !multiline {
				var long_items uint64
				proc_item := func(item any) {
					if len(_value_to_string(item)) >= 16 {
						long_items++
					}
					if long_items >= 6 {
						multiline = true
					}
				}
				switch raw_list := tgt_node.Value.(type) {
				case []string:
					for _, item := range raw_list {
						if len(item)+2 >= 16 {
							long_items++
						}
						if long_items >= 6 {
							multiline = true
							break
						}
					}
				case []int64:
					for _, item := range raw_list {
						if proc_item(item); multiline {
							break
						}
					}
				case []int:
					for _, item := range raw_list {
						if proc_item(item); multiline {
							break
						}
					}
				case []float64:
					for _, item := range raw_list {
						if proc_item(item); multiline {
							break
						}
					}
				}
			}
		}

		value_indent := indent + ctx.indent_unit
		value_level := level + 1

		if ctx.full_width && _sb_last_empty(ret) {
			_sb_trim_last(ret)
		}
		ret.WriteRune(ctx.list_begin)
		if multiline {
			ret.WriteString(ctx.newline)
		}

		proc_item := func(item any, is_back bool) {
			if multiline {
				ret.WriteString(value_indent)
			}
			_to_string_value(ctx, item, ret, value_indent, value_level)
			if (multiline && ctx.force_sep) || (!multiline && is_back) {
				ret.WriteRune(ctx.item_sep)
				if !multiline && !ctx.full_width && !ctx.minify {
					ret.WriteByte(' ')
				}
			}
			if multiline {
				ret.WriteString(ctx.newline)
			}
		}
		switch raw_list := tgt_node.Value.(type) {
		case []bool:
			for idx, item := range raw_list {
				proc_item(item, idx != len(raw_list)-1)
			}
		case []int64:
			for idx, item := range raw_list {
				proc_item(item, idx != len(raw_list)-1)
			}
		case []int:
			for idx, item := range raw_list {
				proc_item(item, idx != len(raw_list)-1)
			}
		case []float64:
			for idx, item := range raw_list {
				proc_item(item, idx != len(raw_list)-1)
			}
		case []string:
			for idx, item := range raw_list {
				proc_item(item, idx != len(raw_list)-1)
			}
		case []*FVVV:
			for idx, item := range raw_list {
				proc_item(item, idx != len(raw_list)-1)
			}
		}

		if multiline {
			ret.WriteString(indent)
		}
		ret.WriteRune(ctx.list_end)
	}

	if !ctx.no_descs && tgt_node.Desc != "" &&
		((tgt_node.IsNodesEmpty() && !tgt_node.IsFVVVList()) ||
			tgt_node.Link != "" || !ctx.fww_style) {
		if !ctx.minify &&
			(!ctx.full_width || tgt_node.Link != "" ||
				(tgt_node.IsNodesEmpty() && !tgt_node.IsList() && !tgt_node.IsString()) ||
				(tgt_node.IsString() && _sb_last_is(ret, '`'))) {
			ret.WriteByte(' ')
		}
		ret.WriteString(_escape_string(tgt_node.Desc, true, ctx.full_width))
	}

	if ctx.minify || ctx.force_sep {
		ret.WriteRune(ctx.stmt_sep)
	}
	if !ctx.minify && !is_back {
		ret.WriteString(ctx.newline)
	}
}

func _to_string_value(ctx *FormatCtx, tgt_val any, ret *strings.Builder, indent string, level int) {
	switch tgt_val := tgt_val.(type) {
	case bool:
		ret.WriteString(strconv.FormatBool(tgt_val))

	case int64, int, float64:
		_, is_int64 := tgt_val.(int64)
		_, is_int := tgt_val.(int)
		_, is_float := tgt_val.(float64)
		var tgt_int int64
		var tgt_float float64
		if is_int64 {
			tgt_int = tgt_val.(int64)
		} else if is_int {
			tgt_int = int64(tgt_val.(int))
		} else if is_float {
			tgt_float = tgt_val.(float64)
		}

		if !is_float && ctx.int_base != 10 {
			if tgt_int == 0 {
				switch ctx.int_base {
				case 16:
					ret.WriteString("0x0")
				case 8:
					ret.WriteString("0o0")
				case 2:
					ret.WriteString("0b0")
				}
				break
			}

			var tgt_uint uint64
			if tgt_int < 0 {
				ret.WriteByte('-')
				tgt_uint = uint64(-tgt_int)
			} else {
				tgt_uint = uint64(tgt_int)
			}

			switch ctx.int_base {
			case 2:
				ret.WriteString("0b")
			case 8:
				ret.WriteString("0o")
			case 16:
				ret.WriteString("0x")
			}
			ret.WriteString(strconv.FormatUint(tgt_uint, ctx.int_base))
			break
		}

		var raw_num string
		if is_float {
			raw_num = strconv.FormatFloat(tgt_float, 'f', -1, 64)
		} else {
			raw_num = strconv.FormatInt(tgt_int, 10)
		}
		if ctx.digit_sep_step == 0 {
			ret.WriteString(raw_num)
			break
		}

		parts := strings.Split(raw_num, ".")
		int_part := parts[0]
		var has_sign bool
		if int_part[0] == '-' || int_part[0] == '+' {
			has_sign = true
			int_part = int_part[1:]
		}
		int_len := len(int_part)

		if int_len <= ctx.digit_sep_step {
			ret.WriteString(raw_num)
			break
		}

		ret.Grow(len(raw_num) + int_len/ctx.digit_sep_step + 1)
		if has_sign {
			ret.WriteByte(raw_num[0])
		}
		for idx, ch := range int_part {
			if idx > 0 && (int_len-idx)%ctx.digit_sep_step == 0 {
				ret.WriteRune(ctx.digit_sep_char)
			}
			ret.WriteRune(ch)
		}

		if len(parts) >= 2 {
			ret.WriteByte('.')
			ret.WriteString(parts[1])
		}

	case string:
		if !ctx.minify && ctx.raw_str && len(tgt_val) >= 3 &&
			!strings.ContainsRune(tgt_val, '`') && strings.ContainsAny(strings.TrimSpace(tgt_val), "\r\n") {
			str_indent := indent + ctx.indent_unit
			ret.Grow(len(tgt_val) + len(str_indent)*6)

			ret.WriteByte('`')
			ret.WriteString(ctx.newline)
			scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(_trim_indent(tgt_val))))
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					ret.WriteString(str_indent)
				}
				ret.WriteString(line)
				ret.WriteString(ctx.newline)
			}
			ret.WriteString(indent)
			ret.WriteByte('`')
			return
		}

		if level == 0 && ctx.full_width && _sb_last_empty(ret) {
			_sb_trim_last(ret)
		}
		ret.WriteString(_escape_string(tgt_val, false, ctx.full_width))

	case *FVVV:
		if ctx.fww_style && tgt_val.Desc != "" {
			ret.WriteString(_escape_string(tgt_val.Desc, true, ctx.full_width))
			if !ctx.minify && !ctx.full_width {
				ret.WriteByte(' ')
			}
		}
		_to_string_fwv(ctx, tgt_val, ret, indent, level)
		if !ctx.no_descs && !ctx.fww_style && tgt_val.Desc != "" {
			if !ctx.minify && !ctx.full_width {
				ret.WriteByte(' ')
			}
			ret.WriteString(_escape_string(tgt_val.Desc, true, ctx.full_width))
		}
	}
}

func _to_string_fwv(ctx *FormatCtx, tgt_fwv *FVVV, ret *strings.Builder, indent string, level int) {
	ret.WriteRune(ctx.fwv_begin)
	if !ctx.minify {
		ret.WriteString(ctx.newline)
	}
	tgt_fwv._to_string_root(ctx, ret, level+1)
	if !ctx.minify {
		ret.WriteString(ctx.newline)
		ret.WriteString(indent)
	}
	ret.WriteRune(ctx.fwv_end)
}

func _escape_string(str string, is_desc bool, full_width bool) string {
	var ret strings.Builder
	ret.Grow(len(str) + 6)

	if is_desc {
		ret.WriteByte('<')
	} else {
		if full_width {
			ret.WriteRune('“')
		} else {
			ret.WriteByte('"')
		}
	}

	for _, ch := range str {
		switch ch {
		case '\\':
			ret.WriteString("\\\\")
		case '\b':
			ret.WriteString("\\b")
		case '\f':
			ret.WriteString("\\f")
		case '\n':
			ret.WriteString("\\n")
		case '\r':
			ret.WriteString("\\r")
		case '\t':
			ret.WriteString("\\t")
		case '"':
			if !full_width && !is_desc {
				ret.WriteString("\\\"")
			} else {
				ret.WriteRune(ch)
			}
		case '”':
			if full_width && !is_desc {
				ret.WriteString("\\”")
			} else {
				ret.WriteRune(ch)
			}
		case '>':
			if is_desc {
				ret.WriteString("\\>")
			} else {
				ret.WriteRune(ch)
			}
		default:
			ret.WriteRune(ch)
		}
	}

	if is_desc {
		ret.WriteByte('>')
	} else {
		if full_width {
			ret.WriteRune('”')
		} else {
			ret.WriteByte('"')
		}
	}
	return ret.String()
}

func _to(node *FVVV, rv reflect.Value) error {
	if node == nil {
		return nil
	}
	if rv.Type() == _fwv_type {
		return _to_value(node, rv)
	}
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return _to(node, rv.Elem())
	case reflect.Struct:
		return _to_struct(node, rv)
	case reflect.Slice, reflect.Array:
		return _to_list(node, rv)
	default:
		return _to_value(node, rv)
	}
}

func _to_struct(node *FVVV, rv reflect.Value) error {
	tp := rv.Type()
	for idx := 0; idx < rv.NumField(); idx++ {
		idx_rv := rv.Field(idx)
		idx_tp := tp.Field(idx)
		if !idx_rv.CanSet() {
			continue
		}

		name := idx_tp.Tag.Get("fvv")
		if name == "-" {
			continue
		}
		if name_idx := strings.Index(name, ","); name_idx != -1 {
			name = name[:name_idx]
		}
		if name == "" {
			name = idx_tp.Name
		}

		if tgt_node, ok := node.Nodes[name]; ok {
			if err := _to(tgt_node, idx_rv); err != nil {
				return err
			}
		}
	}
	return nil
}

func _to_list(node *FVVV, rv reflect.Value) error {
	if node.Value == nil {
		return nil
	}
	src_list := reflect.ValueOf(node.Value)
	src_len := src_list.Len()

	if rv.Kind() == reflect.Slice {
		rv.Set(reflect.MakeSlice(rv.Type(), src_len, src_len))
	} else if src_len > rv.Len() { // reflect.Array
		src_len = rv.Len()
	}

	for idx := 0; idx < src_len; idx++ {
		src_val := src_list.Index(idx).Interface()
		dst_val := rv.Index(idx)

		if fvvv, ok := src_val.(*FVVV); ok {
			if err := _to(fvvv, dst_val); err != nil {
				return err
			}
		} else {
			if err := _to(&FVVV{Value: src_val}, dst_val); err != nil {
				return err
			}
		}
	}
	return nil
}

func _to_value(node *FVVV, rv reflect.Value) error {
	if rv.Type() == _fwv_type {
		rv.Set(reflect.ValueOf(*node))
		return nil
	}

	if node.Value == nil {
		return nil
	}
	val := reflect.ValueOf(node.Value)

	if val.Type().ConvertibleTo(rv.Type()) {
		rv.Set(val.Convert(rv.Type()))
		return nil
	}

	switch rv.Kind() {
	case reflect.Bool:
		if val.Kind() == reflect.Bool {
			rv.SetBool(val.Bool())
		} else {
			tmp_str := _value_to_string(node.Value)
			if strings.EqualFold(tmp_str, "true") {
				rv.SetBool(true)
			} else if strings.EqualFold(tmp_str, "false") {
				rv.SetBool(false)
			}
			return errors.New("type mismatch")
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			rv.SetInt(val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			rv.SetInt(int64(val.Uint()))
		case reflect.Float32, reflect.Float64:
			rv.SetInt(int64(val.Float()))
		default:
			return errors.New("type mismatch")
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			rv.SetUint(uint64(val.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			rv.SetUint(val.Uint())
		case reflect.Float32, reflect.Float64:
			rv.SetUint(uint64(val.Float()))
		default:
			return errors.New("type mismatch")
		}
		return nil
	case reflect.Float32, reflect.Float64:
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			rv.SetFloat(float64(val.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			rv.SetFloat(float64(val.Uint()))
		case reflect.Float32, reflect.Float64:
			rv.SetFloat(val.Float())
		default:
			return errors.New("type mismatch")
		}
		return nil
	case reflect.String:
		rv.SetString(_value_to_string(node.Value))
		return nil
	}
	return errors.New("type mismatch")
}

func _from(node *FVVV, rv reflect.Value) error {
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	if rv.Type() == _fwv_type {
		src := rv.Interface().(FVVV)
		node.Value = src.Value
		node.Nodes = src.Nodes
		node.Desc = src.Desc
		node.Link = src.Link
		return nil
	}

	switch rv.Kind() {
	case reflect.Struct:
		return _from_struct(node, rv)
	case reflect.Map:
		return _from_map(node, rv)
	case reflect.Slice, reflect.Array:
		return _from_list(node, rv)
	default:
		return _from_value(node, rv)
	}
}

func _from_struct(node *FVVV, rv reflect.Value) error {
	if node.Nodes == nil {
		node.Nodes = make(map[string]*FVVV)
	}

	tp := rv.Type()
	for idx := 0; idx < rv.NumField(); idx++ {
		idx_rv := rv.Field(idx)
		idx_tp := tp.Field(idx)
		if !idx_rv.CanInterface() {
			continue
		}

		name := idx_tp.Tag.Get("fvv")
		if name == "-" {
			continue
		}
		var opts string
		if before, after, ok := strings.Cut(name, ","); ok {
			name = before
			opts = after
		}
		if name == "" {
			name = idx_tp.Name
		}
		if strings.Contains(opts, "omitempty") {
			switch idx_rv.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				if idx_rv.Len() == 0 {
					continue
				}
			case reflect.Bool:
				if !idx_rv.Bool() {
					continue
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if idx_rv.Int() == 0 {
					continue
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if idx_rv.Uint() == 0 {
					continue
				}
			case reflect.Float32, reflect.Float64:
				if idx_rv.Float() == 0 {
					continue
				}
			case reflect.Interface, reflect.Pointer:
				if idx_rv.IsNil() {
					continue
				}
			}
		}

		tgt_node := &FVVV{}
		if err := _from(tgt_node, idx_rv); err != nil {
			return err
		}
		node.Nodes[name] = tgt_node
	}
	return nil
}

func _from_list(node *FVVV, rv reflect.Value) error {
	list_tp := rv.Type().Elem().Kind()
	if list_tp == reflect.Uint8 {
		node.Value = string(rv.Bytes())
		return nil
	}

	src_len := rv.Len()
	switch list_tp {
	case reflect.Bool:
		list := make([]bool, src_len)
		for idx := 0; idx < src_len; idx++ {
			list[idx] = rv.Index(idx).Bool()
		}
		node.Value = list
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		list := make([]int64, src_len)
		for idx := 0; idx < src_len; idx++ {
			list[idx] = rv.Index(idx).Int()
		}
		node.Value = list
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		list := make([]int64, src_len)
		for idx := 0; idx < src_len; idx++ {
			list[idx] = int64(rv.Index(idx).Uint())
		}
		node.Value = list
	case reflect.Float32, reflect.Float64:
		list := make([]float64, src_len)
		for idx := 0; idx < src_len; idx++ {
			list[idx] = rv.Index(idx).Float()
		}
		node.Value = list
	case reflect.String:
		list := make([]string, src_len)
		for idx := 0; idx < src_len; idx++ {
			list[idx] = rv.Index(idx).String()
		}
		node.Value = list
	default:
		list := make([]*FVVV, src_len)
		for idx := 0; idx < src_len; idx++ {
			tgt_node := &FVVV{}
			if err := _from(tgt_node, rv.Index(idx)); err != nil {
				return err
			}
			list[idx] = tgt_node
		}
		node.Value = list
	}
	return nil
}

func _from_map(node *FVVV, rv reflect.Value) error {
	if rv.Type().Key().Kind() != reflect.String {
		return errors.New("map keys must be strings")
	}
	if node.Nodes == nil {
		node.Nodes = make(map[string]*FVVV)
	}

	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		val := iter.Value()

		tgt_node := &FVVV{}
		if err := _from(tgt_node, val); err != nil {
			return err
		}
		node.Nodes[key] = tgt_node
	}
	return nil
}

func _from_value(node *FVVV, rv reflect.Value) error {
	if rv.IsValid() {
		node.Value = rv.Interface()
	}
	return nil
}

func _trim_indent(str string) string {
	if str == "" {
		return ""
	}
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return ""
	}

	min_indent := -1
	for idx, line := range lines {
		if strings.TrimSpace(line) == "" {
			lines[idx] = ""
			continue
		}
		idx_indent := 0
		for _, char := range line {
			if unicode.IsSpace(char) {
				idx_indent++
			} else {
				break
			}
		}
		if min_indent == -1 || idx_indent < min_indent {
			min_indent = idx_indent
		}
	}

	if min_indent <= 0 {
		return strings.Join(lines, "\n")
	}
	for idx, line := range lines {
		if len(line) >= min_indent {
			lines[idx] = line[min_indent:]
		}
	}
	return strings.Join(lines, "\n")
}

func _value_to_string(val any) string {
	switch tgt_val := val.(type) {
	case string:
		return tgt_val
	case bool:
		return strconv.FormatBool(tgt_val)
	case int64:
		return strconv.FormatInt(tgt_val, 10)
	case float64:
		return strconv.FormatFloat(tgt_val, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func _is_same_type(a, b any) bool {
	switch a.(type) {
	case bool, *bool:
		_, ok := b.(bool)
		return ok
	case int64, *int64:
		_, ok := b.(int64)
		return ok
	case float64, *float64:
		_, ok := b.(float64)
		return ok
	case string, *string:
		_, ok := b.(string)
		return ok
	case FVVV, *FVVV:
		_, ok := b.(FVVV)
		return ok
	default:
		return false
	}
}

func _get_type_ptr(val any) any {
	switch val.(type) {
	case bool, *bool:
		return (*bool)(nil)
	case int64, *int64:
		return (*int64)(nil)
	case float64, *float64:
		return (*float64)(nil)
	case string, *string:
		return (*string)(nil)
	case FVVV, *FVVV:
		return (*FVVV)(nil)
	default:
		return nil
	}
}

func _sb_last_empty(sb *strings.Builder) bool {
	str := sb.String()
	return len(str) > 0 && str[len(str)-1] == ' '
}

func _sb_last_is(sb *strings.Builder, b byte) bool {
	str := sb.String()
	return len(str) > 0 && str[len(str)-1] == b
}

func _sb_trim_last(sb *strings.Builder) {
	str := sb.String()
	if len(str) > 0 && str[len(str)-1] == ' ' {
		trimmed := str[:len(str)-1]
		sb.Reset()
		sb.WriteString(trimmed)
	}
}
