package db_impl

import (
	"database/sql"
	"errors"
	"fmt"
	"hutool/reflectx"
	"strings"
	"sync-data/pkg/sync_def"

	"github.com/jmoiron/sqlx"
)

type MysqlHandler[T any] struct {
	*sync_def.DbConfig[T]
	Db *sqlx.DB

	loadOneSql  string
	loadOneArgs []*sync_def.DbFieldConfig

	saveOneSql  string
	saveOneArgs []*sync_def.DbFieldConfig

	delOneSql  string
	delOneArgs []*sync_def.DbFieldConfig

	loadBatchSql1  string // select * from %s where
	loadBatchSql2  string // (key1=? and key2=? and key3=?...)
	loadBatchArgs2 []*sync_def.DbFieldConfig
}

func NewMysqlHandler[T any](db *sqlx.DB, option sync_def.DbOption[T]) *MysqlHandler[T] {
	if option.TableName == "" {
		option.TableName = reflectx.GenericTypeOf[T]().Name()
	}

	h := &MysqlHandler[T]{
		DbConfig: &sync_def.DbConfig[T]{
			BaseConfig: &sync_def.BaseConfig[T]{
				Config: sync_def.BuildFieldConfig[T](),
				Coder:  option.Coder,
			},
			TableName: option.TableName,
		},
		Db: db,
	}
	h.Init()
	return h
}

func (m *MysqlHandler[T]) Init() {
	m.initLoadOne()
	m.initSaveOne()
	m.initLoadBatch()
	m.initDelOne()
}

func (m *MysqlHandler[T]) initLoadOne() {
	// select * from ? where key1=? and key2=?
	args := make([]*sync_def.DbFieldConfig, 0, len(m.Config.PrimaryFields))
	for _, fieldCfg := range m.Config.PrimaryFields {
		args = append(args, fieldCfg)
	}

	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("select * from `%s` where ", m.TableName))
	for index, arg := range args {
		if index != 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(fmt.Sprintf("`%s`=?", arg.DbFieldName))
	}

	m.loadOneSql = sb.String()
	m.loadOneArgs = args
}

func (m *MysqlHandler[T]) initSaveOne() {
	args := make([]*sync_def.DbFieldConfig, 0, len(m.Config.AllFields))
	for _, fieldConfig := range m.Config.AllFields {
		args = append(args, fieldConfig)
	}
	for _, fieldConfig := range m.Config.UpdateFields {
		args = append(args, fieldConfig)
	}

	sb := &strings.Builder{}

	sb.WriteString(fmt.Sprintf("insert into `%s` (", m.TableName))
	for index, filedConfig := range m.Config.AllFields {
		if index != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("`%s`", filedConfig.DbFieldName))
	}
	sb.WriteString(") ")
	sb.WriteString("values (")
	for index := range m.Config.AllFields {
		if index != 0 {
			sb.WriteString(",")
		}
		sb.WriteString("?")
	}
	sb.WriteString(") ")
	sb.WriteString("on duplicate key update ")

	for index, fieldConfig := range m.Config.UpdateFields {
		if index != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=?", fieldConfig.DbFieldName))
	}

	m.saveOneSql = sb.String()
	m.saveOneArgs = args
}

func (m *MysqlHandler[T]) initDelOne() {
	args := make([]*sync_def.DbFieldConfig, 0, len(m.Config.PrimaryFields))
	for _, fieldConfig := range m.Config.PrimaryFields {
		args = append(args, fieldConfig)
	}
	sb := &strings.Builder{}

	sb.WriteString(fmt.Sprintf("delete from `%s` where ", m.TableName))
	for index, fieldConfig := range m.Config.PrimaryFields {
		if index != 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(fmt.Sprintf("`%s`=?", fieldConfig.DbFieldName))
	}

	m.delOneSql = sb.String()
	m.delOneArgs = args
}

func (m *MysqlHandler[T]) initLoadBatch() {
	// select * from table where (key1=? and key2=? and key3=?) or (key1=? and key2=? and key3 = ? and key4 = ?)
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("select * from `%s` where ", m.TableName))
	m.loadBatchSql1 = sb.String()

	sb.Reset()
	args := make([]*sync_def.DbFieldConfig, 0, len(m.Config.PrimaryFields))
	for _, fieldConfig := range m.Config.PrimaryFields {
		args = append(args, fieldConfig)
	}

	sb.WriteString("(")
	for index, fieldConfig := range m.Config.PrimaryFields {
		if index != 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(fmt.Sprintf("`%s`=?", fieldConfig.DbFieldName))
	}
	sb.WriteString(")")
	m.loadBatchArgs2 = args
	m.loadBatchSql2 = sb.String()
}

func (m *MysqlHandler[T]) LoadOne(key *T) (*T, error) {
	sqlArgs := make([]interface{}, 0, 4)
	rv := reflectx.IndirectValue(key)

	for _, arg := range m.loadOneArgs {
		fieldV := rv.FieldByName(arg.GoFieldName).Interface()
		sqlArgs = append(sqlArgs, fieldV)
	}

	ret := new(T)
	err := m.Db.Unsafe().Get(ret, m.loadOneSql, sqlArgs...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (m *MysqlHandler[T]) SaveOne(data *T) error {
	sqlArgs := make([]interface{}, 0, 4)
	rv := reflectx.IndirectValue(data)

	for _, arg := range m.saveOneArgs {
		fieldV := rv.FieldByName(arg.GoFieldName).Interface()
		sqlArgs = append(sqlArgs, fieldV)
	}

	_, err := m.Db.Unsafe().Exec(m.saveOneSql, sqlArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (m *MysqlHandler[T]) DelOne(key *T) error {
	sqlArgs := make([]interface{}, 0, 4)
	rv := reflectx.IndirectValue(key)
	for _, arg := range m.delOneArgs {
		fieldV := rv.FieldByName(arg.GoFieldName).Interface()
		sqlArgs = append(sqlArgs, fieldV)
	}

	_, err := m.Db.Unsafe().Exec(m.delOneSql, sqlArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (m *MysqlHandler[T]) LoadBatch(keys []*T) ([]*T, error) {
	sqlArgs := make([]interface{}, 0, 4)

	sb := &strings.Builder{}
	sb.WriteString(m.loadBatchSql1)
	for index, key := range keys {
		if index != 0 {
			sb.WriteString(" or ")
		}
		sb.WriteString(m.loadBatchSql2)

		rv := reflectx.IndirectValue(key)
		for _, arg := range m.loadBatchArgs2 {
			fieldV := rv.FieldByName(arg.GoFieldName).Interface()
			sqlArgs = append(sqlArgs, fieldV)
		}
	}

	ret := make([]*T, 0)
	err := m.Db.Unsafe().Select(&ret, sb.String(), sqlArgs...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}
