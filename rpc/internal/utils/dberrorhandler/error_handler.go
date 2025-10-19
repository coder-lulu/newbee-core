package dberrorhandler

import (
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/v2/msg/logmsg"

	"github.com/coder-lulu/newbee-common/v2/i18n"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// DefaultEntError returns errors dealing with default functions.
func DefaultEntError(logger logx.Logger, err error, detail any) error {
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.TargetNotFound)
		case ent.IsConstraintError(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.ConstraintError)
		case ent.IsValidationError(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.ValidationError)
		case ent.IsNotSingular(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.NotSingularError)
		default:
			logger.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
			return errorx.NewInternalError(i18n.DatabaseError)
		}
	}
	return err
}

// AuthUserEntError returns errors dealing with authenticated user operations.
// When an authenticated user is not found, it returns 401 Unauthorized instead of "target not found".
func AuthUserEntError(logger logx.Logger, err error, detail any) error {
	if err != nil {
		switch {
		case ent.IsNotFound(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewCodeError(401, "Token is invalid")
		case ent.IsConstraintError(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.ConstraintError)
		case ent.IsValidationError(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.ValidationError)
		case ent.IsNotSingular(err):
			logger.Errorw(err.Error(), logx.Field("detail", detail))
			return errorx.NewInvalidArgumentError(i18n.NotSingularError)
		default:
			logger.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
			return errorx.NewInternalError(i18n.DatabaseError)
		}
	}
	return err
}
