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
	"bytes"
	"strconv"
	"strings"
)

func eqOr[Tp comparable](a Tp, args ...Tp) bool {
	for _, v := range args {
		if a == v {
			return true
		}
	}
	return false
}

func removeLastChar(builder *strings.Builder) {
	runes := []rune(builder.String())
	if len(runes) > 0 {
		runes = runes[:len(runes)-1]
		builder.Reset()
		builder.WriteString(string(runes))
	}
}

type FVVV struct {
	Value      any
	Sub        map[string]*FVVV
	Desc, Link string
}

func (_fvvv *FVVV) IsEmpty() bool {
	return _fvvv.Value == nil
}

func (_fvvv *FVVV) IsNotEmpty() bool {
	return !_fvvv.IsEmpty()
}

func (_fvvv *FVVV) SubIsEmpty() bool {
	return len(_fvvv.Sub) == 0
}

func (_fvvv *FVVV) SubIsNotEmpty() bool {
	return !_fvvv.SubIsEmpty()
}

func writeList[T any](
	result *strings.Builder,
	list []T,
	is_biglist bool,
	is_min bool,
	list_indent string,
	printer func(*strings.Builder, T),
) bool {
	for _, value := range list {
		if is_biglist {
			result.WriteString(list_indent)
		}
		printer(result, value)
		if is_biglist {
			result.WriteRune('\n')
		} else {
			result.WriteRune(',')
			if !is_min {
				result.WriteRune(' ')
			}
		}
	}
	return len(list) == 0
}

func (_fvvv *FVVV) Print(tp ...string) string {
	var is_min, is_biglist, is_nodesc bool
	indent_lv := 0
	if len(tp) > 0 {
		is_min, is_biglist, is_nodesc = tp[0] == "min", tp[0] == "biglist", tp[0] == "nodesc"
	}
	if len(tp) > 1 {
		indent_lv, _ = strconv.Atoi(tp[1])
	}
	var result strings.Builder
	var print_func func(path string, node *FVVV, indent_lv int)
	print_func = func(path string, node *FVVV, indent_lv int) {
		if path == "" || (node.IsEmpty() && node.SubIsEmpty()) {
			return
		}
		indent := strings.Repeat("  ", indent_lv)
		if node.SubIsNotEmpty() && node.Link == "" {
			if is_min {
				result.WriteString(path + "={")
			} else {
				result.WriteString(indent + path + " = {\n")
			}
		}
		if node.Link != "" || node.IsNotEmpty() {
			if is_min {
				result.WriteString(path + "=")
			} else {
				result.WriteString(indent + path + " = ")
			}
			if node.Link != "" {
				result.WriteString(node.Link)
			} else {
				switch v := node.Value.(type) {
				case string:
					result.WriteString(`"` + strings.ReplaceAll(v, `"`, `\"`) + `"`)
				case bool:
					result.WriteString(strconv.FormatBool(v))
				case int:
					result.WriteString(strconv.Itoa(v))
				case float64:
					result.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
				case []string, []bool, []int, []float64, []FVVV:
					list_indent := strings.Repeat("  ", indent_lv+1)
					result.WriteRune('[')
					if is_biglist {
						result.WriteRune('\n')
					} else if _, ok := v.([]FVVV); ok {
						result.WriteRune('\n')
					}
					is_empty_list := true
					switch v := node.Value.(type) {
					case []string:
						is_empty_list = writeList(&result, v, is_biglist, is_min, list_indent, func(sb *strings.Builder, v string) {
							sb.WriteString(`"` + strings.ReplaceAll(v, `"`, `\"`) + `"`)
						})
					case []bool:
						is_empty_list = writeList(&result, v, is_biglist, is_min, list_indent, func(sb *strings.Builder, v bool) {
							sb.WriteString(strconv.FormatBool(v))
						})
					case []int:
						is_empty_list = writeList(&result, v, is_biglist, is_min, list_indent, func(sb *strings.Builder, v int) {
							sb.WriteString(strconv.Itoa(v))
						})
					case []float64:
						is_empty_list = writeList(&result, v, is_biglist, is_min, list_indent, func(sb *strings.Builder, v float64) {
							sb.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
						})
					case []FVVV:
						is_empty_list = len(v) == 0
						for _, value := range v {
							if !is_min {
								result.WriteString(list_indent)
							}
							result.WriteRune('{')
							if !is_min {
								result.WriteRune('\n')
							}
							sub_type := ""
							if len(tp) > 0 {
								sub_type = tp[0]
							}
							result.WriteString(value.Print(sub_type, strconv.Itoa(indent_lv+2)))
							if is_min {
								result.WriteString(";}")
							} else {
								result.WriteString("\n" + list_indent + "}")
							}
							if value.Desc != "" {
								if !is_min {
									result.WriteRune(' ')
								}
								result.WriteString("<" + strings.ReplaceAll(value.Desc, `>`, `\>`) + ">")
							}
							if is_min {
								result.WriteRune(',')
							} else {
								result.WriteRune('\n')
							}
						}
					}
					if is_biglist {
						result.WriteString(indent)
					} else if _, ok := v.([]FVVV); ok {
						result.WriteString(indent)
					} else if !is_empty_list {
						removeLastChar(&result)
						if !is_min {
							removeLastChar(&result)
						}
					}
					result.WriteRune(']')
				}
			}
		} else {
			for key, value := range node.Sub {
				print_func(key, value, indent_lv+1)
			}
		}
		if node.SubIsNotEmpty() && node.Link == "" {
			if !is_min {
				result.WriteString(indent)
			}
			result.WriteRune('}')
		}
		if node.Desc != "" && !is_min && !is_nodesc {
			result.WriteString(" <" + strings.ReplaceAll(node.Desc, `>`, `\>`) + ">")
		}
		if is_min {
			result.WriteRune(';')
		} else {
			result.WriteRune('\n')
		}
	}
	for key, value := range _fvvv.Sub {
		print_func(key, value, indent_lv)
	}
	removeLastChar(&result)
	return result.String()
}

