package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	
service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type scanner interface {
    Scan(dest ...any) error
}

type entityMapper struct {
    table string

    byIDQuery string
    allQuery  string
    insertSQL string
    deleteSQL string

    scanOne func(s scanner) (service.ICommonObject, error)
    argsForInsert func(obj service.ICommonObject) ([]any, error)
}

type CommonDBRepo struct {
    db     *sql.DB
    mapper entityMapper
}


func NewCommonDBRepo(db *sql.DB, m entityMapper) *CommonDBRepo {
    return &CommonDBRepo{db: db, mapper: m}
}

func (r *CommonDBRepo) ByID(ctx context.Context, id service.ObjectID) (service.ICommonObject, error) {
    row := r.db.QueryRowContext(ctx, r.mapper.byIDQuery, id)
    obj, err := r.mapper.scanOne(row)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("%s not found: %w", r.mapper.table, err)
        }
        return nil, err
    }
    return obj, nil
}

func (r *CommonDBRepo) All(ctx context.Context) ([]service.ICommonObject, error) {
    rows, err := r.db.QueryContext(ctx, r.mapper.allQuery)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []service.ICommonObject
    for rows.Next() {
        obj, err := r.mapper.scanOne(rows)
        if err != nil {
            return nil, err
        }
        out = append(out, obj)
    }
    return out, rows.Err()
}

func (r *CommonDBRepo) Save(ctx context.Context, obj service.ICommonObject) error {
    args, err := r.mapper.argsForInsert(obj)
    if err != nil {
        return err
    }
    _, err = r.db.ExecContext(ctx, r.mapper.insertSQL, args...)
    return err
}

func (r *CommonDBRepo) Delete(ctx context.Context, id service.ObjectID) error {
    _, err := r.db.ExecContext(ctx, r.mapper.deleteSQL, id)
    return err
}
