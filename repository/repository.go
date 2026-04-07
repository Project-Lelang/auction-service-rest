package repository

import (
	"context"
	"database/sql"

	"auction-service/constant"
	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var stmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func dbtx(db infrastructure.DBTX, ctx context.Context) infrastructure.DBTX {
	dbtx, err := model.GetDbtxCtx(ctx)
	if err == nil || dbtx != nil {
		return dbtx
	}
	return db
}

func exec(db infrastructure.DBTX, ctx context.Context, stmt squirrel.Sqlizer) error {
	query, args, err := stmt.ToSql()
	if err != nil {
		return translateSqlError(err)
	}

	dt := dbtx(db, ctx)
	_, err = dt.ExecContext(ctx, query, args...)
	return translateSqlError(err)
}

func insert(db infrastructure.DBTX, ctx context.Context, tableName string, arg map[string]interface{}) error {
	stmt := stmtBuilder.Insert(tableName).SetMap(arg)
	return exec(db, ctx, stmt)
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
	return exec(db, ctx, stmt)
}

func destroy(db infrastructure.DBTX, ctx context.Context, tableName string, whereStmt squirrel.Sqlizer) error {
	stmt := stmtBuilder.Delete(tableName).Where(whereStmt)
	return exec(db, ctx, stmt)
}

func translateSqlError(err error) error {
	switch v := err.(type) {
	case *pgconn.PgError:
		switch v.Code {
		case pgerrcode.UniqueViolation:
			return constant.ErrDuplicateData
		case pgerrcode.ForeignKeyViolation:
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
