package dbrepo

import (
	"errors"
	"database/sql"
	
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
	category "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service/Category"
)


func NewCategoryDBRepo(db *sql.DB) *CommonDBRepo {
    m := entityMapper{
        table:     "categories",
        byIDQuery: `SELECT id, name, type FROM categories WHERE id = $1`,
        allQuery:  `SELECT id, name, type FROM categories`,
        insertSQL: `INSERT INTO categories(id,name,type) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, type=EXCLUDED.type`,
        deleteSQL: `DELETE FROM categories WHERE id = $1`,
        scanOne: func(s scanner) (service.ICommonObject, error) {
            var id service.ObjectID
            var name string
            var t int
            if err := s.Scan(&id, &name, &t); err != nil {
                return nil, err
            }
            cat, err := category.NewCopyCategory(id, name, category.CategoryType(t))
            if err != nil {
                return nil, err
            }
            return cat, nil
        },
        argsForInsert: func(obj service.ICommonObject) ([]any, error) {
            cat, ok := obj.(category.ICategory)
            if !ok {
                return nil, errors.New("expected ICategory")
            }
            return []any{cat.ID(), cat.Name(), int(cat.Type())}, nil
        },
    }
    return NewCommonDBRepo(db, m)
}
