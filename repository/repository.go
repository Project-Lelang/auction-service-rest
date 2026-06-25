package repository

import (
	"context"
	"database/sql"
	"reflect"

	"auction-service/constant"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
)

var stmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

func dbtx(db infrastructure.DBTX, ctx context.Context) infrastructure.DBTX {
	dbtx, err := model.GetDbtxCtx(ctx)
	if err == nil || dbtx != nil {
		return dbtx
	}
	return db
}

func exec(db infrastructure.DBTX, ctx context.Context, stmt squirrel.Sqlizer) (sql.Result, error) {
	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, translateSqlError(err)
	}

	dt := dbtx(db, ctx)
	result, err := dt.ExecContext(ctx, query, args...)
	return result, translateSqlError(err)
}

func insert(db infrastructure.DBTX, ctx context.Context, tableName string, arg map[string]interface{}) error {
	stmt := stmtBuilder.Insert(tableName).SetMap(arg)
	_, err := exec(db, ctx, stmt)
	return err
}

func defaultInsert(db infrastructure.DBTX, ctx context.Context, m model.BaseModel) error {
	now := util.CurrentDateTime()
	if m.GetCreatedAt().IsZero() {
		m.SetCreatedAt(now)
	}
	if m.GetUpdatedAt().IsZero() {
		m.SetUpdatedAt(now)
	}

	arg := m.ToMap()
	if v, ok := arg["id"]; ok && isZeroInt(v) {
		delete(arg, "id")
	}

	stmt := stmtBuilder.Insert(m.TableName()).SetMap(arg)
	result, err := exec(db, ctx, stmt)
	if err != nil {
		return err
	}
	return setLastInsertId(m, result)
}

func fetch(db infrastructure.DBTX, ctx context.Context, dest interface{}, stmt squirrel.SelectBuilder) error {
	query, args, err := stmt.ToSql()
	if err != nil {
		return translateSqlError(err)
	}
	return translateSqlError(dbtx(db, ctx).SelectContext(ctx, dest, query, args...))
}

func get(db infrastructure.DBTX, ctx context.Context, dest interface{}, stmt squirrel.SelectBuilder) error {
	query, args, err := stmt.ToSql()
	if err != nil {
		return translateSqlError(err)
	}
	return translateSqlError(dbtx(db, ctx).GetContext(ctx, dest, query, args...))
}

func update(db infrastructure.DBTX, ctx context.Context, tableName string, arg map[string]interface{}, whereStmt squirrel.Sqlizer) error {
	stmt := stmtBuilder.Update(tableName).SetMap(arg).Where(whereStmt)
	_, err := exec(db, ctx, stmt)
	return err
}

func destroy(db infrastructure.DBTX, ctx context.Context, tableName string, whereStmt squirrel.Sqlizer) error {
	stmt := stmtBuilder.Delete(tableName).Where(whereStmt)
	_, err := exec(db, ctx, stmt)
	return err
}

func isZeroInt(v interface{}) bool {
	switch n := v.(type) {
	case int:
		return n == 0
	case int64:
		return n == 0
	}
	return false
}

func setLastInsertId(m model.BaseModel, result sql.Result) error {
	id, err := result.LastInsertId()
	if err != nil || id == 0 {
		return nil
	}

	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return nil
	}
	field := elem.FieldByName("Id")
	if !field.IsValid() || !field.CanSet() {
		return nil
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(id)
	}
	return nil
}

func translateSqlError(err error) error {
	switch v := err.(type) {
	case *mysql.MySQLError:
		switch v.Number {
		case 1062: // ER_DUP_ENTRY
			return constant.ErrDuplicateData
		case 1451, 1452: // ER_ROW_IS_REFERENCED_2, ER_NO_REFERENCED_ROW_2
			return constant.ErrForeignKeyViolation
		default:
			return err
		}
	default:
		switch v {
		case sql.ErrNoRows:
			return constant.ErrNoData
		default:
			return err
		}
	}
}
