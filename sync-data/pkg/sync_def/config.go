package sync_def

import (
	"fmt"
	"hutool/convert"
	"hutool/reflectx"
	"sort"
	"strings"
)

type DbFieldConfig struct {
	GoFieldName     string
	DbFieldName     string
	AsPrimaryKey    bool
	PrimaryKeyIndex int
	AsCacheKey      bool
	CacheKeyIndex   int
	IsSeqField      bool
	IsScoreField    bool
}

type DataConfig struct {
	Fields         map[string]*DbFieldConfig
	PrimaryFields  []*DbFieldConfig
	UpdateFields   []*DbFieldConfig
	CacheKeyFields []*DbFieldConfig
	AllFields      []*DbFieldConfig
	ScoreField     *DbFieldConfig
	SeqField       *DbFieldConfig
}

func BuildFieldConfig[T any]() *DataConfig {
	fields := make(map[string]*DbFieldConfig)
	rt := reflectx.GenericTypeOf[T]()
	ret := &DataConfig{}

	for fieldIndex := 0; fieldIndex < rt.NumField(); fieldIndex++ {
		field := rt.Field(fieldIndex)
		if !field.IsExported() {
			continue
		}

		df := &DbFieldConfig{
			GoFieldName: field.Name,
			DbFieldName: field.Name,
		}

		tagData, ok := field.Tag.Lookup("sync")
		if ok {
			details := strings.Split(tagData, ",")
			for _, detail := range details {
				kv := strings.Split(detail, "=")
				if len(kv) != 2 {
					continue
				}

				key := kv[0]
				value := kv[1]
				switch key {
				case "primary":
					df.PrimaryKeyIndex = convert.Int(value)
					df.AsPrimaryKey = true
				case "cache":
					df.CacheKeyIndex = convert.Int(value)
					df.AsCacheKey = true
				case "isSeq":
					df.IsSeqField = convert.Bool(value)
				case "isScore":
					df.IsScoreField = convert.Bool(value)
				}
			}
		}

		fields[field.Name] = df
	}

	for _, field := range fields {
		// 主键
		if field.AsPrimaryKey {
			ret.PrimaryFields = append(ret.PrimaryFields, field)
		} else {
			ret.UpdateFields = append(ret.UpdateFields, field)
		}

		// 缓存键
		if field.AsCacheKey {
			ret.CacheKeyFields = append(ret.CacheKeyFields, field)
		}

		if field.IsSeqField {
			ret.SeqField = field
		}
		if field.IsScoreField {
			ret.ScoreField = field
		}

		ret.AllFields = append(ret.AllFields, field)
	}
	sort.Slice(ret.PrimaryFields, func(i, j int) bool {
		return ret.PrimaryFields[i].PrimaryKeyIndex < ret.PrimaryFields[j].PrimaryKeyIndex
	})
	sort.Slice(ret.CacheKeyFields, func(i, j int) bool {
		return ret.CacheKeyFields[i].CacheKeyIndex < ret.CacheKeyFields[j].CacheKeyIndex
	})
	if len(ret.CacheKeyFields) == 0 {
		ret.CacheKeyFields = make([]*DbFieldConfig, len(ret.PrimaryFields))
		copy(ret.CacheKeyFields, ret.PrimaryFields)
	}

	if len(ret.PrimaryFields) == 0 {
		panic("primary key is empty")
	}

	ret.Fields = fields
	return ret
}

func (d *DbFieldConfig) String() string {
	if d == nil {
		return "<nil>"
	}
	return fmt.Sprintf("DbFieldConfig{DbField: %s, GoField: %s, PrimaryKeyIndex: %d, CacheKeyIndex: %d, SeqFieldIndex: %v, IsScore: %v}",
		d.DbFieldName, d.GoFieldName, d.PrimaryKeyIndex, d.CacheKeyIndex, d.IsSeqField, d.IsScoreField)
}

func (d *DbFieldConfig) GoString() string {
	if d == nil {
		return "<nil>"
	}
	return fmt.Sprintf("&DbFieldConfig{DbField: %q, GoField: %q, PrimaryKeyIndex: %d, CacheKeyIndex: %d, SeqFieldIndex: %v, IsScore: %v}",
		d.DbFieldName, d.GoFieldName, d.PrimaryKeyIndex, d.CacheKeyIndex, d.IsSeqField, d.IsScoreField)
}

func (d *DbFieldConfig) Format(f fmt.State, verb rune) {
	if d == nil {
		fmt.Fprintf(f, "<nil>")
		return
	}

	switch verb {
	case 's', 'v':
		fmt.Fprintf(f, "DbFieldConfig{DbField: %s, GoField: %s, PrimaryKeyIndex: %d, CacheKeyIndex: %d, SeqFieldIndex: %v, IsScore: %v}",
			d.DbFieldName, d.GoFieldName, d.PrimaryKeyIndex, d.CacheKeyIndex, d.IsScoreField, d.IsScoreField)
	case 'q':
		fmt.Fprintf(f, "%q", d.String())
	case '#':
		fmt.Fprintf(f, "&DbFieldConfig{DbField: %q, GoField: %q, PrimaryKeyIndex: %d, CacheKeyIndex: %d, SeqFieldIndex: %v, IsScore: %v}",
			d.DbFieldName, d.GoFieldName, d.PrimaryKeyIndex, d.CacheKeyIndex, d.IsSeqField, d.IsScoreField)
	default:
		fmt.Fprintf(f, "%%!%c(DbFieldConfig=%s)", verb, d.String())
	}
}
