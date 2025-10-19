package dbfunc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/zeromicro/go-zero/core/logx"

	"entgo.io/ent/dialect/sql"
	"github.com/coder-lulu/newbee-core/rpc/ent/department"
)

func GetDepartmentAncestors(departmentID *uint64, db *ent.Client, logger logx.Logger, ctx context.Context) (*string, error) {
	ancestors, err := db.Department.Query().
		Modify(func(s *sql.Selector) {
			t1, t2 := sql.Table(department.Table), sql.Table(department.Table)
			with := sql.WithRecursive("ancestors")
			with.As(
				sql.Select(
					t1.C(department.FieldID),
					t1.C(department.FieldParentID),
				).AppendSelectExpr(sql.ExprFunc(func(b *sql.Builder) {
					b.Ident("1 as level")
				})).
					From(t1).
					Where(sql.EQ(department.FieldID, *departmentID)).
					UnionAll(
						sql.Select(t2.Columns(department.FieldID, department.FieldParentID)...).
							AppendSelectExpr(sql.ExprFunc(func(b *sql.Builder) {
								b.Ident(with.Name() + ".level + 1 as level")
							})).
							From(t2).
							Join(with).
							On(t2.C(department.FieldID), with.C(department.FieldParentID)),
					),
			)
			s.Prefix(with).Select(with.C(department.FieldID)).From(with)
		}).
		Select(department.FieldParentID).Strings(ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(logger, err, fmt.Sprintf("failed to get the department ancestors of %d", departmentID))
	}

	if len(ancestors) == 0 {
		return nil, nil
	}

	return pointy.GetPointer(strings.Join(ancestors, ",")), nil
}

func GetSubDepartment(departmentID uint64, db *ent.Client, logger logx.Logger, ctx context.Context) (string, error) {
	subDepts, err := db.Department.Query().
		Modify(func(s *sql.Selector) {
			t1, t2 := sql.Table(department.Table), sql.Table(department.Table).As("t")
			with := sql.WithRecursive("sub_departments")
			with.As(
				sql.Select(
					t1.C(department.FieldID),
					t1.C(department.FieldParentID),
				).
					From(t1).
					Where(sql.EQ(department.FieldParentID, departmentID)).
					UnionAll(
						sql.Select(t2.Columns(department.FieldID, department.FieldParentID)...).
							From(t2).
							Join(with).
							On(t2.C(department.FieldParentID), with.C(department.FieldID)),
					),
			)
			s.Prefix(with).Select(with.C(department.FieldID)).From(with)
		}).
		Select(department.FieldID).Strings(ctx)
	if err != nil {
		return "", dberrorhandler.DefaultEntError(logger, err, fmt.Sprintf("failed to get the sub department of %d", departmentID))
	}

	if len(subDepts) == 0 {
		return "", nil
	}

	subDepts = append(subDepts, strconv.Itoa(int(departmentID)))

	return strings.Join(subDepts, ","), nil
}