func (_fvvv *FVVV) AddFromString(txt string) {
	txtBytes := []byte(txt)
	if len(txtBytes) >= 3 && txtBytes[0] == 0xEF && txtBytes[1] == 0xBB && txtBytes[2] == 0xBF {
		txtBytes = txtBytes[3:]
	}
	txtBytes = bytes.ReplaceAll(txtBytes, []byte("\r\n"), []byte("\n"))
	txtBytes = bytes.ReplaceAll(txtBytes, []byte("\r"), []byte("\n"))
	txt = strings.TrimSpace(string(txtBytes)) + "\n"
	if txt == "" {
		return
	}

	type FVVVDat struct {
		value_name                 strings.Builder
		idx_desc                   string
		value_names, group_names   []string
		last_group_names           [][]string
		tmp_fvvs                   []FVVV
		in_value, in_list, is_list bool
		group_num                  uint64
		root_key                   FVVV
		idx_key                    *FVVV
	}
	get_key := func(paths []string, root_key *FVVV) *FVVV {
		tmp_key := root_key
		for _, path := range paths {
			if tmp_key.Sub == nil {
				tmp_key.Sub = make(map[string]*FVVV)
			}
			if tmp_key.Sub[path] == nil {
				tmp_key.Sub[path] = &FVVV{
					Value: nil,
					Sub:   make(map[string]*FVVV),
				}
			}
			tmp_key = tmp_key.Sub[path]
		}
		return tmp_key
	}
	find_key := func(path string, stack_dat []FVVVDat) *FVVV {
		tmp_names := strings.Split(strings.TrimSpace(path), ".")
		var tmp_key *FVVV
		for idx := len(stack_dat) - 1; idx >= 0; idx-- {
			idx_dat := &(stack_dat)[idx]
			tmp_key = idx_dat.idx_key
			for _, key := range tmp_names {
				if tmp_key.Sub[key] != nil {
					tmp_key = tmp_key.Sub[key]
				} else {
					if idx == 0 {
						tmp_key = _fvvv
					} else {
						tmp_key = &idx_dat.root_key
					}
					for _, key := range tmp_names {
						if tmp_key.Sub[key] != nil {
							tmp_key = tmp_key.Sub[key]
						} else {
							tmp_key = nil
							break
						}
					}
					break
				}
			}
			if tmp_key != nil {
				return tmp_key
			}
		}
		return nil
	}
	var end_group, old_fvv, is_real_char, in_desc, in_str, is_str, is_all_str, is_empty_str bool
	var tmp_desc, value strings.Builder
	var values []string
	var last_char rune
	fvv_stack := []FVVVDat{{}}
	fvv_stack[0].idx_key = _fvvv
	for idx, idx_char := range txt {
		is_real_char = last_char != '\\'
		idx_dat := &fvv_stack[len(fvv_stack)-1]
		if len(fvv_stack) > 1 {
			idx_dat.idx_key = &idx_dat.root_key
		} else {
			idx_dat.idx_key = _fvvv
		}
		if func() bool {
			if in_desc {
				if idx_char != '>' || !is_real_char {
					if idx_char == '>' {
						removeLastChar(&tmp_desc)
					}
					if idx_dat.in_value || idx_dat.group_num > 0 {
						tmp_desc.WriteRune(idx_char)
					}
					return false
				} else {
					idx_dat.idx_desc = tmp_desc.String()
					tmp_desc.Reset()
					if key := find_key(idx_dat.idx_desc, fvv_stack); key != nil {
						if desc, ok := key.Value.(string); ok {
							idx_dat.idx_desc = desc
						}
					}
					in_desc = false
					return false
				}
			} else {
				if !in_str && eqOr(idx_char, ' ', '\t') {
					return false
				}
				if idx_char == '<' {
					in_desc = true
					return false
				}
			}
			if idx_dat.in_value {
				if in_str {
					if idx_char == '"' {
						if is_real_char {
							is_empty_str = value.String() == ""
							in_str = false
							return false
						} else {
							removeLastChar(&value)
							value.WriteRune(idx_char)
							return false
						}
					} else {
						value.WriteRune(idx_char)
						return false
					}
				} else {
					if idx_char == '"' {
						in_str, is_str, is_all_str = true, true, true
						return false
					} else if idx_char == '[' {
						idx_dat.in_list, idx_dat.is_list = true, true
						return false
					} else if idx_dat.in_list && idx_char == '{' {
						fvv_stack = append(fvv_stack, FVVVDat{})
						return false
					} else if idx_dat.in_list && eqOr(idx_char, ',', ']', '\n') {
						value_str := value.String()
						if idx_char == ']' {
							idx_dat.in_list = false
							pos := 1
							in_list_desc := false
							for {
								if func() bool {
									switch txt[idx-pos] {
									case '<':
										if !in_list_desc {
											return true
										}
										if idx-pos < 1 || txt[idx-pos-1] != '\\' {
											in_list_desc = false
										}
										return false
									case '>':
										in_list_desc = true
										return false
									case ' ', '\t':
										return false
									case ',', '\n':
										return !in_list_desc
									default:
										return !in_list_desc
									}
								}() {
									break
								}
								pos++
							}
							if txt[idx-pos] == ',' || txt[idx-pos] == '\n' {
								return false
							}
						} else if len(idx_dat.tmp_fvvs) == 0 && value_str == "" && (!is_all_str || !is_empty_str) {
							return false
						}
						if (is_all_str && is_str) || eqOr(value_str, "true", "false") {
							values = append(values, value_str)
						} else if _, err := strconv.Atoi(value_str); err == nil {
							values = append(values, value_str)
						} else if _, err := strconv.ParseFloat(value_str, 64); err == nil {
							values = append(values, value_str)
						} else {
							idx_dat.idx_key = get_key([]string{value_str}, get_key(idx_dat.group_names, idx_dat.idx_key))
							if idx_dat.idx_key.IsNotEmpty() {
								switch v := idx_dat.idx_key.Value.(type) {
								case string:
									values = append(values, v)
								case bool:
									values = append(values, strconv.FormatBool(v))
								case int:
									values = append(values, strconv.Itoa(v))
								case float64:
									values = append(values, strconv.FormatFloat(v, 'f', -1, 64))
								case []string:
									values = append(values, v...)
								case []bool:
									for _, tmp := range v {
										values = append(values, strconv.FormatBool(tmp))
									}
								case []int:
									for _, tmp := range v {
										values = append(values, strconv.Itoa(tmp))
									}
								case []float64:
									for _, tmp := range v {
										values = append(values, strconv.FormatFloat(tmp, 'f', -1, 64))
									}
								case []FVVV:
									idx_dat.tmp_fvvs = append(idx_dat.tmp_fvvs, v...)
								}
							}
						}
						if len(idx_dat.tmp_fvvs) > 0 && idx_dat.idx_desc != "" {
							idx_dat.tmp_fvvs[len(idx_dat.tmp_fvvs)-1].Desc = idx_dat.idx_desc
							idx_dat.idx_desc = ""
						}
						if is_empty_str {
							is_empty_str = false
						} else {
							value.Reset()
						}
						is_str = false
						return false
					} else if idx_char == '{' {
						idx_dat.group_names = append(idx_dat.group_names, idx_dat.value_names...)
						idx_dat.last_group_names = append(idx_dat.last_group_names, idx_dat.value_names)
						idx_dat.value_names = nil
						idx_dat.group_num++
						idx_dat.in_value = false
						return false
					} else if !idx_dat.in_list && eqOr(idx_char, ';', '\n') {
						idx_dat.idx_key = get_key(idx_dat.value_names, get_key(idx_dat.group_names, idx_dat.idx_key))
						if idx_dat.is_list {
							if len(values) == 0 && len(idx_dat.tmp_fvvs) == 0 {
								idx_dat.idx_key.Value = nil
							} else if len(idx_dat.tmp_fvvs) > 0 {
								idx_dat.idx_key.Value = idx_dat.tmp_fvvs
							} else if is_all_str {
								idx_dat.idx_key.Value = values
							} else {
								tmp_str := values[0]
								if eqOr(tmp_str, "true", "false") {
									tmps := make([]bool, len(values))
									for i, s := range values {
										tmps[i] = s == "true"
									}
									idx_dat.idx_key.Value = tmps
								} else if _, err := strconv.Atoi(tmp_str); err == nil {
									tmps := make([]int, len(values))
									for i, s := range values {
										if tmp, err := strconv.Atoi(s); err == nil {
											tmps[i] = tmp
										}
									}
									idx_dat.idx_key.Value = tmps
								} else if _, err := strconv.ParseFloat(tmp_str, 64); err == nil {
									tmps := make([]float64, len(values))
									for i, s := range values {
										if tmp, err := strconv.ParseFloat(s, 64); err == nil {
											tmps[i] = tmp
										}
									}
									idx_dat.idx_key.Value = tmps
								}
							}
						} else {
							value_str := value.String()
							if is_all_str {
								idx_dat.idx_key.Value = value_str
							} else if eqOr(value_str, "true", "false") {
								idx_dat.idx_key.Value = value_str == "true"
							} else if tmp, err := strconv.Atoi(value_str); err == nil {
								idx_dat.idx_key.Value = tmp
							} else if tmp, err := strconv.ParseFloat(value_str, 64); err == nil {
								idx_dat.idx_key.Value = tmp
							} else if tmp_key := find_key(value_str, fvv_stack); tmp_key != nil {
								if len(tmp_key.Sub) == 0 {
									idx_dat.idx_key.Value = tmp_key.Value
								} else {
									idx_dat.idx_key.Sub = tmp_key.Sub
								}
								idx_dat.idx_key.Link = value_str
							}
						}
						idx_dat.idx_key.Desc = idx_dat.idx_desc
						value.Reset()
						idx_dat.idx_desc = ""
						values, idx_dat.value_names, idx_dat.tmp_fvvs = nil, nil, nil
						idx_dat.in_value, is_str, is_all_str, idx_dat.is_list = false, false, false, false
						return false
					} else {
						value.WriteRune(idx_char)
						return false
					}
				}
			} else if !old_fvv && idx_char == '{' && idx_dat.value_name.String() == "" {
				old_fvv = true
				return false
			} else if idx_char == '=' {
				idx_dat.value_names = strings.Split(strings.TrimSpace(idx_dat.value_name.String()), ".")
				idx_dat.value_name.Reset()
				idx_dat.in_value = true
				return false
			} else if end_group && eqOr(idx_char, ';', '\n') && idx_dat.group_num > 0 {
				end_group = false
				if idx_dat.idx_desc != "" {
					get_key(idx_dat.group_names, idx_dat.idx_key).Desc = idx_dat.idx_desc
					idx_dat.idx_desc = ""
				}
				for range idx_dat.last_group_names[len(idx_dat.last_group_names)-1] {
					idx_dat.group_names = idx_dat.group_names[:len(idx_dat.group_names)-1]
				}
				idx_dat.last_group_names = idx_dat.last_group_names[:len(idx_dat.last_group_names)-1]
				idx_dat.group_num--
				return false
			} else if idx_char == '}' {
				if idx_dat.group_num == 0 {
					if len(fvv_stack) > 1 {
						target_fvvv := fvv_stack[len(fvv_stack)-1].root_key
						fvv_stack = fvv_stack[:len(fvv_stack)-1]
						idx_dat = &fvv_stack[len(fvv_stack)-1]
						idx_dat.tmp_fvvs = append(idx_dat.tmp_fvvs, target_fvvv)
						return false
					}
					return true
				} else {
					end_group = true
					return false
				}
			} else {
				idx_dat.value_name.WriteRune(idx_char)
				return false
			}
		}() {
			break
		}
		last_char = idx_char
	}
}

func NewFVVV() *FVVV {
	return &FVVV{
		Value: nil,
		Sub:   make(map[string]*FVVV),
	}
}
